package usecase

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"vrchat-tweaker/internal/domain/vrchatconfig"
)

// Sentinel message strings for Asset cache clear (must stay in sync with frontend/src/utils/assetCacheErrors.ts).
const (
	MsgAssetCacheVRChatRunning       = "vrchat is running"
	MsgAssetCacheVolumeRoot          = "cache path is volume root"
	MsgAssetCacheNotDirectory        = "cache path is not a directory"
	MsgAssetCachePathMissing         = "cache path does not exist"
	MsgAssetCacheEqualsPictureFolder = "cache path equals picture folder"
	MsgAssetCacheEqualsVRChatDataDir = "cache path equals vrchat data directory"
	MsgAssetCacheEmptyPath           = "cache path is empty"
	MsgAssetCacheRemoveFailed        = "cache remove failed"
	MsgAssetCacheFailed              = "asset cache clear failed"
)

// Sentinel errors for Asset cache clear (stable English phrases for frontend mapping).
var (
	ErrAssetCacheVRChatRunning       = errors.New(MsgAssetCacheVRChatRunning)
	ErrAssetCacheVolumeRoot          = errors.New(MsgAssetCacheVolumeRoot)
	ErrAssetCacheNotDirectory        = errors.New(MsgAssetCacheNotDirectory)
	ErrAssetCachePathMissing         = errors.New(MsgAssetCachePathMissing)
	ErrAssetCacheEqualsPictureFolder = errors.New(MsgAssetCacheEqualsPictureFolder)
	ErrAssetCacheEqualsVRChatDataDir = errors.New(MsgAssetCacheEqualsVRChatDataDir)
	ErrAssetCacheEmptyPath           = errors.New(MsgAssetCacheEmptyPath)
	ErrAssetCacheRemoveFailed        = errors.New(MsgAssetCacheRemoveFailed)
	ErrAssetCacheFailed              = errors.New(MsgAssetCacheFailed)
)

// VRChatRunningChecker reports whether the VRChat client process is running.
type VRChatRunningChecker interface {
	VRChatRunning() (bool, error)
}

// VRChatAssetCacheUseCase clears the resolved VRChat asset cache directory contents.
type VRChatAssetCacheUseCase struct {
	readConfig     func() (*vrchatconfig.VRChatConfig, error)
	running        VRChatRunningChecker
	defaultCache   func() (string, error)
	defaultPicture func() (string, error)
	vrchatDataDir  func() (string, error)
}

// NewVRChatAssetCacheUseCase wires Asset cache clear against saved config.json.
func NewVRChatAssetCacheUseCase(
	readConfig func() (*vrchatconfig.VRChatConfig, error),
	running VRChatRunningChecker,
	defaultCache func() (string, error),
	defaultPicture func() (string, error),
	vrchatDataDir func() (string, error),
) *VRChatAssetCacheUseCase {
	return &VRChatAssetCacheUseCase{
		readConfig:     readConfig,
		running:        running,
		defaultCache:   defaultCache,
		defaultPicture: defaultPicture,
		vrchatDataDir:  vrchatDataDir,
	}
}

// ResolvePath returns the absolute path that Clear would target (saved config, not UI draft).
func (uc *VRChatAssetCacheUseCase) ResolvePath() (string, error) {
	cfg, err := uc.readConfig()
	if err != nil {
		log.Printf("asset cache: read vrchat config: %v", err)
		return "", ErrAssetCacheFailed
	}
	return uc.resolveFromConfig(cfg)
}

func (uc *VRChatAssetCacheUseCase) resolveFromConfig(cfg *vrchatconfig.VRChatConfig) (string, error) {
	if cfg == nil {
		cfg = &vrchatconfig.VRChatConfig{}
	}
	if p := strings.TrimSpace(cfg.CacheDirectory); p != "" {
		abs, err := filepath.Abs(filepath.Clean(p))
		if err != nil {
			log.Printf("asset cache: abs cache_directory: %v", err)
			return "", ErrAssetCacheFailed
		}
		return abs, nil
	}
	def, err := uc.defaultCache()
	if err != nil {
		log.Printf("asset cache: default cache path: %v", err)
		return "", ErrAssetCacheFailed
	}
	abs, err := filepath.Abs(filepath.Clean(def))
	if err != nil {
		log.Printf("asset cache: abs default cache: %v", err)
		return "", ErrAssetCacheFailed
	}
	return abs, nil
}

func (uc *VRChatAssetCacheUseCase) resolvePicturePath(cfg *vrchatconfig.VRChatConfig) (string, error) {
	if cfg == nil {
		cfg = &vrchatconfig.VRChatConfig{}
	}
	if p := strings.TrimSpace(cfg.PictureOutputFolder); p != "" {
		abs, err := filepath.Abs(filepath.Clean(p))
		if err != nil {
			log.Printf("asset cache: abs picture folder: %v", err)
			return "", ErrAssetCacheFailed
		}
		return abs, nil
	}
	def, err := uc.defaultPicture()
	if err != nil {
		log.Printf("asset cache: default picture folder: %v", err)
		return "", ErrAssetCacheFailed
	}
	abs, err := filepath.Abs(filepath.Clean(def))
	if err != nil {
		log.Printf("asset cache: abs default picture: %v", err)
		return "", ErrAssetCacheFailed
	}
	return abs, nil
}

func (uc *VRChatAssetCacheUseCase) ensureVRChatNotRunning() error {
	if uc.running == nil {
		return nil
	}
	running, err := uc.running.VRChatRunning()
	if err != nil {
		log.Printf("asset cache: check vrchat running: %v", err)
		return ErrAssetCacheFailed
	}
	if running {
		return ErrAssetCacheVRChatRunning
	}
	return nil
}

// Clear deletes all entries inside the resolved VRChat asset cache directory.
// The directory itself remains. Returns the number of top-level entries removed.
// Returned errors are path-free sentinels suitable for UI mapping.
func (uc *VRChatAssetCacheUseCase) Clear() (int64, error) {
	if err := uc.ensureVRChatNotRunning(); err != nil {
		return 0, err
	}

	cfg, err := uc.readConfig()
	if err != nil {
		log.Printf("asset cache: read vrchat config: %v", err)
		return 0, ErrAssetCacheFailed
	}
	cachePath, err := uc.resolveFromConfig(cfg)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(cachePath) == "" {
		return 0, ErrAssetCacheEmptyPath
	}
	if isVolumeRoot(cachePath) {
		return 0, ErrAssetCacheVolumeRoot
	}

	picPath, err := uc.resolvePicturePath(cfg)
	if err != nil {
		return 0, err
	}
	if samePath(cachePath, picPath) {
		return 0, ErrAssetCacheEqualsPictureFolder
	}

	if uc.vrchatDataDir != nil {
		dataDir, dataErr := uc.vrchatDataDir()
		if dataErr != nil {
			log.Printf("asset cache: resolve vrchat data dir: %v", dataErr)
			return 0, ErrAssetCacheFailed
		}
		if dataAbs, absErr := filepath.Abs(filepath.Clean(dataDir)); absErr == nil && samePath(cachePath, dataAbs) {
			return 0, ErrAssetCacheEqualsVRChatDataDir
		}
	}

	info, err := os.Lstat(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, ErrAssetCachePathMissing
		}
		log.Printf("asset cache: lstat %s: %v", cachePath, err)
		return 0, ErrAssetCacheFailed
	}
	if !info.IsDir() {
		return 0, ErrAssetCacheNotDirectory
	}

	// Re-check after path resolution / before mutation (narrow TOCTOU window).
	if runErr := uc.ensureVRChatNotRunning(); runErr != nil {
		return 0, runErr
	}

	entries, err := os.ReadDir(cachePath)
	if err != nil {
		log.Printf("asset cache: readdir %s: %v", cachePath, err)
		return 0, ErrAssetCacheFailed
	}
	var n int64
	for _, e := range entries {
		child := filepath.Join(cachePath, e.Name())
		if err := removeCacheEntry(child); err != nil {
			log.Printf("asset cache: remove %s: %v", child, err)
			return n, ErrAssetCacheRemoveFailed
		}
		n++
	}
	return n, nil
}

// removeCacheEntry removes one top-level entry.
// Symlinks and Windows reparse points (junctions) use os.Remove so the target tree is not walked.
func removeCacheEntry(path string) error {
	fi, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 || isReparsePoint(path) {
		return os.Remove(path)
	}
	return os.RemoveAll(path)
}

func isVolumeRoot(path string) bool {
	clean := filepath.Clean(path)
	vol := filepath.VolumeName(clean)
	if vol != "" {
		rest := strings.TrimPrefix(clean, vol)
		rest = strings.Trim(rest, `/\`)
		return rest == ""
	}
	return clean == string(os.PathSeparator)
}

func samePath(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	if errA == nil {
		a = aa
	}
	if errB == nil {
		b = bb
	}
	// Windows paths are case-insensitive; elsewhere match matchAbsPaths (case-sensitive).
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
