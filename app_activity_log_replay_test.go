package main

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/infrastructure/logwatcher"
	"vrchat-tweaker/internal/usecase"
)

type countingFriendJoined struct {
	n atomic.Int32
}

func (c *countingFriendJoined) OnFriendJoined(context.Context, string) error {
	c.n.Add(1)
	return nil
}

func TestLogFileSwitchReplay_doesNotFireAutomation(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ctx := context.Background()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "output_log.txt")
	joinLine := "2026.03.21 11:32:16 Debug      -  [Behaviour] OnPlayerJoined Alice (" + testBootstrapUserID + ")\n"
	if err := os.WriteFile(logPath, []byte(joinLine), 0600); err != nil {
		t.Fatal(err)
	}

	auto := &countingFriendJoined{}
	parser := activity.NewLogParser()
	logger := appDiagLogger()

	// Control: MultiHandler (live-tail shape) must fire automation on the same Join line.
	controlUC := newReplayTestActivityUseCase(t)
	controlAdapter := logwatcher.NewActivityIngestAdapter(controlUC, ctx, logger, nil, absLogPath(logPath))
	controlHandler := logwatcher.NewMultiHandler(
		controlAdapter,
		logwatcher.NewAutomationTriggerHandler(auto, ctx, logger),
	)
	if _, err := logwatcher.ProcessOutputLogFileFromOffset(ctx, logPath, 0, parser, controlHandler, logger, nil); err != nil {
		t.Fatal(err)
	}
	if auto.n.Load() != 1 {
		t.Fatalf("control MultiHandler automation calls = %d, want 1", auto.n.Load())
	}
	auto.n.Store(0)

	// SUT: Log replay path used after file switch (Activity only).
	app, _ := newTestAppWithActivity(t)
	absWatch, err := filepath.Abs(logPath)
	if err != nil {
		t.Fatal(err)
	}
	adapter := app.activityIngestAdapterForPath(ctx, logger, nil, logPath)
	deps := activityLogWatchDeps{
		watchPath: absWatch,
		parser:    parser,
		logger:    logger,
	}
	app.replayActivityLogAfterFileSwitch(ctx, deps, logPath, adapter)
	if auto.n.Load() != 0 {
		t.Fatalf("Log replay automation calls = %d, want 0", auto.n.Load())
	}

	rows, listErr := app.activity.ListEncounters(ctx, nil)
	if listErr != nil {
		t.Fatal(listErr)
	}
	if len(rows) < 1 {
		t.Fatal("expected encounter from Log replay")
	}
}

func newReplayTestActivityUseCase(t *testing.T) *usecase.ActivityUseCase {
	t.Helper()
	app, _ := newTestAppWithActivity(t)
	return app.activity
}
