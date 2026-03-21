package picturewatcher

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestRun_scheduleIngest_coalescesSamePath(t *testing.T) {
	ctx := context.Background()
	var mu sync.Mutex
	calls := 0
	ing := func(context.Context, string) error {
		mu.Lock()
		calls++
		mu.Unlock()
		return nil
	}
	r := &run{
		ctx:      ctx,
		ingest:   ing,
		log:      nopLogger{},
		debounce: 50 * time.Millisecond,
	}
	r.scheduleIngest("/tmp/a.png")
	r.scheduleIngest("/tmp/a.png")
	time.Sleep(120 * time.Millisecond)
	mu.Lock()
	n := calls
	mu.Unlock()
	if n != 1 {
		t.Fatalf("ingest calls = %d, want 1", n)
	}
}

func TestRun_scheduleIngest_flushesMultiplePaths(t *testing.T) {
	ctx := context.Background()
	var mu sync.Mutex
	var paths []string
	ing := func(_ context.Context, p string) error {
		mu.Lock()
		paths = append(paths, p)
		mu.Unlock()
		return nil
	}
	r := &run{
		ctx:      ctx,
		ingest:   ing,
		log:      nopLogger{},
		debounce: 50 * time.Millisecond,
	}
	r.scheduleIngest("/a.png")
	r.scheduleIngest("/b.png")
	time.Sleep(120 * time.Millisecond)
	mu.Lock()
	got := append([]string(nil), paths...)
	mu.Unlock()
	if len(got) != 2 {
		t.Fatalf("ingest paths = %d, want 2: %v", len(got), got)
	}
}

func TestRun_stopFlushTimer_clearsPending(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var mu sync.Mutex
	calls := 0
	ing := func(context.Context, string) error {
		mu.Lock()
		calls++
		mu.Unlock()
		return nil
	}
	r := &run{
		ctx:      ctx,
		ingest:   ing,
		log:      nopLogger{},
		debounce: 200 * time.Millisecond,
	}
	r.scheduleIngest("/x.png")
	cancel()
	r.stopFlushTimer()
	time.Sleep(250 * time.Millisecond)
	mu.Lock()
	n := calls
	mu.Unlock()
	if n != 0 {
		t.Fatalf("ingest calls = %d, want 0 after stop", n)
	}
}
