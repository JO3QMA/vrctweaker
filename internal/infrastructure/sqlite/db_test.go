package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
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

func TestOpen_setsBusyTimeoutAndWAL(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	var journalMode string
	if err := db.QueryRow(`PRAGMA journal_mode`).Scan(&journalMode); err != nil {
		t.Fatal(err)
	}
	if strings.ToLower(journalMode) != "wal" {
		t.Fatalf("journal_mode: got %q, want wal", journalMode)
	}

	var busyMs int
	if err := db.QueryRow(`PRAGMA busy_timeout`).Scan(&busyMs); err != nil {
		t.Fatal(err)
	}
	if busyMs != defaultBusyTimeoutMs {
		t.Fatalf("busy_timeout: got %d, want %d", busyMs, defaultBusyTimeoutMs)
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

func TestOpen_seedsDefaults(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()

	var retention string
	if scanErr := db.QueryRowContext(ctx, `SELECT value FROM app_settings WHERE key = 'log_retention_days'`).Scan(&retention); scanErr != nil {
		t.Fatal(scanErr)
	}
	if retention != "30" {
		t.Fatalf("log_retention_days=%q", retention)
	}

	launcherRepo := NewLauncherProfileRepository(db)
	list, err := launcherRepo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("seed profiles count=%d want 2: %#v", len(list), list)
	}

	desktop, err := launcherRepo.GetByID(ctx, "default-desktop")
	if err != nil || desktop == nil {
		t.Fatalf("default-desktop: %#v err=%v", desktop, err)
	}
	if desktop.Name != "Desktop" || desktop.Arguments != "--no-vr" || !desktop.IsDefault {
		t.Fatalf("default-desktop fields: %#v", desktop)
	}

	vr, err := launcherRepo.GetByID(ctx, "default-vr")
	if err != nil || vr == nil {
		t.Fatalf("default-vr: %#v err=%v", vr, err)
	}
	if vr.Name != "VR" || vr.Arguments != "" || vr.IsDefault {
		t.Fatalf("default-vr fields: %#v", vr)
	}

	def, err := launcherRepo.GetDefault(ctx)
	if err != nil || def == nil || def.ID != "default-desktop" {
		t.Fatalf("GetDefault: %#v err=%v", def, err)
	}
}

func TestOpen_doesNotAddVRWhenProfilesExist(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "vrchat-tweaker.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	if applyErr := applySchema(db); applyErr != nil {
		t.Fatal(applyErr)
	}
	if _, delErr := db.Exec(`DELETE FROM launch_profiles`); delErr != nil {
		t.Fatal(delErr)
	}
	if _, insErr := db.Exec(`INSERT INTO launch_profiles (id, name, arguments, is_default, created_at, updated_at)
		VALUES ('legacy-desktop', 'Desktop', '--no-vr', 1, datetime('now'), datetime('now'))`); insErr != nil {
		t.Fatal(insErr)
	}
	if closeErr := db.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reopened.Close() })

	ctx := context.Background()
	repo := NewLauncherProfileRepository(reopened)
	list, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected no backfill, got %d profiles: %#v", len(list), list)
	}
	if list[0].ID != "legacy-desktop" {
		t.Fatalf("unexpected profile: %#v", list[0])
	}
	vr, err := repo.GetByID(ctx, "default-vr")
	if err != nil {
		t.Fatal(err)
	}
	if vr != nil {
		t.Fatalf("default-vr must not be added to existing DB: %#v", vr)
	}
}

func TestApplySchema_idempotent(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := applySchema(db); err != nil {
		t.Fatal(err)
	}
	if err := applySchema(db); err != nil {
		t.Fatalf("second applySchema: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM launch_profiles`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("launch_profiles count=%d want 2 after idempotent applySchema", count)
	}
}

func TestOpen_rejectsInvalidDataDir(t *testing.T) {
	_, err := Open("/proc/self/mem/not-a-directory")
	if err == nil {
		t.Fatal("expected Open to fail for invalid data dir")
	}
}
