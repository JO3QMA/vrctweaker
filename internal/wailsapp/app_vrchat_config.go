package wailsapp

import (
	"errors"
	"os"
	"path/filepath"
)

// ResolveVRChatAssetCachePath returns the absolute path Asset cache clear would target
// (saved config.json; empty cache_directory → platform default Cache-WindowsPlayer).
func (a *App) ResolveVRChatAssetCachePath() (string, error) {
	if a.assetCache == nil {
		return "", errors.New("asset cache not initialized")
	}
	return a.assetCache.ResolvePath()
}

// ClearVRChatAssetCache deletes all entries inside the resolved VRChat asset cache directory.
// Returns the number of top-level entries removed. See ADR 0011.
func (a *App) ClearVRChatAssetCache() (int64, error) {
	if a.assetCache == nil {
		return 0, errors.New("asset cache not initialized")
	}
	return a.assetCache.Clear()
}

// VRChatConfigExists checks if config.json exists.
func (a *App) VRChatConfigExists() (bool, error) {
	return a.vrchatConfigRepo.Exists()
}

// GetVRChatConfig reads the current config.json.
func (a *App) GetVRChatConfig() (VRChatConfigDTO, error) {
	cfg, err := a.vrchatConfigRepo.Read()
	if err != nil {
		return VRChatConfigDTO{}, err
	}
	return toVRChatConfigDTO(cfg), nil
}

// SaveVRChatConfig writes config.json.
func (a *App) SaveVRChatConfig(dto VRChatConfigDTO) error {
	return a.vrchatConfigRepo.Write(fromVRChatConfigDTO(dto))
}

// DeleteVRChatConfig removes config.json.
func (a *App) DeleteVRChatConfig() error {
	return a.vrchatConfigRepo.Delete()
}

// DefaultVRChatPictureFolder returns the folder VRChat uses when picture_output_folder
// is unset: filepath.Join(home, "Pictures", "VRChat") (e.g. ~/Pictures/VRChat on Unix,
// %USERPROFILE%\Pictures\VRChat on Windows).
//
// Limitation: this does not resolve the shell “My Pictures” location (e.g. Windows folder
// redirection to another drive). VRChat may follow the OS special folder; Go’s stdlib has
// no direct equivalent, so this path is a conventional default and may differ from the
// actual save location when Pictures is redirected.
func (a *App) DefaultVRChatPictureFolder() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Pictures", "VRChat"), nil
}
