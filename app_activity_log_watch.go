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
	ingestAdapter  *logwatcher.ActivityIngestAdapter
	handler        logwatcher.EventHandler
	logger         logwatcher.Logger
	emitEncounters func()
}

func (a *App) finalizeOpenActivityAtLogPath(ctx context.Context, logPath string) {
	if a.activity == nil || logPath == "" {
		return
	}
	lastTime, err := logwatcher.LastVRChatLineTimeInFile(logPath)
	if err != nil {
		runtime.LogWarning(ctx, "activity finalize log time: "+err.Error())
	}
	if lastTime.IsZero() {
		if cp, cpErr := a.activity.GetActivityLogCheckpoint(ctx); cpErr == nil && cp != nil && cp.VRChatLineTime != "" {
			if t, parseErr := time.Parse(time.RFC3339, cp.VRChatLineTime); parseErr == nil {
				lastTime = t
			}
		}
	}
	if lastTime.IsZero() {
		return
	}
	_ = a.activity.CloseOpenPlaySessionAtLastLogLine(ctx, lastTime)
	_ = a.activity.CloseOpenEncountersAtLastLogLine(ctx, lastTime)
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, activityEncountersChangedEvent, struct{}{})
	}
}

func (a *App) handleActivityLogFileSwitch(ctx context.Context, deps activityLogWatchDeps, previousPath, newPath string) error {
	if previousPath != "" {
		a.finalizeOpenActivityAtLogPath(ctx, previousPath)
	}
	if deps.ingestAdapter != nil {
		deps.ingestAdapter.ResetSessionContextForNewLogFile()
	}
	if newPath == "" || deps.parser == nil || deps.handler == nil {
		return nil
	}

	var lastVRLineTime time.Time
	endOff, err := logwatcher.ProcessOutputLogFileFromOffset(ctx, newPath, 0, deps.parser, deps.handler, deps.logger, func(_ int64, line string) {
		if ts := activity.ParseVRChatTimestamp(line, time.Time{}); !ts.IsZero() {
			lastVRLineTime = ts
		}
	})
	if err != nil {
		return err
	}

	if a.activity != nil && deps.watchPath != "" {
		absNew, absErr := filepath.Abs(filepath.Clean(newPath))
		if absErr != nil {
			absNew = newPath
		}
		_ = a.activity.SetActivityLogCheckpoint(ctx, &usecase.ActivityLogCheckpoint{
			WatchPath:      deps.watchPath,
			File:           absNew,
			ByteOffset:     endOff,
			VRChatLineTime: formatActivityCheckpointLineTime(lastVRLineTime),
		})
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
			logPath, err := a.resolveActiveOutputLogFilePath(ctx, watchPath)
			if err != nil || logPath == "" {
				a.finalizeOpenActivityFromCheckpoint(ctx)
				return
			}
			a.finalizeOpenActivityAtLogPath(ctx, logPath)
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

func (a *App) finalizeOpenActivityFromCheckpoint(ctx context.Context) {
	if a.activity == nil {
		return
	}
	cp, err := a.activity.GetActivityLogCheckpoint(ctx)
	if err != nil || cp == nil || cp.VRChatLineTime == "" {
		return
	}
	lastTime, parseErr := time.Parse(time.RFC3339, cp.VRChatLineTime)
	if parseErr != nil || lastTime.IsZero() {
		return
	}
	_ = a.activity.CloseOpenPlaySessionAtLastLogLine(ctx, lastTime)
	_ = a.activity.CloseOpenEncountersAtLastLogLine(ctx, lastTime)
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, activityEncountersChangedEvent, struct{}{})
	}
}

func (a *App) resolveActiveOutputLogFilePath(ctx context.Context, watchPath string) (string, error) {
	p := watchPath
	if p == "" {
		var err error
		p, err = a.resolveEffectiveOutputLogWatchPath(ctx)
		if err != nil {
			return "", err
		}
	}
	info, err := os.Stat(p)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return logwatcher.ResolveLatestOutputLogFile(p)
	}
	return filepath.Abs(filepath.Clean(p))
}
