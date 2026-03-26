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
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "t.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if schemaErr := applySchema(db); schemaErr != nil {
		t.Fatal(schemaErr)
	}

	repo := NewUserEncounterRepository(db)
	t0 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Hour)
	t2 := t0.Add(2 * time.Hour)
	lt1 := t0.Add(30 * time.Minute)
	rows := []activity.UserEncounter{
		{ID: "e1", VRCUserID: "usr_a", DisplayName: "A", InstanceID: "i1", WorldID: "wrld_1", JoinedAt: t0, LeftAt: &lt1},
		{ID: "e2", VRCUserID: "usr_b", DisplayName: "B", InstanceID: "i2", WorldID: "wrld_1", JoinedAt: t1, LeftAt: nil},
		{ID: "e3", VRCUserID: "usr_a", DisplayName: "A", InstanceID: "i3", WorldID: "wrld_2", JoinedAt: t2, LeftAt: nil},
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
		t.Fatalf("order/want e3,e1 got %#v", list)
	}
}

func TestUserEncounterRepository_List_FilterWorldID(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "t.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if schemaErr := applySchema(db); schemaErr != nil {
		t.Fatal(schemaErr)
	}

	repo := NewUserEncounterRepository(db)
	t0 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	lt := t0.Add(time.Minute)
	rows := []activity.UserEncounter{
		{ID: "e1", VRCUserID: "usr_a", DisplayName: "A", InstanceID: "i1", WorldID: "wrld_x", JoinedAt: t0, LeftAt: &lt},
		{ID: "e2", VRCUserID: "usr_b", DisplayName: "B", InstanceID: "i2", WorldID: "wrld_y", JoinedAt: t0.Add(time.Hour), LeftAt: nil},
		{ID: "e3", VRCUserID: "usr_a", DisplayName: "A", InstanceID: "i3", WorldID: "wrld_x", JoinedAt: t0.Add(2 * time.Hour), LeftAt: nil},
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
}

func TestUserEncounterRepository_ListWithContext_FilterWorldID(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "t.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if schemaErr := applySchema(db); schemaErr != nil {
		t.Fatal(schemaErr)
	}

	repo := NewUserEncounterRepository(db)
	t0 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	rows := []activity.UserEncounter{
		{ID: "e1", VRCUserID: "usr_a", DisplayName: "A", InstanceID: "i1", WorldID: "wrld_only", JoinedAt: t0, LeftAt: nil},
		{ID: "e2", VRCUserID: "usr_b", DisplayName: "B", InstanceID: "i2", WorldID: "wrld_other", JoinedAt: t0.Add(time.Hour), LeftAt: nil},
	}
	for i := range rows {
		if saveErr := repo.Save(ctx, &rows[i]); saveErr != nil {
			t.Fatal(saveErr)
		}
	}

	list, err := repo.ListWithContext(ctx, &activity.EncounterFilter{WorldID: "wrld_only"})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("ListWithContext: got %d rows, want 1", len(list))
	}
	if list[0].Encounter.ID != "e1" {
		t.Fatalf("got encounter id %q, want e1", list[0].Encounter.ID)
	}
}

func TestUserEncounterRepository_BackfillMissingWorldContext(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "t.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if schemaErr := applySchema(db); schemaErr != nil {
		t.Fatal(schemaErr)
	}

	repo := NewUserEncounterRepository(db)
	t0 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	rows := []activity.UserEncounter{
		{ID: "e1", VRCUserID: "usr_a", DisplayName: "A", InstanceID: "inst_full", WorldID: "wrld_fill", JoinedAt: t0, LeftAt: nil},
		{ID: "e2", VRCUserID: "usr_b", DisplayName: "B", InstanceID: "", WorldID: "", JoinedAt: t0.Add(time.Minute), LeftAt: nil},
		{ID: "e3", VRCUserID: "usr_c", DisplayName: "C", InstanceID: "", WorldID: "", JoinedAt: t0.Add(2 * time.Minute), LeftAt: nil},
	}
	for i := range rows {
		if saveErr := repo.Save(ctx, &rows[i]); saveErr != nil {
			t.Fatal(saveErr)
		}
	}

	n, err := repo.BackfillMissingWorldContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("Backfill updated %d rows, want 2", n)
	}

	list, err := repo.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	byID := make(map[string]*activity.UserEncounter, len(list))
	for _, e := range list {
		byID[e.ID] = e
	}
	for _, id := range []string{"e2", "e3"} {
		e := byID[id]
		if e == nil {
			t.Fatalf("missing encounter %s", id)
		}
		if e.WorldID != "wrld_fill" || e.InstanceID != "inst_full" {
			t.Fatalf("encounter %s: world_id=%q instance_id=%q", id, e.WorldID, e.InstanceID)
		}
	}
}

func TestUserEncounterRepository_CloseEncounterLeave(t *testing.T) {
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
	repo := NewUserEncounterRepository(db)
	t0 := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	if saveErr := repo.Save(ctx, &activity.UserEncounter{
		ID: "x1", VRCUserID: "usr_x", DisplayName: "X", InstanceID: "inst", WorldID: "wrld_w",
		JoinedAt: t0, LeftAt: nil,
	}); saveErr != nil {
		t.Fatal(saveErr)
	}
	left := t0.Add(time.Hour)
	n, err := repo.CloseEncounterLeave(ctx, "usr_x", left)
	if err != nil || n != 1 {
		t.Fatalf("CloseEncounterLeave: n=%d err=%v", n, err)
	}
	list, _ := repo.List(ctx, nil)
	if len(list) != 1 || list[0].LeftAt == nil || !list[0].LeftAt.Equal(left) {
		t.Fatalf("after close: %+v", list)
	}
}

func TestUserEncounterRepository_CloseOpenEncountersAt(t *testing.T) {
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
	repo := NewUserEncounterRepository(db)
	t0 := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	for _, id := range []string{"a", "b"} {
		if saveErr := repo.Save(ctx, &activity.UserEncounter{
			ID: id, VRCUserID: "usr_" + id, DisplayName: id, InstanceID: "i", WorldID: "w",
			JoinedAt: t0, LeftAt: nil,
		}); saveErr != nil {
			t.Fatal(saveErr)
		}
	}
	at := t0.Add(2 * time.Hour)
	n, err := repo.CloseOpenEncountersAt(ctx, at)
	if err != nil || n != 2 {
		t.Fatalf("CloseOpenEncountersAt: n=%d err=%v", n, err)
	}
	list, _ := repo.List(ctx, nil)
	for _, e := range list {
		if e.LeftAt == nil || !e.LeftAt.Equal(at) {
			t.Fatalf("row %s not closed: %+v", e.ID, e.LeftAt)
		}
	}
}
