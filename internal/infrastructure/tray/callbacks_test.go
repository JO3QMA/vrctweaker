package tray

import (
	"sync"
	"testing"
	"time"
)

func TestRunAsync_invokesCallback(t *testing.T) {
	var mu sync.Mutex
	done := false
	runAsync(func() {
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
	runAsync(nil)
}
