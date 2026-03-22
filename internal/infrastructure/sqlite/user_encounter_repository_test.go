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

func TestUserEncounterRepository_List_FilterVRCUserID(t *testing.T) {
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

	repo := NewUserEncounterRepository(db)
	ctx := context.Background()
	t0 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	rows := []activity.UserEncounter{
		{ID: "e1", VRCUserID: "usr_a", DisplayName: "A", Action: "join", InstanceID: "i1", WorldID: "wrld_1", EncounteredAt: t0},
		{ID: "e2", VRCUserID: "usr_b", DisplayName: "B", Action: "join", InstanceID: "i2", WorldID: "wrld_1", EncounteredAt: t0.Add(time.Hour)},
		{ID: "e3", VRCUserID: "usr_a", DisplayName: "A", Action: "leave", InstanceID: "i3", WorldID: "wrld_2", EncounteredAt: t0.Add(2 * time.Hour)},
	}
	for i := range rows {
		if saveErr := repo.Save(ctx, &rows[i]); saveErr != nil {
			t.Fatal(saveErr)
		}
	}

	list, err := repo.List(ctx, &activity.EncounterFilter{VRCUserID: "usr_a"})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("List VRCUserID=usr_a: got %d rows, want 2", len(list))
	}
	if list[0].ID != "e3" || list[1].ID != "e1" {
		t.Fatalf("List order: got ids %q, %q, want e3, e1 (desc by time)", list[0].ID, list[1].ID)
	}
}

func TestUserEncounterRepository_List_FilterWorldID(t *testing.T) {
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

	repo := NewUserEncounterRepository(db)
	ctx := context.Background()
	t0 := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	rows := []activity.UserEncounter{
		{ID: "e1", VRCUserID: "usr_a", DisplayName: "A", Action: "join", InstanceID: "i1", WorldID: "wrld_x", EncounteredAt: t0},
		{ID: "e2", VRCUserID: "usr_b", DisplayName: "B", Action: "join", InstanceID: "i2", WorldID: "wrld_y", EncounteredAt: t0.Add(time.Hour)},
		{ID: "e3", VRCUserID: "usr_a", DisplayName: "A", Action: "leave", InstanceID: "i3", WorldID: "wrld_x", EncounteredAt: t0.Add(2 * time.Hour)},
	}
	for i := range rows {
		if saveErr := repo.Save(ctx, &rows[i]); saveErr != nil {
			t.Fatal(saveErr)
		}
	}

	list, err := repo.List(ctx, &activity.EncounterFilter{WorldID: "wrld_x"})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("List WorldID=wrld_x: got %d rows, want 2", len(list))
	}
	for _, e := range list {
		if e.WorldID != "wrld_x" {
			t.Errorf("row %s world_id = %q, want wrld_x", e.ID, e.WorldID)
		}
	}
}

func TestUserEncounterRepository_ListWithContext_FilterWorldID(t *testing.T) {
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

	repo := NewUserEncounterRepository(db)
	ctx := context.Background()
	t0 := time.Date(2025, 1, 3, 12, 0, 0, 0, time.UTC)

	rows := []activity.UserEncounter{
		{ID: "e1", VRCUserID: "usr_a", DisplayName: "A", Action: "join", InstanceID: "i1", WorldID: "wrld_only", EncounteredAt: t0},
		{ID: "e2", VRCUserID: "usr_b", DisplayName: "B", Action: "join", InstanceID: "i2", WorldID: "wrld_other", EncounteredAt: t0.Add(time.Hour)},
	}
	for i := range rows {
		if saveErr := repo.Save(ctx, &rows[i]); saveErr != nil {
			t.Fatal(err)
		}
	}

	list, err := repo.ListWithContext(ctx, &activity.EncounterFilter{WorldID: "wrld_only"})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("ListWithContext WorldID=wrld_only: got %d rows, want 1", len(list))
	}
	if list[0].Encounter.ID != "e1" {
		t.Fatalf("got encounter id %q, want e1", list[0].Encounter.ID)
	}
}
