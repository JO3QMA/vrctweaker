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

func Test_ingestActivityLogsBootstrap_checkpointResume_assignsWorldRoomName(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ctx := context.Background()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "output_log.txt")

	const (
		homeWorld = "wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b"
		cozyWorld = "wrld_6041ba53-0ac0-4b5b-9ecb-890ea2b0aefa"
		buddyID   = "usr_buddy"
	)
	homeInst := homeWorld + ":95147~private(usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e)~region(jp)"
	cozyInst := cozyWorld + ":48580~friends(usr_b4cb47f9-ca01-43db-baa3-ce3fb98ff0d4)~region(jp)"

	prefix := []byte(
		"2026.06.24 08:25:00 Debug      -  [Behaviour] Destination set: " + homeInst + "\n" +
			"2026.06.24 08:25:01 Debug      -  [Behaviour] OnLeftRoom\n" +
			"2026.06.24 08:25:03 Debug      -  [Behaviour] Entering Room: HomeCheck\n" +
			"2026.06.24 08:25:03 Debug      -  [Behaviour] Joining " + homeInst + "\n",
	)
	suffix := []byte(
		"2026.06.24 08:26:40 Debug      -  [Behaviour] Destination set: " + cozyInst + "\n" +
			"2026.06.24 08:26:41 Debug      -  [Behaviour] OnLeftRoom\n" +
			"2026.06.24 08:26:44 Debug      -  [Behaviour] Entering Room: Cozy with.\n" +
			"2026.06.24 08:26:44 Debug      -  [Behaviour] Joining " + cozyInst + "\n" +
			"2026.06.24 08:26:50 Debug      -  [Behaviour] OnPlayerJoined Buddy (" + buddyID + ")\n",
	)
	if err := os.WriteFile(logPath, prefix, 0600); err != nil {
		t.Fatal(err)
	}

	app, _ := newTestAppWithActivity(t)
	absLogPath, err := filepath.Abs(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if cpErr := app.activity.SetActivityLogCheckpoint(ctx, &usecase.ActivityLogCheckpoint{
		WatchPath:      absLogPath,
		File:           absLogPath,
		ByteOffset:     int64(len(prefix)),
		VRChatLineTime: time.Date(2026, 6, 24, 8, 25, 3, 0, time.UTC).Format(time.RFC3339),
	}); cpErr != nil {
		t.Fatal(cpErr)
	}

	if writeErr := os.WriteFile(logPath, append(prefix, suffix...), 0600); writeErr != nil {
		t.Fatal(writeErr)
	}

	runActivityBootstrap(t, app, ctx, absLogPath)

	rows, err := app.activity.ListEncountersWithContext(ctx, &activity.EncounterFilter{VRCUserID: buddyID})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("ListEncountersWithContext() = %d rows, want 1", len(rows))
	}
	row := rows[0]
	if row.Encounter.WorldID != cozyWorld {
		t.Fatalf("world_id = %q, want %q", row.Encounter.WorldID, cozyWorld)
	}
	if row.WorldDisplayName != "Cozy with." {
		t.Fatalf("world display = %q, want %q", row.WorldDisplayName, "Cozy with.")
	}
}

func Test_ingestActivityLogsBootstrap_checkpointResume_homeEncounterKeepsWorldID(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ctx := context.Background()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "output_log.txt")

	const (
		homeWorld = "wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b"
		hostID    = "usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e"
	)
	homeInst := homeWorld + ":95147~private(" + hostID + ")~region(jp)"

	prefix := []byte(
		"2026.06.24 08:25:00 Debug      -  [Behaviour] Destination set: " + homeInst + "\n" +
			"2026.06.24 08:25:01 Debug      -  [Behaviour] OnLeftRoom\n" +
			"2026.06.24 08:25:03 Debug      -  [Behaviour] Entering Room: ホームチェックv6․0\n" +
			"2026.06.24 08:25:03 Debug      -  [Behaviour] Joining " + homeInst + "\n",
	)
	suffix := []byte(
		"2026.06.24 08:25:18 Debug      -  [Behaviour] OnPlayerJoined ぶっちゃん！ (" + hostID + ")\n",
	)
	if writeErr := os.WriteFile(logPath, prefix, 0600); writeErr != nil {
		t.Fatal(writeErr)
	}

	app, _ := newTestAppWithActivity(t)
	absLogPath, err := filepath.Abs(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if cpErr := app.activity.SetActivityLogCheckpoint(ctx, &usecase.ActivityLogCheckpoint{
		WatchPath:      absLogPath,
		File:           absLogPath,
		ByteOffset:     int64(len(prefix)),
		VRChatLineTime: time.Date(2026, 6, 24, 8, 25, 3, 0, time.UTC).Format(time.RFC3339),
	}); cpErr != nil {
		t.Fatal(cpErr)
	}
	if writeErr := os.WriteFile(logPath, append(prefix, suffix...), 0600); writeErr != nil {
		t.Fatal(writeErr)
	}

	runActivityBootstrap(t, app, ctx, absLogPath)

	rows, err := app.activity.ListEncountersWithContext(ctx, &activity.EncounterFilter{VRCUserID: hostID})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("ListEncountersWithContext() = %d rows, want 1", len(rows))
	}
	row := rows[0]
	if row.Encounter.WorldID != homeWorld {
		t.Fatalf("world_id = %q, want %q", row.Encounter.WorldID, homeWorld)
	}
	if row.Encounter.InstanceID != homeInst {
		t.Fatalf("instance_id = %q, want %q", row.Encounter.InstanceID, homeInst)
	}
}

func Test_ingestActivityLogsBootstrap_checkpointResume_worldNamesNotCrossAssigned(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ctx := context.Background()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "output_log.txt")

	const (
		homeWorld = "wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b"
		cozyWorld = "wrld_6041ba53-0ac0-4b5b-9ecb-890ea2b0aefa"
		hostID    = "usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e"
		buddyID   = "usr_buddy"
	)
	homeInst := homeWorld + ":95147~private(" + hostID + ")~region(jp)"
	cozyInst := cozyWorld + ":48580~friends(usr_b4cb47f9-ca01-43db-baa3-ce3fb98ff0d4)~region(jp)"

	prefix := []byte(
		"2026.06.24 08:25:00 Debug      -  [Behaviour] Destination set: " + homeInst + "\n" +
			"2026.06.24 08:25:01 Debug      -  [Behaviour] OnLeftRoom\n" +
			"2026.06.24 08:25:03 Debug      -  [Behaviour] Entering Room: ホームチェックv6․0\n" +
			"2026.06.24 08:25:03 Debug      -  [Behaviour] Joining " + homeInst + "\n" +
			"2026.06.24 08:25:18 Debug      -  [Behaviour] OnPlayerJoined ぶっちゃん！ (" + hostID + ")\n",
	)
	suffix := []byte(
		"2026.06.24 08:26:40 Debug      -  [Behaviour] Destination set: " + cozyInst + "\n" +
			"2026.06.24 08:26:41 Debug      -  [Behaviour] OnLeftRoom\n" +
			"2026.06.24 08:26:44 Debug      -  [Behaviour] Entering Room: Cozy with․\n" +
			"2026.06.24 08:26:44 Debug      -  [Behaviour] Joining " + cozyInst + "\n" +
			"2026.06.24 08:26:54 Debug      -  [Behaviour] OnPlayerJoined Buddy (" + buddyID + ")\n",
	)
	if writeErr := os.WriteFile(logPath, prefix, 0600); writeErr != nil {
		t.Fatal(writeErr)
	}

	app, encounterRepo := newTestAppWithActivity(t)
	homeJoinedAt := time.Date(2026, 6, 24, 8, 25, 18, 0, time.UTC)
	if seedErr := app.activity.UpsertWorldRoomName(ctx, homeWorld, "ホームチェックv6․0", homeJoinedAt); seedErr != nil {
		t.Fatal(seedErr)
	}
	if seedErr := encounterRepo.Save(ctx, &activity.UserEncounter{
		ID:          "enc-home-seed",
		VRCUserID:   hostID,
		DisplayName: "ぶっちゃん！",
		InstanceID:  homeInst,
		WorldID:     homeWorld,
		JoinedAt:    homeJoinedAt,
	}); seedErr != nil {
		t.Fatal(seedErr)
	}

	absLogPath, err := filepath.Abs(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if cpErr := app.activity.SetActivityLogCheckpoint(ctx, &usecase.ActivityLogCheckpoint{
		WatchPath:      absLogPath,
		File:           absLogPath,
		ByteOffset:     int64(len(prefix)),
		VRChatLineTime: homeJoinedAt.Format(time.RFC3339),
	}); cpErr != nil {
		t.Fatal(cpErr)
	}
	if writeErr := os.WriteFile(logPath, append(prefix, suffix...), 0600); writeErr != nil {
		t.Fatal(writeErr)
	}

	runActivityBootstrap(t, app, ctx, absLogPath)

	homeRows, err := app.activity.ListEncountersWithContext(ctx, &activity.EncounterFilter{VRCUserID: hostID})
	if err != nil {
		t.Fatal(err)
	}
	if len(homeRows) != 1 {
		t.Fatalf("home encounters = %d, want 1", len(homeRows))
	}
	if homeRows[0].Encounter.WorldID != homeWorld {
		t.Fatalf("home world_id = %q, want %q", homeRows[0].Encounter.WorldID, homeWorld)
	}
	if homeRows[0].WorldDisplayName != "ホームチェックv6․0" {
		t.Fatalf("home world display = %q, want ホームチェックv6․0 (must not show Cozy with.)", homeRows[0].WorldDisplayName)
	}

	cozyRows, err := app.activity.ListEncountersWithContext(ctx, &activity.EncounterFilter{VRCUserID: buddyID})
	if err != nil {
		t.Fatal(err)
	}
	if len(cozyRows) != 1 {
		t.Fatalf("cozy encounters = %d, want 1", len(cozyRows))
	}
	if cozyRows[0].Encounter.WorldID != cozyWorld {
		t.Fatalf("cozy world_id = %q, want %q", cozyRows[0].Encounter.WorldID, cozyWorld)
	}
	if cozyRows[0].WorldDisplayName != "Cozy with․" {
		t.Fatalf("cozy world display = %q, want Cozy with․", cozyRows[0].WorldDisplayName)
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
