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
	if migrateErr := migrate(db); migrateErr != nil {
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
