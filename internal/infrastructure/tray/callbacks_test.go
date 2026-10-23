package tray

import (
	"sync"
	"testing"
	"time"
)

func TestRunAsync_invokesCallback(t *testing.T) {
	var mu sync.Mutex
	done := false
	runAsync(nil, func() {
		mu.Lock()
		done = true
		mu.Unlock()
	})

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		ok := done
		mu.Unlock()
		if ok {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("callback was not invoked")
}

func TestRunAsync_nilSafe(t *testing.T) {
	runAsync(nil, nil)
}

func TestRunAsync_recoversPanic(t *testing.T) {
	done := make(chan struct{})
	runAsync(nil, func() {
		defer func() { close(done) }()
		panic("tray test panic")
	})
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not finish after panic")
	}
}
