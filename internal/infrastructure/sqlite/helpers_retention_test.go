package sqlite

import (
	"context"
	"testing"
	"time"
)

func TestDeleteOlderThanColumn_rejectsMismatchedPair(t *testing.T) {
	_, err := deleteOlderThanColumn(context.Background(), nil, "play_sessions", "observed_at", time.Now())
	if err == nil {
		t.Fatal("mismatched table/column should fail before SQL")
	}
	_, err = deleteOlderThanColumn(context.Background(), nil, "not_a_table", "start_time", time.Now())
	if err == nil {
		t.Fatal("unknown table should fail before SQL")
	}
}
