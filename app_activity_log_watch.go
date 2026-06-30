package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/infrastructure/logwatcher"
	"vrchat-tweaker/internal/infrastructure/sleepsuppress"
)

type activityLogWatchDeps struct {
	watchPath      string
	parser         *activity.LogParser
	logger         logwatcher.Logger
	emitEncounters func()
}

func (a *App) finalizeOpenActivityForLogSource(ctx context.Context, logPath string) {
	if a.activity == nil || logPath == "" {
		return
	}
	lastTime, err := logwatcher.LastVRChatLineTimeInFile(logPath)
	if err != nil {
		runtime.LogWarning(ctx, "activity finalize log time: "+err.Error())
	}
	if lastTime.IsZero() {
		absPath, absErr := filepath.Abs(filepath.Clean(logPath))
		if absErr != nil {
			absPath = logPath
		}
		if cp, cpErr := a.activity.GetActivityLogCheckpoint(ctx); cpErr == nil && cp != nil {
			if fc, ok := cp.FileCheckpoint(absPath); ok && fc.VRChatLineTime != "" {
				if t, parseErr := time.Parse(time.RFC3339, fc.VRChatLineTime); parseErr == nil {
					lastTime = t
				}
			}
		}
	}
	if lastTime.IsZero() {
		return
	}
	_ = a.activity.FinalizeOpenActivityForLogSource(ctx, absLogPath(logPath), lastTime)
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, activityEncountersChangedEvent, struct{}{})
	}
}

func (a *App) handleActivityLogRotationHandoff(ctx context.Context, deps activityLogWatchDeps, oldPath string) error {
	if oldPath != "" {
		a.finalizeOpenActivityForLogSource(ctx, oldPath)
	}
	if deps.emitEncounters != nil {
		deps.emitEncounters()
	}
	return nil
}

func (a *App) startVRChatActivityMonitor(ctx context.Context, watchPath string) {
	a.activityWatchMu.Lock()
	if a.activityWatchCancel != nil {
		a.activityWatchCancel()
		a.activityWatchWG.Wait()
	}
	runCtx, cancel := context.WithCancel(ctx)
	a.activityWatchCancel = cancel
	a.activityWatchMu.Unlock()

	a.activityWatchWG.Add(1)
	go func() {
		defer a.activityWatchWG.Done()
		checker := sleepsuppress.NewVRChatProcessChecker()
		_ = logwatcher.MonitorVRChatRunning(runCtx, 4*time.Second, checker, func() {
			a.finalizeAllLogSourcesOnVRChatExit(ctx, watchPath)
		})
	}()
}

func (a *App) stopVRChatActivityMonitor() {
	var cancel context.CancelFunc
	a.activityWatchMu.Lock()
	cancel = a.activityWatchCancel
	a.activityWatchCancel = nil
	a.activityWatchMu.Unlock()
	if cancel != nil {
		cancel()
		a.activityWatchWG.Wait()
	}
}

func (a *App) finalizeAllLogSourcesOnVRChatExit(ctx context.Context, watchPath string) {
	if a.activity == nil {
		return
	}
	lastLine := a.latestCheckpointLineTime(ctx)
	if lastLine.IsZero() {
		lastLine = time.Now().UTC()
	}
	_ = a.activity.FinalizeAllOpenActivity(ctx, lastLine)

	cp, err := a.activity.GetActivityLogCheckpoint(ctx)
	if err == nil && cp != nil {
		cp.NormalizeFiles()
		for path := range cp.Files {
			if t, tErr := logwatcher.LastVRChatLineTimeInFile(path); tErr == nil && t.After(lastLine) {
				lastLine = t
			}
			_ = a.activity.FinalizeOpenActivityForLogSource(ctx, absLogPath(path), lastLine)
		}
	}

	if watchPath != "" {
		if info, statErr := os.Stat(watchPath); statErr == nil && info.IsDir() {
			if files, listErr := logwatcher.ListOutputLogFiles(watchPath); listErr == nil {
				for _, path := range files {
					_ = a.activity.FinalizeOpenActivityForLogSource(ctx, absLogPath(path), lastLine)
				}
			}
		}
	}

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, activityEncountersChangedEvent, struct{}{})
	}
}

func (a *App) latestCheckpointLineTime(ctx context.Context) time.Time {
	cp, err := a.activity.GetActivityLogCheckpoint(ctx)
	if err != nil || cp == nil {
		return time.Time{}
	}
	cp.NormalizeFiles()
	var max time.Time
	for _, fc := range cp.Files {
		if fc.VRChatLineTime == "" {
			continue
		}
		t, parseErr := time.Parse(time.RFC3339, fc.VRChatLineTime)
		if parseErr != nil || t.IsZero() {
			continue
		}
		if t.After(max) {
			max = t
		}
	}
	return max
}

func absLogPath(p string) string {
	abs, err := filepath.Abs(filepath.Clean(p))
	if err != nil {
		return p
	}
	return abs
}

func formatActivityCheckpointLineTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// activityTailCheckpoint persists incremental ingest progress for one log file.
func (a *App) activityTailCheckpoint(ctx context.Context, deps activityLogWatchDeps, path string, offset int64, lineTime time.Time) {
	if a.activity == nil {
		return
	}
	_ = a.activity.SetActivityLogFileCheckpoint(ctx, deps.watchPath, absLogPath(path), offset, formatActivityCheckpointLineTime(lineTime))
}
