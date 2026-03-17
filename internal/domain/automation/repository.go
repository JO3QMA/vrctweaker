package automation

import "context"

// AutomationRuleRepository defines persistence operations for automation rules.
type AutomationRuleRepository interface {
	// List returns all automation rules.
	List(ctx context.Context) ([]*AutomationRule, error)
	// ListEnabled returns only enabled rules.
	ListEnabled(ctx context.Context) ([]*AutomationRule, error)
	// GetByID returns a rule by ID.
	GetByID(ctx context.Context, id string) (*AutomationRule, error)
	// Save persists a rule (create or update).
	Save(ctx context.Context, r *AutomationRule) error
	// Delete removes a rule by ID.
	Delete(ctx context.Context, id string) error
}
