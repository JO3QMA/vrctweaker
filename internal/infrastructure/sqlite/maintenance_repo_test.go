package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/domain/media"
)

func TestMaintenance_ClearEncounters_Integration(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err := migrate(db); err != nil {
		t.Fatal(err)
	}

	encounterRepo := NewUserEncounterRepository(db)
	ctx := context.Background()

	// Insert encounters
	e := &activity.UserEncounter{
		ID:            "enc-1",
		VRCUserID:     "usr_xxx",
		DisplayName:   "TestUser",
		Action:        "join",
		InstanceID:    "inst_yyy",
		EncounteredAt: time.Now().UTC(),
	}
	if err := encounterRepo.Save(ctx, e); err != nil {
		t.Fatal(err)
	}

	list, err := encounterRepo.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 encounter before clear, got %d", len(list))
	}

	n, err := encounterRepo.DeleteAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("expected RowsAffected=1, got %d", n)
	}

	list, err = encounterRepo.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 encounters after clear, got %d", len(list))
	}
}

func TestMaintenance_ClearScreenshots_Integration(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err := migrate(db); err != nil {
		t.Fatal(err)
	}

	screenshotRepo := NewScreenshotRepository(db)
	ctx := context.Background()

	s := &media.Screenshot{
		ID:        "scr-1",
		FilePath:  "/path/to/screenshot.png",
		WorldID:   "wrld_xxx",
		WorldName: "Test World",
	}
	if err := screenshotRepo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}

	list, err := screenshotRepo.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 screenshot before clear, got %d", len(list))
	}

	n, err := screenshotRepo.DeleteAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("expected RowsAffected=1, got %d", n)
	}

	list, err = screenshotRepo.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 screenshots after clear, got %d", len(list))
	}
}

func TestMaintenance_ClearFriendsCache_Integration(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err := migrate(db); err != nil {
		t.Fatal(err)
	}

	friendRepo := NewFriendCacheRepository(db)
	ctx := context.Background()

	f := &identity.FriendCache{
		VRCUserID:   "usr_xxx",
		DisplayName: "FriendUser",
		Status:      "active",
		IsFavorite:  false,
		LastUpdated: time.Now().UTC(),
	}
	if err := friendRepo.Save(ctx, f); err != nil {
		t.Fatal(err)
	}

	list, err := friendRepo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 friend before clear, got %d", len(list))
	}

	n, err := friendRepo.DeleteAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("expected RowsAffected=1, got %d", n)
	}

	list, err = friendRepo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 friends after clear, got %d", len(list))
	}
}

func TestMaintenance_Vacuum_Integration(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err := migrate(db); err != nil {
		t.Fatal(err)
	}

	maintenanceRepo := NewMaintenanceRepository(db)
	ctx := context.Background()

	if err := maintenanceRepo.Vacuum(ctx); err != nil {
		t.Fatal(err)
	}
}
