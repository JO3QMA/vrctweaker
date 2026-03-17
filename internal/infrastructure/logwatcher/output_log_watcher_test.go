package logwatcher

import (
	"context"
	"os"
	"path/filepath"
	"sync"
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
	lines := []byte("OnPlayerJoined TestUser (usr_abc123)\nOnPlayerLeft TestUser (usr_abc123)\n")
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
