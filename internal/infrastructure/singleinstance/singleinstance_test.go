package singleinstance

import (
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

	if err := holder.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

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

	if err := holder.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

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

func TestDefaultWindowTitle(t *testing.T) {
	if DefaultWindowTitle == "" {
		t.Fatal("DefaultWindowTitle must not be empty")
	}
}
