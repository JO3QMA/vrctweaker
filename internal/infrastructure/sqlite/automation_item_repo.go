package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"

	"vrchat-tweaker/internal/domain/automation"
)

// AutomationItemRepository persists automation items.
type AutomationItemRepository struct {
	db *sql.DB
}

// NewAutomationItemRepository creates a repository.
func NewAutomationItemRepository(db *sql.DB) *AutomationItemRepository {
	return &AutomationItemRepository{db: db}
}

const automationItemCols = `id, name, kind, is_enabled, trigger_type, schedule_json, conditions_json, actions_json, script_source`

// List returns all automation items ordered by name.
func (r *AutomationItemRepository) List(ctx context.Context) ([]*automation.AutomationItem, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+automationItemCols+` FROM automation_items ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var list []*automation.AutomationItem
	for rows.Next() {
		item, err := scanAutomationItem(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

// ListEnabled returns enabled items.
func (r *AutomationItemRepository) ListEnabled(ctx context.Context) ([]*automation.AutomationItem, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+automationItemCols+` FROM automation_items WHERE is_enabled = 1 ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var list []*automation.AutomationItem
	for rows.Next() {
		item, err := scanAutomationItem(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

// GetByID returns one item.
func (r *AutomationItemRepository) GetByID(ctx context.Context, id string) (*automation.AutomationItem, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+automationItemCols+` FROM automation_items WHERE id = ?`, id)
	return scanAutomationItemRow(row)
}

// Save persists an item.
func (r *AutomationItemRepository) Save(ctx context.Context, item *automation.AutomationItem) error {
	isEnabled := 0
	if item.IsEnabled {
		isEnabled = 1
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO automation_items (id, name, kind, is_enabled, trigger_type, schedule_json, conditions_json, actions_json, script_source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET
		name = excluded.name, kind = excluded.kind, is_enabled = excluded.is_enabled,
		trigger_type = excluded.trigger_type, schedule_json = excluded.schedule_json,
		conditions_json = excluded.conditions_json, actions_json = excluded.actions_json,
		script_source = excluded.script_source`,
		item.ID, item.Name, item.Kind, isEnabled, item.TriggerType, item.ScheduleJSON,
		item.ConditionsJSON, item.ActionsJSON, item.ScriptSource)
	return err
}

// Delete removes an item.
func (r *AutomationItemRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM automation_items WHERE id = ?`, id)
	return err
}

func scanAutomationItem(rows *sql.Rows) (*automation.AutomationItem, error) {
	var id, name, kind, triggerType, scheduleJSON, conditionsJSON, actionsJSON, scriptSource sql.NullString
	var isEnabled int
	if err := rows.Scan(&id, &name, &kind, &isEnabled, &triggerType, &scheduleJSON, &conditionsJSON, &actionsJSON, &scriptSource); err != nil {
		return nil, err
	}
	return &automation.AutomationItem{
		ID:             id.String,
		Name:           name.String,
		Kind:           kind.String,
		IsEnabled:      isEnabled == 1,
		TriggerType:    triggerType.String,
		ScheduleJSON:   scheduleJSON.String,
		ConditionsJSON: conditionsJSON.String,
		ActionsJSON:    actionsJSON.String,
		ScriptSource:   scriptSource.String,
	}, nil
}

func scanAutomationItemRow(row *sql.Row) (*automation.AutomationItem, error) {
	var id, name, kind, triggerType, scheduleJSON, conditionsJSON, actionsJSON, scriptSource sql.NullString
	var isEnabled int
	err := row.Scan(&id, &name, &kind, &isEnabled, &triggerType, &scheduleJSON, &conditionsJSON, &actionsJSON, &scriptSource)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &automation.AutomationItem{
		ID:             id.String,
		Name:           name.String,
		Kind:           kind.String,
		IsEnabled:      isEnabled == 1,
		TriggerType:    triggerType.String,
		ScheduleJSON:   scheduleJSON.String,
		ConditionsJSON: conditionsJSON.String,
		ActionsJSON:    actionsJSON.String,
		ScriptSource:   scriptSource.String,
	}, nil
}

// MigrateAutomationRules copies legacy automation_rules into automation_items once.
func MigrateAutomationRules(ctx context.Context, db *sql.DB) error {
	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM automation_items`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	rows, err := db.QueryContext(ctx, `SELECT id, name, trigger_type, condition_json, action_type, action_payload, is_enabled FROM automation_rules`)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	repo := NewAutomationItemRepository(db)
	for rows.Next() {
		var rule automation.AutomationRule
		var isEnabled int
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.TriggerType, &rule.ConditionJSON, &rule.ActionType, &rule.ActionPayload, &isEnabled); err != nil {
			return err
		}
		rule.IsEnabled = isEnabled == 1
		item := automation.RuleToItem(&rule)
		if item == nil {
			continue
		}
		if err := repo.Save(ctx, item); err != nil {
			return err
		}
	}
	return rows.Err()
}

// ItemToLegacyRule maps a rule item to the old DTO for compat bindings.
func ItemToLegacyRule(item *automation.AutomationItem) *automation.AutomationRule {
	if item == nil || item.Kind != automation.KindRule {
		return nil
	}
	steps, err := automation.ParseActions(item.ActionsJSON)
	if err != nil || len(steps) == 0 {
		return &automation.AutomationRule{
			ID: item.ID, Name: item.Name, TriggerType: item.TriggerType,
			ConditionJSON: item.ConditionsJSON, IsEnabled: item.IsEnabled,
		}
	}
	payload, _ := json.Marshal(steps[0].Payload)
	return &automation.AutomationRule{
		ID:            item.ID,
		Name:          item.Name,
		TriggerType:   item.TriggerType,
		ConditionJSON: item.ConditionsJSON,
		ActionType:    steps[0].Type,
		ActionPayload: string(payload),
		IsEnabled:     item.IsEnabled,
	}
}
