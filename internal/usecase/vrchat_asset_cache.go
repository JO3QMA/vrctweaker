package usecase

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"vrchat-tweaker/internal/domain/vrchatconfig"
)

// Sentinel errors for Asset cache clear (stable English phrases for frontend mapping).
var (
	ErrAssetCacheVRChatRunning       = errors.New("vrchat is running")
	ErrAssetCacheVolumeRoot          = errors.New("cache path is volume root")
	ErrAssetCacheNotDirectory        = errors.New("cache path is not a directory")
	ErrAssetCachePathMissing         = errors.New("cache path does not exist")
	ErrAssetCacheEqualsPictureFolder = errors.New("cache path equals picture folder")
	ErrAssetCacheEqualsVRChatDataDir = errors.New("cache path equals vrchat data directory")
	ErrAssetCacheEmptyPath           = errors.New("cache path is empty")
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
		return "", fmt.Errorf("read vrchat config: %w", err)
	}
	return uc.resolveFromConfig(cfg)
}

func (uc *VRChatAssetCacheUseCase) resolveFromConfig(cfg *vrchatconfig.VRChatConfig) (string, error) {
	if cfg != nil {
		if p := strings.TrimSpace(cfg.CacheDirectory); p != "" {
			abs, err := filepath.Abs(filepath.Clean(p))
			if err != nil {
				return "", err
			}
			return abs, nil
		}
	}
	def, err := uc.defaultCache()
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(filepath.Clean(def))
	if err != nil {
		return "", err
	}
	return abs, nil
}

func (uc *VRChatAssetCacheUseCase) resolvePicturePath(cfg *vrchatconfig.VRChatConfig) (string, error) {
	if cfg != nil {
		if p := strings.TrimSpace(cfg.PictureOutputFolder); p != "" {
			abs, err := filepath.Abs(filepath.Clean(p))
			if err != nil {
				return "", err
			}
			return abs, nil
		}
	}
	def, err := uc.defaultPicture()
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(filepath.Clean(def))
	if err != nil {
		return "", err
	}
	return abs, nil
}

// Clear deletes all entries inside the resolved VRChat asset cache directory.
// The directory itself remains. Returns the number of top-level entries removed.
func (uc *VRChatAssetCacheUseCase) Clear() (int64, error) {
	if uc.running != nil {
		running, err := uc.running.VRChatRunning()
		if err != nil {
			return 0, fmt.Errorf("check vrchat running: %w", err)
		}
		if running {
			return 0, ErrAssetCacheVRChatRunning
		}
	}

	cfg, err := uc.readConfig()
	if err != nil {
		return 0, fmt.Errorf("read vrchat config: %w", err)
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
		return 0, fmt.Errorf("resolve picture folder: %w", err)
	}
	if samePath(cachePath, picPath) {
		return 0, ErrAssetCacheEqualsPictureFolder
	}

	if uc.vrchatDataDir != nil {
		dataDir, dataErr := uc.vrchatDataDir()
		if dataErr != nil {
			return 0, fmt.Errorf("resolve vrchat data dir: %w", dataErr)
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
		return 0, err
	}
	if !info.IsDir() {
		return 0, ErrAssetCacheNotDirectory
	}

	entries, err := os.ReadDir(cachePath)
	if err != nil {
		return 0, err
	}
	var n int64
	for _, e := range entries {
		child := filepath.Join(cachePath, e.Name())
		if err := removeCacheEntry(child); err != nil {
			return n, fmt.Errorf("remove %s: %w", child, err)
		}
		n++
	}
	return n, nil
}

// removeCacheEntry removes one top-level entry.
// Symlinks are removed with os.Remove (do not follow into the target).
func removeCacheEntry(path string) error {
	fi, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
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
	return clean == string(os.PathSeparator) || clean == "/"
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
	return strings.EqualFold(a, b)
}
