// Package singleinstance ensures only one VRChat Tweaker process runs at a time.
package singleinstance

import (
	"errors"
	"log"
	"sync"
)

const defaultName = "VRChatTweaker"

var errGuardReleased = errors.New("singleinstance: guard released")

// Guard coordinates single-instance locking and second-launch activation.
type Guard struct {
	name        string
	windowTitle string
	onActivate  func()
	activateMu  sync.Mutex
	invokeMu    sync.Mutex
	pending     bool
	startMu     sync.Mutex
	started     bool
	startErr    error
	stopCh      chan struct{}
	releaseMu   sync.Mutex
	released    bool
	platformMu  sync.Mutex
	platform    platformGuard
}

// New returns a Guard using the default application identifier.
func New() *Guard {
	return NewNamed(defaultName)
}

// NewWithWindowTitle returns a Guard and stores the window title for platform fallbacks.
func NewWithWindowTitle(windowTitle string) *Guard {
	return NewNamedWithWindowTitle(defaultName, windowTitle)
}

// NewNamed returns a Guard for tests or custom identifiers.
func NewNamed(name string) *Guard {
	return NewNamedWithWindowTitle(name, "")
}

// NewNamedWithWindowTitle returns a Guard for tests with an optional window title.
func NewNamedWithWindowTitle(name, windowTitle string) *Guard {
	return &Guard{
		name:        name,
		windowTitle: windowTitle,
		stopCh:      make(chan struct{}),
		platform:    newPlatformGuard(name, windowTitle),
	}
}

// Acquire tries to become the sole running instance.
// When acquired is false, another instance already holds the lock.
// On success the activation listener is bound immediately.
func (g *Guard) Acquire() (acquired bool, err error) {
	g.releaseMu.Lock()
	defer g.releaseMu.Unlock()
	if g.released {
		return false, errGuardReleased
	}
	g.platformMu.Lock()
	defer g.platformMu.Unlock()
	if g.platform == nil {
		return false, errGuardReleased
	}
	acquired, err = g.platform.acquire()
	if err != nil || !acquired {
		return acquired, err
	}
	if err := g.startListenerLocked(); err != nil {
		g.platform.release()
		return false, err
	}
	return true, nil
}

// ActivateWindow brings the configured main window to the foreground using native APIs.
func (g *Guard) ActivateWindow() error {
	return ActivateWindowByTitle(g.windowTitle)
}

// NotifyExisting asks the running instance to activate its main window.
func (g *Guard) NotifyExisting() error {
	g.releaseMu.Lock()
	defer g.releaseMu.Unlock()
	if g.released {
		return errGuardReleased
	}
	g.platformMu.Lock()
	defer g.platformMu.Unlock()
	if g.platform == nil {
		return errGuardReleased
	}
	return g.platform.notifyExisting()
}

// SetOnActivate registers a callback for second-launch activation requests.
// Any activation requests received before registration are coalesced to at most one
// pending activation and drained once when the callback is registered.
// Passing nil clears the callback and discards any queued activation.
// Callbacks are serialized; they must not block indefinitely.
func (g *Guard) SetOnActivate(fn func()) {
	g.activateMu.Lock()
	g.onActivate = fn
	pending := g.pending
	g.pending = false
	g.activateMu.Unlock()
	if fn == nil || !pending {
		return
	}
	g.runActivate(fn)
}

// Start binds the activation listener. Acquire already binds the listener on success;
// Start remains for idempotent re-entry and tests.
func (g *Guard) Start() error {
	g.releaseMu.Lock()
	defer g.releaseMu.Unlock()
	if g.released {
		return errGuardReleased
	}
	g.platformMu.Lock()
	defer g.platformMu.Unlock()
	if g.platform == nil {
		return errGuardReleased
	}
	return g.startListenerLocked()
}

func (g *Guard) startListenerLocked() error {
	g.startMu.Lock()
	defer g.startMu.Unlock()
	if g.started {
		return g.startErr
	}
	g.started = true
	if g.platform == nil {
		g.startErr = errGuardReleased
		return g.startErr
	}
	g.startErr = g.platform.start(g.stopCh, g.dispatchActivate)
	return g.startErr
}

func (g *Guard) dispatchActivate() {
	g.activateMu.Lock()
	fn := g.onActivate
	if fn == nil {
		g.pending = true
		g.activateMu.Unlock()
		return
	}
	g.activateMu.Unlock()
	g.runActivate(fn)
}

func (g *Guard) runActivate(fn func()) {
	g.invokeMu.Lock()
	defer g.invokeMu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("singleinstance: activation callback panic: %v", r)
		}
	}()
	fn()
}

// Release releases the lock and stops the activation listener.
func (g *Guard) Release() {
	g.releaseMu.Lock()
	defer g.releaseMu.Unlock()
	if g.released {
		return
	}
	g.released = true
	select {
	case <-g.stopCh:
	default:
		close(g.stopCh)
	}

	g.platformMu.Lock()
	defer g.platformMu.Unlock()
	if g.platform != nil {
		g.platform.release()
		g.platform = nil
	}
}

type platformGuard interface {
	acquire() (bool, error)
	notifyExisting() error
	start(stop <-chan struct{}, dispatch func()) error
	release()
}
