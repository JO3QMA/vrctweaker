package discordrpc

import (
	"fmt"
	"sync"
	"time"
)

type frameWaiter struct {
	mu      sync.Mutex
	pending map[string]chan frameMessage
}

func newFrameWaiter() *frameWaiter {
	return &frameWaiter{pending: make(map[string]chan frameMessage)}
}

func (w *frameWaiter) register(nonce string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pending[nonce] = make(chan frameMessage, 1)
}

func (w *frameWaiter) unregister(nonce string) {
	w.mu.Lock()
	delete(w.pending, nonce)
	w.mu.Unlock()
}

func (w *frameWaiter) wait(nonce string, timeout time.Duration) (frameMessage, error) {
	w.mu.Lock()
	ch, ok := w.pending[nonce]
	w.mu.Unlock()
	if !ok {
		return frameMessage{}, fmt.Errorf("discord_rpc_no_waiter")
	}
	select {
	case msg := <-ch:
		return msg, nil
	case <-time.After(timeout):
		return frameMessage{}, fmt.Errorf("discord_rpc_timeout")
	}
}

func (w *frameWaiter) deliver(msg frameMessage) bool {
	if msg.Nonce == "" {
		return false
	}
	w.mu.Lock()
	ch, ok := w.pending[msg.Nonce]
	w.mu.Unlock()
	if !ok {
		return false
	}
	select {
	case ch <- msg:
		return true
	default:
		return false
	}
}
