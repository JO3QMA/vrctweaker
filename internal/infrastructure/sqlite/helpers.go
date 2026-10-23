package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// activityRetentionTimeColumn is the only legal table/column pair for a retention delete.
// Separate allowlists would accept a mismatched pair and fail later inside SQLite.
var activityRetentionTimeColumn = map[string]string{
	"play_sessions":              "start_time",
	"user_encounters":            "joined_at",
	"video_playback_history":     "attempted_at",
	"friend_avatar_observations": "observed_at",
}

// deleteOlderThanColumn removes rows whose stored RFC3339 timestamp is strictly
// before cutoff. SQLite datetime() converts numeric offsets and Z to UTC.
// A plain text comparison does not: a +09 local clock can sort after a Z cutoff
// while the instant is earlier, and a negative offset can sort before it while
// the instant is later.
//
// datetime(column) cannot use idx_play_sessions_start_time, idx_user_encounters_joined_at,
// idx_video_playback_attempted_at, or idx_friend_avatar_obs_observed_at, so each purge
// scans the table. The stored values keep the log line's offset, so a raw range on the
// text is not chronological. These tables are one person's activity log; a later
// UTC-normalized column would be the index-friendly form.
func deleteOlderThanColumn(ctx context.Context, db *sql.DB, table, column string, before time.Time) (int64, error) {
	want, ok := activityRetentionTimeColumn[table]
	if !ok || want != column {
		return 0, fmt.Errorf("delete older than: unsupported table/column %s/%s", table, column)
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
