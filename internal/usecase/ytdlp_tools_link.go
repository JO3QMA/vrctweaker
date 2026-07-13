//go:build windows

package usecase

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

// OfficialYTDLPCachePath returns …/Local/vrchat-tweaker/ytdlp/yt-dlp.exe (outside LocalLow).
func OfficialYTDLPCachePath() (string, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return "", fmt.Errorf("LOCALAPPDATA is not set")
	}
	return officialYTDLPCachePathFromLocal(localAppData), nil
}

func officialYTDLPCachePathFromLocal(localAppData string) string {
	return filepath.Join(localAppData, "vrchat-tweaker", "ytdlp", ytdlpExeName)
}

// VRChatYTDLPToolsPath returns …/LocalLow/VRChat/VRChat/Tools/yt-dlp.exe on Windows
// (same Local→LocalLow rule as getVRChatConfigPath).
func VRChatYTDLPToolsPath() (string, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return "", fmt.Errorf("LOCALAPPDATA is not set")
	}
	return vrchatYTDLPToolsPathFromLocal(localAppData), nil
}

func vrchatYTDLPToolsPathFromLocal(localAppData string) string {
	base := filepath.Join(filepath.Dir(localAppData), "LocalLow", "VRChat", "VRChat", "Tools")
	return filepath.Join(base, ytdlpExeName)
}

// NeedsOfficialLink reports whether Tools/yt-dlp.exe should be (re)linked to cache.
// Missing cache returns ErrYTDLPCacheMissing. Plain file or wrong symlink → true.
func NeedsOfficialLink(toolsPath, cachePath string) (bool, error) {
	if _, err := os.Stat(cachePath); err != nil {
		if os.IsNotExist(err) {
			return false, ErrYTDLPCacheMissing
		}
		return false, err
	}

	fi, err := os.Lstat(toolsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		return true, nil
	}

	target, err := os.Readlink(toolsPath)
	if err != nil {
		return false, err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(toolsPath), target)
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return false, err
	}
	absCache, err := filepath.Abs(cachePath)
	if err != nil {
		return false, err
	}
	return !strings.EqualFold(filepath.Clean(absTarget), filepath.Clean(absCache)), nil
}

// EffectiveOfficialLink is true when Tools is a symlink to the Official yt-dlp cache.
func EffectiveOfficialLink(toolsPath, cachePath string) (bool, error) {
	need, err := NeedsOfficialLink(toolsPath, cachePath)
	if err != nil {
		if errors.Is(err, ErrYTDLPCacheMissing) {
			return false, nil
		}
		return false, err
	}
	return !need, nil
}

func isTransientLockErr(err error) bool {
	if err == nil || os.IsNotExist(err) || os.IsPermission(err) {
		return false
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == windows.ERROR_SHARING_VIOLATION || errno == windows.ERROR_LOCK_VIOLATION
	}
	return false
}

func waitUntilUnlocked(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		if err := tryOpenToolsPath(path); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if !isTransientLockErr(lastErr) {
			return fmt.Errorf("cannot open %s: %w", path, lastErr)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for unlock: %s: %w", path, lastErr)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func tryOpenToolsPath(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err == nil {
		return f.Close()
	}
	if !os.IsPermission(err) {
		return err
	}
	f, err = os.OpenFile(path, os.O_RDONLY, 0)
	if err == nil {
		return f.Close()
	}
	return err
}

// LinkToolsToCache removes Tools/yt-dlp.exe (file or old link) and creates a symlink to cache.
// Callers must serialize concurrent invocations (YTDLPMaintainUseCase holds its mutex during link).
func LinkToolsToCache(toolsPath, cachePath string, unlockTimeout time.Duration) error {
	if _, err := os.Stat(cachePath); err != nil {
		return fmt.Errorf("cache missing: %w", err)
	}
	absCache, err := filepath.Abs(cachePath)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if err := linkToolsToCacheOnce(toolsPath, absCache, unlockTimeout); err == nil {
			return nil
		} else {
			lastErr = err
			if attempt == 0 {
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
	return lastErr
}

func linkToolsToCacheOnce(toolsPath, absCache string, unlockTimeout time.Duration) error {
	restore, err := backupToolsEntry(toolsPath, unlockTimeout)
	if err != nil {
		return err
	}
	symlinkFailed := true
	defer func() {
		if symlinkFailed {
			restore()
		}
	}()

	if err := os.MkdirAll(filepath.Dir(toolsPath), 0o755); err != nil {
		return err
	}
	if err := os.Symlink(absCache, toolsPath); err != nil {
		return fmt.Errorf("symlink tools yt-dlp: enable Windows Developer Mode or run as administrator: %w", err)
	}
	symlinkFailed = false
	return nil
}

func backupToolsEntry(toolsPath string, unlockTimeout time.Duration) (restore func(), err error) {
	fi, statErr := os.Lstat(toolsPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return func() {}, nil
		}
		return nil, fmt.Errorf("lstat tools yt-dlp: %w", statErr)
	}
	if unlockErr := waitUntilUnlocked(toolsPath, unlockTimeout); unlockErr != nil {
		return nil, fmt.Errorf("unlock tools yt-dlp: %w", unlockErr)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		if rmErr := os.Remove(toolsPath); rmErr != nil && !os.IsNotExist(rmErr) {
			return nil, fmt.Errorf("remove tools yt-dlp: %w", rmErr)
		}
		return func() {}, nil
	}
	bak := toolsPath + ".vrctweaker.bak"
	_ = os.Remove(bak)
	if err := os.Rename(toolsPath, bak); err != nil {
		return nil, fmt.Errorf("backup tools yt-dlp: %w", err)
	}
	return func() {
		if _, statErr := os.Stat(bak); statErr != nil {
			return
		}
		_ = os.Remove(toolsPath)
		if err := os.Rename(bak, toolsPath); err != nil {
			data, readErr := os.ReadFile(bak)
			if readErr != nil {
				log.Printf("ytdlp maintain: restore tools from backup: %v", err)
				return
			}
			if writeErr := os.WriteFile(toolsPath, data, fi.Mode().Perm()); writeErr != nil {
				log.Printf("ytdlp maintain: restore tools from backup: %v", writeErr)
			}
		}
	}, nil
}
