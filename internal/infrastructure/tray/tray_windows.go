//go:build windows

package tray

import (
	"sync"

	"github.com/tadvi/systray"
)

type windowsManager struct {
	mu    sync.Mutex
	tray  *systray.Systray
	start bool
}

func newManager() Manager {
	return &windowsManager{}
}

func (m *windowsManager) Supported() bool { return true }

func (m *windowsManager) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.start
}

func (m *windowsManager) Start(cfg Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.start {
		return nil
	}

	tray, err := systray.New()
	if err != nil {
		return err
	}
	if err := tray.ShowCustom(cfg.IconPath, cfg.Tooltip); err != nil {
		_ = tray.Stop()
		return err
	}

	onShow := cfg.OnShow
	if onShow == nil {
		onShow = func() {}
	}
	onQuit := cfg.OnQuit
	if onQuit == nil {
		onQuit = func() {}
	}

	tray.OnClick(onShow)
	tray.AppendMenu(cfg.MenuShowLabel, onShow)
	tray.AppendSeparator()
	tray.AppendMenu(cfg.MenuQuitLabel, onQuit)

	m.tray = tray
	m.start = true

	go func() {
		_ = tray.Run()
	}()
	return nil
}

func (m *windowsManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.start || m.tray == nil {
		return nil
	}
	err := m.tray.Stop()
	m.tray = nil
	m.start = false
	return err
}
