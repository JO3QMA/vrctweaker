package logwatcher

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestPollWatcherState_tryStart_exclusive(t *testing.T) {
	t.Parallel()
	s := newPollWatcherState()
	var started atomic.Int32
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if s.tryStart() {
				started.Add(1)
			}
		}()
	}
	wg.Wait()
	if started.Load() != 1 {
		t.Fatalf("started = %d, want 1", started.Load())
	}
	status, _ := s.Status()
	if status != string(statusRunning) {
		t.Fatalf("status = %q, want running", status)
	}
}

func TestPollWatcherState_restartCycle(t *testing.T) {
	t.Parallel()
	s := newPollWatcherState()

	if !s.tryStart() {
		t.Fatal("first tryStart should succeed")
	}
	s.setStopped()
	status, _ := s.Status()
	if status != string(statusStopped) {
		t.Fatalf("after stop status = %q, want stopped", status)
	}

	if !s.tryStart() {
		t.Fatal("restart tryStart should succeed from stopped")
	}
	status, _ = s.Status()
	if status != string(statusRunning) {
		t.Fatalf("after restart status = %q, want running", status)
	}
}

func TestPollWatcherState_setErrStatus_concurrent(t *testing.T) {
	t.Parallel()
	s := newPollWatcherState()
	if !s.tryStart() {
		t.Fatal("tryStart")
	}
	want := errors.New("open failed")
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			s.setErr(want)
		}()
		go func() {
			defer wg.Done()
			_, _ = s.Status()
		}()
	}
	wg.Wait()
	_, lastErr := s.Status()
	if !errors.Is(lastErr, want) {
		t.Fatalf("lastErr = %v, want %v", lastErr, want)
	}
}

func TestPollWatcherState_setStopped_retainsLastErr(t *testing.T) {
	t.Parallel()
	s := newPollWatcherState()
	if !s.tryStart() {
		t.Fatal("tryStart")
	}
	want := errors.New("resolve failed")
	s.setErr(want)
	s.setStopped()
	status, lastErr := s.Status()
	if status != string(statusStopped) {
		t.Fatalf("status = %q, want stopped", status)
	}
	if !errors.Is(lastErr, want) {
		t.Fatalf("lastErr = %v, want retained %v", lastErr, want)
	}
}

func TestPollWatcherState_setStopped_onlyFromRunning(t *testing.T) {
	t.Parallel()
	s := newPollWatcherState()
	s.setStopped()
	status, _ := s.Status()
	if status != string(statusIdle) {
		t.Fatalf("idle setStopped status = %q, want idle", status)
	}

	if !s.tryStart() {
		t.Fatal("tryStart")
	}
	s.setStopped()
	status, _ = s.Status()
	if status != string(statusStopped) {
		t.Fatalf("running setStopped status = %q, want stopped", status)
	}

	s.setStopped()
	status, _ = s.Status()
	if status != string(statusStopped) {
		t.Fatalf("stopped setStopped status = %q, want stopped", status)
	}
}
