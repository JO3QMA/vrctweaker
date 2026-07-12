package ytdlpmaintain

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

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
		watcher   *fsnotify.Watcher
		watching  bool
		watchPath string
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

	ensureWatch := func() {
		dir, err := toolsDir.ToolsDir()
		if err != nil {
			log.Printf("ytdlp maintain: ToolsDir error: %v", err)
			closeWatcher()
			return
		}
		if watching && watchPath == dir {
			return
		}
		closeWatcher()
		w, err := fsnotify.NewWatcher()
		if err != nil {
			log.Printf("ytdlp maintain: fsnotify.NewWatcher error: %v", err)
			return
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Printf("ytdlp maintain: MkdirAll(%s) error: %v", dir, err)
			_ = w.Close()
			return
		}
		if err := w.Add(dir); err != nil {
			log.Printf("ytdlp maintain: watcher Add(%s) error: %v", dir, err)
			_ = w.Close()
			return
		}
		watcher = w
		watching = true
		watchPath = dir
	}

	reapply := func() {
		go func() {
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
			return ctx.Err()
		case ev, ok := <-watchCh:
			if !ok {
				closeWatcher()
				continue
			}
			if filepath.Base(ev.Name) != "yt-dlp.exe" {
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
			ensureWatch()
			reapply()
		}
	}
}
