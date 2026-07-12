package ytdlpmaintain

import (
	"context"
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
			return
		}
		if watching && watchPath == dir {
			return
		}
		closeWatcher()
		w, err := fsnotify.NewWatcher()
		if err != nil {
			return
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			_ = w.Close()
			return
		}
		if err := w.Add(dir); err != nil {
			_ = w.Close()
			return
		}
		watcher = w
		watching = true
		watchPath = dir
	}

	reapply := func() {
		_ = reapplier.ReapplyIfNeeded(ctx)
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
		case _, ok := <-errCh:
			if !ok {
				closeWatcher()
				continue
			}
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
