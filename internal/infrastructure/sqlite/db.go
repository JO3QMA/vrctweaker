package sqlite

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const defaultLogRetentionDays = 30

// Open opens or creates the SQLite database and applies the canonical schema.
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

	if err := applySchema(conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return conn, nil
}

func applySchema(db *sql.DB) error {
	for _, stmt := range schemaStatements() {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("schema: %w", err)
		}
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

func schemaStatements() []string {
	return []string{
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
			taken_at TEXT,
			file_size_bytes INTEGER,
			author_vrc_user_id TEXT
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
			world_id TEXT,
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
			last_contact_at TEXT,
			user_kind TEXT NOT NULL DEFAULT 'contact',
			session_fingerprint TEXT,
			username TEXT,
			status_description TEXT,
			user_state TEXT,
			avatar_thumbnail_url TEXT,
			user_icon_url TEXT,
			profile_pic_override_thumbnail TEXT
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
		`CREATE TABLE IF NOT EXISTS screenshot_thumbnails (
			screenshot_id TEXT PRIMARY KEY,
			jpeg_blob BLOB NOT NULL,
			source_size INTEGER NOT NULL,
			source_mod_unix INTEGER NOT NULL,
			FOREIGN KEY (screenshot_id) REFERENCES screenshots(id) ON DELETE CASCADE
		)`,
	}
}
