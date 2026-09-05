package main

import (
	"context"
	_ "embed"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/infrastructure/tray"
	"vrchat-tweaker/internal/locale"
)

//go:embed build/windows/icon.ico
var trayIconICO []byte

func (a *App) initTrayManager() {
	if a.tray == nil {
		a.tray = tray.NewManager()
	}
}

func (a *App) trayIconPath() (string, error) {
	dataDir, err := getDataDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dataDir, "tray.ico")
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, trayIconICO, 0600); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) closeToTrayEffective(ctx context.Context) bool {
	a.initTrayManager()
	if a.settings == nil || !a.tray.Supported() {
		return false
	}
	on, err := a.settings.GetCloseToTray(ctx)
	if err != nil {
		runtime.LogWarning(ctx, "close_to_tray: "+err.Error())
		return false
	}
	return on
}

func (a *App) syncTrayFromSettings(ctx context.Context) {
	a.initTrayManager()
	if !a.tray.Supported() {
		return
	}
	if a.closeToTrayEffective(ctx) {
		if err := a.startTray(ctx); err != nil {
			runtime.LogWarning(ctx, "tray start: "+err.Error())
		}
		return
	}
	if err := a.tray.Stop(); err != nil {
		runtime.LogWarning(ctx, "tray stop: "+err.Error())
	}
}

func (a *App) startTray(ctx context.Context) error {
	if !a.tray.Supported() || a.tray.Running() {
		return nil
	}
	iconPath, err := a.trayIconPath()
	if err != nil {
		return err
	}
	lang, _ := a.settings.GetLanguage(ctx)
	if lang == "" {
		lang = locale.Detect()
	}
	showLabel, quitLabel := locale.TrayMenuLabels(lang)
	return a.tray.Start(tray.Config{
		Tooltip:       "VRChat Tweaker",
		IconPath:      iconPath,
		MenuShowLabel: showLabel,
		MenuQuitLabel: quitLabel,
		OnShow:        func() { a.showMainWindow() },
		OnQuit:        func() { a.quitApplication() },
	})
}

func (a *App) showMainWindow() {
	if a.ctx == nil {
		return
	}
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
}

func (a *App) quitApplication() {
	if a.ctx == nil {
		return
	}
	// runtime.Quit also invokes OnBeforeClose; bypass close-to-tray hide for explicit quit.
	a.quitPending.Store(true)
	a.stopTray()
	runtime.Quit(a.ctx)
}

func (a *App) hideMainWindow(ctx context.Context) {
	if ctx == nil {
		return
	}
	runtime.WindowHide(ctx)
}

// handleBeforeClose hides the window when close-to-tray is enabled; returns true to prevent quit.
func (a *App) handleBeforeClose(ctx context.Context) bool {
	if a.quitPending.Load() {
		return false
	}
	if !a.closeToTrayEffective(ctx) {
		return false
	}
	a.hideMainWindow(ctx)
	return true
}

// RequestClose closes the main window or quits depending on close-to-tray setting.
func (a *App) RequestClose() {
	if a.ctx == nil {
		return
	}
	if a.handleBeforeClose(a.ctx) {
		return
	}
	a.quitApplication()
}

// GetCloseToTray returns whether closing hides to the system tray (default true).
func (a *App) GetCloseToTray() (bool, error) {
	return a.settings.GetCloseToTray(a.ctx)
}

// SetCloseToTray enables or disables close-to-tray behavior and syncs the tray icon.
func (a *App) SetCloseToTray(on bool) error {
	if err := a.settings.SetCloseToTray(a.ctx, on); err != nil {
		return err
	}
	a.syncTrayFromSettings(a.ctx)
	return nil
}

func (a *App) stopTray() {
	if a.tray == nil {
		return
	}
	if err := a.tray.Stop(); err != nil && a.ctx != nil {
		runtime.LogWarning(a.ctx, "tray stop: "+err.Error())
	}
}
