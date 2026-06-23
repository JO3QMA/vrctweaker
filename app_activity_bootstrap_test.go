package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/infrastructure/logwatcher"
	"vrchat-tweaker/internal/infrastructure/sqlite"
	"vrchat-tweaker/internal/usecase"
)

const (
	testBootstrapWorldID = "wrld_db637cfb-64f8-4109-977b-6b755482f133"
	testBootstrapInstID  = testBootstrapWorldID + ":88577~region(jp)"
	testBootstrapUserID  = "usr_abc"
)

func Test_shouldFinalizeOpenActivityAtLogFileEnd(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name            string
		isDirectoryMode bool
		fileIndex       int
		fileCount       int
		want            bool
	}{
		{
			name:            "single file watch never finalizes",
			isDirectoryMode: false,
			fileIndex:       0,
			fileCount:       1,
			want:            false,
		},
		{
			name:            "directory with one file is still active tail",
			isDirectoryMode: true,
			fileIndex:       0,
			fileCount:       1,
			want:            false,
		},
		{
			name:            "directory historical file before tail finalizes",
			isDirectoryMode: true,
			fileIndex:       0,
			fileCount:       3,
			want:            true,
		},
		{
			name:            "directory active tail file does not finalize",
			isDirectoryMode: true,
			fileIndex:       2,
			fileCount:       3,
			want:            false,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := shouldFinalizeOpenActivityAtLogFileEnd(tc.isDirectoryMode, tc.fileIndex, tc.fileCount); got != tc.want {
				t.Fatalf("shouldFinalizeOpenActivityAtLogFileEnd() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_ingestActivityLogsBootstrap_restart_keepsOpenEncounters_singleFile(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ctx := context.Background()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "output_log.txt")
	initial := []byte(
		"2026.03.21 11:32:04 Debug      -  [Behaviour] Joining " + testBootstrapInstID + "\n" +
			"2026.03.21 11:32:16 Debug      -  [Behaviour] OnPlayerJoined Alice (" + testBootstrapUserID + ")\n",
	)
	if err := os.WriteFile(logPath, initial, 0600); err != nil {
		t.Fatal(err)
	}

	app, encounterRepo := newTestAppWithActivity(t)
	joinedAt := time.Date(2026, 3, 21, 11, 32, 16, 0, time.UTC)
	if err := encounterRepo.Save(ctx, &activity.UserEncounter{
		ID:          "enc-open",
		VRCUserID:   testBootstrapUserID,
		DisplayName: "Alice",
		InstanceID:  testBootstrapInstID,
		WorldID:     testBootstrapWorldID,
		JoinedAt:    joinedAt,
	}); err != nil {
		t.Fatal(err)
	}
	absLogPath, err := filepath.Abs(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.activity.SetActivityLogCheckpoint(ctx, &usecase.ActivityLogCheckpoint{
		WatchPath:      absLogPath,
		File:           absLogPath,
		ByteOffset:     int64(len(initial)),
		VRChatLineTime: joinedAt.Format(time.RFC3339),
	}); err != nil {
		t.Fatal(err)
	}

	appendLine := "2026.03.21 11:40:00 Debug      -  [EOSManager] heartbeat while VRCTweaker was closed\n"
	if err := os.WriteFile(logPath, append(initial, []byte(appendLine)...), 0600); err != nil {
		t.Fatal(err)
	}

	runActivityBootstrap(t, app, ctx, absLogPath)
	assertEncounterStillOpen(t, app, ctx, testBootstrapUserID)
}

func Test_ingestActivityLogsBootstrap_restart_keepsOpenEncounters_directoryTailFile(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ctx := context.Background()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "output_log_2026-03-21_11-00-00.txt")
	initial := []byte(
		"2026.03.21 11:32:04 Debug      -  [Behaviour] Joining " + testBootstrapInstID + "\n" +
			"2026.03.21 11:32:16 Debug      -  [Behaviour] OnPlayerJoined Alice (" + testBootstrapUserID + ")\n",
	)
	if err := os.WriteFile(logPath, initial, 0600); err != nil {
		t.Fatal(err)
	}

	app, encounterRepo := newTestAppWithActivity(t)
	joinedAt := time.Date(2026, 3, 21, 11, 32, 16, 0, time.UTC)
	if err := encounterRepo.Save(ctx, &activity.UserEncounter{
		ID:          "enc-open",
		VRCUserID:   testBootstrapUserID,
		DisplayName: "Alice",
		InstanceID:  testBootstrapInstID,
		WorldID:     testBootstrapWorldID,
		JoinedAt:    joinedAt,
	}); err != nil {
		t.Fatal(err)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	absLogPath, err := filepath.Abs(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.activity.SetActivityLogCheckpoint(ctx, &usecase.ActivityLogCheckpoint{
		WatchPath:      absDir,
		File:           absLogPath,
		ByteOffset:     int64(len(initial)),
		VRChatLineTime: joinedAt.Format(time.RFC3339),
	}); err != nil {
		t.Fatal(err)
	}

	appendLine := "2026.03.21 11:40:00 Debug      -  [EOSManager] heartbeat while VRCTweaker was closed\n"
	if err := os.WriteFile(logPath, append(initial, []byte(appendLine)...), 0600); err != nil {
		t.Fatal(err)
	}

	runActivityBootstrap(t, app, ctx, absDir)
	assertEncounterStillOpen(t, app, ctx, testBootstrapUserID)
}

func Test_ingestActivityLogsBootstrap_finalizesOpenEncounters_onHistoricalDirectoryFile(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ctx := context.Background()
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "output_log_2026-03-20_23-00-00.txt")
	newPath := filepath.Join(dir, "output_log_2026-03-21_11-00-00.txt")
	oldContent := []byte("2026.03.20 23:50:00 Debug      -  [EOSManager] old file tail\n")
	if err := os.WriteFile(oldPath, oldContent, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte("2026.03.21 11:00:00 Debug      -  [EOSManager] new file start\n"), 0600); err != nil {
		t.Fatal(err)
	}

	app, encounterRepo := newTestAppWithActivity(t)
	joinedAt := time.Date(2026, 3, 20, 22, 0, 0, 0, time.UTC)
	closeAt := time.Date(2026, 3, 20, 23, 50, 0, 0, time.UTC)
	if err := encounterRepo.Save(ctx, &activity.UserEncounter{
		ID:          "enc-historical",
		VRCUserID:   testBootstrapUserID,
		DisplayName: "Alice",
		InstanceID:  testBootstrapInstID,
		WorldID:     testBootstrapWorldID,
		JoinedAt:    joinedAt,
	}); err != nil {
		t.Fatal(err)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	runActivityBootstrap(t, app, ctx, absDir)

	rows, err := app.activity.ListEncounters(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("ListEncounters() = %d rows, want 1", len(rows))
	}
	if rows[0].LeftAt == nil {
		t.Fatal("historical file bootstrap should close open encounters at last line of completed file")
	}
	if !rows[0].LeftAt.Equal(closeAt) {
		t.Fatalf("LeftAt = %v, want %v", rows[0].LeftAt, closeAt)
	}
}

func newTestAppWithActivity(t *testing.T) (*App, *sqlite.UserEncounterRepository) {
	t.Helper()
	dataDir := t.TempDir()
	db, err := sqlite.Open(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	encounterRepo := sqlite.NewUserEncounterRepository(db)
	app := &App{
		activity: usecase.NewActivityUseCase(
			sqlite.NewPlaySessionRepository(db),
			encounterRepo,
			sqlite.NewAppSettingsRepository(db),
			sqlite.NewUserCacheRepository(db),
			sqlite.NewWorldInfoRepository(db),
		),
	}
	return app, encounterRepo
}

func runActivityBootstrap(t *testing.T, app *App, ctx context.Context, watchPath string) {
	t.Helper()
	parser := activity.NewLogParser()
	adapter := logwatcher.NewActivityIngestAdapter(app.activity, ctx, logLogger{}, nil)
	app.ingestActivityLogsBootstrap(ctx, watchPath, parser, adapter, logLogger{})
}

func assertEncounterStillOpen(t *testing.T, app *App, ctx context.Context, vrcUserID string) {
	t.Helper()
	rows, err := app.activity.ListEncounters(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.VRCUserID != vrcUserID {
			continue
		}
		if row.LeftAt != nil {
			t.Fatalf("encounter %s: left_at = %v, want nil (user still in world after VRCTweaker restart)", vrcUserID, row.LeftAt)
		}
		return
	}
	t.Fatalf("encounter for %s not found", vrcUserID)
}
