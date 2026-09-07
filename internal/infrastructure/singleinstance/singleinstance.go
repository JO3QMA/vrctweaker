// Package singleinstance ensures only one VRChat Tweaker process runs at a time.
package singleinstance

import "sync"

const defaultName = "VRChatTweaker"

// Guard coordinates single-instance locking and second-launch activation.
type Guard struct {
	name       string
	onActivate func()
	startOnce  sync.Once
	stopCh     chan struct{}
	releaseMu  sync.Mutex
	released   bool
	platform   platformGuard
}

// New returns a Guard using the default application identifier.
func New() *Guard {
	return NewNamed(defaultName)
}

// NewNamed returns a Guard for tests or custom identifiers.
func NewNamed(name string) *Guard {
	return &Guard{
		name:   name,
		stopCh: make(chan struct{}),
	}
}

// Acquire tries to become the sole running instance.
// When acquired is false, another instance already holds the lock.
func (g *Guard) Acquire() (acquired bool, err error) {
	if g.platform == nil {
		g.platform = newPlatformGuard(g.name)
	}
	return g.platform.acquire()
}

// NotifyExisting asks the running instance to activate its main window.
func (g *Guard) NotifyExisting() error {
	if g.platform == nil {
		g.platform = newPlatformGuard(g.name)
	}
	return g.platform.notifyExisting()
}

// SetOnActivate registers a callback for second-launch activation requests.
func (g *Guard) SetOnActivate(fn func()) {
	g.onActivate = fn
}

// Start listens for activation requests from another process.
func (g *Guard) Start() {
	g.startOnce.Do(func() {
		if g.platform == nil {
			return
		}
		g.platform.start(g.stopCh, func() {
			if fn := g.onActivate; fn != nil {
				fn()
			}
		})
	})
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
	if g.platform != nil {
		g.platform.release()
		g.platform = nil
	}
}

type platformGuard interface {
	acquire() (bool, error)
	notifyExisting() error
	start(stop <-chan struct{}, onActivate func())
	release()
}
