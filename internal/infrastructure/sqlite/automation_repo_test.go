package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"vrchat-tweaker/internal/domain/automation"
)

func TestAutomationRuleRepository_CRUD(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if schemaErr := applySchema(db); schemaErr != nil {
		t.Fatal(schemaErr)
	}

	repo := NewAutomationRuleRepository(db)
	ctx := context.Background()

	rule := &automation.AutomationRule{
		ID:            "rule-1",
		Name:          "AFK Status",
		TriggerType:   "afk_detected",
		ConditionJSON: `{"minutes":5}`,
		ActionType:    "change_status",
		ActionPayload: `{"status":"ask me"}`,
		IsEnabled:     true,
	}
	if saveErr := repo.Save(ctx, rule); saveErr != nil {
		t.Fatal(saveErr)
	}

	got, err := repo.GetByID(ctx, "rule-1")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Name != "AFK Status" || !got.IsEnabled {
		t.Fatalf("GetByID: %#v", got)
	}

	list, err := repo.List(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("List: err=%v len=%d", err, len(list))
	}

	enabled, err := repo.ListEnabled(ctx)
	if err != nil || len(enabled) != 1 {
		t.Fatalf("ListEnabled: err=%v len=%d", err, len(enabled))
	}

	rule.IsEnabled = false
	if saveErr := repo.Save(ctx, rule); saveErr != nil {
		t.Fatal(saveErr)
	}
	enabled, err = repo.ListEnabled(ctx)
	if err != nil || len(enabled) != 0 {
		t.Fatalf("ListEnabled after disable: len=%d err=%v", len(enabled), err)
	}

	if delErr := repo.Delete(ctx, "rule-1"); delErr != nil {
		t.Fatal(delErr)
	}
	miss, err := repo.GetByID(ctx, "rule-1")
	if err != nil || miss != nil {
		t.Fatalf("after delete: got=%#v err=%v", miss, err)
	}
}

func TestAutomationRuleRepository_List_empty(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if schemaErr := applySchema(db); schemaErr != nil {
		t.Fatal(schemaErr)
	}
	repo := NewAutomationRuleRepository(db)
	ctx := context.Background()
	list, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("List empty: len=%d", len(list))
	}
	enabled, err := repo.ListEnabled(ctx)
	if err != nil || len(enabled) != 0 {
		t.Fatalf("ListEnabled empty: len=%d err=%v", len(enabled), err)
	}
}
