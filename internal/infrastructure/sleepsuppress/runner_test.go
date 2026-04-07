package sleepsuppress

import (
	"context"
	"sync"
	"testing"
	"time"
)

type fakeSettings struct {
	mu  sync.Mutex
	val bool
	err error
}

func (f *fakeSettings) SuppressSleepWhileVRChat(ctx context.Context) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.val, f.err
}

type fakeProc struct {
	mu  sync.Mutex
	val bool
	err error
}

func (f *fakeProc) VRChatRunning() (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.val, f.err
}

func (f *fakeProc) set(v bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.val = v
}

type fakeExec struct {
	mu    sync.Mutex
	calls []bool
	err   error
}

func (f *fakeExec) SetSuppress(on bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, on)
	return nil
}

func (f *fakeExec) snapshot() []bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]bool, len(f.calls))
	copy(out, f.calls)
	return out
}

func TestRun_settingOff_neverSuppresses(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := &fakeSettings{val: false}
	p := &fakeProc{val: true}
	e := &fakeExec{}

	done := make(chan struct{})
	go func() {
		_ = Run(ctx, 15*time.Millisecond, s, p, e)
		close(done)
	}()

	time.Sleep(45 * time.Millisecond)
	cancel()
	<-done

	calls := e.snapshot()
	for _, c := range calls {
		if c {
			t.Fatalf("expected no SetSuppress(true), got calls %v", calls)
		}
	}
}

func TestRun_settingOnAndRunning_setsSuppressTrue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := &fakeSettings{val: true}
	p := &fakeProc{val: true}
	e := &fakeExec{}

	done := make(chan struct{})
	go func() {
		_ = Run(ctx, 15*time.Millisecond, s, p, e)
		close(done)
	}()

	time.Sleep(40 * time.Millisecond)
	cancel()
	<-done

	calls := e.snapshot()
	foundTrue := false
	for _, c := range calls {
		if c {
			foundTrue = true
			break
		}
	}
	if !foundTrue {
		t.Fatalf("expected SetSuppress(true) at least once, calls=%v", calls)
	}
}

func TestRun_vrchatStops_clearsSuppress(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := &fakeSettings{val: true}
	p := &fakeProc{val: true}
	e := &fakeExec{}

	done := make(chan struct{})
	go func() {
		_ = Run(ctx, 15*time.Millisecond, s, p, e)
		close(done)
	}()

	time.Sleep(35 * time.Millisecond)
	p.set(false)
	time.Sleep(35 * time.Millisecond)
	cancel()
	<-done

	calls := e.snapshot()
	if len(calls) == 0 {
		t.Fatal("expected some SetSuppress calls")
	}
	last := calls[len(calls)-1]
	if last {
		t.Fatalf("expected final SetSuppress to be false, calls=%v", calls)
	}
}

func TestRun_cancelAlwaysClearsSuppress(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	s := &fakeSettings{val: true}
	p := &fakeProc{val: true}
	e := &fakeExec{}

	done := make(chan struct{})
	go func() {
		_ = Run(ctx, 15*time.Millisecond, s, p, e)
		close(done)
	}()

	time.Sleep(25 * time.Millisecond)
	cancel()
	<-done

	calls := e.snapshot()
	if len(calls) == 0 {
		t.Fatal("expected at least one SetSuppress on shutdown")
	}
	if calls[len(calls)-1] {
		t.Fatalf("expected last call SetSuppress(false), calls=%v", calls)
	}
}
