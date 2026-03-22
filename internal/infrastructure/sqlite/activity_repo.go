package sqlite

import (
	"context"
	"database/sql"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

var _ activity.PlaySessionRepository = (*PlaySessionRepository)(nil)
var _ activity.UserEncounterRepository = (*UserEncounterRepository)(nil)

// PlaySessionRepository implements activity.PlaySessionRepository.
type PlaySessionRepository struct {
	db *sql.DB
}

// NewPlaySessionRepository creates a new PlaySessionRepository.
func NewPlaySessionRepository(db *sql.DB) *PlaySessionRepository {
	return &PlaySessionRepository{db: db}
}

// List returns play sessions within the time range.
func (r *PlaySessionRepository) List(ctx context.Context, from, to time.Time) ([]*activity.PlaySession, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, start_time, end_time, duration_sec FROM play_sessions WHERE start_time >= ? AND start_time <= ? ORDER BY start_time DESC`,
		from.Format(time.RFC3339), to.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var list []*activity.PlaySession
	for rows.Next() {
		s, err := scanPlaySession(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

// FindLatestWithoutEndTime returns the most recent play session with no end time.
func (r *PlaySessionRepository) FindLatestWithoutEndTime(ctx context.Context) (*activity.PlaySession, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, start_time, end_time, duration_sec FROM play_sessions WHERE end_time IS NULL OR end_time = '' ORDER BY start_time DESC LIMIT 1`)
	return scanPlaySessionRow(row)
}

// Count returns the number of play sessions.
func (r *PlaySessionRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM play_sessions`).Scan(&n)
	return n, err
}

// GetByID returns a play session by ID.
func (r *PlaySessionRepository) GetByID(ctx context.Context, id string) (*activity.PlaySession, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, start_time, end_time, duration_sec FROM play_sessions WHERE id = ?`, id)
	return scanPlaySessionRow(row)
}

// Save persists a play session.
func (r *PlaySessionRepository) Save(ctx context.Context, s *activity.PlaySession) error {
	endTime := interface{}(nil)
	if s.EndTime != nil {
		endTime = s.EndTime.Format(time.RFC3339)
	}
	dur := interface{}(nil)
	if s.DurationSec != nil {
		dur = *s.DurationSec
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO play_sessions (id, start_time, end_time, duration_sec) VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET start_time = excluded.start_time, end_time = excluded.end_time, duration_sec = excluded.duration_sec`,
		s.ID, s.StartTime.Format(time.RFC3339), endTime, dur)
	return err
}

func scanPlaySession(rows *sql.Rows) (*activity.PlaySession, error) {
	var id string
	var startTime string
	var endTime sql.NullString
	var durSec sql.NullInt64
	if err := rows.Scan(&id, &startTime, &endTime, &durSec); err != nil {
		return nil, err
	}
	st, _ := time.Parse(time.RFC3339, startTime)
	var et *time.Time
	if endTime.Valid && endTime.String != "" {
		t, _ := time.Parse(time.RFC3339, endTime.String)
		et = &t
	}
	var ds *int
	if durSec.Valid {
		d := int(durSec.Int64)
		ds = &d
	}
	return &activity.PlaySession{
		ID:          id,
		StartTime:   st,
		EndTime:     et,
		DurationSec: ds,
	}, nil
}

// UserEncounterRepository implements activity.UserEncounterRepository.
type UserEncounterRepository struct {
	db *sql.DB
}

// NewUserEncounterRepository creates a new UserEncounterRepository.
func NewUserEncounterRepository(db *sql.DB) *UserEncounterRepository {
	return &UserEncounterRepository{db: db}
}

// List returns user encounters with optional filters.
func (r *UserEncounterRepository) List(ctx context.Context, filter *activity.EncounterFilter) ([]*activity.UserEncounter, error) {
	query := `SELECT id, vrc_user_id, display_name, action, instance_id, world_id, encountered_at FROM user_encounters WHERE 1=1`
	args := []interface{}{}
	if filter != nil {
		if filter.VRCUserID != "" {
			query += ` AND vrc_user_id = ?`
			args = append(args, filter.VRCUserID)
		}
		if filter.DisplayName != "" {
			query += ` AND display_name LIKE ?`
			args = append(args, "%"+filter.DisplayName+"%")
		}
		if filter.InstanceID != "" {
			query += ` AND instance_id = ?`
			args = append(args, filter.InstanceID)
		}
		if filter.WorldID != "" {
			query += ` AND world_id = ?`
			args = append(args, filter.WorldID)
		}
		if filter.From != nil {
			query += ` AND encountered_at >= ?`
			args = append(args, filter.From.Format(time.RFC3339))
		}
		if filter.To != nil {
			query += ` AND encountered_at <= ?`
			args = append(args, filter.To.Format(time.RFC3339))
		}
	}
	query += ` ORDER BY encountered_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var list []*activity.UserEncounter
	for rows.Next() {
		e, err := scanUserEncounter(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

// ListWithContext returns encounters with world display name and user cache timestamps.
func (r *UserEncounterRepository) ListWithContext(ctx context.Context, filter *activity.EncounterFilter) ([]*activity.EncounterWithContext, error) {
	query := `SELECT e.id, e.vrc_user_id, e.display_name, e.action, e.instance_id, e.world_id, e.encountered_at,
		w.display_name, u.first_seen_at, u.last_contact_at
		FROM user_encounters e
		LEFT JOIN world_info w ON w.world_id = e.world_id
		LEFT JOIN users_cache u ON u.vrc_user_id = e.vrc_user_id
		WHERE 1=1`
	args := []interface{}{}
	if filter != nil {
		if filter.VRCUserID != "" {
			query += ` AND e.vrc_user_id = ?`
			args = append(args, filter.VRCUserID)
		}
		if filter.DisplayName != "" {
			query += ` AND e.display_name LIKE ?`
			args = append(args, "%"+filter.DisplayName+"%")
		}
		if filter.InstanceID != "" {
			query += ` AND e.instance_id = ?`
			args = append(args, filter.InstanceID)
		}
		if filter.WorldID != "" {
			query += ` AND e.world_id = ?`
			args = append(args, filter.WorldID)
		}
		if filter.From != nil {
			query += ` AND e.encountered_at >= ?`
			args = append(args, filter.From.Format(time.RFC3339))
		}
		if filter.To != nil {
			query += ` AND e.encountered_at <= ?`
			args = append(args, filter.To.Format(time.RFC3339))
		}
	}
	query += ` ORDER BY e.encountered_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var list []*activity.EncounterWithContext
	for rows.Next() {
		row, err := scanEncounterWithContextRow(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

// Save persists a user encounter.
func (r *UserEncounterRepository) Save(ctx context.Context, e *activity.UserEncounter) error {
	var wid interface{}
	if e.WorldID != "" {
		wid = e.WorldID
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO user_encounters (id, vrc_user_id, display_name, action, instance_id, world_id, encountered_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.VRCUserID, e.DisplayName, e.Action, e.InstanceID, wid, e.EncounteredAt.Format(time.RFC3339))
	return err
}

// DeleteOlderThan removes encounters older than before.
func (r *UserEncounterRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM user_encounters WHERE encountered_at < ?`, before.Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DeleteAll removes all encounters.
func (r *UserEncounterRepository) DeleteAll(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM user_encounters`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Count returns the number of encounters.
func (r *UserEncounterRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_encounters`).Scan(&n)
	return n, err
}

func scanUserEncounter(rows *sql.Rows) (*activity.UserEncounter, error) {
	var id, vrcUserID, displayName, action, instanceID string
	var worldID sql.NullString
	var encounteredAt string
	if err := rows.Scan(&id, &vrcUserID, &displayName, &action, &instanceID, &worldID, &encounteredAt); err != nil {
		return nil, err
	}
	t, _ := time.Parse(time.RFC3339, encounteredAt)
	wid := ""
	if worldID.Valid {
		wid = worldID.String
	}
	return &activity.UserEncounter{
		ID:            id,
		VRCUserID:     vrcUserID,
		DisplayName:   displayName,
		Action:        action,
		InstanceID:    instanceID,
		WorldID:       wid,
		EncounteredAt: t,
	}, nil
}

func scanEncounterWithContextRow(rows *sql.Rows) (*activity.EncounterWithContext, error) {
	var id, vrcUserID, displayName, action, instanceID string
	var worldID sql.NullString
	var encounteredAt string
	var worldDN, firstSeen, lastContact sql.NullString
	if err := rows.Scan(&id, &vrcUserID, &displayName, &action, &instanceID, &worldID, &encounteredAt,
		&worldDN, &firstSeen, &lastContact); err != nil {
		return nil, err
	}
	t, _ := time.Parse(time.RFC3339, encounteredAt)
	wid := ""
	if worldID.Valid {
		wid = worldID.String
	}
	enc := &activity.UserEncounter{
		ID:            id,
		VRCUserID:     vrcUserID,
		DisplayName:   displayName,
		Action:        action,
		InstanceID:    instanceID,
		WorldID:       wid,
		EncounteredAt: t,
	}
	out := &activity.EncounterWithContext{Encounter: enc}
	if worldDN.Valid {
		out.WorldDisplayName = worldDN.String
	}
	if firstSeen.Valid {
		if ft, err := time.Parse(time.RFC3339, firstSeen.String); err == nil {
			out.UserFirstSeenAt = &ft
		}
	}
	if lastContact.Valid {
		if lt, err := time.Parse(time.RFC3339, lastContact.String); err == nil {
			out.UserLastContactAt = &lt
		}
	}
	if out.UserFirstSeenAt != nil && enc.EncounteredAt.Equal(*out.UserFirstSeenAt) {
		out.IsFirstEncounter = true
	} else if out.UserFirstSeenAt != nil {
		d := enc.EncounteredAt.Sub(*out.UserFirstSeenAt)
		if d >= 0 && d < time.Second {
			out.IsFirstEncounter = true
		}
	}
	return out, nil
}

func scanPlaySessionRow(row *sql.Row) (*activity.PlaySession, error) {
	var id string
	var startTime string
	var endTime sql.NullString
	var durSec sql.NullInt64
	err := row.Scan(&id, &startTime, &endTime, &durSec)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	st, _ := time.Parse(time.RFC3339, startTime)
	var et *time.Time
	if endTime.Valid && endTime.String != "" {
		t, _ := time.Parse(time.RFC3339, endTime.String)
		et = &t
	}
	var ds *int
	if durSec.Valid {
		d := int(durSec.Int64)
		ds = &d
	}
	return &activity.PlaySession{
		ID:          id,
		StartTime:   st,
		EndTime:     et,
		DurationSec: ds,
	}, nil
}
