package wailsapp

import (
	"errors"
	"log"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/appversion"
	"vrchat-tweaker/internal/usecase"
)

var errAppUpdateNotInitialized = errors.New("app update not initialized")

const appUpdateAvailableEvent = "app:update-available"

// GetAppUpdateAutoCheck returns whether startup should check for updates.
func (a *App) GetAppUpdateAutoCheck() (bool, error) {
	if a.appUpdate == nil || a.settings == nil {
		return true, nil
	}
	return a.appUpdate.GetAutoCheckEnabled(a.ctx)
}

// SetAppUpdateAutoCheck persists the auto-check toggle.
func (a *App) SetAppUpdateAutoCheck(enabled bool) error {
	if a.appUpdate == nil {
		return errAppUpdateNotInitialized
	}
	return a.appUpdate.SetAutoCheckEnabled(a.ctx, enabled)
}

// CheckForAppUpdate queries GitHub for a newer stable release (manual check).
func (a *App) CheckForAppUpdate() (usecase.AppUpdateStatus, error) {
	if a.appUpdate == nil {
		return usecase.AppUpdateStatus{CurrentVersion: appversion.Version}, errAppUpdateNotInitialized
	}
	return a.appUpdate.CheckForUpdate(a.ctx, true)
}

// DismissAppUpdate skips prompts for the given release version.
func (a *App) DismissAppUpdate(version string) error {
	if a.appUpdate == nil {
		return errAppUpdateNotInitialized
	}
	return a.appUpdate.DismissUpdate(a.ctx, version)
}

// ApplyAppUpdate downloads and applies a Windows release zip, then quits for restart.
func (a *App) ApplyAppUpdate(downloadURL, zipSHA256 string) error {
	if a.appUpdate == nil {
		return errAppUpdateNotInitialized
	}
	if err := a.appUpdate.ApplyUpdate(a.ctx, downloadURL, zipSHA256); err != nil {
		return err
	}
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
	return nil
}

// OpenAppReleasePage opens the release page in the default browser.
func (a *App) OpenAppReleasePage(url string) error {
	if a.ctx == nil {
		return errAppUpdateNotInitialized
	}
	runtime.BrowserOpenURL(a.ctx, url)
	return nil
}

func (a *App) startAppUpdateCheckOnStartup() {
	if a.ctx == nil || a.appUpdate == nil {
		return
	}
	go func() {
		time.Sleep(2 * time.Second)
		enabled, err := a.appUpdate.GetAutoCheckEnabled(a.ctx)
		if err != nil {
			log.Printf("app update: auto check setting: %v", err)
			return
		}
		if !enabled {
			return
		}
		st, err := a.appUpdate.CheckForUpdate(a.ctx, false)
		if err != nil {
			log.Printf("app update: startup check: %v", err)
			return
		}
		if !st.ShouldPrompt {
			return
		}
		runtime.EventsEmit(a.ctx, appUpdateAvailableEvent, st)
	}()
}
