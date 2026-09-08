// Package singleinstance ensures only one VRChat Tweaker process runs at a time.
package singleinstance

import (
	"errors"
	"sync"
)

const (
	defaultName = "VRChatTweaker"

	// DefaultWindowTitle is the Wails window title and Windows FindWindow fallback.
	DefaultWindowTitle = "VRChat Tweaker"
)

var errGuardReleased = errors.New("singleinstance: guard released")

// Guard coordinates single-instance locking and second-launch activation.
type Guard struct {
	name       string
	onActivate func()
	activateMu sync.Mutex
	pending    int
	startMu    sync.Mutex
	started    bool
	startErr   error
	stopCh     chan struct{}
	releaseMu  sync.Mutex
	released   bool
	platformMu sync.Mutex
	platform   platformGuard
}

// New returns a Guard using the default application identifier.
func New() *Guard {
	return NewNamed(defaultName)
}

// NewNamed returns a Guard for tests or custom identifiers.
func NewNamed(name string) *Guard {
	return &Guard{
		name:     name,
		stopCh:   make(chan struct{}),
		platform: newPlatformGuard(name),
	}
}

// Acquire tries to become the sole running instance.
// When acquired is false, another instance already holds the lock.
func (g *Guard) Acquire() (acquired bool, err error) {
	g.platformMu.Lock()
	defer g.platformMu.Unlock()
	if g.released || g.platform == nil {
		return false, errGuardReleased
	}
	return g.platform.acquire()
}

// NotifyExisting asks the running instance to activate its main window.
func (g *Guard) NotifyExisting() error {
	g.platformMu.Lock()
	defer g.platformMu.Unlock()
	if g.platform == nil {
		return errGuardReleased
	}
	return g.platform.notifyExisting()
}

// SetOnActivate registers a callback for second-launch activation requests.
// Any activation requests received before registration are queued and drained once.
// Passing nil clears the callback and discards any queued activations.
func (g *Guard) SetOnActivate(fn func()) {
	g.activateMu.Lock()
	g.onActivate = fn
	pending := g.pending
	g.pending = 0
	g.activateMu.Unlock()
	if fn == nil {
		return
	}
	for i := 0; i < pending; i++ {
		fn()
	}
}

// Start binds the activation listener. Call immediately after a successful Acquire,
// before the UI is ready, so second launches can connect while the first is starting.
func (g *Guard) Start() error {
	g.startMu.Lock()
	defer g.startMu.Unlock()
	if g.started {
		return g.startErr
	}
	g.started = true

	g.platformMu.Lock()
	defer g.platformMu.Unlock()
	if g.released || g.platform == nil {
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
		g.pending++
		g.activateMu.Unlock()
		return
	}
	g.activateMu.Unlock()
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
