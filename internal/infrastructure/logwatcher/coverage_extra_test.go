package logwatcher

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/event"
	"vrchat-tweaker/internal/usecase"
)

type errWorldInfoRepo struct {
	spyWorldInfoRepo
	visitErr   error
	displayErr error
}

func (e *errWorldInfoRepo) UpsertVisit(context.Context, string, time.Time) error {
	return e.visitErr
}

func (e *errWorldInfoRepo) UpsertDisplayName(context.Context, string, string, time.Time) error {
	return e.displayErr
}

func TestActivityIngestAdapter_LogsUpsertErrors(t *testing.T) {
	ctx := context.Background()
	base := time.Now()

	worldRepo := &errWorldInfoRepo{
		visitErr:   errors.New("visit fail"),
		displayErr: errors.New("display fail"),
	}
	uc := usecase.NewActivityUseCase(stubPlaySessionRepo{}, stubEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, worldRepo)
	buf := &raceSafeLogBuffer{}
	h := NewActivityIngestAdapter(uc, ctx, buf.logger(), nil)

	h.Handle(nil)
	h.Handle(&activity.DestinationSetEvent{WorldID: testWorldID, OccurredAt: base})
	h.Handle(&activity.RoomNameEvent{RoomName: "Room", OccurredAt: base})

	if buf.len() < 2 {
		t.Fatalf("logs = %d, want visit and display errors", buf.len())
	}
}

func TestActivityIngestAdapter_SessionStartEmptyInstanceIgnored(t *testing.T) {
	ctx := context.Background()
	h := NewActivityIngestAdapter(
		usecase.NewActivityUseCase(stubPlaySessionRepo{}, stubEncounterRepo{}, &fakeAppSettingsRepo{m: map[string]string{}}, nil, nil),
		ctx, nil, nil,
	)
	h.Handle(&activity.SessionEvent{Type: activity.SessionEventStart, InstanceID: "", OccurredAt: time.Now()})
}

func TestListOutputLogFiles_EmptyDirectory(t *testing.T) {
	_, err := ListOutputLogFiles(t.TempDir())
	if !errors.Is(err, ErrNoOutputLogFiles) {
		t.Fatalf("err = %v", err)
	}
}

func TestOutputLogPathValid_missingPath(t *testing.T) {
	if OutputLogPathValid(filepath.Join(t.TempDir(), "nope")) {
		t.Fatal("missing path should be invalid")
	}
}

func TestResolveLatestOutputLogFile_skipsDirectoryNamedLikeLog(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "output_log_2026-01-01_00-00-00.txt")
	if err := os.WriteFile(target, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "output_log_dir.txt"), 0755); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveLatestOutputLogFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != target {
		t.Fatalf("got %q want %q", got, target)
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

func TestOutputLogWatcher_resolveActivePath_fixedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output_log.txt")
	w := NewOutputLogWatcher(path, stubParser{}, EventHandlerFunc(func(activity.ParsedEvent) {}), nil)
	got, err := w.resolveActivePath()
	if err != nil || got != path {
		t.Fatalf("resolveActivePath() = %q, %v", got, err)
	}
}

func TestEventHandlerFunc_Handle(t *testing.T) {
	var called bool
	EventHandlerFunc(func(activity.ParsedEvent) { called = true }).Handle(&activity.EncounterEvent{})
	if !called {
		t.Fatal("expected handle call")
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

func TestLogWriterLogger_Printf(t *testing.T) {
	logWriterLogger{log.New(io.Discard, "", 0)}.Printf("ignored %d", 1)
}

func TestActivityIngestAdapter_EncounterRecordErrorLogged(t *testing.T) {
	ctx := context.Background()
	buf := &raceSafeLogBuffer{}
	encRepo := &errEncounterRepo{}
	uc := usecase.NewActivityUseCase(stubPlaySessionRepo{}, encRepo, &fakeAppSettingsRepo{m: map[string]string{}}, nil, nil)
	h := NewActivityIngestAdapter(uc, ctx, buf.logger(), nil)
	h.Handle(&activity.SessionEvent{Type: activity.SessionEventStart, InstanceID: testFullInstance, OccurredAt: time.Now()})
	h.Handle(&activity.EncounterEvent{
		VRCUserID: "usr_x", DisplayName: "X", Action: activity.EncounterActionJoin, EncounteredAt: time.Now(),
	})
	if buf.len() == 0 {
		t.Fatal("expected encounter error log")
	}
}

type errEncounterRepo struct{ stubEncounterRepo }

func (errEncounterRepo) Save(context.Context, *activity.UserEncounter) error {
	return errors.New("save fail")
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

func TestActivityIngestAdapter_SessionErrorsLogged(t *testing.T) {
	ctx := context.Background()
	buf := &raceSafeLogBuffer{}
	playRepo := &errPlaySessionRepo{}
	encRepo := &errEncounterRepo{}
	uc := usecase.NewActivityUseCase(playRepo, encRepo, &fakeAppSettingsRepo{m: map[string]string{}}, nil, nil)
	h := NewActivityIngestAdapter(uc, ctx, buf.logger(), nil)
	h.Handle(&activity.SessionEvent{Type: activity.SessionEventStart, InstanceID: testFullInstance, OccurredAt: time.Now()})
	h.Handle(&activity.SessionEvent{Type: activity.SessionEventEnd, OccurredAt: time.Now()})
	if buf.len() < 2 {
		t.Fatalf("expected session error logs, got %d", buf.len())
	}
}

type errPlaySessionRepo struct{ stubPlaySessionRepo }

func (errPlaySessionRepo) Save(context.Context, *activity.PlaySession) error {
	return errors.New("save fail")
}

func (errPlaySessionRepo) FindLatestWithoutEndTime(context.Context) (*activity.PlaySession, error) {
	return &activity.PlaySession{ID: "open", StartTime: time.Now().Add(-time.Minute)}, nil
}

func (errEncounterRepo) CloseOpenEncountersAt(context.Context, time.Time) (int64, error) {
	return 0, errors.New("close encounters fail")
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

func TestNewEventPublishingHandler_defaultLogger(t *testing.T) {
	h := NewEventPublishingHandler(event.NewChannelEventBus(), context.Background(), nil)
	if h.logger == nil {
		t.Fatal("expected default logger")
	}
}

func TestListOutputLogFiles_onlyAuxiliaryFiles(t *testing.T) {
	dir := t.TempDir()
	aux := filepath.Join(dir, "output_log_2026-03-18_12-52-26.parsed_lines.txt")
	if err := os.WriteFile(aux, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := ListOutputLogFiles(dir)
	if !errors.Is(err, ErrNoOutputLogFiles) {
		t.Fatalf("err = %v", err)
	}
}

func TestResolveLatestOutputLogFile_onlyBrokenSymlink(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "output_log_broken.txt")
	if err := os.Symlink(filepath.Join(dir, "missing-target.txt"), link); err != nil {
		t.Skip("symlink unsupported")
	}
	_, err := ResolveLatestOutputLogFile(dir)
	if !errors.Is(err, ErrNoOutputLogFiles) {
		t.Fatalf("err = %v", err)
	}
}

func TestActivityIngestAdapter_DefaultLoggerOnUpsertError(t *testing.T) {
	ctx := context.Background()
	worldRepo := &errWorldInfoRepo{visitErr: errors.New("visit fail")}
	uc := usecase.NewActivityUseCase(stubPlaySessionRepo{}, stubEncounterRepo{}, &fakeAppSettingsRepo{m: map[string]string{}}, nil, worldRepo)
	h := NewActivityIngestAdapter(uc, ctx, nil, nil)
	h.Handle(&activity.DestinationSetEvent{WorldID: testWorldID, OccurredAt: time.Now()})
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
