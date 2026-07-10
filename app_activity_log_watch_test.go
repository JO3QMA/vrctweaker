package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"vrchat-tweaker/internal/usecase"
)

func Test_finalizeAllLogSourcesOnVRChatExit_usesPerFileLastLine(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ctx := context.Background()
	dir := t.TempDir()

	const instA = testBootstrapInstID
	const instB = "wrld_other:12345~friends(host)~region(jp)"
	logA := filepath.Join(dir, "output_log_a.txt")
	logB := filepath.Join(dir, "output_log_b.txt")
	timeA := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	timeB := time.Date(2026, 3, 21, 11, 0, 0, 0, time.UTC)
	contentA := []byte("2026.03.21 10:00:00 Debug      -  [EOSManager] client A tail\n")
	contentB := []byte("2026.03.21 11:00:00 Debug      -  [EOSManager] client B tail\n")
	if err := os.WriteFile(logA, contentA, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logB, contentB, 0600); err != nil {
		t.Fatal(err)
	}

	app, _ := newTestAppWithActivity(t)
	absA, err := filepath.Abs(logA)
	if err != nil {
		t.Fatal(err)
	}
	absB, err := filepath.Abs(logB)
	if err != nil {
		t.Fatal(err)
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}

	startA := time.Date(2026, 3, 21, 9, 0, 0, 0, time.UTC)
	startB := time.Date(2026, 3, 21, 10, 30, 0, 0, time.UTC)
	if startErr := app.activity.StartPlaySession(ctx, absA, instA, startA); startErr != nil {
		t.Fatal(startErr)
	}
	if startErr := app.activity.StartPlaySession(ctx, absB, instB, startB); startErr != nil {
		t.Fatal(startErr)
	}

	if setErr := app.activity.SetActivityLogCheckpoint(ctx, &usecase.ActivityLogCheckpoint{
		WatchPath: absDir,
		Files: map[string]usecase.ActivityLogFileCheckpoint{
			absA: {ByteOffset: int64(len(contentA)), VRChatLineTime: timeA.Format(time.RFC3339)},
			absB: {ByteOffset: int64(len(contentB)), VRChatLineTime: timeB.Format(time.RFC3339)},
		},
	}); setErr != nil {
		t.Fatal(setErr)
	}

	app.finalizeAllLogSourcesOnVRChatExit(ctx, absDir)

	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	sessions, err := app.activity.ListPlaySessions(ctx, from, to)
	if err != nil {
		t.Fatal(err)
	}
	ends := map[string]time.Time{}
	for _, s := range sessions {
		if s.EndTime == nil {
			t.Fatalf("session %s still open", s.LogSourcePath)
		}
		ends[s.LogSourcePath] = *s.EndTime
	}
	if !ends[absA].Equal(timeA) {
		t.Fatalf("log A end = %v, want %v", ends[absA], timeA)
	}
	if !ends[absB].Equal(timeB) {
		t.Fatalf("log B end = %v, want %v (must not use log A time)", ends[absB], timeB)
	}
}

func Test_finalizeAllLogSourcesOnVRChatExit_closesNullLogSourceRows(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ctx := context.Background()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "output_log.txt")
	closeAt := time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)
	content := []byte("2026.03.21 12:00:00 Debug      -  [EOSManager] tail\n")
	if err := os.WriteFile(logPath, content, 0600); err != nil {
		t.Fatal(err)
	}

	app, _ := newTestAppWithActivity(t)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	absLog, err := filepath.Abs(logPath)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 3, 21, 11, 0, 0, 0, time.UTC)
	// Empty log source → NULL in DB (legacy rows).
	if startErr := app.activity.StartPlaySession(ctx, "", testBootstrapInstID, start); startErr != nil {
		t.Fatal(startErr)
	}
	// Checkpoint supplies LastObservedLogTime so global finalize does not use time.Now()
	// (which would day-split across months).
	if setErr := app.activity.SetActivityLogCheckpoint(ctx, &usecase.ActivityLogCheckpoint{
		WatchPath: absDir,
		Files: map[string]usecase.ActivityLogFileCheckpoint{
			absLog: {ByteOffset: int64(len(content)), VRChatLineTime: closeAt.Format(time.RFC3339)},
		},
	}); setErr != nil {
		t.Fatal(setErr)
	}

	app.finalizeAllLogSourcesOnVRChatExit(ctx, absDir)

	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	sessions, listErr := app.activity.ListPlaySessions(ctx, from, to)
	if listErr != nil {
		t.Fatal(listErr)
	}
	var nullClosed int
	for _, s := range sessions {
		if s.LogSourcePath != "" {
			continue
		}
		if s.EndTime == nil {
			t.Fatal("legacy NULL log_source_path session still open; global finalize should close it")
		}
		nullClosed++
		if !s.EndTime.Equal(closeAt) {
			t.Fatalf("null session end = %v, want %v", s.EndTime, closeAt)
		}
	}
	if nullClosed != 1 {
		t.Fatalf("null log_source sessions closed = %d, want 1 (total sessions %d)", nullClosed, len(sessions))
	}
}

func Test_closeTimeForLogFile_prefersFileTimestamp(t *testing.T) {
	t.Setenv("TZ", "UTC")
	dir := t.TempDir()
	logPath := filepath.Join(dir, "output_log.txt")
	lineTime := time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)
	fallback := time.Date(2026, 3, 21, 13, 0, 0, 0, time.UTC)
	if err := os.WriteFile(logPath, []byte("2026.03.21 12:00:00 Debug      -  [EOSManager] tail\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got := closeTimeForLogFile(logPath, fallback, nil)
	if !got.Equal(lineTime) {
		t.Fatalf("closeTimeForLogFile = %v, want %v", got, lineTime)
	}
}
