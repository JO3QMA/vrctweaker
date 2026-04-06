package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestWorldInfoRepository_UpsertVisit_emptyWorldID(t *testing.T) {
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
	repo := NewWorldInfoRepository(db)
	at := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
	if emptyErr := repo.UpsertVisit(ctx, "", at); emptyErr != nil {
		t.Fatal(emptyErr)
	}
}

func TestWorldInfoRepository_UpsertVisit_GetByWorldID(t *testing.T) {
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
	repo := NewWorldInfoRepository(db)
	wid := "wrld_test_1"
	at := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
	if visitErr := repo.UpsertVisit(ctx, wid, at); visitErr != nil {
		t.Fatal(visitErr)
	}
	got, err := repo.GetByWorldID(ctx, wid)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.WorldID != wid {
		t.Fatalf("GetByWorldID: %#v", got)
	}
	if !got.LastVisitedAt.Equal(at) {
		t.Fatalf("LastVisitedAt: got %v want %v", got.LastVisitedAt, at)
	}
}

func TestWorldInfoRepository_GetByWorldID_missing(t *testing.T) {
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
	repo := NewWorldInfoRepository(db)
	got, err := repo.GetByWorldID(ctx, "wrld_none")
	if err != nil || got != nil {
		t.Fatalf("want nil,nil got %#v err=%v", got, err)
	}
}

func TestWorldInfoRepository_UpsertDisplayName(t *testing.T) {
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
	repo := NewWorldInfoRepository(db)
	wid := "wrld_named"
	at0 := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
	if visitErr := repo.UpsertVisit(ctx, wid, at0); visitErr != nil {
		t.Fatal(visitErr)
	}
	at1 := at0.Add(time.Hour)
	if dnErr := repo.UpsertDisplayName(ctx, wid, "My World", at1); dnErr != nil {
		t.Fatal(dnErr)
	}
	got, err := repo.GetByWorldID(ctx, wid)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.DisplayName != "My World" {
		t.Fatalf("DisplayName: %#v", got)
	}
	if !got.LastVisitedAt.Equal(at1) {
		t.Fatalf("LastVisitedAt: got %v want %v", got.LastVisitedAt, at1)
	}
}
