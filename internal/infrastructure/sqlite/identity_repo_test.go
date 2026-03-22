package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestUserCacheRepository_UpsertFromLog_LastContactNoRegression(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, dbErr := sql.Open("sqlite", dbPath)
	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer func() { _ = db.Close() }()
	if migrateErr := migrate(db); migrateErr != nil {
		t.Fatal(migrateErr)
	}

	repo := NewUserCacheRepository(db)
	ctx := context.Background()
	const vrcID = "usr_lc_regress_test"

	t1 := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

	err := repo.UpsertFromLog(ctx, vrcID, "UserA", t1)
	if err != nil {
		t.Fatal(err)
	}
	u, err := repo.GetByVRCUserID(ctx, vrcID)
	if err != nil {
		t.Fatal(err)
	}
	if u.LastContactAt == nil || !u.LastContactAt.Equal(t1) {
		t.Fatalf("after first upsert LastContactAt = %v, want %v", u.LastContactAt, t1)
	}

	err = repo.UpsertFromLog(ctx, vrcID, "UserA", t3)
	if err != nil {
		t.Fatal(err)
	}
	u, err = repo.GetByVRCUserID(ctx, vrcID)
	if err != nil {
		t.Fatal(err)
	}
	if u.LastContactAt == nil || !u.LastContactAt.Equal(t3) {
		t.Fatalf("after newer upsert LastContactAt = %v, want %v", u.LastContactAt, t3)
	}

	err = repo.UpsertFromLog(ctx, vrcID, "UserA", t2)
	if err != nil {
		t.Fatal(err)
	}
	u, err = repo.GetByVRCUserID(ctx, vrcID)
	if err != nil {
		t.Fatal(err)
	}
	if u.LastContactAt == nil || !u.LastContactAt.Equal(t3) {
		t.Fatalf("after older upsert LastContactAt = %v, want %v (no regression)", u.LastContactAt, t3)
	}
}
