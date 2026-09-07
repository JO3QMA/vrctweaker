// Package paths resolves OS-specific application and VRChat filesystem locations.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

// AppDataDir returns the directory for VRChat Tweaker application data (SQLite, tray icon, etc.).
func AppDataDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "vrchat-tweaker"), nil
}

// VRChatDataDir returns the directory that holds config.json (LocalLow/VRChat/VRChat on Windows).
func VRChatDataDir() (string, error) {
	if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
		return filepath.Join(filepath.Dir(dir), "LocalLow", "VRChat", "VRChat"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home: %w", err)
	}
	return filepath.Join(home, ".local", "share", "VRChat", "VRChat"), nil
}

// VRChatConfigPath returns the path to VRChat's config.json.
func VRChatConfigPath() (string, error) {
	dir, err := VRChatDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// VRChatConfigPathOrFallback returns VRChatConfigPath or ./config.json when resolution fails.
func VRChatConfigPathOrFallback() string {
	p, err := VRChatConfigPath()
	if err != nil {
		return filepath.Join(".", "config.json")
	}
	return p
}

// DefaultVRChatOutputLogDir returns the default directory watched for output_log files.
func DefaultVRChatOutputLogDir() string {
	p, err := VRChatConfigPath()
	if err != nil {
		return ""
	}
	return filepath.Dir(p)
}
