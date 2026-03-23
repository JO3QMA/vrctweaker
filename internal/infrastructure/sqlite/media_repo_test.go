package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"vrchat-tweaker/internal/domain/media"
)

func TestScreenshotRepository_ThumbnailRoundTrip(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, openErr := sql.Open("sqlite", dbPath)
	if openErr != nil {
		t.Fatal(openErr)
	}
	t.Cleanup(func() { _ = db.Close() })

	if migErr := applySchema(db); migErr != nil {
		t.Fatal(migErr)
	}

	repo := NewScreenshotRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	s := &media.Screenshot{
		ID:       "s1",
		FilePath: "/tmp/a.png",
		WorldID:  "w",
		TakenAt:  &now,
	}
	sz := int64(42)
	s.FileSizeBytes = &sz
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}

	jpegHdr := []byte{0xff, 0xd8, 0xff, 0xe0} // minimal JPEG SOI + marker
	th := &media.ScreenshotThumbnail{
		JpegBlob:      jpegHdr,
		SourceSize:    100,
		SourceModUnix: 1700000000,
	}
	if err := repo.UpsertThumbnail(ctx, "s1", th); err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetThumbnail(ctx, "s1")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected thumbnail")
	}
	if string(got.JpegBlob) != string(jpegHdr) {
		t.Fatalf("thumbnail blob mismatch")
	}
	if got.SourceSize != 100 || got.SourceModUnix != 1700000000 {
		t.Fatalf("meta mismatch: %+v", got)
	}

	if delErr := repo.DeleteThumbnail(ctx, "s1"); delErr != nil {
		t.Fatal(delErr)
	}
	got2, err := repo.GetThumbnail(ctx, "s1")
	if err != nil {
		t.Fatal(err)
	}
	if got2 != nil {
		t.Fatal("expected no thumbnail after delete")
	}
}

func TestScreenshotRepository_ListJoinsWorldAndAuthorDisplayNames(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if mErr := applySchema(db); mErr != nil {
		t.Fatal(mErr)
	}
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	_, err = db.ExecContext(ctx, `INSERT INTO world_info (world_id, display_name, last_visited_at) VALUES (?, ?, ?)`,
		"wrld_join", "Joined World", now.Format(time.RFC3339))
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO users_cache (vrc_user_id, display_name, status, is_favorite, last_updated, user_kind) VALUES (?, ?, '', 0, ?, 'contact')`,
		"usr_join", "Author Joined", now.Format(time.RFC3339))
	if err != nil {
		t.Fatal(err)
	}

	repo := NewScreenshotRepository(db)
	sz := int64(10)
	s := &media.Screenshot{
		ID:              "sj",
		FilePath:        "/tmp/join.png",
		WorldID:         "wrld_join",
		AuthorVRCUserID: "usr_join",
		TakenAt:         &now,
		FileSizeBytes:   &sz,
	}
	if saveErr := repo.Save(ctx, s); saveErr != nil {
		t.Fatal(saveErr)
	}

	list, err := repo.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("len(list)=%d", len(list))
	}
	got := list[0]
	if got.WorldName != "Joined World" {
		t.Errorf("WorldName = %q, want Joined World", got.WorldName)
	}
	if got.AuthorDisplayName != "Author Joined" {
		t.Errorf("AuthorDisplayName = %q", got.AuthorDisplayName)
	}

	pat := "%Joined%"
	list2, err := repo.List(ctx, &media.ScreenshotFilter{WorldName: pat})
	if err != nil {
		t.Fatal(err)
	}
	if len(list2) != 1 {
		t.Fatalf("WorldName filter: len=%d", len(list2))
	}
}
