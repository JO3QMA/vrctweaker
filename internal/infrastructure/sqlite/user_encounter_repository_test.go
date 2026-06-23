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
	n, err := repo.CloseEncounterLeave(ctx, "usr_x", "", left)
	if err != nil || n != 1 {
		t.Fatalf("CloseEncounterLeave: n=%d err=%v", n, err)
	}
	list, _ := repo.List(ctx, nil)
	if len(list) != 1 || list[0].LeftAt == nil || !list[0].LeftAt.Equal(left) {
		t.Fatalf("after close: %+v", list)
	}
}

func TestUserEncounterRepository_List_AllFilters(t *testing.T) {
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
	t0 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Hour)
	rows := []activity.UserEncounter{
		{ID: "e1", VRCUserID: "usr_a", DisplayName: "Alice", InstanceID: "inst_1", WorldID: "wrld_x", JoinedAt: t0, LeftAt: nil},
		{ID: "e2", VRCUserID: "usr_b", DisplayName: "Bob", InstanceID: "inst_2", WorldID: "wrld_y", JoinedAt: t1, LeftAt: nil},
		{ID: "e3", VRCUserID: "usr_c", DisplayName: "Alice2", InstanceID: "inst_1", WorldID: "wrld_x", JoinedAt: t1.Add(15 * time.Minute), LeftAt: nil},
	}
	for i := range rows {
		if saveErr := repo.Save(ctx, &rows[i]); saveErr != nil {
			t.Fatal(saveErr)
		}
	}

	from := t0.Add(30 * time.Minute)
	to := t1.Add(30 * time.Minute)
	list, err := repo.List(ctx, &activity.EncounterFilter{
		DisplayName: "Ali",
		InstanceID:  "inst_1",
		From:        &from,
		To:          &to,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != "e3" {
		t.Fatalf("filtered list: %#v", list)
	}
}

func TestUserEncounterRepository_ListWithContext_joinsWorldAndUserCache(t *testing.T) {
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

	worldRepo := NewWorldInfoRepository(db)
	userRepo := NewUserCacheRepository(db)
	encRepo := NewUserEncounterRepository(db)

	t0 := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)
	if dnErr := worldRepo.UpsertDisplayName(ctx, "wrld_ctx", "Context World", t0); dnErr != nil {
		t.Fatal(dnErr)
	}
	if saveErr := userRepo.Save(ctx, &identity.UserCache{
		VRCUserID:     "usr_ctx",
		DisplayName:   "CtxUser",
		Status:        "active",
		UserKind:      identity.UserKindFriend,
		LastUpdated:   t0,
		FirstSeenAt:   &t0,
		LastContactAt: &t0,
	}); saveErr != nil {
		t.Fatal(saveErr)
	}
	left := t0.Add(time.Hour)
	if encErr := encRepo.Save(ctx, &activity.UserEncounter{
		ID: "enc_ctx", VRCUserID: "usr_ctx", DisplayName: "CtxUser",
		InstanceID: "inst_ctx", WorldID: "wrld_ctx", JoinedAt: t0, LeftAt: &left,
	}); encErr != nil {
		t.Fatal(encErr)
	}

	list, err := encRepo.ListWithContext(ctx, &activity.EncounterFilter{
		VRCUserID:   "usr_ctx",
		DisplayName: "Ctx",
		InstanceID:  "inst_ctx",
		WorldID:     "wrld_ctx",
		From:        &t0,
		To:          &t0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("len=%d", len(list))
	}
	row := list[0]
	if row.WorldDisplayName != "Context World" {
		t.Fatalf("WorldDisplayName=%q", row.WorldDisplayName)
	}
	if row.UserFirstSeenAt == nil || !row.UserFirstSeenAt.Equal(t0) {
		t.Fatalf("UserFirstSeenAt=%v", row.UserFirstSeenAt)
	}
	if row.UserLastContactAt == nil || !row.UserLastContactAt.Equal(t0) {
		t.Fatalf("UserLastContactAt=%v", row.UserLastContactAt)
	}
	if !row.IsFirstEncounter {
		t.Fatal("expected IsFirstEncounter")
	}
}

func TestUserEncounterRepository_ListWithContext_firstEncounterWithinOneSecond(t *testing.T) {
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

	userRepo := NewUserCacheRepository(db)
	encRepo := NewUserEncounterRepository(db)
	t0 := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)
	firstSeen := t0
	joinedAt := t0.Add(500 * time.Millisecond)
	if saveErr := userRepo.Save(ctx, &identity.UserCache{
		VRCUserID: "usr_edge", DisplayName: "Edge", Status: "active",
		UserKind: identity.UserKindFriend, LastUpdated: t0,
		FirstSeenAt: &firstSeen, LastContactAt: &t0,
	}); saveErr != nil {
		t.Fatal(saveErr)
	}
	if encErr := encRepo.Save(ctx, &activity.UserEncounter{
		ID: "enc_edge", VRCUserID: "usr_edge", DisplayName: "Edge",
		InstanceID: "i", WorldID: "w", JoinedAt: joinedAt,
	}); encErr != nil {
		t.Fatal(encErr)
	}

	list, err := encRepo.ListWithContext(ctx, nil)
	if err != nil || len(list) != 1 {
		t.Fatalf("ListWithContext: len=%d err=%v", len(list), err)
	}
	if !list[0].IsFirstEncounter {
		t.Fatal("expected IsFirstEncounter within one second")
	}
}

func TestUserEncounterRepository_List_nilFilter(t *testing.T) {
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
	t0 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if saveErr := repo.Save(ctx, &activity.UserEncounter{
		ID: "only", VRCUserID: "u", DisplayName: "U", InstanceID: "i", WorldID: "w", JoinedAt: t0,
	}); saveErr != nil {
		t.Fatal(saveErr)
	}
	list, err := repo.List(ctx, nil)
	if err != nil || len(list) != 1 {
		t.Fatalf("List nil filter: len=%d err=%v", len(list), err)
	}
}

func TestUserEncounterRepository_DeleteOlderThan_and_Count(t *testing.T) {
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
	old := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	newT := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	for _, spec := range []struct {
		id string
		at time.Time
	}{
		{"old", old},
		{"new", newT},
	} {
		if saveErr := repo.Save(ctx, &activity.UserEncounter{
			ID: spec.id, VRCUserID: "u", DisplayName: "U", InstanceID: "i", WorldID: "w",
			JoinedAt: spec.at,
		}); saveErr != nil {
			t.Fatal(saveErr)
		}
	}

	n, err := repo.Count(ctx)
	if err != nil || n != 2 {
		t.Fatalf("Count: n=%d err=%v", n, err)
	}

	cutoff := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	deleted, err := repo.DeleteOlderThan(ctx, cutoff)
	if err != nil || deleted != 1 {
		t.Fatalf("DeleteOlderThan: deleted=%d err=%v", deleted, err)
	}

	n, err = repo.Count(ctx)
	if err != nil || n != 1 {
		t.Fatalf("Count after delete: n=%d err=%v", n, err)
	}

	allDeleted, err := repo.DeleteAll(ctx)
	if err != nil || allDeleted != 1 {
		t.Fatalf("DeleteAll: n=%d err=%v", allDeleted, err)
	}
}

func TestUserEncounterRepository_BackfillMissingWorldContext_noAnchor(t *testing.T) {
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
	t0 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if saveErr := repo.Save(ctx, &activity.UserEncounter{
		ID: "orphan", VRCUserID: "u", DisplayName: "U", JoinedAt: t0,
	}); saveErr != nil {
		t.Fatal(saveErr)
	}

	n, err := repo.BackfillMissingWorldContext(ctx)
	if err != nil || n != 0 {
		t.Fatalf("Backfill with no anchor: n=%d err=%v", n, err)
	}
}

func TestUserEncounterRepository_BackfillMissingWorldContext_preservesRowInstance(t *testing.T) {
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
	t0 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	rows := []activity.UserEncounter{
		{ID: "e1", VRCUserID: "a", DisplayName: "A", InstanceID: "inst_anchor", WorldID: "wrld_anchor", JoinedAt: t0},
		{ID: "e2", VRCUserID: "b", DisplayName: "B", InstanceID: "inst_own", WorldID: "", JoinedAt: t0.Add(time.Minute)},
	}
	for i := range rows {
		if saveErr := repo.Save(ctx, &rows[i]); saveErr != nil {
			t.Fatal(saveErr)
		}
	}

	n, err := repo.BackfillMissingWorldContext(ctx)
	if err != nil || n != 1 {
		t.Fatalf("Backfill: n=%d err=%v", n, err)
	}
	list, err := repo.List(ctx, &activity.EncounterFilter{VRCUserID: "b"})
	if err != nil || len(list) != 1 {
		t.Fatalf("List: %#v err=%v", list, err)
	}
	if list[0].WorldID != "wrld_anchor" || list[0].InstanceID != "inst_own" {
		t.Fatalf("backfill row: world=%q inst=%q", list[0].WorldID, list[0].InstanceID)
	}
}

func TestUserEncounterRepository_CloseEncounterLeave_noMatch(t *testing.T) {
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
	n, err := repo.CloseEncounterLeave(ctx, "nobody", "", time.Now().UTC())
	if err != nil || n != 0 {
		t.Fatalf("CloseEncounterLeave no match: n=%d err=%v", n, err)
	}
}

func TestUserEncounterRepository_DeleteOlderThan_noneRemoved(t *testing.T) {
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
	t0 := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	if saveErr := repo.Save(ctx, &activity.UserEncounter{
		ID: "keep", VRCUserID: "u", DisplayName: "U", InstanceID: "i", WorldID: "w", JoinedAt: t0,
	}); saveErr != nil {
		t.Fatal(saveErr)
	}
	n, err := repo.DeleteOlderThan(ctx, t0.Add(-time.Hour))
	if err != nil || n != 0 {
		t.Fatalf("DeleteOlderThan none: n=%d err=%v", n, err)
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

func TestUserEncounterRepository_CloseEncounterLeave_closesLatestOpenOnly(t *testing.T) {
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
	t1 := t0.Add(time.Hour)
	for _, spec := range []struct {
		id, inst string
		join     time.Time
	}{
		{"old", "inst_a", t0},
		{"new", "inst_b", t1},
	} {
		if saveErr := repo.Save(ctx, &activity.UserEncounter{
			ID: spec.id, VRCUserID: "usr_x", DisplayName: "X", InstanceID: spec.inst, WorldID: "w",
			JoinedAt: spec.join, LeftAt: nil,
		}); saveErr != nil {
			t.Fatal(saveErr)
		}
	}
	left := t1.Add(30 * time.Minute)
	n, err := repo.CloseEncounterLeave(ctx, "usr_x", "inst_b", left)
	if err != nil || n != 1 {
		t.Fatalf("CloseEncounterLeave: n=%d err=%v", n, err)
	}
	list, _ := repo.List(ctx, &activity.EncounterFilter{VRCUserID: "usr_x"})
	for _, e := range list {
		switch e.ID {
		case "new":
			if e.LeftAt == nil || !e.LeftAt.Equal(left) {
				t.Fatalf("new row left_at = %v, want %v", e.LeftAt, left)
			}
		case "old":
			if e.LeftAt != nil {
				t.Fatalf("old row should stay open, left_at=%v", e.LeftAt)
			}
		}
	}
}

func TestUserEncounterRepository_CloseOpenEncountersAt_skipsFutureJoins(t *testing.T) {
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
	early := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	late := early.Add(2 * time.Hour)
	if saveErr := repo.Save(ctx, &activity.UserEncounter{
		ID: "late", VRCUserID: "usr_late", DisplayName: "Late", InstanceID: "i", WorldID: "w",
		JoinedAt: late, LeftAt: nil,
	}); saveErr != nil {
		t.Fatal(saveErr)
	}
	n, err := repo.CloseOpenEncountersAt(ctx, early.Add(time.Hour))
	if err != nil || n != 0 {
		t.Fatalf("CloseOpenEncountersAt: n=%d err=%v, want 0", n, err)
	}
	list, _ := repo.List(ctx, nil)
	if list[0].LeftAt != nil {
		t.Fatalf("future join should remain open, left_at=%v", list[0].LeftAt)
	}
}

func TestUserEncounterRepository_DeduplicateEncounters(t *testing.T) {
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
	join := time.Date(2026, 6, 22, 22, 51, 17, 0, time.FixedZone("JST", 9*3600))
	validLeave := join.Add(17 * time.Minute)
	invalidLeave := join.Add(-4 * time.Hour)
	for _, spec := range []struct {
		id, inst string
		left     *time.Time
	}{
		{"dup_a", "", &invalidLeave},
		{"dup_b", "wrld_x:1~region(jp)", &validLeave},
	} {
		if saveErr := repo.Save(ctx, &activity.UserEncounter{
			ID: spec.id, VRCUserID: "usr_dup", DisplayName: "User A", InstanceID: spec.inst, WorldID: "wrld_x",
			JoinedAt: join, LeftAt: spec.left,
		}); saveErr != nil {
			t.Fatal(saveErr)
		}
	}
	n, err := repo.DeduplicateEncounters(ctx)
	if err != nil || n < 1 {
		t.Fatalf("DeduplicateEncounters: n=%d err=%v", n, err)
	}
	list, _ := repo.List(ctx, &activity.EncounterFilter{VRCUserID: "usr_dup"})
	if len(list) != 1 {
		t.Fatalf("want 1 row after dedupe, got %d", len(list))
	}
	kept := list[0]
	if kept.InstanceID == "" {
		t.Fatalf("kept row should prefer filled instance_id, got %+v", kept)
	}
	if kept.LeftAt == nil || !kept.LeftAt.Equal(validLeave) {
		t.Fatalf("kept left_at = %v, want %v", kept.LeftAt, validLeave)
	}
}
