package tray

import (
	"sync"
	"testing"
	"time"
)

// Regression: OnQuit calls Stop() from an async handler. If runSession waits on the
// same WaitGroup before signaling stop waiters, shutdown deadlocks.
func TestSessionShutdown_signalsStopWaitersBeforeDrainingShowHandlers(t *testing.T) {
	waiters := &stopWaiterSet{}
	stopCh := waiters.add()

	var showHandlers sync.WaitGroup
	showHandlers.Add(1)

	unblocked := make(chan struct{})
	go func() {
		waiters.complete(nil)
		close(unblocked)
	}()

	select {
	case <-unblocked:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stop waiters were not completed")
	}

	select {
	case err := <-stopCh:
		if err != nil {
			t.Fatalf("unexpected stop error: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stop waiter did not receive completion")
	}

	showHandlers.Done()
	showHandlers.Wait()
}

func TestStopWaiterSet_broadcastsToConcurrentWaiters(t *testing.T) {
	waiters := &stopWaiterSet{}
	a := waiters.add()
	b := waiters.add()

	waiters.complete(nil)

	if err := <-a; err != nil {
		t.Fatalf("waiter a: %v", err)
	}
	if err := <-b; err != nil {
		t.Fatalf("waiter b: %v", err)
	}
}
