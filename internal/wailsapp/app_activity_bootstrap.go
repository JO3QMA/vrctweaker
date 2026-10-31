package wailsapp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/infrastructure/logwatcher"
	"vrchat-tweaker/internal/platform/paths"
	"vrchat-tweaker/internal/usecase"
)

func (a *App) resolveEffectiveOutputLogWatchPath(ctx context.Context) (string, error) {
	dir, cleared, err := a.settings.EnsureOutputLogWatchDir(ctx)
	if err != nil {
		return "", err
	}
	if cleared {
		runtime.LogWarning(ctx, "cleared stale output_log_path setting; using default log folder")
	}
	if dir != "" {
		// EnsureOutputLogWatchDir already returns an existing absolute directory.
		return dir, nil
	}
	defaultDir := paths.DefaultVRChatOutputLogDir()
	if defaultDir == "" {
		return "", os.ErrNotExist
	}
	absDir, err := filepath.Abs(defaultDir)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(absDir); err != nil {
		return "", err
	}
	return absDir, nil
}

// ingestActivityLogsBootstrap replays existing output_log files under absWatch.
// Callers must pass a directory that has already been resolved and validated
// (resolveEffectiveOutputLogWatchPath).
func (a *App) ingestActivityLogsBootstrap(ctx context.Context, absWatch string, parser *activity.LogParser, logger logwatcher.Logger, emitEncounters, emitVideoPlayback func()) {
	cp, _ := a.activity.GetActivityLogCheckpoint(ctx)

	files, listErr := logwatcher.ListOutputLogFiles(absWatch)
	if listErr != nil {
		return
	}
	live := logwatcher.BootstrapLiveLogFiles(files, 5*time.Second)
	for _, fp := range files {
		finalize := live == nil || !live[fp]
		a.ingestOneActivityLogBootstrap(ctx, absWatch, fp, parser, logger, emitEncounters, emitVideoPlayback, cp, finalize, nil)
	}
	a.retryScreenshotEnrichment()
}

func (a *App) ingestOneActivityLogBootstrap(
	ctx context.Context,
	absWatch, filePath string,
	parser *activity.LogParser,
	logger logwatcher.Logger,
	emitEncounters, emitVideoPlayback func(),
	cp *usecase.ActivityLogCheckpoint,
	finalizeAtEnd bool,
	ingestAdapter *logwatcher.ActivityIngestAdapter,
) {
	absFile := logwatcher.AbsLogPath(filePath)
	if ingestAdapter == nil {
		ingestAdapter = a.activityIngestAdapterForPath(ctx, logger, emitEncounters, emitVideoPlayback, filePath)
	}
	off := int64(0)
	if cp != nil && logwatcher.MatchAbsPaths(cp.WatchPath, absWatch) {
		if fc, ok := cp.FileCheckpoint(absFile); ok {
			off = fc.ByteOffset
			if st, statErr := os.Stat(filePath); statErr == nil && st != nil && st.Size() > 0 && off >= st.Size() {
				if finalizeAtEnd {
					lastVRLineTime := time.Time{}
					if fc.VRChatLineTime != "" {
						if t, parseErr := time.Parse(time.RFC3339, fc.VRChatLineTime); parseErr == nil {
							lastVRLineTime = t
						}
					}
					if lastVRLineTime.IsZero() {
						if t, tErr := logwatcher.LastVRChatLineTimeInFile(filePath); tErr == nil {
							lastVRLineTime = t
						}
					}
					if !lastVRLineTime.IsZero() {
						_ = a.activity.FinalizeOpenActivityForLogSource(ctx, ingestAdapter.LogSourcePath(), lastVRLineTime)
					}
				}
				return
			}
		}
	}
	pathCopy := filePath
	ingestAdapter.SetSuppressEncounterNotify(true)
	defer ingestAdapter.SetSuppressEncounterNotify(false)
	checkpointLines := 0
	var lastVRLineTime time.Time
	if off == 0 {
		ingestAdapter.ResetSessionContextForNewLogFile()
	} else if warmErr := logwatcher.WarmSessionCorrelatorFromLogFile(ctx, pathCopy, off, parser, ingestAdapter, logger); warmErr != nil {
		runtime.LogWarning(ctx, "activity log correlator warm: "+warmErr.Error())
	}
	_, procErr := logwatcher.ProcessOutputLogFileFromOffset(ctx, pathCopy, off, parser, ingestAdapter, logger, func(pos int64, line string) {
		ts := activity.ParseVRChatTimestamp(line, time.Time{})
		if !ts.IsZero() {
			lastVRLineTime = ts
		}
		checkpointLines++
		if checkpointLines != 1 && checkpointLines%32 != 0 {
			return
		}
		_ = a.activity.SetActivityLogFileCheckpoint(ctx, absWatch, logwatcher.AbsLogPath(pathCopy), pos, logwatcher.CheckpointVRTime(ts))
	})
	if procErr != nil {
		if errors.Is(procErr, context.Canceled) {
			return
		}
		runtime.LogWarning(ctx, "activity log ingest: "+procErr.Error())
		return
	}
	if finalizeAtEnd && !lastVRLineTime.IsZero() {
		_ = a.activity.FinalizeOpenActivityForLogSource(ctx, ingestAdapter.LogSourcePath(), lastVRLineTime)
	}
	st, statErr := os.Stat(pathCopy)
	endOff := int64(0)
	if statErr == nil && st != nil {
		endOff = st.Size()
	}
	_ = a.activity.SetActivityLogFileCheckpoint(ctx, absWatch, logwatcher.AbsLogPath(pathCopy), endOff, logwatcher.CheckpointVRTime(lastVRLineTime))
}

func (a *App) startOutputLogWatcher(ctx context.Context) {
	watchPath, err := a.resolveEffectiveOutputLogWatchPath(ctx)
	if err != nil || watchPath == "" {
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			runtime.LogWarning(ctx, "output_log watch path: "+err.Error())
		}
		return
	}

	parser := activity.NewLogParser()
	logger := appDiagLogger()
	emitEncounters := func() {
		runtime.EventsEmit(a.ctx, activityEncountersChangedEvent, struct{}{})
	}
	emitVideoPlayback := func() {
		runtime.EventsEmit(a.ctx, activityVideoPlaybackChangedEvent, struct{}{})
	}

	a.resetActivityIngestAdapterCache()
	a.ingestActivityLogsBootstrap(ctx, watchPath, parser, logger, emitEncounters, emitVideoPlayback)
	if _, dedupeErr := a.activity.DeduplicateEncounters(ctx); dedupeErr != nil {
		runtime.LogWarning(ctx, "activity encounter dedupe: "+dedupeErr.Error())
	}
	if _, backfillErr := a.activity.BackfillEncounterWorldContext(ctx); backfillErr != nil {
		runtime.LogWarning(ctx, "activity encounter world backfill: "+backfillErr.Error())
	}

	watchDeps := activityLogWatchDeps{
		watchPath:         watchPath,
		parser:            parser,
		logger:            logger,
		emitEncounters:    emitEncounters,
		emitVideoPlayback: emitVideoPlayback,
	}

	watcher := logwatcher.NewMultiOutputLogWatcher(watchPath, parser, func(logPath string) logwatcher.EventHandler {
		adapter := a.activityIngestAdapterForPath(ctx, logger, emitEncounters, emitVideoPlayback, logPath)
		triggerHandler := logwatcher.NewAutomationTriggerHandler(a.automation, ctx, logger)
		return logwatcher.FanoutHandler{adapter, triggerHandler}
	}, logwatcher.MultiOutputLogWatcherCallbacks{
		OnLogRotationHandoff: func(c context.Context, oldPath string) error {
			return a.handleActivityLogRotationHandoff(c, watchDeps, oldPath)
		},
		OnTailCheckpoint: func(c context.Context, path string, offset int64, lineTime time.Time) {
			if a.activity == nil {
				return
			}
			_ = a.activity.SetActivityLogFileCheckpoint(c, watchDeps.watchPath, logwatcher.AbsLogPath(path), offset, logwatcher.CheckpointVRTime(lineTime))
		},
	}, logger)
	if startErr := watcher.Start(ctx); startErr != nil {
		runtime.LogError(ctx, "failed to start multi output_log watcher: "+startErr.Error())
		return
	}

	a.startVRChatActivityMonitor(ctx, watchPath)
	runtime.LogInfo(ctx, "output_log watcher started for "+watchPath)
}

// ValidateOutputLogPath checks if path is an existing directory (empty dirs allowed).
func (a *App) ValidateOutputLogPath(path string) bool {
	return logwatcher.OutputLogPathValid(path)
}
