package sqlite

import (
	"context"
	"database/sql"
	"time"

	"vrchat-tweaker/internal/domain/identity"
)

var _ identity.FriendCacheRepository = (*FriendCacheRepository)(nil)

// FriendCacheRepository implements identity.FriendCacheRepository.
type FriendCacheRepository struct {
	db *sql.DB
}

// NewFriendCacheRepository creates a new FriendCacheRepository.
func NewFriendCacheRepository(db *sql.DB) *FriendCacheRepository {
	return &FriendCacheRepository{db: db}
}

// List returns all cached friends.
func (r *FriendCacheRepository) List(ctx context.Context) ([]*identity.FriendCache, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT vrc_user_id, display_name, status, is_favorite, last_updated FROM friends_cache ORDER BY display_name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var list []*identity.FriendCache
	for rows.Next() {
		f, err := scanFriendCache(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, rows.Err()
}

// GetByVRCUserID returns a cached friend by VRChat user ID.
func (r *FriendCacheRepository) GetByVRCUserID(ctx context.Context, vrcUserID string) (*identity.FriendCache, error) {
	row := r.db.QueryRowContext(ctx, `SELECT vrc_user_id, display_name, status, is_favorite, last_updated FROM friends_cache WHERE vrc_user_id = ?`, vrcUserID)
	return scanFriendCacheRow(row)
}

// ListFavorites returns cached friends marked as favorite.
func (r *FriendCacheRepository) ListFavorites(ctx context.Context) ([]*identity.FriendCache, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT vrc_user_id, display_name, status, is_favorite, last_updated FROM friends_cache WHERE is_favorite = 1 ORDER BY display_name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var list []*identity.FriendCache
	for rows.Next() {
		f, err := scanFriendCache(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, rows.Err()
}

// Save persists a friend cache (upsert).
func (r *FriendCacheRepository) Save(ctx context.Context, f *identity.FriendCache) error {
	isFav := 0
	if f.IsFavorite {
		isFav = 1
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO friends_cache (vrc_user_id, display_name, status, is_favorite, last_updated)
		VALUES (?, ?, ?, ?, ?) ON CONFLICT(vrc_user_id) DO UPDATE SET
		display_name = excluded.display_name, status = excluded.status, is_favorite = excluded.is_favorite, last_updated = excluded.last_updated`,
		f.VRCUserID, f.DisplayName, f.Status, isFav, f.LastUpdated.Format(time.RFC3339))
	return err
}

// SaveBatch persists multiple friends.
func (r *FriendCacheRepository) SaveBatch(ctx context.Context, friends []*identity.FriendCache) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO friends_cache (vrc_user_id, display_name, status, is_favorite, last_updated)
		VALUES (?, ?, ?, ?, ?) ON CONFLICT(vrc_user_id) DO UPDATE SET
		display_name = excluded.display_name, status = excluded.status, is_favorite = excluded.is_favorite, last_updated = excluded.last_updated`)
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()

	for _, f := range friends {
		isFav := 0
		if f.IsFavorite {
			isFav = 1
		}
		_, err = stmt.ExecContext(ctx, f.VRCUserID, f.DisplayName, f.Status, isFav, f.LastUpdated.Format(time.RFC3339))
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Delete removes a friend cache.
func (r *FriendCacheRepository) Delete(ctx context.Context, vrcUserID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM friends_cache WHERE vrc_user_id = ?`, vrcUserID)
	return err
}

// DeleteAll removes all cached friends.
func (r *FriendCacheRepository) DeleteAll(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM friends_cache`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func scanFriendCache(rows *sql.Rows) (*identity.FriendCache, error) {
	var vrcUserID, displayName, status string
	var isFav int
	var lastUpdated string
	if err := rows.Scan(&vrcUserID, &displayName, &status, &isFav, &lastUpdated); err != nil {
		return nil, err
	}
	t, _ := time.Parse(time.RFC3339, lastUpdated)
	return &identity.FriendCache{
		VRCUserID:   vrcUserID,
		DisplayName: displayName,
		Status:      status,
		IsFavorite:  isFav == 1,
		LastUpdated: t,
	}, nil
}

func scanFriendCacheRow(row *sql.Row) (*identity.FriendCache, error) {
	var vrcUserID, displayName, status string
	var isFav int
	var lastUpdated string
	err := row.Scan(&vrcUserID, &displayName, &status, &isFav, &lastUpdated)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t, _ := time.Parse(time.RFC3339, lastUpdated)
	return &identity.FriendCache{
		VRCUserID:   vrcUserID,
		DisplayName: displayName,
		Status:      status,
		IsFavorite:  isFav == 1,
		LastUpdated: t,
	}, nil
}
