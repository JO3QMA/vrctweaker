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
	starting     bool
	stopping     bool
	stopPending  bool
	trayThreadID uint32
	stopWaiters  stopWaiterSet

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
	return m.running || m.starting
}

func (m *windowsManager) Start(cfg Config) error {
	m.mu.Lock()
	if m.running || m.starting || m.stopping {
		m.mu.Unlock()
		return nil
	}
	m.starting = true
	m.mu.Unlock()

	resp := make(chan error, 1)
	m.startReq <- startRequest{cfg: cfg, resp: resp}
	return <-resp
}

func (m *windowsManager) Stop() error {
	m.mu.Lock()
	if !m.running && !m.starting {
		m.mu.Unlock()
		return nil
	}
	if m.stopping {
		waiter := m.stopWaiters.add()
		m.mu.Unlock()
		return <-waiter
	}
	m.stopping = true
	m.stopPending = true
	waiter := m.stopWaiters.add()
	tid := m.trayThreadID
	onlyStarting := m.starting && !m.running
	m.mu.Unlock()

	if onlyStarting {
		// runSession will observe stopPending before going live.
		return <-waiter
	}

	if tid == 0 {
		m.abortStop(fmt.Errorf("tray: thread id not ready"))
		return <-waiter
	}
	if err := w32systray.PostThreadShutdown(tid); err != nil {
		m.abortStop(fmt.Errorf("tray: PostThreadShutdown: %w", err))
		return <-waiter
	}
	return <-waiter
}

func (m *windowsManager) abortStop(err error) {
	m.mu.Lock()
	m.stopPending = false
	m.stopping = false
	m.running = false
	m.starting = false
	waiters := m.stopWaiters
	m.stopWaiters = stopWaiterSet{}
	m.mu.Unlock()
	waiters.complete(err)
}

func (m *windowsManager) finishStop(err error) {
	m.mu.Lock()
	m.running = false
	m.starting = false
	m.stopping = false
	m.stopPending = false
	waiters := m.stopWaiters
	m.stopWaiters = stopWaiterSet{}
	m.mu.Unlock()
	waiters.complete(err)
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
		m.finishStop(err)
		req.resp <- err
		return
	}
	if err := tray.ShowCustom(req.cfg.IconPath, req.cfg.Tooltip); err != nil {
		_ = tray.Destroy()
		m.finishStop(err)
		req.resp <- err
		return
	}

	if m.consumeStopPending(tray) {
		req.resp <- nil
		return
	}

	onShow := req.cfg.OnShow
	onQuit := req.cfg.OnQuit

	var showHandlers sync.WaitGroup
	tray.OnClick(func() { runAsync(&showHandlers, onShow) })
	tray.AppendMenu(req.cfg.MenuShowLabel, func() { runAsync(&showHandlers, onShow) })
	tray.AppendSeparator()
	// Quit must not join showHandlers: it calls Stop() and would deadlock if we
	// Wait() on the group before signaling stop waiters.
	tray.AppendMenu(req.cfg.MenuQuitLabel, func() { runAsync(nil, onQuit) })

	tray.EnsureMessageQueue()

	if m.consumeStopPending(tray) {
		req.resp <- nil
		return
	}

	m.mu.Lock()
	m.starting = false
	m.running = true
	m.mu.Unlock()
	req.resp <- nil

	_ = tray.Run()
	_ = tray.Destroy()

	m.finishStop(nil)
	showHandlers.Wait()
}

func (m *windowsManager) consumeStopPending(tray *w32systray.Systray) bool {
	m.mu.Lock()
	pending := m.stopPending
	if pending {
		m.stopPending = false
	}
	m.mu.Unlock()
	if !pending {
		return false
	}
	_ = tray.Destroy()
	m.finishStop(nil)
	return true
}
