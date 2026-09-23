package wailsapp

import "vrchat-tweaker/internal/appversion"

// GetAppVersion returns the application release version (from wails.json).
func (a *App) GetAppVersion() string {
	return appversion.Version
}
