package sqlite_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/infrastructure/sqlite"
	"vrchat-tweaker/internal/usecase"
)

// Activity retention must drop rows older than the configured days, keep newer
// rows, and refuse to recreate expired rows when the same log events are applied again.
// Timestamps are stored with the zone of the log line (local). Comparison has to be
// chronological: a +09 clock can sort after a UTC cutoff string while still being earlier.
func TestActivityRetention_purgesExpiredRowsAndBlocksReingest(t *testing.T) {
	ctx := context.Background()
	db := openRetentionDB(t)
	settings := sqlite.NewAppSettingsRepository(db)
	if err := settings.Set(ctx, "log_retention_days", "7"); err != nil {
		t.Fatal(err)
	}

	playRepo := sqlite.NewPlaySessionRepository(db)
	encRepo := sqlite.NewUserEncounterRepository(db)
	videoRepo := sqlite.NewVideoPlaybackRepository(db)
	avatarRepo := sqlite.NewFriendAvatarObservationRepository(db)
	worldRepo := sqlite.NewWorldInfoRepository(db)
	uc := usecase.NewActivityUseCase(playRepo, encRepo, settings, nil, worldRepo).
		WithVideoPlaybackRepo(videoRepo).
		WithFriendAvatarObservationRepo(avatarRepo)

	const logSrc = `C:\VRChat\output_log.txt`
	now := time.Now().UTC().Truncate(time.Second)
	oldAt := now.AddDate(0, 0, -30)
	freshAt := now.Add(-2 * time.Hour)
	// Two hours before the 7-day cutoff, written with a +09 offset so the text
	// sorts after the UTC cutoff string.
	offsetExpired := now.AddDate(0, 0, -7).Add(-2 * time.Hour).In(time.FixedZone("JST", 9*60*60))
	// Two hours after the cutoff, written with a -14 offset so the text sorts
	// before the UTC cutoff string.
	offsetFresh := now.AddDate(0, 0, -7).Add(2 * time.Hour).In(time.FixedZone("WEST", -14*60*60))

	apply := func(cmd any) {
		t.Helper()
		if err := uc.ApplyCommand(ctx, logSrc, cmd); err != nil {
			t.Fatalf("apply %#v: %v", cmd, err)
		}
	}
	seed := func() {
		t.Helper()
		apply(activity.RecordEncounterJoinCmd{
			VRCUserID: "vrc-old", DisplayName: "Old", InstanceID: "inst-old", WorldID: "wrld_old", At: oldAt,
		})
		apply(activity.RecordEncounterJoinCmd{
			VRCUserID: "vrc-fresh", DisplayName: "Fresh", InstanceID: "inst-fresh", WorldID: "wrld_fresh", At: freshAt,
		})
		apply(activity.StartPlaySessionCmd{InstanceID: "inst-old", At: oldAt})
		apply(activity.StartPlaySessionCmd{InstanceID: "inst-fresh", At: freshAt})
		apply(activity.RecordVideoPlaybackAttemptCmd{URL: "https://example.com/old", WorldID: "wrld_old", At: oldAt})
		apply(activity.RecordVideoPlaybackAttemptCmd{URL: "https://example.com/fresh", WorldID: "wrld_fresh", At: freshAt})
		apply(activity.RecordFriendAvatarSwitchCmd{
			VRCUserID: "vrc-old", DisplayName: "Old", AvatarName: "OldAv", At: oldAt,
		})
		apply(activity.RecordFriendAvatarSwitchCmd{
			VRCUserID: "vrc-fresh", DisplayName: "Fresh", AvatarName: "FreshAv", At: freshAt,
		})
		apply(activity.UpsertWorldVisitCmd{WorldID: "wrld_both", At: oldAt})
		apply(activity.UpsertWorldVisitCmd{WorldID: "wrld_both", At: freshAt})

		if err := encRepo.Save(ctx, &activity.UserEncounter{
			ID: "enc-offset-expired", VRCUserID: "vrc-off-old", DisplayName: "OffOld", JoinedAt: offsetExpired,
		}); err != nil {
			t.Fatal(err)
		}
		if err := encRepo.Save(ctx, &activity.UserEncounter{
			ID: "enc-offset-fresh", VRCUserID: "vrc-off-fresh", DisplayName: "OffFresh", JoinedAt: offsetFresh,
		}); err != nil {
			t.Fatal(err)
		}
		if err := playRepo.Save(ctx, &activity.PlaySession{
			ID: "play-offset-expired", StartTime: offsetExpired, InstanceID: "inst-offset-old",
		}); err != nil {
			t.Fatal(err)
		}
		if err := playRepo.Save(ctx, &activity.PlaySession{
			ID: "play-offset-fresh", StartTime: offsetFresh, InstanceID: "inst-offset-fresh",
		}); err != nil {
			t.Fatal(err)
		}
		if err := videoRepo.Save(ctx, &activity.VideoPlaybackAttempt{
			ID: "vid-offset-expired", AttemptedAt: offsetExpired, URL: "https://example.com/offset-old",
		}); err != nil {
			t.Fatal(err)
		}
		if err := videoRepo.Save(ctx, &activity.VideoPlaybackAttempt{
			ID: "vid-offset-fresh", AttemptedAt: offsetFresh, URL: "https://example.com/offset-fresh",
		}); err != nil {
			t.Fatal(err)
		}
		if err := avatarRepo.Save(ctx, &activity.FriendAvatarObservation{
			ID: "av-offset-expired", VRCUserID: "vrc-off-old", DisplayName: "OffOld", AvatarName: "OffOldAv", ObservedAt: offsetExpired,
		}); err != nil {
			t.Fatal(err)
		}
		if err := avatarRepo.Save(ctx, &activity.FriendAvatarObservation{
			ID: "av-offset-fresh", VRCUserID: "vrc-off-fresh", DisplayName: "OffFresh", AvatarName: "OffFreshAv", ObservedAt: offsetFresh,
		}); err != nil {
			t.Fatal(err)
		}
	}
	seed()

	if n := countQuery(t, db, `SELECT COUNT(*) FROM user_encounters`); n < 4 {
		t.Fatalf("seed encounters = %d, want at least 4", n)
	}

	if _, err := uc.RotateEncounters(ctx); err != nil {
		t.Fatal(err)
	}
	assertRetentionShape(t, db, worldRepo)

	// Replay the expired log events. They must not come back.
	apply(activity.RecordEncounterJoinCmd{
		VRCUserID: "vrc-old", DisplayName: "Old", InstanceID: "inst-old", WorldID: "wrld_old", At: oldAt,
	})
	apply(activity.RecordEncounterJoinCmd{
		VRCUserID: "vrc-off-old", DisplayName: "OffOld", At: offsetExpired,
	})
	apply(activity.StartPlaySessionCmd{InstanceID: "inst-old", At: oldAt})
	apply(activity.StartPlaySessionCmd{InstanceID: "inst-offset-old", At: offsetExpired})
	apply(activity.RecordVideoPlaybackAttemptCmd{URL: "https://example.com/old", At: oldAt})
	apply(activity.RecordVideoPlaybackAttemptCmd{URL: "https://example.com/offset-old", At: offsetExpired})
	apply(activity.RecordFriendAvatarSwitchCmd{
		VRCUserID: "vrc-old", DisplayName: "Old", AvatarName: "OldAv", At: oldAt,
	})
	apply(activity.RecordFriendAvatarSwitchCmd{
		VRCUserID: "vrc-off-old", DisplayName: "OffOld", AvatarName: "OffOldAv", At: offsetExpired,
	})
	apply(activity.UpsertWorldVisitCmd{WorldID: "wrld_both", At: oldAt})
	assertRetentionShape(t, db, worldRepo)

	wi, err := worldRepo.GetByWorldID(ctx, "wrld_both")
	if err != nil || wi == nil {
		t.Fatalf("world_info after reingest: %#v err=%v", wi, err)
	}
	if !wi.LastVisitedAt.Equal(freshAt) {
		t.Fatalf("world last_visited_at = %s, want fresh %s (old visit must not roll it back)", wi.LastVisitedAt, freshAt)
	}
}

func assertRetentionShape(t *testing.T, db *sql.DB, worldRepo *sqlite.WorldInfoRepository) {
	t.Helper()
	ctx := context.Background()
	if n := countQuery(t, db, `SELECT COUNT(*) FROM user_encounters WHERE vrc_user_id = 'vrc-old'`); n != 0 {
		t.Fatalf("expired encounters = %d", n)
	}
	if n := countQuery(t, db, `SELECT COUNT(*) FROM user_encounters WHERE id = 'enc-offset-expired' OR vrc_user_id = 'vrc-off-old'`); n != 0 {
		t.Fatalf("offset-expired encounter still stored (text compare kept a row that is before the cutoff, or re-ingest revived it)")
	}
	if n := countQuery(t, db, `SELECT COUNT(*) FROM user_encounters WHERE vrc_user_id = 'vrc-fresh'`); n != 1 {
		t.Fatalf("fresh encounters = %d, want 1", n)
	}
	if n := countQuery(t, db, `SELECT COUNT(*) FROM user_encounters WHERE id = 'enc-offset-fresh'`); n != 1 {
		t.Fatalf("offset-fresh encounter = %d, want 1 (text compare would drop a row still inside retention)", n)
	}

	if n := countQuery(t, db, `SELECT COUNT(*) FROM play_sessions WHERE instance_id = 'inst-old' OR instance_id = 'inst-offset-old'`); n != 0 {
		t.Fatalf("expired play sessions = %d", n)
	}
	if n := countQuery(t, db, `SELECT COUNT(*) FROM play_sessions WHERE instance_id = 'inst-fresh'`); n != 1 {
		t.Fatalf("fresh play sessions = %d, want 1", n)
	}
	if n := countQuery(t, db, `SELECT COUNT(*) FROM play_sessions WHERE id = 'play-offset-fresh'`); n != 1 {
		t.Fatalf("offset-fresh play session = %d, want 1", n)
	}

	if n := countQuery(t, db, `SELECT COUNT(*) FROM video_playback_history WHERE url IN ('https://example.com/old', 'https://example.com/offset-old')`); n != 0 {
		t.Fatalf("expired video rows = %d", n)
	}
	if n := countQuery(t, db, `SELECT COUNT(*) FROM video_playback_history WHERE url = 'https://example.com/fresh'`); n != 1 {
		t.Fatalf("fresh video rows = %d, want 1", n)
	}
	if n := countQuery(t, db, `SELECT COUNT(*) FROM video_playback_history WHERE id = 'vid-offset-fresh'`); n != 1 {
		t.Fatalf("offset-fresh video rows = %d, want 1", n)
	}

	if n := countQuery(t, db, `SELECT COUNT(*) FROM friend_avatar_observations WHERE vrc_user_id IN ('vrc-old', 'vrc-off-old') OR id = 'av-offset-expired'`); n != 0 {
		t.Fatalf("expired avatar observations = %d", n)
	}
	if n := countQuery(t, db, `SELECT COUNT(*) FROM friend_avatar_observations WHERE vrc_user_id = 'vrc-fresh'`); n != 1 {
		t.Fatalf("fresh avatar observations = %d, want 1", n)
	}
	if n := countQuery(t, db, `SELECT COUNT(*) FROM friend_avatar_observations WHERE id = 'av-offset-fresh'`); n != 1 {
		t.Fatalf("offset-fresh avatar observations = %d, want 1", n)
	}

	wi, err := worldRepo.GetByWorldID(ctx, "wrld_both")
	if err != nil || wi == nil {
		t.Fatalf("world_info is not part of activity retention deletion: %#v err=%v", wi, err)
	}
}

func openRetentionDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sqlite.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func countQuery(t *testing.T, db *sql.DB, query string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(query).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
