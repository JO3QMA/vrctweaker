package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestEnsureScreenshotThumbnailJpegBlobColumn_LegacyWebpBlob(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "legacy.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.Exec(q, args...); err != nil {
			t.Fatal(err)
		}
	}

	exec(`CREATE TABLE screenshots (
		id TEXT PRIMARY KEY,
		file_path TEXT UNIQUE NOT NULL,
		world_id TEXT,
		world_name TEXT,
		taken_at TEXT
	)`)
	exec(`INSERT INTO screenshots (id, file_path) VALUES ('s1', '/x')`)
	exec(`CREATE TABLE screenshot_thumbnails (
		screenshot_id TEXT PRIMARY KEY,
		webp_blob BLOB NOT NULL,
		source_size INTEGER NOT NULL,
		source_mod_unix INTEGER NOT NULL,
		FOREIGN KEY (screenshot_id) REFERENCES screenshots(id) ON DELETE CASCADE
	)`)
	hdr := []byte{0xff, 0xd8, 0xff}
	exec(`INSERT INTO screenshot_thumbnails (screenshot_id, webp_blob, source_size, source_mod_unix) VALUES (?, ?, ?, ?)`,
		"s1", hdr, int64(10), int64(20))

	if err := ensureScreenshotThumbnailJpegBlobColumn(db); err != nil {
		t.Fatal(err)
	}

	repo := NewScreenshotRepository(db)
	got, err := repo.GetThumbnail(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected thumbnail after column rename")
	}
	if string(got.JpegBlob) != string(hdr) {
		t.Fatalf("blob mismatch after migrate")
	}
}
