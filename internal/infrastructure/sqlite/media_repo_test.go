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

	if migErr := migrate(db); migErr != nil {
		t.Fatal(migErr)
	}

	repo := NewScreenshotRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	s := &media.Screenshot{
		ID:        "s1",
		FilePath:  "/tmp/a.png",
		WorldID:   "w",
		WorldName: "W",
		TakenAt:   &now,
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
