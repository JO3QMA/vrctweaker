package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"vrchat-tweaker/internal/domain/activity"
)

func TestPlaySessionRepository_Save_List_GetByID_Count(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if schemaErr := applySchema(db); schemaErr != nil {
		t.Fatal(schemaErr)
	}

	repo := NewPlaySessionRepository(db)
	t0 := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Hour)
	dur := 3600
	s0 := &activity.PlaySession{
		ID:          "ps1",
		StartTime:   t0,
		EndTime:     &t1,
		DurationSec: &dur,
	}
	if saveErr := repo.Save(ctx, s0); saveErr != nil {
		t.Fatal(saveErr)
	}

	n, err := repo.Count(ctx)
	if err != nil || n != 1 {
		t.Fatalf("Count: n=%d err=%v", n, err)
	}

	got, err := repo.GetByID(ctx, "ps1")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.ID != "ps1" || !got.StartTime.Equal(t0) {
		t.Fatalf("GetByID: %#v", got)
	}

	list, err := repo.List(ctx, t0.Add(-time.Minute), t0.Add(2*time.Hour))
	if err != nil || len(list) != 1 || list[0].ID != "ps1" {
		t.Fatalf("List: %v %#v", err, list)
	}
}

func TestPlaySessionRepository_FindLatestWithoutEndTime(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if schemaErr := applySchema(db); schemaErr != nil {
		t.Fatal(schemaErr)
	}

	repo := NewPlaySessionRepository(db)
	t0 := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	if saveErr := repo.Save(ctx, &activity.PlaySession{ID: "open", StartTime: t0}); saveErr != nil {
		t.Fatal(saveErr)
	}

	got, err := repo.FindLatestWithoutEndTime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.ID != "open" {
		t.Fatalf("FindLatestWithoutEndTime: %#v", got)
	}
}
