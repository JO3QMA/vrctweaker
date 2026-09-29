package sqlite

import (
	"context"
	"strings"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/identity"
)

func TestAvatarNameCacheRepository_upsertAndGet(t *testing.T) {
	db := openTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewAvatarNameCacheRepository(db)
	ctx := context.Background()
	at := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	row := &identity.AvatarNameCache{
		AvatarID:  "avtr_11111111-2222-3333-4444-555555555555",
		Name:      "Fox",
		FetchedAt: at,
	}
	if err := repo.Upsert(ctx, row); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(ctx, row.AvatarID)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Name != "Fox" || !got.FetchedAt.Equal(at) {
		t.Fatalf("got %+v", got)
	}

	later := at.Add(time.Hour)
	row.Name = "Fox v2"
	row.FetchedAt = later
	if err = repo.Upsert(ctx, row); err != nil {
		t.Fatal(err)
	}
	got, err = repo.Get(ctx, " "+row.AvatarID+" ")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Name != "Fox v2" || !got.FetchedAt.Equal(later) {
		t.Fatalf("updated %+v", got)
	}
}

func TestAvatarNameCacheRepository_getMiss(t *testing.T) {
	db := openTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewAvatarNameCacheRepository(db)
	got, err := repo.Get(context.Background(), "avtr_11111111-2222-3333-4444-555555555555")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("got %+v, want nil", got)
	}
}

func TestAvatarNameCacheRepository_rejectsNewlines(t *testing.T) {
	db := openTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewAvatarNameCacheRepository(db)
	ctx := context.Background()
	base := &identity.AvatarNameCache{
		AvatarID:  "avtr_11111111-2222-3333-4444-555555555555",
		Name:      "Fox",
		FetchedAt: time.Now().UTC(),
	}
	for _, row := range []*identity.AvatarNameCache{
		{AvatarID: "avtr_bad\nid", Name: "Fox", FetchedAt: base.FetchedAt},
		{AvatarID: base.AvatarID, Name: "Fox\rname", FetchedAt: base.FetchedAt},
	} {
		err := repo.Upsert(ctx, row)
		if err == nil || !strings.Contains(err.Error(), "invalid") {
			t.Fatalf("Upsert(%q, %q) err = %v", row.AvatarID, row.Name, err)
		}
	}
	got, err := repo.Get(ctx, base.AvatarID)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("rejected rows must not persist, got %+v", got)
	}
}
