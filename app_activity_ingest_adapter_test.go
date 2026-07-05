package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

// Regression: multi-log tail must reuse the bootstrap adapter so SessionCorrelator keeps world context.
func Test_activityIngestAdapterForPath_tailAfterBootstrap_hasWorldContext(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ctx := context.Background()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "output_log.txt")

	joiningOnly := []byte(
		"2026.07.02 21:48:00 Debug      -  [Behaviour] Joining " + testBootstrapInstID + "\n",
	)
	if err := os.WriteFile(logPath, joiningOnly, 0600); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if err := os.Chtimes(logPath, now, now); err != nil {
		t.Fatal(err)
	}

	app, _ := newTestAppWithActivity(t)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.resetActivityIngestAdapterCache()
	parser := activity.NewLogParser()
	logger := appDiagLogger()
	app.ingestActivityLogsBootstrap(ctx, absDir, parser, logger, nil)

	tailAdapter := app.activityIngestAdapterForPath(ctx, logger, nil, logPath)
	joinLine := "2026.07.02 21:49:04 Debug      -  [Behaviour] OnPlayerJoined Alice (" + testBootstrapUserID + ")"
	events, err := parser.ParseLine(joinLine, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("ParseLine() = %d events, want 1", len(events))
	}
	tailAdapter.Handle(events[0])

	rows, err := app.activity.ListEncountersWithContext(ctx, &activity.EncounterFilter{VRCUserID: testBootstrapUserID})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("ListEncountersWithContext() = %d rows, want 1", len(rows))
	}
	if rows[0].Encounter.WorldID != testBootstrapWorldID {
		t.Fatalf("world_id = %q, want %q", rows[0].Encounter.WorldID, testBootstrapWorldID)
	}
	if rows[0].Encounter.InstanceID != testBootstrapInstID {
		t.Fatalf("instance_id = %q, want %q", rows[0].Encounter.InstanceID, testBootstrapInstID)
	}
}

func Test_activityIngestAdapterForPath_sameInstanceAcrossCalls(t *testing.T) {
	ctx := context.Background()
	app, _ := newTestAppWithActivity(t)
	logger := appDiagLogger()
	path := filepath.Join(t.TempDir(), "output_log.txt")
	if err := os.WriteFile(path, []byte("x\n"), 0600); err != nil {
		t.Fatal(err)
	}

	app.resetActivityIngestAdapterCache()
	a1 := app.activityIngestAdapterForPath(ctx, logger, nil, path)
	a2 := app.activityIngestAdapterForPath(ctx, logger, nil, path)
	if a1 != a2 {
		t.Fatal("expected same adapter instance for one log source path")
	}
	app.evictActivityIngestAdapter(path)
	a3 := app.activityIngestAdapterForPath(ctx, logger, nil, path)
	if a3 == a1 {
		t.Fatal("expected new adapter after evict")
	}
}
