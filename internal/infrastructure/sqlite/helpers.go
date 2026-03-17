package sqlite

import (
	"database/sql"
	"time"
)

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
