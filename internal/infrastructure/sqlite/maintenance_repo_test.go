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
	if migrateErr := applySchema(db); migrateErr != nil {
		t.Fatal(migrateErr)
	}

	encounterRepo := NewUserEncounterRepository(db)
	ctx := context.Background()

	// Insert encounters
	e := &activity.UserEncounter{
		ID:          "enc-1",
		VRCUserID:   "usr_xxx",
		DisplayName: "TestUser",
		InstanceID:  "inst_yyy",
		WorldID:     "wrld_z",
		JoinedAt:    time.Now().UTC(),
		LeftAt:      nil,
	}
	if saveErr := encounterRepo.Save(ctx, e); saveErr != nil {
		t.Fatal(saveErr)
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
	if migrateErr := applySchema(db); migrateErr != nil {
		t.Fatal(migrateErr)
	}

	screenshotRepo := NewScreenshotRepository(db)
	ctx := context.Background()

	s := &media.Screenshot{
		ID:       "scr-1",
		FilePath: "/path/to/screenshot.png",
		WorldID:  "wrld_xxx",
	}
	if saveErr := screenshotRepo.Save(ctx, s); saveErr != nil {
		t.Fatal(saveErr)
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
	if migrateErr := applySchema(db); migrateErr != nil {
		t.Fatal(migrateErr)
	}

	userRepo := NewUserCacheRepository(db)
	ctx := context.Background()

	f := &identity.UserCache{
		VRCUserID:   "usr_xxx",
		DisplayName: "FriendUser",
		Status:      "active",
		UserKind:    identity.UserKindFriend,
		IsFavorite:  false,
		LastUpdated: time.Now().UTC(),
	}
	if saveErr := userRepo.Save(ctx, f); saveErr != nil {
		t.Fatal(saveErr)
	}

	list, err := userRepo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 friend before clear, got %d", len(list))
	}

	n, err := userRepo.DeleteAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("expected RowsAffected=1, got %d", n)
	}

	list, err = userRepo.List(ctx)
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
	if err := applySchema(db); err != nil {
		t.Fatal(err)
	}

	maintenanceRepo := NewMaintenanceRepository(db)
	ctx := context.Background()

	if err := maintenanceRepo.Vacuum(ctx); err != nil {
		t.Fatal(err)
	}
}
