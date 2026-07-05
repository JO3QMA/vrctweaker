package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
	"vrchat-tweaker/internal/domain/launcher"
)

func TestLauncherProfileRepository_CRUD(t *testing.T) {
	// avoid circular import - use raw sql.DB
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
	repo := NewLauncherProfileRepository(db)
	ctx := context.Background()

	p := &launcher.LaunchProfile{
		ID:        "test-1",
		Name:      "Desktop",
		Arguments: "--no-vr",
		IsDefault: true,
	}
	if saveErr := repo.Save(ctx, p); saveErr != nil {
		t.Fatal(saveErr)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) < 1 {
		t.Error("expected at least 1 profile")
	}

	got, err := repo.GetByID(ctx, "test-1")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Name != "Desktop" {
		t.Errorf("got = %v", got)
	}
}

func TestLauncherProfileRepository_GetDefault_and_Delete(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if migrateErr := applySchema(db); migrateErr != nil {
		t.Fatal(migrateErr)
	}
	repo := NewLauncherProfileRepository(db)
	ctx := context.Background()

	def, err := repo.GetDefault(ctx)
	if err != nil || def == nil || def.ID != "default-desktop" {
		t.Fatalf("GetDefault: %#v err=%v", def, err)
	}

	p := &launcher.LaunchProfile{ID: "temp", Name: "Temp", Arguments: ""}
	if saveErr := repo.Save(ctx, p); saveErr != nil {
		t.Fatal(saveErr)
	}
	if delErr := repo.Delete(ctx, "temp"); delErr != nil {
		t.Fatal(delErr)
	}
	if delErr := repo.Delete(ctx, "no-such"); delErr == nil {
		t.Fatal("expected error deleting missing profile")
	}

	miss, err := repo.GetByID(ctx, "no-such")
	if err != nil || miss != nil {
		t.Fatalf("GetByID missing: %#v err=%v", miss, err)
	}
}
