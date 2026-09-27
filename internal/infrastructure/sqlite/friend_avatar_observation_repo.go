package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

// FriendAvatarObservationRepository persists log-derived friend avatar switches.
type FriendAvatarObservationRepository struct {
	db *sql.DB
}

// NewFriendAvatarObservationRepository creates a repository.
func NewFriendAvatarObservationRepository(db *sql.DB) *FriendAvatarObservationRepository {
	return &FriendAvatarObservationRepository{db: db}
}

// Save inserts one observation row.
func (r *FriendAvatarObservationRepository) Save(ctx context.Context, o *activity.FriendAvatarObservation) error {
	if o == nil || o.ID == "" || o.AvatarName == "" || o.DisplayName == "" {
		return nil
	}
	if strings.ContainsAny(o.AvatarName, "\n\r") || strings.ContainsAny(o.DisplayName, "\n\r") {
		return fmt.Errorf("friend avatar observation: invalid name")
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO friend_avatar_observations (
		id, vrc_user_id, display_name, avatar_name, instance_id, world_id, log_source_path, observed_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		o.ID, o.VRCUserID, o.DisplayName, o.AvatarName, o.InstanceID, o.WorldID, o.LogSourcePath, o.ObservedAt.Format(time.RFC3339))
	return err
}

// ListUsageSummariesByVRCUserID returns aggregated avatar usage for a friend (newest last_seen first).
func (r *FriendAvatarObservationRepository) ListUsageSummariesByVRCUserID(ctx context.Context, vrcUserID, displayName string) ([]*activity.FriendAvatarUsageSummary, error) {
	vrcUserID = strings.TrimSpace(vrcUserID)
	displayName = strings.TrimSpace(displayName)
	if vrcUserID == "" {
		return []*activity.FriendAvatarUsageSummary{}, nil
	}
	query := `SELECT avatar_name, COUNT(*), MIN(observed_at), MAX(observed_at)
		FROM friend_avatar_observations
		WHERE vrc_user_id = ?`
	args := []any{vrcUserID}
	if displayName != "" {
		query += ` OR (vrc_user_id = '' AND display_name = ?)`
		args = append(args, displayName)
	}
	query += ` GROUP BY avatar_name ORDER BY MAX(observed_at) DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var list []*activity.FriendAvatarUsageSummary
	for rows.Next() {
		var name string
		var count int64
		var firstISO, lastISO string
		if err := rows.Scan(&name, &count, &firstISO, &lastISO); err != nil {
			return nil, err
		}
		first, err := time.Parse(time.RFC3339, firstISO)
		if err != nil {
			return nil, fmt.Errorf("parse first_seen: %w", err)
		}
		last, err := time.Parse(time.RFC3339, lastISO)
		if err != nil {
			return nil, fmt.Errorf("parse last_seen: %w", err)
		}
		list = append(list, &activity.FriendAvatarUsageSummary{
			AvatarName:  name,
			UseCount:    count,
			FirstSeenAt: first,
			LastSeenAt:  last,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if list == nil {
		return []*activity.FriendAvatarUsageSummary{}, nil
	}
	return list, nil
}

// DeleteOlderThan removes observations before cutoff (by observed_at).
func (r *FriendAvatarObservationRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM friend_avatar_observations WHERE observed_at < ?`, before.Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}

// DeleteAll removes all friend avatar observations.
func (r *FriendAvatarObservationRepository) DeleteAll(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM friend_avatar_observations`)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}
