package logwatcher

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"vrchat-tweaker/internal/domain/activity"
)

func TestProcessOutputLogFile_DispatchesEvents(t *testing.T) {
	t.Setenv("TZ", "UTC")
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	content := "2026.03.21 11:32:04 Debug      -  [Behaviour] Joining wrld_db637cfb-64f8-4109-977b-6b755482f133:88577~region(jp)\n" +
		"2026.03.21 11:32:16 Debug      -  [Behaviour] OnPlayerJoined Alice (usr_abc)\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	parser := activity.NewLogParser()
	var mu sync.Mutex
	var kinds []activity.EventKind
	h := EventHandlerFunc(func(ev activity.ParsedEvent) {
		mu.Lock()
		kinds = append(kinds, ev.Kind())
		mu.Unlock()
	})

	if err := ProcessOutputLogFile(context.Background(), path, parser, h, nil); err != nil {
		t.Fatalf("ProcessOutputLogFile: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(kinds) != 2 {
		t.Fatalf("got %d events, want 2: %v", len(kinds), kinds)
	}
	if kinds[0] != activity.EventKindSession || kinds[1] != activity.EventKindEncounter {
		t.Fatalf("unexpected kinds: %v", kinds)
	}
}

func TestProcessOutputLogFile_topLevelWrapper(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := os.WriteFile(path, []byte("2026.03.21 11:32:16 Debug      -  [Behaviour] OnPlayerJoined Alice (usr_abc)\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var count atomic.Int32
	err := ProcessOutputLogFile(context.Background(), path, activity.NewLogParser(), EventHandlerFunc(func(activity.ParsedEvent) {
		count.Add(1)
	}), nil)
	if err != nil {
		t.Fatal(err)
	}
	if count.Load() != 1 {
		t.Fatalf("count = %d", count.Load())
	}
}

func TestProcessOutputLogFileFromOffset_ResumeAndProgress(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	content := "line1\nline2\nline3\n"
	if err := writeTestFile(path, content); err != nil {
		t.Fatal(err)
	}

	parser := activity.NewLogParser()
	h := EventHandlerFunc(func(activity.ParsedEvent) {})
	var progress []string
	pos, err := ProcessOutputLogFileFromOffset(
		context.Background(), path, int64(len("line1\n")), parser, h, nil,
		func(_ int64, line string) { progress = append(progress, line) },
	)
	if err != nil {
		t.Fatalf("ProcessOutputLogFileFromOffset: %v", err)
	}
	if pos != int64(len(content)) {
		t.Fatalf("pos = %d, want %d", pos, len(content))
	}
	if len(progress) != 2 {
		t.Fatalf("progress lines = %d, want 2", len(progress))
	}
}

func TestProcessOutputLogFileFromOffset_ParseErrorContinues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := writeTestFile(path, "bad line\n"); err != nil {
		t.Fatal(err)
	}
	parser := stubParser{err: errors.New("parse fail")}
	buf := &raceSafeLogBuffer{}
	_, err := ProcessOutputLogFileFromOffset(
		context.Background(), path, 0, parser, EventHandlerFunc(func(activity.ParsedEvent) {}),
		buf.logger(),
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if buf.len() == 0 {
		t.Fatal("expected parse error log")
	}
}

func TestProcessOutputLogFileFromOffset_ContextCancel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := writeTestFile(path, "a\nb\nc\n"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ProcessOutputLogFileFromOffset(ctx, path, 0, stubParser{}, EventHandlerFunc(func(activity.ParsedEvent) {}), nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
}

func TestProcessOutputLogFileFromOffset_NegativeOffsetClamped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := writeTestFile(path, "only\n"); err != nil {
		t.Fatal(err)
	}
	parser := stubParser{events: []activity.ParsedEvent{&activity.EncounterEvent{DisplayName: "X"}}}
	var got activity.ParsedEvent
	h := EventHandlerFunc(func(ev activity.ParsedEvent) { got = ev })
	pos, err := ProcessOutputLogFileFromOffset(context.Background(), path, -5, parser, h, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected event")
	}
	if pos != int64(len("only\n")) {
		t.Fatalf("pos = %d", pos)
	}
}

func TestProcessOutputLogFileFromOffset_openError(t *testing.T) {
	_, err := ProcessOutputLogFileFromOffset(
		context.Background(),
		filepath.Join(t.TempDir(), "missing.txt"),
		0, stubParser{}, EventHandlerFunc(func(activity.ParsedEvent) {}), nil, nil,
	)
	if err == nil {
		t.Fatal("expected open error")
	}
}

func TestProcessOutputLogFileFromOffset_offsetBeyondSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "log.txt")
	if err := os.WriteFile(path, []byte("ab"), 0600); err != nil {
		t.Fatal(err)
	}
	pos, err := ProcessOutputLogFileFromOffset(
		context.Background(), path, 100, stubParser{}, EventHandlerFunc(func(activity.ParsedEvent) {}), nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if pos != 2 {
		t.Fatalf("pos = %d", pos)
	}
}

func TestProcessOutputLogFileFromOffset_readError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fifo")
	if err := os.Mkdir(path, 0555); err != nil {
		t.Fatal(err)
	}
	_, err := ProcessOutputLogFileFromOffset(
		context.Background(), path, 0, stubParser{}, EventHandlerFunc(func(activity.ParsedEvent) {}), nil, nil,
	)
	if err == nil {
		t.Fatal("expected read error on directory path")
	}
}

func TestProcessOutputLogFileFromOffset_progressCallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "log.txt")
	if err := os.WriteFile(path, []byte("x\n"), 0600); err != nil {
		t.Fatal(err)
	}
	pos, err := ProcessOutputLogFileFromOffset(
		context.Background(), path, 0, stubParser{}, EventHandlerFunc(func(activity.ParsedEvent) {}), nil,
		func(offset int64, line string) {
			if offset != 2 || line != "x" {
				t.Fatalf("progress offset=%d line=%q", offset, line)
			}
		},
	)
	if err != nil || pos != 2 {
		t.Fatalf("pos=%d err=%v", pos, err)
	}
}
