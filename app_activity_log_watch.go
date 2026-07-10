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
	"vrchat-tweaker/internal/usecase"
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
		a.evictActivityIngestAdapter(oldPath)
	}
	if deps.emitEncounters != nil {
		deps.emitEncounters()
	}
	return nil
}

// replayActivityLogAfterFileSwitch runs Log replay after single-file switch/truncate.
// Activity ingest only — automation stays on live tail (ADR 0005 Decision 12).
func (a *App) replayActivityLogAfterFileSwitch(
	ctx context.Context,
	deps activityLogWatchDeps,
	newPath string,
	ingestAdapter *logwatcher.ActivityIngestAdapter,
) {
	if newPath == "" || deps.parser == nil || ingestAdapter == nil || a.activity == nil {
		return
	}
	a.ingestOneActivityLogBootstrap(
		ctx, deps.watchPath, newPath, deps.parser, deps.logger, deps.emitEncounters,
		nil, false, ingestAdapter,
	)
	if deps.emitEncounters != nil {
		deps.emitEncounters()
	}
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
	lastLine := a.activity.LastObservedLogTime(ctx)
	if lastLine.IsZero() {
		lastLine = time.Now().UTC()
	}

	// Stage 1: per-source finalize for known paths (checkpoint ∪ dir listing).
	// Stage 2: global FinalizeAllOpenActivity for log_source_path IS NULL legacy rows (ADR 0005 Decision 13).
	type pathClose struct {
		fc *usecase.ActivityLogFileCheckpoint
	}
	paths := make(map[string]pathClose)

	cp, err := a.activity.GetActivityLogCheckpoint(ctx)
	if err == nil && cp != nil {
		cp.NormalizeFiles()
		for path, fc := range cp.Files {
			fcCopy := fc
			paths[absLogPath(path)] = pathClose{fc: &fcCopy}
		}
	}

	if watchPath != "" {
		if info, statErr := os.Stat(watchPath); statErr == nil && info.IsDir() {
			if files, listErr := logwatcher.ListOutputLogFiles(watchPath); listErr == nil {
				for _, path := range files {
					abs := absLogPath(path)
					if _, ok := paths[abs]; ok {
						continue
					}
					paths[abs] = pathClose{}
				}
			}
		}
	}

	for path, pc := range paths {
		closeAt := closeTimeForLogFile(path, lastLine, pc.fc)
		if closeAt.IsZero() {
			continue
		}
		_ = a.activity.FinalizeOpenActivityForLogSource(ctx, path, closeAt)
	}

	_ = a.activity.FinalizeAllOpenActivity(ctx, lastLine)

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, activityEncountersChangedEvent, struct{}{})
	}
}

func absLogPath(p string) string {
	abs, err := filepath.Abs(filepath.Clean(p))
	if err != nil {
		return p
	}
	return abs
}

func checkpointVRTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func closeTimeForLogFile(path string, fallback time.Time, fc *usecase.ActivityLogFileCheckpoint) time.Time {
	if t, err := logwatcher.LastVRChatLineTimeInFile(path); err == nil && !t.IsZero() {
		return t
	}
	if fc != nil && fc.VRChatLineTime != "" {
		if t, err := time.Parse(time.RFC3339, fc.VRChatLineTime); err == nil && !t.IsZero() {
			return t
		}
	}
	return fallback
}
