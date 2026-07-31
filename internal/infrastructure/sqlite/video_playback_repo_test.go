package sqlite

import (
	"context"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

func TestVideoPlaybackRepository_SaveListCompleteAndDelete(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVideoPlaybackRepository(db)
	worldRepo := NewWorldInfoRepository(db)
	ctx := context.Background()
	base := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	const logSrc = "/logs/output_log.txt"
	const worldID = "wrld_videotest-0000-0000-0000-000000000001"

	if upsertErr := worldRepo.UpsertDisplayName(ctx, worldID, "Video World", base); upsertErr != nil {
		t.Fatal(upsertErr)
	}

	if saveErr := repo.Save(ctx, &activity.VideoPlaybackAttempt{
		ID: "vp-fail", AttemptedAt: base, URL: "https://youtu.be/fail",
		WorldID: worldID, LogSourcePath: logSrc,
	}); saveErr != nil {
		t.Fatal(saveErr)
	}
	n, err := repo.CompleteFailure(ctx, logSrc, "https://youtu.be/fail", "format missing", base.Add(time.Minute))
	if err != nil || n != 1 {
		t.Fatalf("CompleteFailure n=%d err=%v", n, err)
	}
	n, err = repo.CompleteSuccess(ctx, logSrc, "https://youtu.be/fail", "https://cdn/x", base.Add(2*time.Minute))
	if err != nil || n != 0 {
		t.Fatalf("CompleteSuccess after failure n=%d err=%v, want 0 (no overwrite)", n, err)
	}

	if saveErr := repo.Save(ctx, &activity.VideoPlaybackAttempt{
		ID: "vp-ok", AttemptedAt: base.Add(time.Hour), URL: "https://youtu.be/ok",
		WorldID: worldID, LogSourcePath: logSrc,
	}); saveErr != nil {
		t.Fatal(saveErr)
	}
	n, err = repo.CompleteSuccess(ctx, logSrc, "https://youtu.be/ok", "https://cdn/ok", base.Add(2*time.Hour))
	if err != nil || n != 1 {
		t.Fatalf("CompleteSuccess n=%d err=%v", n, err)
	}

	if saveErr := repo.Save(ctx, &activity.VideoPlaybackAttempt{
		ID: "vp-open", AttemptedAt: base.Add(3 * time.Hour), URL: "https://other", LogSourcePath: logSrc,
	}); saveErr != nil {
		t.Fatal(saveErr)
	}
	n, err = repo.CompleteFailure(ctx, logSrc, "", "no-url-error", base.Add(4*time.Hour))
	if err != nil || n != 1 {
		t.Fatalf("CompleteFailure empty url n=%d err=%v", n, err)
	}

	list, err := repo.ListWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("list len=%d", len(list))
	}
	if list[0].Attempt.ID != "vp-open" {
		t.Fatalf("newest first want vp-open, got %s", list[0].Attempt.ID)
	}
	if list[0].Attempt.Outcome != activity.VideoPlaybackOutcomeFailure {
		t.Errorf("vp-open outcome = %q", list[0].Attempt.Outcome)
	}
	if list[1].Attempt.Outcome != activity.VideoPlaybackOutcomeSuccess {
		t.Errorf("vp-ok outcome = %q", list[1].Attempt.Outcome)
	}
	if list[2].WorldDisplayName != "Video World" {
		t.Errorf("world name = %q", list[2].WorldDisplayName)
	}
	if list[2].Attempt.FailureReason != "format missing" {
		t.Errorf("failure reason = %q", list[2].Attempt.FailureReason)
	}

	deleted, err := repo.DeleteOlderThan(ctx, base.Add(30*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 1 {
		t.Fatalf("DeleteOlderThan = %d, want 1 (vp-fail only)", deleted)
	}
}

func TestVideoPlaybackRepository_CompleteFailure_noOpenReturnsZero(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewVideoPlaybackRepository(db)
	n, err := repo.CompleteFailure(context.Background(), "/x", "https://a", "e", time.Now())
	if err != nil || n != 0 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}
