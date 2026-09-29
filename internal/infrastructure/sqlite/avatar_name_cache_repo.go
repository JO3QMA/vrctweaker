package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"vrchat-tweaker/internal/domain/identity"
)

// AvatarNameCacheRepository persists GET /avatars/{id} display names.
type AvatarNameCacheRepository struct {
	db *sql.DB
}

// NewAvatarNameCacheRepository creates a repository for avatar_cache.
func NewAvatarNameCacheRepository(db *sql.DB) *AvatarNameCacheRepository {
	return &AvatarNameCacheRepository{db: db}
}

// Get returns a cached avatar name, or (nil, nil) when the id is unknown.
func (r *AvatarNameCacheRepository) Get(ctx context.Context, avatarID string) (*identity.AvatarNameCache, error) {
	avatarID = strings.TrimSpace(avatarID)
	if avatarID == "" {
		return nil, nil
	}
	var name, fetched string
	err := r.db.QueryRowContext(ctx, `SELECT name, fetched_at FROM avatar_cache WHERE avatar_id = ?`, avatarID).Scan(&name, &fetched)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	at, err := time.Parse(time.RFC3339, fetched)
	if err != nil {
		return nil, fmt.Errorf("avatar cache fetched_at: %w", err)
	}
	return &identity.AvatarNameCache{AvatarID: avatarID, Name: name, FetchedAt: at}, nil
}

// Upsert stores an avatar display name. Newlines in the id or name are rejected.
func (r *AvatarNameCacheRepository) Upsert(ctx context.Context, row *identity.AvatarNameCache) error {
	if row == nil {
		return fmt.Errorf("avatar cache: invalid row")
	}
	avatarID := strings.TrimSpace(row.AvatarID)
	name := strings.TrimSpace(row.Name)
	if avatarID == "" || name == "" || row.FetchedAt.IsZero() {
		return fmt.Errorf("avatar cache: invalid row")
	}
	if strings.ContainsAny(avatarID, "\n\r") || strings.ContainsAny(name, "\n\r") {
		return fmt.Errorf("avatar cache: invalid name")
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO avatar_cache (avatar_id, name, fetched_at)
		VALUES (?, ?, ?)
		ON CONFLICT(avatar_id) DO UPDATE SET name = excluded.name, fetched_at = excluded.fetched_at`,
		avatarID, name, row.FetchedAt.UTC().Format(time.RFC3339))
	return err
}
