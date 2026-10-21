package wailsapp

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/infrastructure/sqlite"
	"vrchat-tweaker/internal/usecase"
)

func TestSetLogRetentionDays_purgesExpiredPlaybackAndEncounters(t *testing.T) {
	ctx := context.Background()
	app, db := newRetentionApp(t, ctx)
	oldAt := time.Now().UTC().AddDate(0, 0, -40).Truncate(time.Second)
	freshAt := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	if err := sqlite.NewUserEncounterRepository(db).Save(ctx, &activity.UserEncounter{
		ID: "enc-old", VRCUserID: "vrc-old", DisplayName: "Old", JoinedAt: oldAt,
	}); err != nil {
		t.Fatal(err)
	}
	if err := sqlite.NewUserEncounterRepository(db).Save(ctx, &activity.UserEncounter{
		ID: "enc-fresh", VRCUserID: "vrc-fresh", DisplayName: "Fresh", JoinedAt: freshAt,
	}); err != nil {
		t.Fatal(err)
	}
	if err := sqlite.NewVideoPlaybackRepository(db).Save(ctx, &activity.VideoPlaybackAttempt{
		ID: "vid-old", AttemptedAt: oldAt, URL: "https://example.com/old",
	}); err != nil {
		t.Fatal(err)
	}
	if err := sqlite.NewVideoPlaybackRepository(db).Save(ctx, &activity.VideoPlaybackAttempt{
		ID: "vid-fresh", AttemptedAt: freshAt, URL: "https://example.com/fresh",
	}); err != nil {
		t.Fatal(err)
	}

	if err := app.SetLogRetentionDays(7); err != nil {
		t.Fatal(err)
	}
	if n := retentionCount(t, db, `SELECT COUNT(*) FROM user_encounters WHERE id = 'enc-old'`); n != 0 {
		t.Fatalf("expired encounters = %d", n)
	}
	if n := retentionCount(t, db, `SELECT COUNT(*) FROM user_encounters WHERE id = 'enc-fresh'`); n != 1 {
		t.Fatalf("fresh encounters = %d, want 1", n)
	}
	if n := retentionCount(t, db, `SELECT COUNT(*) FROM video_playback_history WHERE id = 'vid-old'`); n != 0 {
		t.Fatalf("expired video rows = %d", n)
	}
	if n := retentionCount(t, db, `SELECT COUNT(*) FROM video_playback_history WHERE id = 'vid-fresh'`); n != 1 {
		t.Fatalf("fresh video rows = %d, want 1", n)
	}
}

func TestActivityRetentionLoop_purgesExpiredVideo(t *testing.T) {
	ctx := context.Background()
	app, db := newRetentionApp(t, ctx)
	if err := sqlite.NewAppSettingsRepository(db).Set(ctx, "log_retention_days", "7"); err != nil {
		t.Fatal(err)
	}
	oldAt := time.Now().UTC().AddDate(0, 0, -40).Truncate(time.Second)
	freshAt := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	video := sqlite.NewVideoPlaybackRepository(db)
	if err := video.Save(ctx, &activity.VideoPlaybackAttempt{
		ID: "vid-old", AttemptedAt: oldAt, URL: "https://example.com/old",
	}); err != nil {
		t.Fatal(err)
	}
	if err := video.Save(ctx, &activity.VideoPlaybackAttempt{
		ID: "vid-fresh", AttemptedAt: freshAt, URL: "https://example.com/fresh",
	}); err != nil {
		t.Fatal(err)
	}

	app.startActivityRetentionLoopEvery(20 * time.Millisecond)
	t.Cleanup(app.stopActivityRetentionLoop)

	deadline := time.Now().Add(2 * time.Second)
	for retentionCount(t, db, `SELECT COUNT(*) FROM video_playback_history WHERE id = 'vid-old'`) != 0 {
		if time.Now().After(deadline) {
			t.Fatal("retention loop left expired video playback history in place")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if n := retentionCount(t, db, `SELECT COUNT(*) FROM video_playback_history WHERE id = 'vid-fresh'`); n != 1 {
		t.Fatalf("fresh video rows = %d, want 1", n)
	}
}

func newRetentionApp(t *testing.T, ctx context.Context) (*App, *sql.DB) {
	t.Helper()
	db, err := sqlite.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	settingsRepo := sqlite.NewAppSettingsRepository(db)
	app := &App{
		ctx:      ctx,
		settings: usecase.NewSettingsUseCase(settingsRepo),
		activity: usecase.NewActivityUseCase(
			sqlite.NewPlaySessionRepository(db),
			sqlite.NewUserEncounterRepository(db),
			settingsRepo,
			nil,
			nil,
		).WithVideoPlaybackRepo(sqlite.NewVideoPlaybackRepository(db)).
			WithFriendAvatarObservationRepo(sqlite.NewFriendAvatarObservationRepository(db)),
	}
	return app, db
}

func retentionCount(t *testing.T, db *sql.DB, query string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(query).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
