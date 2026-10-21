package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// deleteOlderThanColumn removes rows whose stored RFC3339 timestamp is strictly
// before cutoff. SQLite datetime() converts numeric offsets and Z to UTC.
// A plain text comparison does not: a +09 local clock can sort after a Z cutoff
// while the instant is earlier, and a negative offset can sort before it while
// the instant is later.
func deleteOlderThanColumn(ctx context.Context, db *sql.DB, table, column string, before time.Time) (int64, error) {
	switch table {
	case "play_sessions", "user_encounters", "video_playback_history", "friend_avatar_observations":
	default:
		return 0, fmt.Errorf("delete older than: unsupported table %s", table)
	}
	switch column {
	case "start_time", "joined_at", "attempted_at", "observed_at":
	default:
		return 0, fmt.Errorf("delete older than: unsupported column %s", column)
	}
	query := `DELETE FROM ` + table + ` WHERE ` + column + ` != '' AND datetime(` + column + `) < datetime(?)`
	res, err := db.ExecContext(ctx, query, before.UTC().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Format(time.RFC3339)
}

func parseTime(n sql.NullString) *time.Time {
	if !n.Valid || n.String == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, n.String)
	if err != nil {
		return nil
	}
	return &t
}

func nullableInt64(p *int64) interface{} {
	if p == nil {
		return nil
	}
	return *p
}

func parseInt64Ptr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}
