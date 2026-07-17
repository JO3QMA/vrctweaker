package logwatcher

import (
	"context"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

func joinLine(name, userID string) string {
	return "2026.03.18 00:17:57 Debug      -  [Behaviour] OnPlayerJoined " + name + " (" + userID + ")\n"
}

func newMultiWatcherForTest(
	t *testing.T,
	dir string,
	handlerFactory func(string) EventHandler,
	callbacks MultiOutputLogWatcherCallbacks,
) *MultiOutputLogWatcher {
	t.Helper()
	w := NewMultiOutputLogWatcher(dir, activity.NewLogParser(), handlerFactory, callbacks, nil)
	w.activeLogStallTimeout = 400 * time.Millisecond
	return w
}

func TestMultiOutputLogWatcher_ParallelAppendTwoFiles(t *testing.T) {
	dir := t.TempDir()
	pathA := filepath.Join(dir, "output_log_a.txt")
	pathB := filepath.Join(dir, "output_log_b.txt")
	if err := writeTestFile(pathA, ""); err != nil {
		t.Fatal(err)
	}
	if err := writeTestFile(pathB, ""); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	byPath := map[string]int{}
	handlerFactory := func(logPath string) EventHandler {
		p := logPath
		return testEventHandler(func(activity.ParsedEvent) {
			mu.Lock()
			byPath[p]++
			mu.Unlock()
		})
	}

	w := newMultiWatcherForTest(t, dir, handlerFactory, MultiOutputLogWatcherCallbacks{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)

	if err := appendToTestFile(pathA, joinLine("UserA", "usr_a001")); err != nil {
		t.Fatal(err)
	}
	if err := appendToTestFile(pathB, joinLine("UserB", "usr_b001")); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		gotA, gotB := byPath[pathA], byPath[pathB]
		mu.Unlock()
		if gotA >= 1 && gotB >= 1 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	t.Fatalf("events by path: A=%d B=%d, want >=1 each", byPath[pathA], byPath[pathB])
}

func TestMultiOutputLogWatcher_StallStopsTailWithoutHandoff(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log_stall.txt")
	if err := writeTestFile(path, ""); err != nil {
		t.Fatal(err)
	}

	var handoff atomic.Int32
	var mu sync.Mutex
	var count int
	handlerFactory := func(string) EventHandler {
		return testEventHandler(func(activity.ParsedEvent) {
			mu.Lock()
			count++
			mu.Unlock()
		})
	}
	callbacks := MultiOutputLogWatcherCallbacks{
		OnLogRotationHandoff: func(context.Context, string) error {
			handoff.Add(1)
			return nil
		},
	}

	w := newMultiWatcherForTest(t, dir, handlerFactory, callbacks)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)

	if err := appendToTestFile(path, joinLine("StallUser", "usr_stall1")); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := count
		mu.Unlock()
		if n >= 1 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	mu.Lock()
	firstCount := count
	mu.Unlock()
	if firstCount < 1 {
		t.Fatal("expected first append to be tailed")
	}

	// Wait past stall timeout with no further growth.
	time.Sleep(w.activeLogStallTimeout + 600*time.Millisecond)

	if err := appendToTestFile(path, joinLine("AfterStall", "usr_stall2")); err != nil {
		t.Fatal(err)
	}

	// Stall should have stopped the tail goroutine; no handoff because no other file grew.
	if handoff.Load() != 0 {
		t.Fatalf("handoff callbacks = %d, want 0 on stall-only", handoff.Load())
	}

	// New growth should restart tail and deliver the second line.
	deadline = time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := count
		mu.Unlock()
		if n >= 2 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	t.Fatalf("event count = %d after stall regrowth, want >= 2", count)
}

func TestMultiOutputLogWatcher_RotationHandoff(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "output_log_2026-01-01_00-00-00.txt")
	newPath := filepath.Join(dir, "output_log_2026-03-22_00-47-45.txt")
	if err := writeTestFile(oldPath, ""); err != nil {
		t.Fatal(err)
	}

	var handoffPaths []string
	var handoffMu sync.Mutex
	handlerFactory := func(string) EventHandler {
		return testEventHandler(func(activity.ParsedEvent) {})
	}
	callbacks := MultiOutputLogWatcherCallbacks{
		OnLogRotationHandoff: func(_ context.Context, oldPath string) error {
			handoffMu.Lock()
			handoffPaths = append(handoffPaths, oldPath)
			handoffMu.Unlock()
			return nil
		},
	}

	w := newMultiWatcherForTest(t, dir, handlerFactory, callbacks)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)

	if err := appendToTestFile(oldPath, joinLine("OldUser", "usr_old001")); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		// allow old file tail to start
		time.Sleep(50 * time.Millisecond)
		break
	}

	if err := writeTestFile(newPath, ""); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)
	if err := appendToTestFile(newPath, joinLine("NewUser", "usr_new001")); err != nil {
		t.Fatal(err)
	}

	deadline = time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		handoffMu.Lock()
		n := len(handoffPaths)
		handoffMu.Unlock()
		if n >= 1 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	handoffMu.Lock()
	defer handoffMu.Unlock()
	if len(handoffPaths) != 1 {
		t.Fatalf("handoff count = %d, want 1", len(handoffPaths))
	}
	if handoffPaths[0] != oldPath {
		t.Fatalf("handoff path = %q, want %q", handoffPaths[0], oldPath)
	}
}

func TestMultiOutputLogWatcher_NewFileAppears(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "output_log_existing.txt")
	newPath := filepath.Join(dir, "output_log_new.txt")
	if err := writeTestFile(existing, ""); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	seen := map[string]bool{}
	handlerFactory := func(logPath string) EventHandler {
		p := logPath
		return testEventHandler(func(activity.ParsedEvent) {
			mu.Lock()
			seen[p] = true
			mu.Unlock()
		})
	}

	w := newMultiWatcherForTest(t, dir, handlerFactory, MultiOutputLogWatcherCallbacks{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)

	if err := writeTestFile(newPath, joinLine("BrandNew", "usr_brand1")); err != nil {
		t.Fatal(err)
	}
	if err := appendToTestFile(newPath, joinLine("BrandNew2", "usr_brand2")); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		ok := seen[newPath]
		mu.Unlock()
		if ok {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("new file events not received")
}

func TestMultiOutputLogWatcher_StopsOnCancel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := writeTestFile(path, ""); err != nil {
		t.Fatal(err)
	}

	w := newMultiWatcherForTest(t, dir, func(string) EventHandler {
		return testEventHandler(func(activity.ParsedEvent) {})
	}, MultiOutputLogWatcherCallbacks{})
	ctx, cancel := context.WithCancel(context.Background())
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	cancel()
	time.Sleep(200 * time.Millisecond)
	status, _ := w.Status()
	if status != "stopped" {
		t.Fatalf("status = %q, want stopped", status)
	}
}

func TestMultiOutputLogWatcher_OnTailCheckpoint(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log_ckpt.txt")
	if err := writeTestFile(path, ""); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var checkpoints []struct {
		path   string
		offset int64
	}
	handlerFactory := func(string) EventHandler {
		return testEventHandler(func(activity.ParsedEvent) {})
	}
	callbacks := MultiOutputLogWatcherCallbacks{
		OnTailCheckpoint: func(_ context.Context, p string, offset int64, _ time.Time) {
			mu.Lock()
			checkpoints = append(checkpoints, struct {
				path   string
				offset int64
			}{p, offset})
			mu.Unlock()
		},
	}

	w := newMultiWatcherForTest(t, dir, handlerFactory, callbacks)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)

	line := joinLine("CkptUser", "usr_ckpt01")
	if err := appendToTestFile(path, line); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(checkpoints)
		mu.Unlock()
		if n >= 1 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(checkpoints) == 0 {
		t.Fatal("expected at least one checkpoint callback")
	}
	last := checkpoints[len(checkpoints)-1]
	if last.path != path {
		t.Fatalf("checkpoint path = %q, want %q", last.path, path)
	}
	if last.offset <= 0 {
		t.Fatalf("checkpoint offset = %d, want > 0", last.offset)
	}
}

func TestMultiOutputLogWatcher_TailGoroutineExitClearsTailing(t *testing.T) {
	dir := t.TempDir()
	missingPath := filepath.Join(dir, "output_log_missing.txt")
	existingPath := filepath.Join(dir, "output_log_existing.txt")

	w := newMultiWatcherForTest(t, dir, func(string) EventHandler {
		return testEventHandler(func(activity.ParsedEvent) {})
	}, MultiOutputLogWatcherCallbacks{})

	state := &trackedLogFile{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w.startTail(ctx, missingPath, state, 0)
	time.Sleep(150 * time.Millisecond)
	if state.tailing.Load() {
		t.Fatal("tailing should be false after tail goroutine exits on open error")
	}

	if err := writeTestFile(existingPath, ""); err != nil {
		t.Fatal(err)
	}
	state = &trackedLogFile{}
	w.startTail(ctx, existingPath, state, 0)
	time.Sleep(50 * time.Millisecond)
	if !state.tailing.Load() {
		t.Fatal("tailing should be true after successful restart")
	}
	cancel()
	time.Sleep(50 * time.Millisecond)
}
