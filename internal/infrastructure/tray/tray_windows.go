//go:build windows

package tray

import (
	"fmt"
	"runtime"
	"sync"

	"golang.org/x/sys/windows"

	"vrchat-tweaker/internal/infrastructure/tray/w32systray"
)

type windowsManager struct {
	mu           sync.Mutex
	running      bool
	trayThreadID uint32
	// stopWaiter receives exactly one value when the in-flight Stop() completes.
	// Concurrent Stop() calls wait on the same channel instead of overwriting it.
	stopWaiter chan error

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
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	resp := make(chan error, 1)
	m.startReq <- startRequest{cfg: cfg, resp: resp}
	return <-resp
}

func (m *windowsManager) Stop() error {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return nil
	}
	if m.stopWaiter != nil {
		waiter := m.stopWaiter
		m.mu.Unlock()
		return <-waiter
	}
	waiter := make(chan error, 1)
	m.stopWaiter = waiter
	tid := m.trayThreadID
	m.mu.Unlock()

	if tid == 0 {
		m.signalStopWaiter(waiter, fmt.Errorf("tray: thread id not ready"))
		return <-waiter
	}
	if err := windows.PostThreadMessage(tid, w32systray.WMTrayShutdown, 0, 0); err != nil {
		m.signalStopWaiter(waiter, fmt.Errorf("tray: PostThreadMessage: %w", err))
		return <-waiter
	}
	return <-waiter
}

func (m *windowsManager) signalStopWaiter(waiter chan error, err error) {
	m.mu.Lock()
	if m.stopWaiter == waiter {
		m.stopWaiter = nil
	}
	m.mu.Unlock()
	waiter <- err
}

// trayThread owns the Win32 HWND and message loop on a single OS thread.
func (m *windowsManager) trayThread() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	m.mu.Lock()
	m.trayThreadID = windows.GetCurrentThreadId()
	m.mu.Unlock()

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
		_ = tray.Destroy()
		req.resp <- err
		return
	}

	onShow := req.cfg.OnShow
	onQuit := req.cfg.OnQuit

	var handlers sync.WaitGroup
	tray.OnClick(func() { runAsync(&handlers, onShow) })
	tray.AppendMenu(req.cfg.MenuShowLabel, func() { runAsync(&handlers, onShow) })
	tray.AppendSeparator()
	tray.AppendMenu(req.cfg.MenuQuitLabel, func() { runAsync(&handlers, onQuit) })

	m.mu.Lock()
	m.running = true
	m.mu.Unlock()
	req.resp <- nil

	_ = tray.Run()
	_ = tray.Destroy()

	handlers.Wait()

	m.mu.Lock()
	m.running = false
	waiter := m.stopWaiter
	m.stopWaiter = nil
	m.mu.Unlock()

	if waiter != nil {
		waiter <- nil
	}
}
