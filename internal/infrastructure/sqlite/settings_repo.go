package sqlite

import (
	"context"
	"database/sql"
	"time"

	"vrchat-tweaker/internal/domain/settings"
)

var _ settings.AppSettingsRepository = (*AppSettingsRepository)(nil)

// AppSettingsRepository implements settings.AppSettingsRepository.
type AppSettingsRepository struct {
	db *sql.DB
}

// NewAppSettingsRepository creates a new AppSettingsRepository.
func NewAppSettingsRepository(db *sql.DB) *AppSettingsRepository {
	return &AppSettingsRepository{db: db}
}

// Get returns the value for the given key.
func (r *AppSettingsRepository) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM app_settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// Set saves a key-value pair.
func (r *AppSettingsRepository) Set(ctx context.Context, key, value string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `INSERT INTO app_settings (key, value, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		key, value, now)
	return err
}

// GetAll returns all settings.
func (r *AppSettingsRepository) GetAll(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM app_settings`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		result[k] = v
	}
	return result, rows.Err()
}
