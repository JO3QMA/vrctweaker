package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestAppSettingsRepository_GetSetGetAll(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if schemaErr := applySchema(db); schemaErr != nil {
		t.Fatal(schemaErr)
	}

	repo := NewAppSettingsRepository(db)
	ctx := context.Background()

	missing, err := repo.Get(ctx, "no_such_key")
	if err != nil {
		t.Fatal(err)
	}
	if missing != "" {
		t.Fatalf("missing key: got %q", missing)
	}

	if setErr := repo.Set(ctx, "theme", "dark"); setErr != nil {
		t.Fatal(setErr)
	}
	if setErr := repo.Set(ctx, "theme", "light"); setErr != nil {
		t.Fatal(setErr)
	}

	got, err := repo.Get(ctx, "theme")
	if err != nil || got != "light" {
		t.Fatalf("Get theme: %q err=%v", got, err)
	}

	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if all["theme"] != "light" {
		t.Fatalf("GetAll theme: %q", all["theme"])
	}
	if all["log_retention_days"] != "30" {
		t.Fatalf("seeded log_retention_days: %q", all["log_retention_days"])
	}
}
