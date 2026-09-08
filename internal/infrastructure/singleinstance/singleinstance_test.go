package singleinstance

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAcquireExclusive(t *testing.T) {
	name := "vrctweaker-test-" + uuid.NewString()

	first := NewNamed(name)
	second := NewNamed(name)

	acquired, err := first.Acquire()
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	if !acquired {
		t.Fatal("first instance should acquire lock")
	}
	t.Cleanup(func() { first.Release() })

	acquired, err = second.Acquire()
	if err != nil {
		t.Fatalf("second Acquire: %v", err)
	}
	if acquired {
		t.Fatal("second instance should not acquire lock")
	}
}

func TestReleaseAllowsReacquire(t *testing.T) {
	name := "vrctweaker-test-" + uuid.NewString()

	first := NewNamed(name)
	acquired, err := first.Acquire()
	if err != nil || !acquired {
		t.Fatalf("first Acquire: acquired=%v err=%v", acquired, err)
	}
	first.Release()

	replacement := NewNamed(name)
	acquired, err = replacement.Acquire()
	if err != nil {
		t.Fatalf("replacement Acquire: %v", err)
	}
	if !acquired {
		t.Fatal("lock should be available after Release")
	}
	replacement.Release()
}

func TestNotifyExistingInvokesActivateCallback(t *testing.T) {
	name := "vrctweaker-test-" + uuid.NewString()

	holder := NewNamed(name)
	acquired, err := holder.Acquire()
	if err != nil || !acquired {
		t.Fatalf("holder Acquire: acquired=%v err=%v", acquired, err)
	}
	t.Cleanup(func() { holder.Release() })

	var activated atomic.Bool
	holder.SetOnActivate(func() { activated.Store(true) })

	launcher := NewNamed(name)
	if acquired, err := launcher.Acquire(); err != nil || acquired {
		t.Fatalf("launcher should see existing instance: acquired=%v err=%v", acquired, err)
	}
	if err := launcher.NotifyExisting(); err != nil {
		t.Fatalf("NotifyExisting: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for !activated.Load() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !activated.Load() {
		t.Fatal("activate callback was not invoked")
	}
}

func TestNotifyExistingQueuesUntilSetOnActivate(t *testing.T) {
	name := "vrctweaker-test-" + uuid.NewString()

	holder := NewNamed(name)
	acquired, err := holder.Acquire()
	if err != nil || !acquired {
		t.Fatalf("holder Acquire: acquired=%v err=%v", acquired, err)
	}
	t.Cleanup(func() { holder.Release() })

	launcher := NewNamed(name)
	if acquired, err := launcher.Acquire(); err != nil || acquired {
		t.Fatalf("launcher should see existing instance: acquired=%v err=%v", acquired, err)
	}
	if err := launcher.NotifyExisting(); err != nil {
		t.Fatalf("NotifyExisting before callback: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	var activated atomic.Bool
	for time.Now().Before(deadline) {
		holder.SetOnActivate(func() { activated.Store(true) })
		if activated.Load() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("queued activation should drain when callback is registered")
}

func TestStartIsIdempotent(t *testing.T) {
	name := "vrctweaker-test-" + uuid.NewString()
	g := NewNamed(name)
	acquired, err := g.Acquire()
	if err != nil || !acquired {
		t.Fatalf("Acquire: acquired=%v err=%v", acquired, err)
	}
	t.Cleanup(func() { g.Release() })

	if err := g.Start(); err != nil {
		t.Fatalf("first Start: %v", err)
	}
	if err := g.Start(); err != nil {
		t.Fatalf("second Start: %v", err)
	}
}

func TestReleaseIsIdempotent(t *testing.T) {
	name := "vrctweaker-test-" + uuid.NewString()
	g := NewNamed(name)
	acquired, err := g.Acquire()
	if err != nil || !acquired {
		t.Fatalf("Acquire: acquired=%v err=%v", acquired, err)
	}
	g.Release()
	g.Release()
}

func TestSetOnActivateNilDiscardsPending(t *testing.T) {
	g := NewNamed("vrctweaker-test-" + uuid.NewString())
	g.dispatchActivate()
	g.dispatchActivate()
	g.SetOnActivate(nil)

	var activated atomic.Bool
	g.SetOnActivate(func() { activated.Store(true) })
	if activated.Load() {
		t.Fatal("pending activations should be discarded when callback cleared with nil")
	}
}

func TestSetOnActivateCoalescesPendingActivations(t *testing.T) {
	g := NewNamed("vrctweaker-test-" + uuid.NewString())
	g.dispatchActivate()
	g.dispatchActivate()
	g.dispatchActivate()

	var count atomic.Int32
	g.SetOnActivate(func() { count.Add(1) })
	if count.Load() != 1 {
		t.Fatalf("pending activations coalesced to one callback, got %d calls", count.Load())
	}
}

func TestAbstractActivateAddrIncludesUIDAndFingerprint(t *testing.T) {
	fp := dataDirFingerprint("/tmp/example-app-data")
	addr := abstractActivateAddr("VRChatTweaker", fp)
	want := fmt.Sprintf("@VRChatTweaker_%d_%s_activate", os.Getuid(), fp)
	if addr != want {
		t.Fatalf("got %q want %q", addr, want)
	}
}

func TestAcquireBindsListener(t *testing.T) {
	name := "vrctweaker-test-" + uuid.NewString()
	holder := NewNamed(name)
	acquired, err := holder.Acquire()
	if err != nil || !acquired {
		t.Fatalf("Acquire: acquired=%v err=%v", acquired, err)
	}
	t.Cleanup(func() { holder.Release() })

	var activated atomic.Bool
	holder.SetOnActivate(func() { activated.Store(true) })

	launcher := NewNamed(name)
	if acquired, err := launcher.Acquire(); err != nil || acquired {
		t.Fatalf("launcher should see existing instance: acquired=%v err=%v", acquired, err)
	}
	if err := launcher.NotifyExisting(); err != nil {
		t.Fatalf("NotifyExisting: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for !activated.Load() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !activated.Load() {
		t.Fatal("listener bound by Acquire should accept activation")
	}
}
