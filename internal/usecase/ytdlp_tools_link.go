package usecase

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const ytdlpExeName = "yt-dlp.exe"

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
// Missing cache → false (caller must download first). Plain file or wrong symlink → true.
func NeedsOfficialLink(toolsPath, cachePath string) (bool, error) {
	if _, err := os.Stat(cachePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
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
	return filepath.Clean(absTarget) != filepath.Clean(absCache), nil
}

// EffectiveOfficialLink is true when Tools is a symlink to the Official yt-dlp cache.
func EffectiveOfficialLink(toolsPath, cachePath string) (bool, error) {
	need, err := NeedsOfficialLink(toolsPath, cachePath)
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(cachePath); err != nil {
		return false, nil
	}
	return !need, nil
}

func waitUntilUnlocked(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		f, err := os.OpenFile(path, os.O_RDWR, 0)
		if err == nil {
			_ = f.Close()
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for unlock: %s: %w", path, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// LinkToolsToCache removes Tools/yt-dlp.exe (file or old link) and creates a symlink to cache.
func LinkToolsToCache(toolsPath, cachePath string, unlockTimeout time.Duration) error {
	if _, err := os.Stat(cachePath); err != nil {
		return fmt.Errorf("cache missing: %w", err)
	}
	absCache, err := filepath.Abs(cachePath)
	if err != nil {
		return err
	}

	if _, statErr := os.Lstat(toolsPath); statErr == nil {
		_ = waitUntilUnlocked(toolsPath, unlockTimeout)
		if rmErr := os.Remove(toolsPath); rmErr != nil && !os.IsNotExist(rmErr) {
			return fmt.Errorf("remove tools yt-dlp: %w", rmErr)
		}
	} else if !os.IsNotExist(statErr) {
		return fmt.Errorf("lstat tools yt-dlp: %w", statErr)
	}

	if err := os.MkdirAll(filepath.Dir(toolsPath), 0o755); err != nil {
		return err
	}
	if err := os.Symlink(absCache, toolsPath); err != nil {
		return fmt.Errorf("symlink Tools -> cache (enable Windows Developer Mode or run elevated): %w", err)
	}
	return nil
}
