package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"vrchat-tweaker/internal/domain/media"
)

func columnNames(t *testing.T, db *sql.DB, table string) map[string]bool {
	t.Helper()
	rows, err := db.Query(fmt.Sprintf(`SELECT name FROM pragma_table_info('%s')`, table))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	out := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		out[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestApplySchema_canonicalColumns(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "t.db")+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := applySchema(db); err != nil {
		t.Fatal(err)
	}

	screenshots := columnNames(t, db, "screenshots")
	for _, col := range []string{
		"id", "file_path", "world_id", "taken_at", "file_size_bytes", "author_vrc_user_id",
	} {
		if !screenshots[col] {
			t.Fatalf("screenshots missing column %q", col)
		}
	}
	if screenshots["world_name"] {
		t.Fatal("screenshots must not define legacy world_name column")
	}

	usersCache := columnNames(t, db, "users_cache")
	for _, col := range []string{
		"vrc_user_id", "display_name", "status", "is_favorite", "last_updated",
		"first_seen_at", "last_contact_at", "user_kind", "session_fingerprint", "username",
		"status_description", "user_state", "avatar_thumbnail_url", "user_icon_url",
		"profile_pic_override_thumbnail",
		"bio", "bio_links_json", "current_avatar_image_url", "current_avatar_tags_json",
		"developer_type", "friend_key", "image_url", "last_platform", "location",
		"last_login", "last_activity", "last_mobile", "platform", "profile_pic_override", "tags_json",
	} {
		if !usersCache[col] {
			t.Fatalf("users_cache missing column %q", col)
		}
	}

	encounters := columnNames(t, db, "user_encounters")
	for _, col := range []string{"world_id", "joined_at", "left_at"} {
		if !encounters[col] {
			t.Fatalf("user_encounters missing column %q", col)
		}
	}
	if encounters["action"] || encounters["encountered_at"] {
		t.Fatal("user_encounters should not use legacy action/encountered_at columns")
	}

	thumbs := columnNames(t, db, "screenshot_thumbnails")
	if !thumbs["jpeg_blob"] || thumbs["webp_blob"] {
		t.Fatalf("screenshot_thumbnails: want jpeg_blob only, got %#v", thumbs)
	}
}

func TestOpen_foreignKeysCascadeDeletesThumbnails(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// Exercise the pool so DELETE may run on a connection other than Ping/applySchema.
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
