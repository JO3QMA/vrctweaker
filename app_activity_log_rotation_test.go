package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/testvrc"
)

func Test_handleActivityLogRotationHandoff_finalizesOldSession(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ctx := context.Background()
	dir := t.TempDir()

	const cozyWorld = "wrld_6041ba53-0ac0-4b5b-9ecb-890ea2b0aefa"
	cozyInst := cozyWorld + ":48580~friends(" + testvrc.FriendsHostUserID + ")~region(jp)"
	joinedAt := time.Date(2026, 6, 24, 8, 20, 0, 0, time.UTC)

	oldPath := filepath.Join(dir, "output_log_2026-06-24_08-00-00.txt")
	oldContent := []byte(
		"2026.06.24 08:20:00 Debug      -  [Behaviour] Joining " + cozyInst + "\n" +
			"2026.06.24 08:20:05 Debug      -  [Behaviour] OnPlayerJoined Alice (usr_abc)\n" +
			"2026.06.24 08:25:00 Debug      -  [EOSManager] old file tail\n",
	)
	if err := os.WriteFile(oldPath, oldContent, 0600); err != nil {
		t.Fatal(err)
	}

	newPath := filepath.Join(dir, "output_log_2026-06-24_08-30-00.txt")
	newContent := []byte(
		"2026.06.24 08:30:10 Debug      -  [Behaviour] Destination set: " + cozyInst + "\n" +
			"2026.06.24 08:30:11 Debug      -  [Behaviour] Entering Room: Cozy with․\n" +
			"2026.06.24 08:30:12 Debug      -  [Behaviour] Joining " + cozyInst + "\n",
	)
	if err := os.WriteFile(newPath, newContent, 0600); err != nil {
		t.Fatal(err)
	}

	app, encounterRepo := newTestAppWithActivity(t)
	absOldPath, err := filepath.Abs(oldPath)
	if err != nil {
		t.Fatal(err)
	}
	if saveErr := encounterRepo.Save(ctx, &activity.UserEncounter{
		ID:            "enc-open",
		VRCUserID:     "usr_abc",
		DisplayName:   "Alice",
		InstanceID:    cozyInst,
		WorldID:       cozyWorld,
		LogSourcePath: absOldPath,
		JoinedAt:      joinedAt,
	}); saveErr != nil {
		t.Fatal(saveErr)
	}
	if startErr := app.activity.StartPlaySession(ctx, absOldPath, cozyInst, joinedAt); startErr != nil {
		t.Fatal(startErr)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	parser := activity.NewLogParser()
	deps := activityLogWatchDeps{
		watchPath: absDir,
		parser:    parser,
		logger:    appDiagLogger(),
	}

	if handoffErr := app.handleActivityLogRotationHandoff(ctx, deps, absOldPath); handoffErr != nil {
		t.Fatal(handoffErr)
	}

	rows, err := app.activity.ListEncounters(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("encounters = %d, want 1", len(rows))
	}
	if rows[0].LeftAt == nil {
		t.Fatal("encounter left_at is nil, want finalized when VRChat log rotated")
	}
	wantClose := time.Date(2026, 6, 24, 8, 25, 0, 0, time.UTC)
	if !rows[0].LeftAt.Equal(wantClose) {
		t.Fatalf("LeftAt = %v, want %v", rows[0].LeftAt, wantClose)
	}

	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	sessions, err := app.activity.ListPlaySessions(ctx, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].EndTime == nil {
		t.Fatalf("play session should be closed: %+v", sessions)
	}
}

func Test_finalizeOpenActivityForLogSource_usesLastTimestamp(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ctx := context.Background()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "output_log.txt")
	const inst = testBootstrapInstID
	joinedAt := time.Date(2026, 3, 21, 11, 32, 4, 0, time.UTC)
	lastLine := time.Date(2026, 3, 21, 11, 45, 0, 0, time.UTC)
	content := []byte(
		"2026.03.21 11:32:04 Debug      -  [Behaviour] Joining " + inst + "\n" +
			"2026.03.21 11:45:00 Debug      -  [EOSManager] heartbeat\n",
	)
	if err := os.WriteFile(logPath, content, 0600); err != nil {
		t.Fatal(err)
	}

	app, _ := newTestAppWithActivity(t)
	absLogPath, err := filepath.Abs(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if startErr := app.activity.StartPlaySession(ctx, absLogPath, inst, joinedAt); startErr != nil {
		t.Fatal(startErr)
	}
	app.finalizeOpenActivityForLogSource(ctx, absLogPath)

	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	sessions, err := app.activity.ListPlaySessions(ctx, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].EndTime == nil || !sessions[0].EndTime.Equal(lastLine) {
		t.Fatalf("sessions = %+v, want one closed at %v", sessions, lastLine)
	}
}
