//go:build windows

package tray

import (
	"runtime"
	"sync"

	"golang.org/x/sys/windows"

	"vrchat-tweaker/internal/infrastructure/tray/w32systray"
)

type windowsManager struct {
	mu           sync.Mutex
	running      bool
	trayThreadID uint32
	stopWaiter   chan error

	startReq chan startRequest
}

type startRequest struct {
	cfg  Config
	resp chan error
}

func newManager() Manager {
	m := &windowsManager{
		startReq: make(chan startRequest),
	}
	go m.trayThread()
	return m
}

func (m *windowsManager) Supported() bool { return true }

func (m *windowsManager) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

func (m *windowsManager) Start(cfg Config) error {
	if m.Running() {
		return nil
	}
	resp := make(chan error, 1)
	m.startReq <- startRequest{cfg: cfg, resp: resp}
	return <-resp
}

func (m *windowsManager) Stop() error {
	if !m.Running() {
		return nil
	}
	waiter := make(chan error, 1)
	m.mu.Lock()
	m.stopWaiter = waiter
	tid := m.trayThreadID
	m.mu.Unlock()

	if tid != 0 {
		_ = windows.PostThreadMessage(tid, w32systray.WMTrayShutdown, 0, 0)
	}
	return <-waiter
}

// trayThread owns the Win32 HWND and message loop on a single OS thread.
func (m *windowsManager) trayThread() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	m.trayThreadID = windows.GetCurrentThreadId()

	for {
		req := <-m.startReq
		m.runSession(req)
	}
}

func (m *windowsManager) runSession(req startRequest) {
	tray, err := w32systray.New()
	if err != nil {
		req.resp <- err
		return
	}
	if err := tray.ShowCustom(req.cfg.IconPath, req.cfg.Tooltip); err != nil {
		_ = tray.Stop()
		req.resp <- err
		return
	}

	onShow := req.cfg.OnShow
	onQuit := req.cfg.OnQuit

	tray.OnClick(func() { runAsync(onShow) })
	tray.AppendMenu(req.cfg.MenuShowLabel, func() { runAsync(onShow) })
	tray.AppendSeparator()
	tray.AppendMenu(req.cfg.MenuQuitLabel, func() { runAsync(onQuit) })

	req.resp <- nil

	m.mu.Lock()
	m.running = true
	m.mu.Unlock()

	_ = tray.Run()

	m.mu.Lock()
	m.running = false
	waiter := m.stopWaiter
	m.stopWaiter = nil
	m.mu.Unlock()

	if waiter != nil {
		waiter <- nil
	}
}
