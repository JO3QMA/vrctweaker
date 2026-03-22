package sqlite

import (
	"context"
	"database/sql"
	"time"

	"vrchat-tweaker/internal/domain/identity"
)

var _ identity.UserCacheRepository = (*UserCacheRepository)(nil)

// UserCacheRepository implements identity.UserCacheRepository.
type UserCacheRepository struct {
	db *sql.DB
}

// NewUserCacheRepository creates a UserCacheRepository.
func NewUserCacheRepository(db *sql.DB) *UserCacheRepository {
	return &UserCacheRepository{db: db}
}

const userCacheSelectCols = `vrc_user_id, display_name, status, is_favorite, last_updated, first_seen_at, last_contact_at`

// List returns cached users that have API status (excludes log-only rows).
func (r *UserCacheRepository) List(ctx context.Context) ([]*identity.UserCache, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+userCacheSelectCols+` FROM users_cache WHERE status IS NOT NULL AND TRIM(status) != '' ORDER BY display_name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanUserCacheRows(rows)
}

// GetByVRCUserID returns a row by id (includes log-only users).
func (r *UserCacheRepository) GetByVRCUserID(ctx context.Context, vrcUserID string) (*identity.UserCache, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+userCacheSelectCols+` FROM users_cache WHERE vrc_user_id = ?`, vrcUserID)
	return scanUserCacheRow(row)
}

// ListFavorites returns favorites among API friends.
func (r *UserCacheRepository) ListFavorites(ctx context.Context) ([]*identity.UserCache, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+userCacheSelectCols+` FROM users_cache WHERE is_favorite = 1 AND status IS NOT NULL AND TRIM(status) != '' ORDER BY display_name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanUserCacheRows(rows)
}

// Save persists user cache from API sync (does not clear log contact columns).
func (r *UserCacheRepository) Save(ctx context.Context, u *identity.UserCache) error {
	isFav := 0
	if u.IsFavorite {
		isFav = 1
	}
	var fs, lc interface{}
	if u.FirstSeenAt != nil {
		fs = u.FirstSeenAt.Format(time.RFC3339)
	}
	if u.LastContactAt != nil {
		lc = u.LastContactAt.Format(time.RFC3339)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO users_cache (vrc_user_id, display_name, status, is_favorite, last_updated, first_seen_at, last_contact_at)
		VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT(vrc_user_id) DO UPDATE SET
		display_name = excluded.display_name, status = excluded.status, is_favorite = excluded.is_favorite, last_updated = excluded.last_updated`,
		u.VRCUserID, u.DisplayName, u.Status, isFav, u.LastUpdated.Format(time.RFC3339), fs, lc)
	return err
}

// SaveBatch persists multiple users from API sync.
func (r *UserCacheRepository) SaveBatch(ctx context.Context, users []*identity.UserCache) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO users_cache (vrc_user_id, display_name, status, is_favorite, last_updated, first_seen_at, last_contact_at)
		VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT(vrc_user_id) DO UPDATE SET
		display_name = excluded.display_name, status = excluded.status, is_favorite = excluded.is_favorite, last_updated = excluded.last_updated`)
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()

	for _, u := range users {
		isFav := 0
		if u.IsFavorite {
			isFav = 1
		}
		var fs, lc interface{}
		if u.FirstSeenAt != nil {
			fs = u.FirstSeenAt.Format(time.RFC3339)
		}
		if u.LastContactAt != nil {
			lc = u.LastContactAt.Format(time.RFC3339)
		}
		_, err = stmt.ExecContext(ctx, u.VRCUserID, u.DisplayName, u.Status, isFav, u.LastUpdated.Format(time.RFC3339), fs, lc)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// UpsertFromLog updates or inserts log-derived contact times without touching API status/favorite.
func (r *UserCacheRepository) UpsertFromLog(ctx context.Context, vrcUserID, displayName string, at time.Time) error {
	if vrcUserID == "" {
		return nil
	}
	ts := at.Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `INSERT INTO users_cache (vrc_user_id, display_name, status, is_favorite, last_updated, first_seen_at, last_contact_at)
		VALUES (?, ?, NULL, 0, ?, ?, ?) ON CONFLICT(vrc_user_id) DO UPDATE SET
		display_name = excluded.display_name,
		last_contact_at = excluded.last_contact_at,
		first_seen_at = COALESCE(users_cache.first_seen_at, excluded.first_seen_at),
		last_updated = CASE WHEN users_cache.status IS NOT NULL AND TRIM(users_cache.status) != '' THEN users_cache.last_updated ELSE excluded.last_updated END`,
		vrcUserID, displayName, ts, ts, ts)
	return err
}

// Delete removes a row by VRChat user ID.
func (r *UserCacheRepository) Delete(ctx context.Context, vrcUserID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users_cache WHERE vrc_user_id = ?`, vrcUserID)
	return err
}

// DeleteAll removes all rows.
func (r *UserCacheRepository) DeleteAll(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM users_cache`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func scanUserCacheRows(rows *sql.Rows) ([]*identity.UserCache, error) {
	var list []*identity.UserCache
	for rows.Next() {
		u, err := scanUserCacheScanner(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

func scanUserCacheScanner(sc interface {
	Scan(dest ...interface{}) error
}) (*identity.UserCache, error) {
	var vrcUserID, displayName, status, lastUpdated string
	var isFav int
	var firstSeen, lastContact sql.NullString
	if err := sc.Scan(&vrcUserID, &displayName, &status, &isFav, &lastUpdated, &firstSeen, &lastContact); err != nil {
		return nil, err
	}
	t, _ := time.Parse(time.RFC3339, lastUpdated)
	u := &identity.UserCache{
		VRCUserID:   vrcUserID,
		DisplayName: displayName,
		Status:      status,
		IsFavorite:  isFav == 1,
		LastUpdated: t,
	}
	if firstSeen.Valid {
		if ft, err := time.Parse(time.RFC3339, firstSeen.String); err == nil {
			u.FirstSeenAt = &ft
		}
	}
	if lastContact.Valid {
		if lt, err := time.Parse(time.RFC3339, lastContact.String); err == nil {
			u.LastContactAt = &lt
		}
	}
	return u, nil
}

func scanUserCacheRow(row *sql.Row) (*identity.UserCache, error) {
	u, err := scanUserCacheScanner(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}
