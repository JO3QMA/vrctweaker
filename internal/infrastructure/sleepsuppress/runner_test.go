package sleepsuppress

import (
	"context"
	"errors"
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
	mu        sync.Mutex
	calls     []bool
	err       error
	errOnTrue bool
}

func (f *fakeExec) SetSuppress(on bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	if f.errOnTrue && on {
		return errSetSuppressTrue
	}
	f.calls = append(f.calls, on)
	return nil
}

var errSetSuppressTrue = errors.New("SetSuppress(true) failed")

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

func TestRun_settingsError_continuesWithoutSuppress(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := &fakeSettings{err: errors.New("settings unavailable")}
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

	for _, c := range e.snapshot() {
		if c {
			t.Fatalf("expected no SetSuppress(true) on settings error, calls=%v", e.snapshot())
		}
	}
}

func TestRun_procError_continuesWithoutSuppress(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := &fakeSettings{val: true}
	p := &fakeProc{err: errors.New("proc check failed")}
	e := &fakeExec{}

	done := make(chan struct{})
	go func() {
		_ = Run(ctx, 15*time.Millisecond, s, p, e)
		close(done)
	}()

	time.Sleep(40 * time.Millisecond)
	cancel()
	<-done

	for _, c := range e.snapshot() {
		if c {
			t.Fatalf("expected no SetSuppress(true) on proc error, calls=%v", e.snapshot())
		}
	}
}

func TestRun_setSuppressTrueError_doesNotMarkSuppressing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := &fakeSettings{val: true}
	p := &fakeProc{val: true}
	e := &fakeExec{errOnTrue: true}

	done := make(chan struct{})
	go func() {
		_ = Run(ctx, 15*time.Millisecond, s, p, e)
		close(done)
	}()

	time.Sleep(40 * time.Millisecond)
	cancel()
	<-done

	calls := e.snapshot()
	for _, c := range calls {
		if c {
			t.Fatalf("SetSuppress(true) should fail and not be recorded, calls=%v", calls)
		}
	}
}

func TestNewExecutionState_stub(t *testing.T) {
	es := NewExecutionState()
	if err := es.SetSuppress(true); err != nil {
		t.Fatalf("SetSuppress(true): %v", err)
	}
	if err := es.SetSuppress(false); err != nil {
		t.Fatalf("SetSuppress(false): %v", err)
	}
}

func TestNewVRChatProcessChecker_stub(t *testing.T) {
	pc := NewVRChatProcessChecker()
	running, err := pc.VRChatRunning()
	if err != nil {
		t.Fatalf("VRChatRunning: %v", err)
	}
	if running {
		t.Fatal("stub checker should report VRChat as not running")
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
