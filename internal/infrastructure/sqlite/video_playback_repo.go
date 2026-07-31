package sqlite

import (
	"context"
	"database/sql"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

// VideoPlaybackRepository persists video playback attempts in SQLite.
type VideoPlaybackRepository struct {
	db *sql.DB
}

// NewVideoPlaybackRepository creates a VideoPlaybackRepository.
func NewVideoPlaybackRepository(db *sql.DB) *VideoPlaybackRepository {
	return &VideoPlaybackRepository{db: db}
}

// Save inserts a video playback attempt row.
func (r *VideoPlaybackRepository) Save(ctx context.Context, a *activity.VideoPlaybackAttempt) error {
	var completed any
	if a.CompletedAt != nil {
		completed = a.CompletedAt.Format(time.RFC3339)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO video_playback_history (
		id, attempted_at, url, outcome, failure_reason, resolved_url, world_id, log_source_path, completed_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID,
		a.AttemptedAt.Format(time.RFC3339),
		a.URL,
		a.Outcome,
		nullIfEmpty(a.FailureReason),
		nullIfEmpty(a.ResolvedURL),
		nullIfEmpty(a.WorldID),
		nullIfEmpty(a.LogSourcePath),
		completed,
	)
	return err
}

// CompleteFailure marks the oldest open attempt matching url (or oldest open if url empty) as failure.
func (r *VideoPlaybackRepository) CompleteFailure(ctx context.Context, logSource, url, failureReason string, at time.Time) (int64, error) {
	id, err := r.findOldestOpenID(ctx, logSource, url)
	if err != nil {
		return 0, err
	}
	if id == "" {
		return 0, nil
	}
	res, err := r.db.ExecContext(ctx, `UPDATE video_playback_history SET
		outcome = ?, failure_reason = ?, completed_at = ?
		WHERE id = ? AND outcome = ''`,
		activity.VideoPlaybackOutcomeFailure, failureReason, at.Format(time.RFC3339), id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// CompleteSuccess marks the oldest open attempt with url as success. Already-failed rows are not updated.
func (r *VideoPlaybackRepository) CompleteSuccess(ctx context.Context, logSource, url, resolvedURL string, at time.Time) (int64, error) {
	if url == "" {
		return 0, nil
	}
	id, err := r.findOldestOpenID(ctx, logSource, url)
	if err != nil {
		return 0, err
	}
	if id == "" {
		return 0, nil
	}
	res, err := r.db.ExecContext(ctx, `UPDATE video_playback_history SET
		outcome = ?, resolved_url = ?, completed_at = ?
		WHERE id = ? AND outcome = ''`,
		activity.VideoPlaybackOutcomeSuccess, nullIfEmpty(resolvedURL), at.Format(time.RFC3339), id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *VideoPlaybackRepository) findOldestOpenID(ctx context.Context, logSource, url string) (string, error) {
	query := `SELECT id FROM video_playback_history WHERE outcome = ''`
	args := []any{}
	if logSource != "" {
		query += ` AND log_source_path = ?`
		args = append(args, logSource)
	}
	if url != "" {
		query += ` AND url = ?`
		args = append(args, url)
	}
	query += ` ORDER BY attempted_at ASC, id ASC LIMIT 1`
	var id string
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

// ListWithContext returns all attempts newest-first, joined with world display name.
func (r *VideoPlaybackRepository) ListWithContext(ctx context.Context) ([]*activity.VideoPlaybackWithContext, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT
		v.id, v.attempted_at, v.url, v.outcome, IFNULL(v.failure_reason, ''), IFNULL(v.resolved_url, ''),
		IFNULL(v.world_id, ''), IFNULL(v.log_source_path, ''), v.completed_at,
		IFNULL(w.display_name, '')
		FROM video_playback_history v
		LEFT JOIN world_info w ON w.world_id = v.world_id
		ORDER BY v.attempted_at DESC, v.id DESC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var list []*activity.VideoPlaybackWithContext
	for rows.Next() {
		var (
			a         activity.VideoPlaybackAttempt
			attempted string
			completed sql.NullString
			worldName string
		)
		if err := rows.Scan(
			&a.ID, &attempted, &a.URL, &a.Outcome, &a.FailureReason, &a.ResolvedURL,
			&a.WorldID, &a.LogSourcePath, &completed, &worldName,
		); err != nil {
			return nil, err
		}
		at, err := time.Parse(time.RFC3339, attempted)
		if err != nil {
			return nil, err
		}
		a.AttemptedAt = at
		if completed.Valid && completed.String != "" {
			ct, err := time.Parse(time.RFC3339, completed.String)
			if err != nil {
				return nil, err
			}
			a.CompletedAt = &ct
		}
		list = append(list, &activity.VideoPlaybackWithContext{
			Attempt:          &a,
			WorldDisplayName: worldName,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if list == nil {
		list = []*activity.VideoPlaybackWithContext{}
	}
	return list, nil
}

// DeleteOlderThan removes attempts with attempted_at before the cutoff.
func (r *VideoPlaybackRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM video_playback_history WHERE attempted_at < ?`, before.Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
