package sqlite

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const defaultLogRetentionDays = 30

// Open opens or creates the SQLite database and runs migrations.
func Open(dataDir string) (*sql.DB, error) {
	dbPath := filepath.Join(dataDir, "vrchat-tweaker.db")
	// Foreign keys are per-connection; database/sql pools connections, so a one-off
	// PRAGMA after Open only affects one handle. _pragma runs on every new connection.
	dsn := dbPath + "?_pragma=foreign_keys(1)"
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := migrate(conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return conn, nil
}

func migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS launch_profiles (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			arguments TEXT,
			is_default INTEGER DEFAULT 0,
			created_at TEXT,
			updated_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS screenshots (
			id TEXT PRIMARY KEY,
			file_path TEXT UNIQUE NOT NULL,
			world_id TEXT,
			world_name TEXT,
			taken_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS play_sessions (
			id TEXT PRIMARY KEY,
			start_time TEXT NOT NULL,
			end_time TEXT,
			duration_sec INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS user_encounters (
			id TEXT PRIMARY KEY,
			vrc_user_id TEXT NOT NULL,
			display_name TEXT NOT NULL,
			action TEXT NOT NULL,
			instance_id TEXT,
			encountered_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_user_encounters_encountered_at ON user_encounters(encountered_at)`,
		`CREATE TABLE IF NOT EXISTS users_cache (
			vrc_user_id TEXT PRIMARY KEY,
			display_name TEXT NOT NULL,
			status TEXT,
			is_favorite INTEGER DEFAULT 0,
			last_updated TEXT NOT NULL,
			first_seen_at TEXT,
			last_contact_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS world_info (
			world_id TEXT PRIMARY KEY,
			display_name TEXT,
			last_visited_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS automation_rules (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			trigger_type TEXT NOT NULL,
			condition_json TEXT,
			action_type TEXT NOT NULL,
			action_payload TEXT,
			is_enabled INTEGER DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS app_settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TEXT
		)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	if err := ensureScreenshotsFileSizeColumn(db); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS screenshot_thumbnails (
			screenshot_id TEXT PRIMARY KEY,
			jpeg_blob BLOB NOT NULL,
			source_size INTEGER NOT NULL,
			source_mod_unix INTEGER NOT NULL,
			FOREIGN KEY (screenshot_id) REFERENCES screenshots(id) ON DELETE CASCADE
		)`); err != nil {
		return fmt.Errorf("migration screenshot_thumbnails: %w", err)
	}
	if err := ensureScreenshotThumbnailJpegBlobColumn(db); err != nil {
		return err
	}
	if err := ensureFriendsCacheRenamedToUsersCache(db); err != nil {
		return err
	}
	if err := ensureUsersCacheLogColumns(db); err != nil {
		return err
	}
	if err := ensureUserEncountersWorldIDColumn(db); err != nil {
		return err
	}

	// Insert default log_retention_days if not present
	if _, err := db.Exec(`INSERT OR IGNORE INTO app_settings (key, value, updated_at) VALUES ('log_retention_days', ?, datetime('now'))`, fmt.Sprintf("%d", defaultLogRetentionDays)); err != nil {
		return err
	}

	// Seed default launch profile if none exist
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM launch_profiles`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		_, err := db.Exec(`INSERT INTO launch_profiles (id, name, arguments, is_default, created_at, updated_at)
			VALUES ('default-desktop', 'Desktop', '--no-vr', 1, datetime('now'), datetime('now'))`)
		if err != nil {
			return err
		}
	}

	return nil
}

// ensureFriendsCacheRenamedToUsersCache migrates legacy friends_cache → users_cache.
func ensureFriendsCacheRenamedToUsersCache(db *sql.DB) error {
	var friendsCount, usersCount int
	_ = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='friends_cache'`).Scan(&friendsCount)
	_ = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users_cache'`).Scan(&usersCount)
	if friendsCount > 0 && usersCount == 0 {
		if _, err := db.Exec(`ALTER TABLE friends_cache RENAME TO users_cache`); err != nil {
			return fmt.Errorf("rename friends_cache to users_cache: %w", err)
		}
	}
	return nil
}

func ensureUsersCacheLogColumns(db *sql.DB) error {
	for _, col := range []struct {
		name string
		sql  string
	}{
		{"first_seen_at", `ALTER TABLE users_cache ADD COLUMN first_seen_at TEXT`},
		{"last_contact_at", `ALTER TABLE users_cache ADD COLUMN last_contact_at TEXT`},
	} {
		if err := addColumnIfMissing(db, "users_cache", col.name, col.sql); err != nil {
			return err
		}
	}
	return nil
}

func ensureUserEncountersWorldIDColumn(db *sql.DB) error {
	return addColumnIfMissing(db, "user_encounters", "world_id", `ALTER TABLE user_encounters ADD COLUMN world_id TEXT`)
}

func addColumnIfMissing(db *sql.DB, table, column, alterSQL string) error {
	rows, err := db.Query(fmt.Sprintf(`SELECT name FROM pragma_table_info('%s')`, table))
	if err != nil {
		return fmt.Errorf("pragma_table_info %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if _, err := db.Exec(alterSQL); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "duplicate column") {
			return nil
		}
		return fmt.Errorf("%s: %w", alterSQL, err)
	}
	return nil
}

func ensureScreenshotsFileSizeColumn(db *sql.DB) error {
	_, err := db.Exec(`ALTER TABLE screenshots ADD COLUMN file_size_bytes INTEGER`)
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "duplicate column") {
		return nil
	}
	return fmt.Errorf("add file_size_bytes: %w", err)
}

// ensureScreenshotThumbnailJpegBlobColumn renames legacy webp_blob → jpeg_blob (data was always JPEG).
func ensureScreenshotThumbnailJpegBlobColumn(db *sql.DB) error {
	rows, err := db.Query(`SELECT name FROM pragma_table_info('screenshot_thumbnails')`)
	if err != nil {
		return fmt.Errorf("pragma screenshot_thumbnails columns: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var hasWebp, hasJpeg bool
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		switch name {
		case "webp_blob":
			hasWebp = true
		case "jpeg_blob":
			hasJpeg = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if hasWebp && !hasJpeg {
		if _, err := db.Exec(`ALTER TABLE screenshot_thumbnails RENAME COLUMN webp_blob TO jpeg_blob`); err != nil {
			return fmt.Errorf("rename webp_blob to jpeg_blob: %w", err)
		}
	}
	return nil
}
