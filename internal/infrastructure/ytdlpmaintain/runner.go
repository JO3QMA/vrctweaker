package ytdlpmaintain

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

const toolsDirLogInterval = 30 * time.Second

// SettingGetter reads whether Tools replace maintain is desired.
type SettingGetter interface {
	YTDLPToolsReplaceMaintain(ctx context.Context) (bool, error)
}

// ProcessChecker reports whether VRChat.exe is running.
type ProcessChecker interface {
	VRChatRunning() (bool, error)
}

// Reapplier re-links Tools when maintain is on and the link drifted.
type Reapplier interface {
	ReapplyIfNeeded(ctx context.Context) error
}

// ToolsDirProvider returns the Tools directory to watch (parent of yt-dlp.exe).
type ToolsDirProvider interface {
	ToolsDir() (string, error)
}

// Run polls until ctx is cancelled. While maintain is desired and VRChat is running,
// it watches Tools for rollbacks and periodically re-applies the Official cache symlink.
func Run(
	ctx context.Context,
	pollInterval time.Duration,
	settings SettingGetter,
	proc ProcessChecker,
	reapplier Reapplier,
	toolsDir ToolsDirProvider,
) error {
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	var (
		watcher      *fsnotify.Watcher
		watching     bool
		watchPath    string
		lastWatchLog time.Time
		reapplyWG    sync.WaitGroup
		reapplyBusy  atomic.Bool
	)
	closeWatcher := func() {
		if watcher != nil {
			_ = watcher.Close()
			watcher = nil
		}
		watching = false
		watchPath = ""
	}
	defer closeWatcher()

	logWatchErr := func(format string, args ...any) {
		if time.Since(lastWatchLog) < toolsDirLogInterval {
			return
		}
		log.Printf(format, args...)
		lastWatchLog = time.Now()
	}

	ensureWatch := func() bool {
		dir, err := toolsDir.ToolsDir()
		if err != nil {
			logWatchErr("ytdlp maintain: ToolsDir error: %v", err)
			return false
		}
		if watching && watchPath == dir {
			return true
		}
		closeWatcher()
		w, err := fsnotify.NewWatcher()
		if err != nil {
			logWatchErr("ytdlp maintain: fsnotify.NewWatcher error: %v", err)
			return false
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			logWatchErr("ytdlp maintain: MkdirAll(%s) error: %v", dir, err)
			_ = w.Close()
			return false
		}
		if err := w.Add(dir); err != nil {
			logWatchErr("ytdlp maintain: watcher Add(%s) error: %v", dir, err)
			_ = w.Close()
			return false
		}
		watcher = w
		watching = true
		watchPath = dir
		return true
	}

	reapply := func() {
		reapplyWG.Add(1)
		go func() {
			defer reapplyWG.Done()
			if !reapplyBusy.CompareAndSwap(false, true) {
				return
			}
			defer reapplyBusy.Store(false)
			if ctx.Err() != nil {
				return
			}
			if err := reapplier.ReapplyIfNeeded(ctx); err != nil {
				log.Printf("ytdlp maintain: ReapplyIfNeeded error: %v", err)
			}
		}()
	}

	for {
		var watchCh <-chan fsnotify.Event
		var errCh <-chan error
		if watcher != nil {
			watchCh = watcher.Events
			errCh = watcher.Errors
		}

		select {
		case <-ctx.Done():
			closeWatcher()
			reapplyWG.Wait()
			return ctx.Err()
		case ev, ok := <-watchCh:
			if !ok {
				closeWatcher()
				continue
			}
			if !strings.EqualFold(filepath.Base(ev.Name), "yt-dlp.exe") {
				continue
			}
			if ev.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) == 0 {
				continue
			}
			reapply()
		case evErr, ok := <-errCh:
			if !ok {
				closeWatcher()
				continue
			}
			log.Printf("ytdlp maintain: watcher error: %v", evErr)
			closeWatcher()
		case <-ticker.C:
			enabled, err := settings.YTDLPToolsReplaceMaintain(ctx)
			if err != nil || !enabled {
				closeWatcher()
				continue
			}
			running, err := proc.VRChatRunning()
			if err != nil || !running {
				closeWatcher()
				continue
			}
			if ensureWatch() {
				reapply()
			}
		}
	}
}
