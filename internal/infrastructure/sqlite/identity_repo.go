package sqlite

import (
	"context"
	"database/sql"
	"fmt"
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

const userCacheSelectCols = `vrc_user_id, display_name, status, is_favorite, last_updated, first_seen_at, last_contact_at,
	user_kind, session_fingerprint, username, status_description, user_state, avatar_thumbnail_url, user_icon_url, profile_pic_override_thumbnail`

// List returns cached VRChat friends (user_kind=friend with API status).
func (r *UserCacheRepository) List(ctx context.Context) ([]*identity.UserCache, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+userCacheSelectCols+` FROM users_cache WHERE user_kind = 'friend' AND status IS NOT NULL AND TRIM(status) != '' ORDER BY display_name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanUserCacheRows(rows)
}

// GetByVRCUserID returns a row by id (any user_kind).
func (r *UserCacheRepository) GetByVRCUserID(ctx context.Context, vrcUserID string) (*identity.UserCache, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+userCacheSelectCols+` FROM users_cache WHERE vrc_user_id = ?`, vrcUserID)
	return scanUserCacheRow(row)
}

// ListFavorites returns favorites among API friends.
func (r *UserCacheRepository) ListFavorites(ctx context.Context) ([]*identity.UserCache, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+userCacheSelectCols+` FROM users_cache WHERE user_kind = 'friend' AND is_favorite = 1 AND status IS NOT NULL AND TRIM(status) != '' ORDER BY display_name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanUserCacheRows(rows)
}

// Save persists a user cache row (favorite toggle, etc.).
func (r *UserCacheRepository) Save(ctx context.Context, u *identity.UserCache) error {
	uk := u.UserKind
	if uk == "" {
		uk = identity.UserKindContact
	}
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
	_, err := r.db.ExecContext(ctx, `INSERT INTO users_cache (vrc_user_id, display_name, status, is_favorite, last_updated, first_seen_at, last_contact_at, user_kind, session_fingerprint, username, status_description, user_state, avatar_thumbnail_url, user_icon_url, profile_pic_override_thumbnail)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(vrc_user_id) DO UPDATE SET
		display_name = excluded.display_name,
		status = excluded.status,
		is_favorite = excluded.is_favorite,
		last_updated = excluded.last_updated,
		user_kind = excluded.user_kind,
		session_fingerprint = excluded.session_fingerprint,
		username = excluded.username,
		status_description = excluded.status_description,
		user_state = excluded.user_state,
		avatar_thumbnail_url = excluded.avatar_thumbnail_url,
		user_icon_url = excluded.user_icon_url,
		profile_pic_override_thumbnail = excluded.profile_pic_override_thumbnail`,
		u.VRCUserID, u.DisplayName, nullString(u.Status), isFav, u.LastUpdated.Format(time.RFC3339), fs, lc, string(uk), nullString(u.SessionFingerprint),
		nullString(u.Username), nullString(u.StatusDescription), nullString(u.UserState),
		nullString(u.AvatarThumbnailURL), nullString(u.UserIconURL), nullString(u.ProfilePicOverrideThumbnail))
	return err
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// SaveBatch persists multiple users from the friends API sync.
func (r *UserCacheRepository) SaveBatch(ctx context.Context, users []*identity.UserCache) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO users_cache (vrc_user_id, display_name, status, is_favorite, last_updated, first_seen_at, last_contact_at, user_kind, session_fingerprint, username, status_description, user_state, avatar_thumbnail_url, user_icon_url, profile_pic_override_thumbnail)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'friend', NULL, NULL, NULL, NULL, NULL, NULL, NULL) ON CONFLICT(vrc_user_id) DO UPDATE SET
		display_name = CASE WHEN users_cache.user_kind = 'self' THEN users_cache.display_name ELSE excluded.display_name END,
		status = CASE WHEN users_cache.user_kind = 'self' THEN users_cache.status ELSE excluded.status END,
		is_favorite = CASE WHEN users_cache.user_kind = 'self' THEN users_cache.is_favorite ELSE excluded.is_favorite END,
		last_updated = CASE WHEN users_cache.user_kind = 'self' THEN users_cache.last_updated ELSE excluded.last_updated END,
		user_kind = CASE WHEN users_cache.user_kind = 'self' THEN users_cache.user_kind ELSE 'friend' END`)
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
		_, err = stmt.ExecContext(ctx, u.VRCUserID, u.DisplayName, nullString(u.Status), isFav, u.LastUpdated.Format(time.RFC3339), fs, lc)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// UpsertFromLog updates or inserts log-derived contact times without downgrading friend/self.
func (r *UserCacheRepository) UpsertFromLog(ctx context.Context, vrcUserID, displayName string, at time.Time) error {
	if vrcUserID == "" {
		return nil
	}
	ts := at.Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `INSERT INTO users_cache (vrc_user_id, display_name, status, is_favorite, last_updated, first_seen_at, last_contact_at, user_kind, session_fingerprint, username, status_description, user_state, avatar_thumbnail_url, user_icon_url, profile_pic_override_thumbnail)
		VALUES (?, ?, NULL, 0, ?, ?, ?, 'contact', NULL, NULL, NULL, NULL, NULL, NULL, NULL) ON CONFLICT(vrc_user_id) DO UPDATE SET
		display_name = excluded.display_name,
		user_kind = CASE WHEN users_cache.user_kind IN ('self', 'friend') THEN users_cache.user_kind ELSE 'contact' END,
		last_contact_at = CASE
			WHEN users_cache.last_contact_at IS NULL THEN excluded.last_contact_at
			WHEN excluded.last_contact_at > users_cache.last_contact_at THEN excluded.last_contact_at
			ELSE users_cache.last_contact_at END,
		first_seen_at = COALESCE(users_cache.first_seen_at, excluded.first_seen_at),
		last_updated = CASE WHEN users_cache.user_kind IN ('self', 'friend') THEN users_cache.last_updated ELSE excluded.last_updated END`,
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

// GetSelfBySessionFingerprint returns the self row for the given auth token fingerprint.
func (r *UserCacheRepository) GetSelfBySessionFingerprint(ctx context.Context, sessionFingerprint string) (*identity.UserCache, error) {
	if sessionFingerprint == "" {
		return nil, nil
	}
	row := r.db.QueryRowContext(ctx, `SELECT `+userCacheSelectCols+` FROM users_cache WHERE user_kind = 'self' AND session_fingerprint = ? LIMIT 1`, sessionFingerprint)
	return scanUserCacheRow(row)
}

// UpsertSelf removes self rows for other VRChat accounts, then inserts or updates this user's row
// to user_kind=self (ON CONFLICT: same vrc_user_id may already exist as friend/contact from logs or API).
func (r *UserCacheRepository) UpsertSelf(ctx context.Context, u *identity.UserCache) error {
	if u.VRCUserID == "" {
		return fmt.Errorf("upsert self: empty vrc_user_id")
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM users_cache WHERE user_kind = 'self' AND vrc_user_id != ?`, u.VRCUserID); err != nil {
		return err
	}
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
	_, err := r.db.ExecContext(ctx, `INSERT INTO users_cache (vrc_user_id, display_name, status, is_favorite, last_updated, first_seen_at, last_contact_at, user_kind, session_fingerprint, username, status_description, user_state, avatar_thumbnail_url, user_icon_url, profile_pic_override_thumbnail)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'self', ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(vrc_user_id) DO UPDATE SET
		display_name = excluded.display_name,
		status = excluded.status,
		is_favorite = excluded.is_favorite,
		last_updated = excluded.last_updated,
		first_seen_at = COALESCE(users_cache.first_seen_at, excluded.first_seen_at),
		last_contact_at = users_cache.last_contact_at,
		user_kind = 'self',
		session_fingerprint = excluded.session_fingerprint,
		username = excluded.username,
		status_description = excluded.status_description,
		user_state = excluded.user_state,
		avatar_thumbnail_url = excluded.avatar_thumbnail_url,
		user_icon_url = excluded.user_icon_url,
		profile_pic_override_thumbnail = excluded.profile_pic_override_thumbnail`,
		u.VRCUserID, u.DisplayName, nullString(u.Status), isFav, u.LastUpdated.Format(time.RFC3339), fs, lc,
		nullString(u.SessionFingerprint),
		nullString(u.Username), nullString(u.StatusDescription), nullString(u.UserState),
		nullString(u.AvatarThumbnailURL), nullString(u.UserIconURL), nullString(u.ProfilePicOverrideThumbnail))
	return err
}

// DeleteSelfRows removes all self profile rows.
func (r *UserCacheRepository) DeleteSelfRows(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users_cache WHERE user_kind = 'self'`)
	return err
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
	var vrcUserID, displayName, lastUpdated, userKindStr string
	var status, sessionFP, username, statusDesc, userState, avatarURL, iconURL, profilePic sql.NullString
	var isFav int
	var firstSeen, lastContact sql.NullString
	if err := sc.Scan(&vrcUserID, &displayName, &status, &isFav, &lastUpdated, &firstSeen, &lastContact,
		&userKindStr, &sessionFP, &username, &statusDesc, &userState, &avatarURL, &iconURL, &profilePic); err != nil {
		return nil, err
	}
	t, _ := time.Parse(time.RFC3339, lastUpdated)
	st := ""
	if status.Valid {
		st = status.String
	}
	uk := identity.UserKind(userKindStr)
	if uk == "" {
		uk = identity.UserKindContact
	}
	u := &identity.UserCache{
		VRCUserID:                   vrcUserID,
		DisplayName:                 displayName,
		Status:                      st,
		IsFavorite:                  isFav == 1,
		LastUpdated:                 t,
		UserKind:                    uk,
		SessionFingerprint:          sessionStr(sessionFP),
		Username:                    sessionStr(username),
		StatusDescription:           sessionStr(statusDesc),
		UserState:                   sessionStr(userState),
		AvatarThumbnailURL:          sessionStr(avatarURL),
		UserIconURL:                 sessionStr(iconURL),
		ProfilePicOverrideThumbnail: sessionStr(profilePic),
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

func sessionStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
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
