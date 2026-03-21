package picturewatcher

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const defaultDebounce = 400 * time.Millisecond

// IngestFunc registers an image file path (e.g. into SQLite). Called after debounce.
type IngestFunc func(ctx context.Context, path string) error

// Logger receives non-fatal diagnostics (optional).
type Logger interface {
	Printf(format string, v ...any)
}

type nopLogger struct{}

func (nopLogger) Printf(string, ...any) {}

// Start watches root and its subdirectories for new or updated image files.
// Events are debounced then passed to ingest. The caller should cancel ctx to stop the watcher.
func Start(ctx context.Context, root string, ingest IngestFunc, log Logger) error {
	if ingest == nil {
		return nil
	}
	root = filepath.Clean(root)
	st, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return fmt.Errorf("picturewatcher: not a directory: %s", root)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if log == nil {
		log = nopLogger{}
	}

	w := &run{
		ctx:      ctx,
		fsw:      watcher,
		ingest:   ingest,
		log:      log,
		debounce: defaultDebounce,
	}

	if err := w.watchDirTree(root); err != nil {
		_ = watcher.Close()
		return err
	}

	go w.loop()
	return nil
}

type run struct {
	ctx      context.Context
	fsw      *fsnotify.Watcher
	ingest   IngestFunc
	log      Logger
	debounce time.Duration

	mu      sync.Mutex
	pending map[string]struct{}
	timer   *time.Timer
}

func (w *run) watchDirTree(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if addErr := w.fsw.Add(path); addErr != nil {
			w.log.Printf("picturewatcher: watch %s: %v", path, addErr)
		}
		return nil
	})
}

func (w *run) loop() {
	defer func() { _ = w.fsw.Close() }()
	for {
		select {
		case <-w.ctx.Done():
			w.stopFlushTimer()
			return
		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			if err != nil {
				w.log.Printf("picturewatcher: %v", err)
			}
		case ev, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			w.handleEvent(ev)
		}
	}
}

func (w *run) handleEvent(ev fsnotify.Event) {
	name := filepath.Clean(ev.Name)
	if ev.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename) == 0 {
		return
	}

	st, err := os.Stat(name)
	if err != nil {
		return
	}
	if st.IsDir() {
		_ = w.watchDirTree(name)
		return
	}
	if !isImageFile(name) {
		return
	}
	w.scheduleIngest(name)
}

func isImageFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg":
		return true
	default:
		return false
	}
}

func (w *run) scheduleIngest(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.pending == nil {
		w.pending = make(map[string]struct{})
	}
	w.pending[path] = struct{}{}
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(w.debounce, w.flush)
}

func (w *run) stopFlushTimer() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timer != nil {
		w.timer.Stop()
		w.timer = nil
	}
	w.pending = nil
}

func (w *run) flush() {
	w.mu.Lock()
	paths := make([]string, 0, len(w.pending))
	for p := range w.pending {
		paths = append(paths, p)
	}
	w.pending = make(map[string]struct{})
	w.timer = nil
	w.mu.Unlock()

	for _, p := range paths {
		if w.ctx.Err() != nil {
			return
		}
		if err := w.ingest(w.ctx, p); err != nil {
			w.log.Printf("picturewatcher: ingest %s: %v", p, err)
		}
	}
	}
}
