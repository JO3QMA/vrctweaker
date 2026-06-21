package picturewatcher

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestIsImageFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want bool
	}{
		{"/a.png", true},
		{"/a.PNG", true},
		{"/a.jpg", true},
		{"/a.jpeg", true},
		{"/a.gif", false},
		{"/a.txt", false},
	}
	for _, tt := range tests {
		if got := isImageFile(tt.path); got != tt.want {
			t.Errorf("isImageFile(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestStart_nilIngest(t *testing.T) {
	t.Parallel()
	err := Start(context.Background(), t.TempDir(), nil, nil)
	if err != nil {
		t.Fatalf("Start nil ingest: %v", err)
	}
}

func TestStart_missingRoot(t *testing.T) {
	t.Parallel()
	err := Start(context.Background(), filepath.Join(t.TempDir(), "missing"), func(context.Context, string) error { return nil }, nil)
	if err == nil {
		t.Fatal("expected error for missing root")
	}
}

func TestStart_notDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := Start(context.Background(), p, func(context.Context, string) error { return nil }, nil)
	if err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("Start file root: %v", err)
	}
}

type testLogger struct {
	mu   sync.Mutex
	msgs []string
}

func (l *testLogger) Printf(format string, v ...any) {
	l.mu.Lock()
	l.msgs = append(l.msgs, format)
	l.mu.Unlock()
}

func TestStart_ingestsNewImage(t *testing.T) {
	if testing.Short() {
		t.Skip("filesystem watcher integration")
	}
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan string, 1)
	ingest := func(_ context.Context, p string) error {
		done <- p
		return nil
	}
	if err := Start(ctx, root, ingest, nil); err != nil {
		t.Fatalf("Start: %v", err)
	}

	path := filepath.Join(root, "photo.png")
	if err := os.WriteFile(path, []byte("fake png"), 0o600); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-done:
		if got != path {
			t.Fatalf("ingested %q, want %q", got, path)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for ingest")
	}
}

func TestStart_ingestsNestedImage(t *testing.T) {
	if testing.Short() {
		t.Skip("filesystem watcher integration")
	}
	root := t.TempDir()
	sub := filepath.Join(root, "nested")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan string, 1)
	ingest := func(_ context.Context, p string) error {
		done <- p
		return nil
	}
	if err := Start(ctx, root, ingest, nil); err != nil {
		t.Fatalf("Start: %v", err)
	}

	path := filepath.Join(sub, "inner.jpg")
	if err := os.WriteFile(path, []byte("fake jpg"), 0o600); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-done:
		if got != path {
			t.Fatalf("ingested %q, want %q", got, path)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for nested ingest")
	}
}

func TestHandleEvent_ignoresNonImageCreateOp(t *testing.T) {
	ctx := context.Background()
	var calls int
	r := &run{
		ctx:      ctx,
		fsw:      mustWatcher(t),
		ingest:   func(context.Context, string) error { calls++; return nil },
		log:      nopLogger{},
		debounce: time.Hour,
	}
	t.Cleanup(func() { _ = r.fsw.Close() })

	r.handleEvent(fsnotify.Event{Name: filepath.Join(t.TempDir(), "readme.txt"), Op: fsnotify.Create})
	time.Sleep(20 * time.Millisecond)
	if calls != 0 {
		t.Fatalf("ingest calls = %d, want 0", calls)
	}
}

func TestHandleEvent_ignoresChmodOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "photo.png")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	var calls int
	r := &run{
		ctx:      context.Background(),
		fsw:      mustWatcher(t),
		ingest:   func(context.Context, string) error { calls++; return nil },
		log:      nopLogger{},
		debounce: time.Hour,
	}
	t.Cleanup(func() { _ = r.fsw.Close() })
	r.handleEvent(fsnotify.Event{Name: path, Op: fsnotify.Chmod})
	time.Sleep(20 * time.Millisecond)
	if calls != 0 {
		t.Fatalf("ingest calls = %d, want 0", calls)
	}
}

func TestHandleEvent_ingestsOnRename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "renamed.png")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	done := make(chan string, 1)
	r := &run{
		ctx:      context.Background(),
		fsw:      mustWatcher(t),
		ingest:   func(_ context.Context, p string) error { done <- p; return nil },
		log:      nopLogger{},
		debounce: 20 * time.Millisecond,
	}
	t.Cleanup(func() { _ = r.fsw.Close() })
	r.handleEvent(fsnotify.Event{Name: path, Op: fsnotify.Rename})
	select {
	case got := <-done:
		if got != path {
			t.Fatalf("got %q", got)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func TestHandleEvent_ignoresNonImage(t *testing.T) {
	ctx := context.Background()
	var calls int
	r := &run{
		ctx:      ctx,
		fsw:      mustWatcher(t),
		ingest:   func(context.Context, string) error { calls++; return nil },
		log:      nopLogger{},
		debounce: time.Hour,
	}
	t.Cleanup(func() { _ = r.fsw.Close() })

	r.handleEvent(fsnotify.Event{Name: filepath.Join(t.TempDir(), "readme.txt"), Op: fsnotify.Create})
	time.Sleep(20 * time.Millisecond)
	if calls != 0 {
		t.Fatalf("ingest calls = %d, want 0", calls)
	}
}

func TestHandleEvent_watchesNewDirectory(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	r := &run{
		ctx:      ctx,
		fsw:      mustWatcher(t),
		ingest:   func(context.Context, string) error { return nil },
		log:      nopLogger{},
		debounce: time.Hour,
	}
	t.Cleanup(func() { _ = r.fsw.Close() })

	if err := r.watchDirTree(root); err != nil {
		t.Fatal(err)
	}
	r.handleEvent(fsnotify.Event{Name: child, Op: fsnotify.Create})
}

func TestFlush_logsIngestError(t *testing.T) {
	ctx := context.Background()
	log := &testLogger{}
	r := &run{
		ctx:    ctx,
		ingest: func(context.Context, string) error { return errors.New("ingest failed") },
		log:    log,
	}
	r.pending = map[string]struct{}{"/a.png": {}}
	r.flush()

	log.mu.Lock()
	n := len(log.msgs)
	log.mu.Unlock()
	if n == 0 {
		t.Fatal("expected log message for ingest error")
	}
}

func TestFlush_respectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var calls int
	r := &run{
		ctx:    ctx,
		ingest: func(context.Context, string) error { calls++; return nil },
		log:    nopLogger{},
	}
	r.pending = map[string]struct{}{"/a.png": {}, "/b.png": {}}
	r.flush()
	if calls != 0 {
		t.Fatalf("ingest calls = %d, want 0 when ctx cancelled", calls)
	}
}

func mustWatcher(t *testing.T) *fsnotify.Watcher {
	t.Helper()
	w, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	return w
}

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

func TestLoop_exitsWhenWatcherClosed(t *testing.T) {
	w := mustWatcher(t)
	r := &run{
		ctx:      context.Background(),
		fsw:      w,
		ingest:   func(context.Context, string) error { return nil },
		log:      nopLogger{},
		debounce: time.Second,
	}
	done := make(chan struct{})
	go func() {
		r.loop()
		close(done)
	}()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("loop did not exit after watcher close")
	}
}

func TestWatchDirTree_skipsInaccessibleEntries(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(sub, 0o000); err != nil {
		t.Skip("chmod not supported")
	}
	t.Cleanup(func() { _ = os.Chmod(sub, 0o755) })

	log := &testLogger{}
	r := &run{
		ctx: context.Background(),
		fsw: mustWatcher(t),
		log: log,
	}
	t.Cleanup(func() { _ = r.fsw.Close() })
	if err := r.watchDirTree(root); err != nil {
		t.Fatalf("watchDirTree: %v", err)
	}
}

func TestNopLogger_Printf(t *testing.T) {
	t.Parallel()
	var l nopLogger
	l.Printf("ignored %d", 1)
}

func TestStart_cancelsCleanly(t *testing.T) {
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	if err := Start(ctx, root, func(context.Context, string) error { return nil }, nil); err != nil {
		t.Fatalf("Start: %v", err)
	}
	cancel()
	time.Sleep(100 * time.Millisecond)
}

func TestHandleEvent_ingestsOnWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "edit.png")
	if err := os.WriteFile(path, []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	done := make(chan string, 1)
	r := &run{
		ctx:      ctx,
		fsw:      mustWatcher(t),
		ingest:   func(_ context.Context, p string) error { done <- p; return nil },
		log:      nopLogger{},
		debounce: 20 * time.Millisecond,
	}
	t.Cleanup(func() { _ = r.fsw.Close() })

	r.handleEvent(fsnotify.Event{Name: path, Op: fsnotify.Write})
	select {
	case got := <-done:
		if got != path {
			t.Fatalf("got %q", got)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout")
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
