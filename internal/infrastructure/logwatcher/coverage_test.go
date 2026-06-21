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
	"vrchat-tweaker/internal/domain/automation"
	"vrchat-tweaker/internal/domain/event"
)

type stubParser struct {
	events []activity.ParsedEvent
	err    error
}

func (p stubParser) ParseLine(string, time.Time) ([]activity.ParsedEvent, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.events, nil
}

func TestMultiHandler_DelegatesToAll(t *testing.T) {
	var count atomic.Int32
	h1 := EventHandlerFunc(func(activity.ParsedEvent) { count.Add(1) })
	h2 := EventHandlerFunc(func(activity.ParsedEvent) { count.Add(1) })
	mh := NewMultiHandler(h1, h2)
	ev := &activity.EncounterEvent{DisplayName: "A"}
	mh.Handle(ev)
	if count.Load() != 2 {
		t.Fatalf("calls = %d, want 2", count.Load())
	}
}

func TestEventPublishingHandler_FriendJoined(t *testing.T) {
	bus := event.NewChannelEventBus()
	ctx := context.Background()
	var published atomic.Int32
	bus.Subscribe(automation.TriggerFriendJoined, func(_ context.Context, ev *event.Event) error {
		published.Add(1)
		payload, _ := ev.Payload.(map[string]interface{})
		if payload["vrc_user_id"] != "usr_join01" {
			t.Errorf("payload = %v", ev.Payload)
		}
		return nil
	})

	h := NewEventPublishingHandler(bus, ctx, nil)
	h.Handle(&activity.EncounterEvent{
		Action:      activity.EncounterActionJoin,
		VRCUserID:   "usr_join01",
		DisplayName: "Friend",
	})
	if published.Load() != 1 {
		t.Fatalf("published = %d", published.Load())
	}

	h.Handle(&activity.EncounterEvent{Action: activity.EncounterActionLeave, VRCUserID: "usr_join01"})
	h.Handle(nil)
	if published.Load() != 1 {
		t.Fatalf("leave/nil should not publish, got %d", published.Load())
	}
}

type failingEventBus struct{}

func (failingEventBus) Publish(context.Context, string, *event.Event) error {
	return errors.New("boom")
}

func (failingEventBus) Subscribe(string, func(context.Context, *event.Event) error) func() {
	return func() {}
}

func TestEventPublishingHandler_PublishErrorLogged(t *testing.T) {
	var logs []string
	h := NewEventPublishingHandler(failingEventBus{}, context.Background(), testLogger{fn: func(format string, args ...interface{}) {
		logs = append(logs, format)
	}})
	h.Handle(&activity.EncounterEvent{
		Action:    activity.EncounterActionJoin,
		VRCUserID: "usr_err",
	})
	if len(logs) == 0 {
		t.Fatal("expected log on publish error")
	}
}

type testLogger struct {
	fn func(format string, args ...interface{})
}

func (l testLogger) Printf(format string, args ...interface{}) { l.fn(format, args...) }

type raceSafeLogBuffer struct {
	mu   sync.Mutex
	logs []string
}

func (b *raceSafeLogBuffer) logger() Logger {
	return testLogger{fn: func(format string, args ...interface{}) {
		b.mu.Lock()
		b.logs = append(b.logs, format)
		b.mu.Unlock()
	}}
}

func (b *raceSafeLogBuffer) len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.logs)
}

func TestProcessOutputLogFileFromOffset_ResumeAndProgress(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	content := "line1\nline2\nline3\n"
	if err := writeTestFile(path, content); err != nil {
		t.Fatal(err)
	}

	parser := activity.NewLogParser()
	var mu sync.Mutex
	var lines []string
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
	mu.Lock()
	_ = lines
	mu.Unlock()
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

func TestOutputLogWatcher_StartAlreadyRunning(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := writeTestFile(path, ""); err != nil {
		t.Fatal(err)
	}
	w := NewOutputLogWatcher(path, activity.NewLogParser(), EventHandlerFunc(func(activity.ParsedEvent) {}), testLogger{fn: func(string, ...interface{}) {}})
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

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0600)
}

func appendToTestFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.Write([]byte(content))
	return err
}
