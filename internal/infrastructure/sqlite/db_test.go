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

func TestMigrate_renamesFriendsCacheBeforeCreatingUsersCache(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "vrchat-tweaker.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	exec := func(q string, args ...any) {
		t.Helper()
		if _, execErr := db.Exec(q, args...); execErr != nil {
			t.Fatal(execErr)
		}
	}

	exec(`CREATE TABLE friends_cache (
		vrc_user_id TEXT PRIMARY KEY,
		display_name TEXT NOT NULL,
		status TEXT,
		is_favorite INTEGER DEFAULT 0,
		last_updated TEXT NOT NULL
	)`)
	exec(`INSERT INTO friends_cache (vrc_user_id, display_name, status, is_favorite, last_updated)
		VALUES ('usr_legacy', 'Legacy Friend', 'join me', 1, '2020-01-01T00:00:00Z')`)

	if migErr := migrate(db); migErr != nil {
		t.Fatal(migErr)
	}

	var gotName string
	if err := db.QueryRow(`SELECT display_name FROM users_cache WHERE vrc_user_id = ?`, "usr_legacy").Scan(&gotName); err != nil {
		t.Fatalf("expected row in users_cache after rename: %v", err)
	}
	if gotName != "Legacy Friend" {
		t.Fatalf("display_name = %q, want Legacy Friend", gotName)
	}

	var userKind string
	if err := db.QueryRow(`SELECT user_kind FROM users_cache WHERE vrc_user_id = ?`, "usr_legacy").Scan(&userKind); err != nil {
		t.Fatalf("user_kind: %v", err)
	}
	if userKind != "friend" {
		t.Fatalf("user_kind = %q, want friend (API-style status backfill)", userKind)
	}

	var friendsN int
	_ = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='friends_cache'`).Scan(&friendsN)
	if friendsN != 0 {
		t.Fatalf("friends_cache should be gone after rename, sqlite_master count=%d", friendsN)
	}
}

func TestEnsureScreenshotThumbnailJpegBlobColumn_LegacyWebpBlob(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "legacy.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	exec := func(q string, args ...any) {
		t.Helper()
		if _, execErr := db.Exec(q, args...); execErr != nil {
			t.Fatal(execErr)
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

	if colErr := ensureScreenshotThumbnailJpegBlobColumn(db); colErr != nil {
		t.Fatal(colErr)
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

func TestOpen_foreignKeysCascadeDeletesThumbnails(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// Exercise the pool so DELETE may run on a connection other than Ping/migrate.
	db.SetMaxOpenConns(4)

	repo := NewScreenshotRepository(db)
	ctx := context.Background()
	taken := time.Now().UTC().Truncate(time.Second)
	s := &media.Screenshot{
		ID:       "s-cascade",
		FilePath: "/tmp/cascade.png",
		TakenAt:  &taken,
	}
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertThumbnail(ctx, s.ID, &media.ScreenshotThumbnail{
		JpegBlob:      []byte{0xff, 0xd8, 0xff},
		SourceSize:    1,
		SourceModUnix: 2,
	}); err != nil {
		t.Fatal(err)
	}

	if err := repo.Delete(ctx, s.ID); err != nil {
		t.Fatal(err)
	}

	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM screenshot_thumbnails`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected CASCADE to remove thumbnail row, got count %d", n)
	}
}
