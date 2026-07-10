package logwatcher

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

func TestOutputLogWatcher_EmitsEvents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := os.WriteFile(path, []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	parser := activity.NewLogParser()
	var mu sync.Mutex
	var received []activity.ParsedEvent
	handler := EventHandlerFunc(func(ev activity.ParsedEvent) {
		mu.Lock()
		received = append(received, ev)
		mu.Unlock()
	})

	watcher := NewOutputLogWatcher(path, parser, handler, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := watcher.Start(ctx); err != nil {
		t.Fatal(err)
	}

	// Give watcher time to open file and seek to end
	time.Sleep(200 * time.Millisecond)

	// Append join/leave lines
	lines := []byte("2026.03.18 00:17:57 Debug      -  [Behaviour] OnPlayerJoined TestUser (usr_abc123)\n2026.03.18 00:17:58 Debug      -  [Behaviour] OnPlayerLeft TestUser (usr_abc123)\n")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(lines); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(received)
		mu.Unlock()
		if n >= 2 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	mu.Lock()
	got := len(received)
	mu.Unlock()
	if got < 2 {
		t.Errorf("handler received %d events, want at least 2", got)
	}
}

func TestOutputLogWatcher_Start_rejectsDirectory(t *testing.T) {
	dir := t.TempDir()
	w := NewOutputLogWatcher(dir, activity.NewLogParser(), EventHandlerFunc(func(activity.ParsedEvent) {}), nil)
	err := w.Start(context.Background())
	if err == nil {
		t.Fatal("Start: want error for directory path")
	}
	if !errors.Is(err, ErrOutputLogWatcherFileOnly) {
		t.Fatalf("Start err = %v, want ErrOutputLogWatcherFileOnly", err)
	}
}

func TestOutputLogWatcher_StopsOnCancel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := os.WriteFile(path, []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	parser := activity.NewLogParser()
	handler := EventHandlerFunc(func(activity.ParsedEvent) {})
	watcher := NewOutputLogWatcher(path, parser, handler, nil)

	ctx, cancel := context.WithCancel(context.Background())
	if err := watcher.Start(ctx); err != nil {
		t.Fatal(err)
	}

	cancel()
	time.Sleep(200 * time.Millisecond)

	status, _ := watcher.Status()
	if status != "stopped" {
		t.Errorf("Status() = %q, want stopped", status)
	}
}

func TestOutputLogWatcher_StartAlreadyRunning(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := writeTestFile(path, ""); err != nil {
		t.Fatal(err)
	}
	w := NewOutputLogWatcher(path, activity.NewLogParser(), EventHandlerFunc(func(activity.ParsedEvent) {}), func(string, ...any) {})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	status, _ := w.Status()
	if status != "running" {
		t.Fatalf("status = %q", status)
	}
}

func TestOutputLogWatcher_FileRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := writeTestFile(path, "seed\n"); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var received []activity.ParsedEvent
	parser := activity.NewLogParser()
	w := NewOutputLogWatcher(path, parser, EventHandlerFunc(func(ev activity.ParsedEvent) {
		mu.Lock()
		received = append(received, ev)
		mu.Unlock()
	}), nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)

	if err := writeTestFile(path, "2026.03.18 00:17:57 Debug      -  [Behaviour] OnPlayerJoined RotUser (usr_rot01)\n"); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(received)
		mu.Unlock()
		if n >= 1 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(received) < 1 {
		t.Fatalf("received %d events after rotation", len(received))
	}
}

func TestOutputLogWatcher_InvalidPathRetries(t *testing.T) {
	buf := &raceSafeLogBuffer{}
	w := NewOutputLogWatcher("", stubParser{}, EventHandlerFunc(func(activity.ParsedEvent) {}), buf.logger())
	ctx, cancel := context.WithCancel(context.Background())
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	cancel()
	time.Sleep(100 * time.Millisecond)
	status, lastErr := w.Status()
	if status != "stopped" {
		t.Fatalf("status = %q", status)
	}
	if lastErr == nil {
		t.Fatal("expected last error")
	}
	if buf.len() == 0 {
		t.Fatal("expected resolve/open error logs")
	}
}

func TestOutputLogWatcher_ParseErrorSkipsLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := writeTestFile(path, ""); err != nil {
		t.Fatal(err)
	}
	parser := stubParser{err: errors.New("bad parse")}
	buf := &raceSafeLogBuffer{}
	w := NewOutputLogWatcher(path, parser, EventHandlerFunc(func(activity.ParsedEvent) {}), buf.logger())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if err := appendToTestFile(path, "broken\n"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && buf.len() == 0 {
		time.Sleep(50 * time.Millisecond)
	}
	if buf.len() == 0 {
		t.Fatal("expected parse error log")
	}
}

func TestOutputLogWatcher_resolveActivePath_fixedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output_log.txt")
	w := NewOutputLogWatcher(path, stubParser{}, EventHandlerFunc(func(activity.ParsedEvent) {}), nil)
	got, err := w.resolveActivePath()
	if err != nil || got != path {
		t.Fatalf("resolveActivePath() = %q, %v", got, err)
	}
}

func TestOutputLogWatcher_ReopensAfterTruncate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := os.WriteFile(path, []byte("seed\n"), 0600); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var received []activity.ParsedEvent
	parser := activity.NewLogParser()
	w := NewOutputLogWatcher(path, parser, EventHandlerFunc(func(ev activity.ParsedEvent) {
		mu.Lock()
		received = append(received, ev)
		mu.Unlock()
	}), nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)

	if err := os.Truncate(path, 0); err != nil {
		t.Fatal(err)
	}
	time.Sleep(800 * time.Millisecond)
	if err := appendToTestFile(path, "2026.03.18 00:17:57 Debug      -  [Behaviour] OnPlayerJoined TruncUser (usr_trunc1)\n"); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(received)
		mu.Unlock()
		if n >= 1 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("received %d events after truncate", len(received))
}

func TestOutputLogWatcher_SkipsEmptyLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := os.WriteFile(path, []byte("\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var count atomic.Int32
	w := NewOutputLogWatcher(path, stubParser{events: []activity.ParsedEvent{&activity.EncounterEvent{}}}, EventHandlerFunc(func(activity.ParsedEvent) {
		count.Add(1)
	}), nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if err := appendToTestFile(path, "\n2026.03.18 00:17:57 Debug      -  [Behaviour] OnPlayerJoined E (usr_e1)\n"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && count.Load() == 0 {
		time.Sleep(50 * time.Millisecond)
	}
	if count.Load() == 0 {
		t.Fatal("expected handler call for non-empty line")
	}
}

func TestOutputLogWatcher_SkipsNilParsedEvents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := os.WriteFile(path, []byte(""), 0600); err != nil {
		t.Fatal(err)
	}
	var count atomic.Int32
	parser := stubParser{events: []activity.ParsedEvent{nil, &activity.EncounterEvent{DisplayName: "Z"}}}
	w := NewOutputLogWatcher(path, parser, EventHandlerFunc(func(activity.ParsedEvent) {
		count.Add(1)
	}), nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if err := appendToTestFile(path, "line\n"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && count.Load() == 0 {
		time.Sleep(50 * time.Millisecond)
	}
	if count.Load() != 1 {
		t.Fatalf("count = %d, want 1 non-nil event", count.Load())
	}
}

func TestOutputLogWatcher_ReopenWhenPathBecomesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := os.WriteFile(path, []byte("seed\n"), 0600); err != nil {
		t.Fatal(err)
	}
	buf := &raceSafeLogBuffer{}
	w := NewOutputLogWatcher(path, stubParser{events: []activity.ParsedEvent{&activity.EncounterEvent{}}}, EventHandlerFunc(func(activity.ParsedEvent) {}), buf.logger())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1200 * time.Millisecond)
	if buf.len() == 0 {
		t.Fatal("expected watcher to log while recovering from path change")
	}
}
