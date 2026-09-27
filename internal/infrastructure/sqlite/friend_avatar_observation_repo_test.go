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

	for _, row := range []activity.FriendAvatarObservation{
		{ID: "1", VRCUserID: "usr_a", DisplayName: "Alpha", AvatarName: "Av1", ObservedAt: t0},
		{ID: "2", VRCUserID: "usr_a", DisplayName: "Alpha", AvatarName: "Av1", ObservedAt: t1},
		{ID: "3", VRCUserID: "", DisplayName: "Alpha", AvatarName: "Av2", ObservedAt: t1},
	} {
		if err := repo.Save(ctx, &row); err != nil {
			t.Fatal(err)
		}
	}

	sums, err := repo.ListUsageSummariesByVRCUserID(ctx, "usr_a", "Alpha")
	if err != nil {
		t.Fatal(err)
	}
	if len(sums) != 2 {
		t.Fatalf("summaries = %d", len(sums))
	}
	if sums[0].AvatarName != "Av2" || sums[0].UseCount != 1 {
		t.Fatalf("newest first: %+v", sums[0])
	}
	if sums[1].AvatarName != "Av1" || sums[1].UseCount != 2 {
		t.Fatalf("older: %+v", sums[1])
	}
}
