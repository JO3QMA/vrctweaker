package sqlite

import (
	"context"
	"database/sql"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

// PlaySessionRepository persists play sessions in SQLite.
type PlaySessionRepository struct {
	db *sql.DB
}

// NewPlaySessionRepository creates a new PlaySessionRepository.
func NewPlaySessionRepository(db *sql.DB) *PlaySessionRepository {
	return &PlaySessionRepository{db: db}
}

// List returns play sessions within the time range.
func (r *PlaySessionRepository) List(ctx context.Context, from, to time.Time) ([]*activity.PlaySession, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, start_time, end_time, duration_sec, IFNULL(instance_id, ''), IFNULL(log_source_path, '') FROM play_sessions WHERE start_time >= ? AND start_time <= ? ORDER BY start_time DESC`,
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
	row := r.db.QueryRowContext(ctx, `SELECT id, start_time, end_time, duration_sec, IFNULL(instance_id, ''), IFNULL(log_source_path, '') FROM play_sessions WHERE end_time IS NULL OR end_time = '' ORDER BY start_time DESC LIMIT 1`)
	return scanPlaySessionRow(row)
}

// FindOpenForLogSource returns the open play session for the given log source path.
func (r *PlaySessionRepository) FindOpenForLogSource(ctx context.Context, logSource string) (*activity.PlaySession, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, start_time, end_time, duration_sec, IFNULL(instance_id, ''), IFNULL(log_source_path, '')
		FROM play_sessions WHERE (end_time IS NULL OR end_time = '') AND log_source_path = ?
		ORDER BY start_time DESC LIMIT 1`, logSource)
	return scanPlaySessionRow(row)
}

// FindAllWithoutEndTime returns all play sessions with no end time.
func (r *PlaySessionRepository) FindAllWithoutEndTime(ctx context.Context) ([]*activity.PlaySession, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, start_time, end_time, duration_sec, IFNULL(instance_id, ''), IFNULL(log_source_path, '')
		FROM play_sessions WHERE end_time IS NULL OR end_time = '' ORDER BY start_time ASC`)
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

// Count returns the number of play sessions.
func (r *PlaySessionRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM play_sessions`).Scan(&n)
	return n, err
}

// DeleteOlderThan removes play sessions older than before (by start_time).
func (r *PlaySessionRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM play_sessions WHERE start_time < ?`, before.Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// GetByID returns a play session by ID.
func (r *PlaySessionRepository) GetByID(ctx context.Context, id string) (*activity.PlaySession, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, start_time, end_time, duration_sec, IFNULL(instance_id, ''), IFNULL(log_source_path, '') FROM play_sessions WHERE id = ?`, id)
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
	_, err := r.db.ExecContext(ctx, `INSERT INTO play_sessions (id, start_time, end_time, duration_sec, instance_id, log_source_path) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET start_time = excluded.start_time, end_time = excluded.end_time, duration_sec = excluded.duration_sec,
		instance_id = excluded.instance_id, log_source_path = excluded.log_source_path`,
		s.ID, s.StartTime.Format(time.RFC3339), endTime, dur, nullIfEmpty(s.InstanceID), nullIfEmpty(s.LogSourcePath))
	return err
}

func scanPlaySession(rows *sql.Rows) (*activity.PlaySession, error) {
	var id, instanceID, logSource string
	var startTime string
	var endTime sql.NullString
	var durSec sql.NullInt64
	if err := rows.Scan(&id, &startTime, &endTime, &durSec, &instanceID, &logSource); err != nil {
		return nil, err
	}
	return buildPlaySession(id, startTime, endTime, durSec, instanceID, logSource)
}

func scanPlaySessionRow(row *sql.Row) (*activity.PlaySession, error) {
	var id, instanceID, logSource string
	var startTime string
	var endTime sql.NullString
	var durSec sql.NullInt64
	if err := row.Scan(&id, &startTime, &endTime, &durSec, &instanceID, &logSource); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return buildPlaySession(id, startTime, endTime, durSec, instanceID, logSource)
}

func buildPlaySession(id, startTime string, endTime sql.NullString, durSec sql.NullInt64, instanceID, logSource string) (*activity.PlaySession, error) {
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
		ID:            id,
		StartTime:     st,
		EndTime:       et,
		DurationSec:   ds,
		InstanceID:    instanceID,
		LogSourcePath: logSource,
	}, nil
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// UserEncounterRepository persists user encounters in SQLite.
type UserEncounterRepository struct {
	db *sql.DB
}

// NewUserEncounterRepository creates a new UserEncounterRepository.
func NewUserEncounterRepository(db *sql.DB) *UserEncounterRepository {
	return &UserEncounterRepository{db: db}
}

// List returns user encounters with optional filters.
func (r *UserEncounterRepository) List(ctx context.Context, filter *activity.EncounterFilter) ([]*activity.UserEncounter, error) {
	query := `SELECT id, vrc_user_id, display_name, instance_id, world_id, joined_at, left_at, IFNULL(log_source_path, '') FROM user_encounters WHERE 1=1`
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
			query += ` AND joined_at >= ?`
			args = append(args, filter.From.Format(time.RFC3339))
		}
		if filter.To != nil {
			query += ` AND joined_at <= ?`
			args = append(args, filter.To.Format(time.RFC3339))
		}
	}
	query += ` ORDER BY joined_at DESC`

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
	query := `SELECT e.id, e.vrc_user_id, e.display_name, e.instance_id, e.world_id, e.joined_at, e.left_at, IFNULL(e.log_source_path, ''),
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
			query += ` AND e.joined_at >= ?`
			args = append(args, filter.From.Format(time.RFC3339))
		}
		if filter.To != nil {
			query += ` AND e.joined_at <= ?`
			args = append(args, filter.To.Format(time.RFC3339))
		}
	}
	query += ` ORDER BY e.joined_at DESC`

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

// Save persists a user encounter (insert only).
func (r *UserEncounterRepository) Save(ctx context.Context, e *activity.UserEncounter) error {
	var wid interface{}
	if e.WorldID != "" {
		wid = e.WorldID
	}
	leftAt := interface{}(nil)
	if e.LeftAt != nil {
		leftAt = e.LeftAt.Format(time.RFC3339)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO user_encounters (id, vrc_user_id, display_name, instance_id, world_id, joined_at, left_at, log_source_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.VRCUserID, e.DisplayName, e.InstanceID, wid, e.JoinedAt.Format(time.RFC3339), leftAt, nullIfEmpty(e.LogSourcePath))
	return err
}

// FindByVRCUserIDAndJoinedAt returns an encounter for the user at the exact join time.
func (r *UserEncounterRepository) FindByVRCUserIDAndJoinedAt(ctx context.Context, vrcUserID string, joinedAt time.Time) (*activity.UserEncounter, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, vrc_user_id, display_name, instance_id, world_id, joined_at, left_at, IFNULL(log_source_path, '')
		FROM user_encounters WHERE vrc_user_id = ? AND joined_at = ? LIMIT 1`,
		vrcUserID, joinedAt.Format(time.RFC3339))
	return scanUserEncounterRow(row)
}

// UpdateEncounter patches display name and fills empty instance/world on an existing row.
func (r *UserEncounterRepository) UpdateEncounter(ctx context.Context, e *activity.UserEncounter) error {
	var wid interface{}
	if e.WorldID != "" {
		wid = e.WorldID
	}
	_, err := r.db.ExecContext(ctx, `UPDATE user_encounters SET
		display_name = ?,
		instance_id = CASE WHEN (instance_id IS NULL OR instance_id = '') AND ? != '' THEN ? ELSE instance_id END,
		world_id = CASE WHEN (world_id IS NULL OR world_id = '') AND ? IS NOT NULL THEN ? ELSE world_id END
		WHERE id = ?`,
		e.DisplayName, e.InstanceID, e.InstanceID, wid, wid, e.ID)
	return err
}

// CloseEncounterLeave sets left_at on the latest eligible open stay for the user.
func (r *UserEncounterRepository) CloseEncounterLeave(ctx context.Context, vrcUserID, instanceID string, leftAt time.Time) (int64, error) {
	leftStr := leftAt.Format(time.RFC3339)
	joinedCutoff := leftStr
	var id string
	if instanceID != "" {
		err := r.db.QueryRowContext(ctx, `SELECT id FROM user_encounters
			WHERE vrc_user_id = ? AND (left_at IS NULL OR left_at = '') AND joined_at <= ?
			AND instance_id = ?
			ORDER BY joined_at DESC LIMIT 1`,
			vrcUserID, joinedCutoff, instanceID).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			return 0, err
		}
	}
	if id == "" {
		err := r.db.QueryRowContext(ctx, `SELECT id FROM user_encounters
			WHERE vrc_user_id = ? AND (left_at IS NULL OR left_at = '') AND joined_at <= ?
			ORDER BY joined_at DESC LIMIT 1`,
			vrcUserID, joinedCutoff).Scan(&id)
		if err == sql.ErrNoRows {
			return 0, nil
		}
		if err != nil {
			return 0, err
		}
	}
	res, err := r.db.ExecContext(ctx, `UPDATE user_encounters SET left_at = ? WHERE id = ?`,
		leftStr, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// CloseOpenEncountersAtForLogSource sets left_at on open rows for the given log source.
func (r *UserEncounterRepository) CloseOpenEncountersAtForLogSource(ctx context.Context, logSource string, at time.Time) (int64, error) {
	atStr := at.Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, `UPDATE user_encounters SET left_at = ?
		WHERE (left_at IS NULL OR left_at = '') AND joined_at <= ? AND log_source_path = ?`,
		atStr, atStr, logSource)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// CloseOpenEncountersAt sets left_at on every row still open with joined_at <= at.
func (r *UserEncounterRepository) CloseOpenEncountersAt(ctx context.Context, at time.Time) (int64, error) {
	atStr := at.Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, `UPDATE user_encounters SET left_at = ?
		WHERE (left_at IS NULL OR left_at = '') AND joined_at <= ?`,
		atStr, atStr)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DeleteOlderThan removes encounters older than before.
func (r *UserEncounterRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM user_encounters WHERE joined_at < ?`, before.Format(time.RFC3339))
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

// BackfillMissingWorldContext sets world_id on rows with missing context.
func (r *UserEncounterRepository) BackfillMissingWorldContext(ctx context.Context) (int64, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, IFNULL(world_id, ''), IFNULL(instance_id, '') FROM user_encounters ORDER BY joined_at ASC, id ASC`)
	if err != nil {
		return 0, err
	}
	type row struct {
		id, wid, inst string
	}
	var list []row
	for rows.Next() {
		var rec row
		if scanErr := rows.Scan(&rec.id, &rec.wid, &rec.inst); scanErr != nil {
			_ = rows.Close()
			return 0, scanErr
		}
		list = append(list, rec)
	}
	if closeErr := rows.Close(); closeErr != nil {
		return 0, closeErr
	}
	if err2 := rows.Err(); err2 != nil {
		return 0, err2
	}

	var lastWid, lastInst string
	var updates [][3]string
	for _, rec := range list {
		if rec.wid != "" {
			lastWid = rec.wid
			if rec.inst != "" {
				lastInst = rec.inst
			}
			continue
		}
		if lastWid == "" {
			continue
		}
		fillInst := rec.inst
		if fillInst == "" {
			fillInst = lastInst
		}
		updates = append(updates, [3]string{rec.id, lastWid, fillInst})
		if fillInst != "" {
			lastInst = fillInst
		}
	}
	if len(updates) == 0 {
		return 0, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	var n int64
	for _, u := range updates {
		res, execErr := tx.ExecContext(ctx, `UPDATE user_encounters SET world_id = ?, instance_id = ? WHERE id = ?`, u[1], u[2], u[0])
		if execErr != nil {
			return 0, execErr
		}
		k, _ := res.RowsAffected()
		n += k
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return n, nil
}

// DeduplicateEncounters merges rows sharing vrc_user_id and joined_at and fixes invalid left_at.
func (r *UserEncounterRepository) DeduplicateEncounters(ctx context.Context) (int64, error) {
	list, err := r.List(ctx, nil)
	if err != nil {
		return 0, err
	}
	type key struct {
		vrcUserID string
		join      string
	}
	groups := make(map[key][]*activity.UserEncounter)
	for _, e := range list {
		k := key{vrcUserID: e.VRCUserID, join: e.JoinedAt.Format(time.RFC3339)}
		groups[k] = append(groups[k], e)
	}

	var affected int64
	for _, rows := range groups {
		if len(rows) <= 1 {
			e := rows[0]
			if e.LeftAt != nil && e.LeftAt.Before(e.JoinedAt) {
				if _, err := r.db.ExecContext(ctx, `UPDATE user_encounters SET left_at = NULL WHERE id = ?`, e.ID); err != nil {
					return affected, err
				}
				affected++
			}
			continue
		}
		keep := pickEncounterToKeep(rows)
		for _, e := range rows {
			if e.ID == keep.ID {
				if keep.LeftAt != nil && keep.LeftAt.Before(keep.JoinedAt) {
					if _, err := r.db.ExecContext(ctx, `UPDATE user_encounters SET left_at = NULL WHERE id = ?`, keep.ID); err != nil {
						return affected, err
					}
					affected++
				}
				continue
			}
			if _, err := r.db.ExecContext(ctx, `DELETE FROM user_encounters WHERE id = ?`, e.ID); err != nil {
				return affected, err
			}
			affected++
		}
	}
	return affected, nil
}

func pickEncounterToKeep(rows []*activity.UserEncounter) *activity.UserEncounter {
	best := rows[0]
	bestScore := encounterKeepScore(best)
	for _, e := range rows[1:] {
		if s := encounterKeepScore(e); s > bestScore {
			best = e
			bestScore = s
		}
	}
	return best
}

func encounterKeepScore(e *activity.UserEncounter) int {
	score := 0
	if e.InstanceID != "" {
		score += 4
	}
	if e.WorldID != "" {
		score += 2
	}
	if e.LeftAt != nil {
		if !e.LeftAt.Before(e.JoinedAt) {
			score += 8
			score += int(e.LeftAt.Sub(e.JoinedAt).Seconds())
		}
	}
	return score
}

func scanUserEncounterRow(row *sql.Row) (*activity.UserEncounter, error) {
	var id, vrcUserID, displayName, joinedAtStr, logSource string
	var instanceID, worldID, leftAt sql.NullString
	err := row.Scan(&id, &vrcUserID, &displayName, &instanceID, &worldID, &joinedAtStr, &leftAt, &logSource)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return buildUserEncounter(id, vrcUserID, displayName, instanceID, worldID, joinedAtStr, leftAt, logSource)
}

func scanUserEncounter(rows *sql.Rows) (*activity.UserEncounter, error) {
	var id, vrcUserID, displayName, joinedAtStr, logSource string
	var instanceID, worldID, leftAt sql.NullString
	if err := rows.Scan(&id, &vrcUserID, &displayName, &instanceID, &worldID, &joinedAtStr, &leftAt, &logSource); err != nil {
		return nil, err
	}
	return buildUserEncounter(id, vrcUserID, displayName, instanceID, worldID, joinedAtStr, leftAt, logSource)
}

func buildUserEncounter(id, vrcUserID, displayName string, instanceID, worldID sql.NullString, joinedAtStr string, leftAt sql.NullString, logSource string) (*activity.UserEncounter, error) {
	jt, _ := time.Parse(time.RFC3339, joinedAtStr)
	inst := ""
	if instanceID.Valid {
		inst = instanceID.String
	}
	wid := ""
	if worldID.Valid {
		wid = worldID.String
	}
	var lt *time.Time
	if leftAt.Valid && leftAt.String != "" {
		t, _ := time.Parse(time.RFC3339, leftAt.String)
		lt = &t
	}
	return &activity.UserEncounter{
		ID:            id,
		VRCUserID:     vrcUserID,
		DisplayName:   displayName,
		InstanceID:    inst,
		WorldID:       wid,
		LogSourcePath: logSource,
		JoinedAt:      jt,
		LeftAt:        lt,
	}, nil
}

func scanEncounterWithContextRow(rows *sql.Rows) (*activity.EncounterWithContext, error) {
	var id, vrcUserID, displayName, joinedAtStr, logSource string
	var instanceID, worldID, leftAt sql.NullString
	var worldDN, firstSeen, lastContact sql.NullString
	if err := rows.Scan(&id, &vrcUserID, &displayName, &instanceID, &worldID, &joinedAtStr, &leftAt, &logSource,
		&worldDN, &firstSeen, &lastContact); err != nil {
		return nil, err
	}
	jt, _ := time.Parse(time.RFC3339, joinedAtStr)
	inst := ""
	if instanceID.Valid {
		inst = instanceID.String
	}
	wid := ""
	if worldID.Valid {
		wid = worldID.String
	}
	var lt *time.Time
	if leftAt.Valid && leftAt.String != "" {
		t, _ := time.Parse(time.RFC3339, leftAt.String)
		lt = &t
	}
	enc := &activity.UserEncounter{
		ID:            id,
		VRCUserID:     vrcUserID,
		DisplayName:   displayName,
		InstanceID:    inst,
		WorldID:       wid,
		LogSourcePath: logSource,
		JoinedAt:      jt,
		LeftAt:        lt,
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
		if lt2, err := time.Parse(time.RFC3339, lastContact.String); err == nil {
			out.UserLastContactAt = &lt2
		}
	}
	if out.UserFirstSeenAt != nil && enc.JoinedAt.Equal(*out.UserFirstSeenAt) {
		out.IsFirstEncounter = true
	} else if out.UserFirstSeenAt != nil {
		d := enc.JoinedAt.Sub(*out.UserFirstSeenAt)
		if d >= 0 && d < time.Second {
			out.IsFirstEncounter = true
		}
	}
	return out, nil
}
