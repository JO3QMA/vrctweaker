//go:build !windows

package usecase

import (
	"os"
	"time"
)

// OfficialYTDLPCachePath is Windows-only.
func OfficialYTDLPCachePath() (string, error) {
	return "", ErrYTDLPUnsupportedPlatform
}

// VRChatYTDLPToolsPath is Windows-only.
func VRChatYTDLPToolsPath() (string, error) {
	return "", ErrYTDLPUnsupportedPlatform
}

// NeedsOfficialLink is a no-op on non-Windows when cache is present.
func NeedsOfficialLink(_, cachePath string) (bool, error) {
	if _, err := os.Stat(cachePath); err != nil {
		if os.IsNotExist(err) {
			return false, ErrYTDLPCacheMissing
		}
		return false, err
	}
	return false, nil
}

// EffectiveOfficialLink is always false on non-Windows.
func EffectiveOfficialLink(_, cachePath string) (bool, error) {
	if _, err := os.Stat(cachePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return false, nil
}

// LinkToolsToCache is unsupported on non-Windows.
func LinkToolsToCache(_, _ string, _ time.Duration) error {
	return ErrYTDLPUnsupportedPlatform
}
