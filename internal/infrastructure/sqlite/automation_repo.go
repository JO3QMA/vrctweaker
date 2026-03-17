package sqlite

import (
	"context"
	"database/sql"

	"vrchat-tweaker/internal/domain/automation"
)

var _ automation.AutomationRuleRepository = (*AutomationRuleRepository)(nil)

// AutomationRuleRepository implements automation.AutomationRuleRepository.
type AutomationRuleRepository struct {
	db *sql.DB
}

// NewAutomationRuleRepository creates a new AutomationRuleRepository.
func NewAutomationRuleRepository(db *sql.DB) *AutomationRuleRepository {
	return &AutomationRuleRepository{db: db}
}

// List returns all automation rules.
func (r *AutomationRuleRepository) List(ctx context.Context) ([]*automation.AutomationRule, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, trigger_type, condition_json, action_type, action_payload, is_enabled FROM automation_rules ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var list []*automation.AutomationRule
	for rows.Next() {
		rule, err := scanAutomationRule(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, rule)
	}
	return list, rows.Err()
}

// ListEnabled returns only enabled rules.
func (r *AutomationRuleRepository) ListEnabled(ctx context.Context) ([]*automation.AutomationRule, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, trigger_type, condition_json, action_type, action_payload, is_enabled FROM automation_rules WHERE is_enabled = 1 ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var list []*automation.AutomationRule
	for rows.Next() {
		rule, err := scanAutomationRule(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, rule)
	}
	return list, rows.Err()
}

// GetByID returns a rule by ID.
func (r *AutomationRuleRepository) GetByID(ctx context.Context, id string) (*automation.AutomationRule, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, trigger_type, condition_json, action_type, action_payload, is_enabled FROM automation_rules WHERE id = ?`, id)
	return scanAutomationRuleRow(row)
}

// Save persists a rule.
func (r *AutomationRuleRepository) Save(ctx context.Context, rule *automation.AutomationRule) error {
	isEnabled := 0
	if rule.IsEnabled {
		isEnabled = 1
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO automation_rules (id, name, trigger_type, condition_json, action_type, action_payload, is_enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET
		name = excluded.name, trigger_type = excluded.trigger_type, condition_json = excluded.condition_json,
		action_type = excluded.action_type, action_payload = excluded.action_payload, is_enabled = excluded.is_enabled`,
		rule.ID, rule.Name, rule.TriggerType, rule.ConditionJSON, rule.ActionType, rule.ActionPayload, isEnabled)
	return err
}

// Delete removes a rule.
func (r *AutomationRuleRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM automation_rules WHERE id = ?`, id)
	return err
}

func scanAutomationRule(rows *sql.Rows) (*automation.AutomationRule, error) {
	var id, name, triggerType, conditionJSON, actionType, actionPayload string
	var isEnabled int
	if err := rows.Scan(&id, &name, &triggerType, &conditionJSON, &actionType, &actionPayload, &isEnabled); err != nil {
		return nil, err
	}
	return &automation.AutomationRule{
		ID:            id,
		Name:          name,
		TriggerType:   triggerType,
		ConditionJSON: conditionJSON,
		ActionType:    actionType,
		ActionPayload: actionPayload,
		IsEnabled:     isEnabled == 1,
	}, nil
}

func scanAutomationRuleRow(row *sql.Row) (*automation.AutomationRule, error) {
	var id, name, triggerType, conditionJSON, actionType, actionPayload string
	var isEnabled int
	err := row.Scan(&id, &name, &triggerType, &conditionJSON, &actionType, &actionPayload, &isEnabled)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &automation.AutomationRule{
		ID:            id,
		Name:          name,
		TriggerType:   triggerType,
		ConditionJSON: conditionJSON,
		ActionType:    actionType,
		ActionPayload: actionPayload,
		IsEnabled:     isEnabled == 1,
	}, nil
}
