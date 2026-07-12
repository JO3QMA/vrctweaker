//go:build !windows

package usecase

import "time"

// OfficialYTDLPCachePath is Windows-only.
func OfficialYTDLPCachePath() (string, error) {
	return "", ErrYTDLPUnsupportedPlatform
}

// VRChatYTDLPToolsPath is Windows-only.
func VRChatYTDLPToolsPath() (string, error) {
	return "", ErrYTDLPUnsupportedPlatform
}

// NeedsOfficialLink is a no-op on non-Windows.
func NeedsOfficialLink(_, _ string) (bool, error) {
	return false, nil
}

// EffectiveOfficialLink is always false on non-Windows.
func EffectiveOfficialLink(_, _ string) (bool, error) {
	return false, nil
}

// LinkToolsToCache is unsupported on non-Windows.
func LinkToolsToCache(_, _ string, _ time.Duration) error {
	return ErrYTDLPUnsupportedPlatform
}
