package usecase

import (
	"path/filepath"

	"vrchat-tweaker/internal/platform/paths"
)

// DefaultVRChatAssetCacheFolder returns the conventional VRChat asset cache directory
// when config.json cache_directory is unset: …/LocalLow/VRChat/VRChat/Cache-WindowsPlayer
// (Windows via LOCALAPPDATA) or ~/.local/share/VRChat/VRChat/Cache-WindowsPlayer otherwise.
//
// This is intentionally Cache-WindowsPlayer under the VRChat data dir — not the data dir
// itself — so clearing contents cannot wipe config.json or Tools.
func DefaultVRChatAssetCacheFolder() (string, error) {
	base, err := paths.VRChatDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "Cache-WindowsPlayer"), nil
}

// VRChatDataDir returns the directory that holds config.json (LocalLow/VRChat/VRChat on Windows).
func VRChatDataDir() (string, error) {
	return paths.VRChatDataDir()
}
