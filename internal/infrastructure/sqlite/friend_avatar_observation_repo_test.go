package sqlite

import (
	"context"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

func TestFriendAvatarObservationRepository_saveAndSummarize(t *testing.T) {
	db := openTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	if err := applySchema(db); err != nil {
		t.Fatal(err)
	}
	repo := NewFriendAvatarObservationRepository(db)
	ctx := context.Background()
	t0 := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	t1 := time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC)

	logPath := "/logs/output_log.txt"
	for _, row := range []activity.FriendAvatarObservation{
		{ID: "1", VRCUserID: "usr_a", DisplayName: "Alpha", AvatarName: "Av1", LogSourcePath: logPath, ObservedAt: t0},
		{ID: "2", VRCUserID: "usr_a", DisplayName: "Alpha", AvatarName: "Av1", LogSourcePath: logPath, ObservedAt: t1},
		{ID: "3", VRCUserID: "", DisplayName: "Alpha", AvatarName: "Av2", LogSourcePath: logPath, ObservedAt: t1},
	} {
		if err := repo.Save(ctx, &row); err != nil {
			t.Fatal(err)
		}
	}

	dup := activity.FriendAvatarObservation{
		ID: "dup", VRCUserID: "usr_a", DisplayName: "Alpha", AvatarName: "Av1",
		LogSourcePath: logPath, ObservedAt: t0,
	}
	if err := repo.Save(ctx, &dup); err != nil {
		t.Fatal(err)
	}

	sums, err := repo.ListUsageSummariesByVRCUserID(ctx, "usr_a")
	if err != nil {
		t.Fatal(err)
	}
	if len(sums) != 1 {
		t.Fatalf("summaries = %d", len(sums))
	}
	if sums[0].AvatarName != "Av1" || sums[0].UseCount != 2 {
		t.Fatalf("resolved rows only: %+v", sums[0])
	}
}

func TestFriendAvatarObservationRepository_listSkipsCorruptObservedAt(t *testing.T) {
	db := openTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	if err := applySchema(db); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	tGood := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	_, err := db.ExecContext(ctx, `INSERT INTO friend_avatar_observations (
		id, vrc_user_id, display_name, avatar_name, instance_id, world_id, log_source_path, observed_at
	) VALUES (?, ?, ?, ?, '', '', '/log.txt', ?), (?, ?, ?, ?, '', '', '/log.txt', ?)`,
		"good", "usr_a", "A", "AvOK", tGood.Format(time.RFC3339),
		"bad", "usr_a", "A", "AvBad", "not-rfc3339",
	)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewFriendAvatarObservationRepository(db)
	sums, err := repo.ListUsageSummariesByVRCUserID(ctx, "usr_a")
	if err != nil {
		t.Fatal(err)
	}
	if len(sums) != 1 || sums[0].AvatarName != "AvOK" {
		t.Fatalf("want only valid timestamp row: %+v", sums)
	}
}
