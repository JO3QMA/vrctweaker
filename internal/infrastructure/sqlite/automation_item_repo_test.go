package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrateAutomationRules_copiesLegacyWithoutOpenCursorWrite(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if schemaErr := applySchema(db); schemaErr != nil {
		t.Fatal(schemaErr)
	}
	ctx := context.Background()
	_, err = db.ExecContext(ctx, `INSERT INTO automation_rules (id, name, trigger_type, condition_json, action_type, action_payload, is_enabled)
		VALUES ('legacy-1', 'L', 'friend_joined', '{"vrc_user_id":"usr_x"}', 'change_status', '{"status":"ask me"}', 1)`)
	if err != nil {
		t.Fatal(err)
	}
	if err := MigrateAutomationRules(ctx, db); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM automation_items WHERE id = 'legacy-1'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("want migrated item, got count=%d", n)
	}
}
