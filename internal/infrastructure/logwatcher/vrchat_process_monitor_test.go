package logwatcher

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type stubVRChatRunningChecker struct {
	running atomic.Bool
}

func (s *stubVRChatRunningChecker) VRChatRunning() (bool, error) {
	return s.running.Load(), nil
}

func TestMonitorVRChatRunning_callsOnStoppedOnRunningToStopped(t *testing.T) {
	checker := &stubVRChatRunningChecker{}
	checker.running.Store(true)
	var stopped atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		_ = MonitorVRChatRunning(ctx, 10*time.Millisecond, checker, func() {
			stopped.Add(1)
		})
		close(done)
	}()

	time.Sleep(25 * time.Millisecond)
	checker.running.Store(false)
	time.Sleep(40 * time.Millisecond)
	cancel()
	<-done

	if stopped.Load() != 1 {
		t.Fatalf("onStopped calls = %d, want 1", stopped.Load())
	}
}

func TestMonitorVRChatRunning_notCalledWhenNeverRunning(t *testing.T) {
	checker := &stubVRChatRunningChecker{}
	var stopped atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		_ = MonitorVRChatRunning(ctx, 10*time.Millisecond, checker, func() {
			stopped.Add(1)
		})
		close(done)
	}()

	time.Sleep(35 * time.Millisecond)
	cancel()
	<-done

	if stopped.Load() != 0 {
		t.Fatalf("onStopped calls = %d, want 0", stopped.Load())
	}
}
