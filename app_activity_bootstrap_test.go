package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/infrastructure/sqlite"
	"vrchat-tweaker/internal/testvrc"
	"vrchat-tweaker/internal/usecase"
)

const (
	testBootstrapWorldID = "wrld_db637cfb-64f8-4109-977b-6b755482f133"
	testBootstrapInstID  = testBootstrapWorldID + ":88577~region(jp)"
	testBootstrapUserID  = "usr_abc"
)

func Test_bootstrapLiveLogFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "output_log_2026-03-20_23-00-00.txt")
	midPath := filepath.Join(dir, "output_log_2026-03-21_08-00-00.txt")
	newPath := filepath.Join(dir, "output_log_2026-03-21_11-00-00.txt")

	base := time.Date(2026, 3, 21, 11, 0, 0, 0, time.UTC)
	for _, spec := range []struct {
		path string
		mod  time.Time
	}{
		{oldPath, base.Add(-12 * time.Hour)},
		{midPath, base.Add(-3 * time.Hour)},
		{newPath, base},
	} {
		if err := os.WriteFile(spec.path, []byte("log\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(spec.path, spec.mod, spec.mod); err != nil {
			t.Fatal(err)
		}
	}

	live := bootstrapLiveLogFiles([]string{oldPath, midPath, newPath})
	if live == nil {
		t.Fatal("expected live map for directory with recent tail file")
	}
	if live[oldPath] || live[midPath] {
		t.Fatalf("historical files should not be live: old=%v mid=%v", live[oldPath], live[midPath])
	}
	if !live[newPath] {
		t.Fatal("newest file within live window should be marked live")
	}

	single := bootstrapLiveLogFiles([]string{newPath})
	if single == nil || !single[newPath] {
		t.Fatalf("single file watch should mark only file live: %+v", single)
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
	base := time.Date(2026, 3, 21, 11, 0, 0, 0, time.UTC)
	if err := os.Chtimes(oldPath, base.Add(-12*time.Hour), base.Add(-12*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(newPath, base, base); err != nil {
		t.Fatal(err)
	}

	app, encounterRepo := newTestAppWithActivity(t)
	absOldPath, err := filepath.Abs(oldPath)
	if err != nil {
		t.Fatal(err)
	}
	joinedAt := time.Date(2026, 3, 20, 22, 0, 0, 0, time.UTC)
	closeAt := time.Date(2026, 3, 20, 23, 50, 0, 0, time.UTC)
	if saveErr := encounterRepo.Save(ctx, &activity.UserEncounter{
		ID:            "enc-historical",
		VRCUserID:     testBootstrapUserID,
		DisplayName:   "Alice",
		InstanceID:    testBootstrapInstID,
		WorldID:       testBootstrapWorldID,
		LogSourcePath: absOldPath,
		JoinedAt:      joinedAt,
	}); saveErr != nil {
		t.Fatal(saveErr)
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
	homeInst := homeWorld + ":95147~private(" + testvrc.PlayerUserID + ")~region(jp)"
	cozyInst := cozyWorld + ":48580~friends(" + testvrc.FriendsHostUserID + ")~region(jp)"

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
		hostID    = testvrc.PlayerUserID
	)
	homeInst := homeWorld + ":95147~private(" + hostID + ")~region(jp)"

	prefix := []byte(
		"2026.06.24 08:25:00 Debug      -  [Behaviour] Destination set: " + homeInst + "\n" +
			"2026.06.24 08:25:01 Debug      -  [Behaviour] OnLeftRoom\n" +
			"2026.06.24 08:25:03 Debug      -  [Behaviour] Entering Room: ホームチェックv6․0\n" +
			"2026.06.24 08:25:03 Debug      -  [Behaviour] Joining " + homeInst + "\n",
	)
	suffix := []byte(
		"2026.06.24 08:25:18 Debug      -  [Behaviour] OnPlayerJoined " + testvrc.PlayerDisplayName + " (" + hostID + ")\n",
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
		hostID    = testvrc.PlayerUserID
		buddyID   = "usr_buddy"
	)
	homeInst := homeWorld + ":95147~private(" + hostID + ")~region(jp)"
	cozyInst := cozyWorld + ":48580~friends(" + testvrc.FriendsHostUserID + ")~region(jp)"

	prefix := []byte(
		"2026.06.24 08:25:00 Debug      -  [Behaviour] Destination set: " + homeInst + "\n" +
			"2026.06.24 08:25:01 Debug      -  [Behaviour] OnLeftRoom\n" +
			"2026.06.24 08:25:03 Debug      -  [Behaviour] Entering Room: ホームチェックv6․0\n" +
			"2026.06.24 08:25:03 Debug      -  [Behaviour] Joining " + homeInst + "\n" +
			"2026.06.24 08:25:18 Debug      -  [Behaviour] OnPlayerJoined " + testvrc.PlayerDisplayName + " (" + hostID + ")\n",
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
		DisplayName: testvrc.PlayerDisplayName,
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
	app.ingestActivityLogsBootstrap(ctx, watchPath, parser, appDiagLogger(), nil)
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
