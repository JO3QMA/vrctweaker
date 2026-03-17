package usecase

import (
	"context"

	"github.com/google/uuid"
	"vrchat-tweaker/internal/domain/automation"
	"vrchat-tweaker/internal/domain/event"
)

// AutomationUseCase handles automation rules and rule engine execution.
type AutomationUseCase struct {
	repo         automation.AutomationRuleRepository
	eventBus     event.EventBus
	actionRunner ActionRunner
}

// NewAutomationUseCase creates a new AutomationUseCase.
func NewAutomationUseCase(repo automation.AutomationRuleRepository, eventBus event.EventBus, actionRunner ActionRunner) *AutomationUseCase {
	return &AutomationUseCase{
		repo:         repo,
		eventBus:     eventBus,
		actionRunner: actionRunner,
	}
}

// ListRules returns all automation rules.
func (uc *AutomationUseCase) ListRules(ctx context.Context) ([]*automation.AutomationRule, error) {
	return uc.repo.List(ctx)
}

// GetRule returns a rule by ID.
func (uc *AutomationUseCase) GetRule(ctx context.Context, id string) (*automation.AutomationRule, error) {
	return uc.repo.GetByID(ctx, id)
}

// SaveRule persists a rule.
func (uc *AutomationUseCase) SaveRule(ctx context.Context, r *automation.AutomationRule) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return uc.repo.Save(ctx, r)
}

// DeleteRule removes a rule.
func (uc *AutomationUseCase) DeleteRule(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

// ToggleRule enables or disables a rule.
func (uc *AutomationUseCase) ToggleRule(ctx context.Context, id string, enabled bool) error {
	r, err := uc.repo.GetByID(ctx, id)
	if err != nil || r == nil {
		return err
	}
	r.IsEnabled = enabled
	return uc.repo.Save(ctx, r)
}

// EvalRules evaluates enabled rules for the given trigger context.
func (uc *AutomationUseCase) EvalRules(ctx context.Context, triggerType string, payload map[string]interface{}) ([]*automation.EvalResult, error) {
	rules, err := uc.repo.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	evalCtx := &automation.EvalContext{TriggerType: triggerType, Payload: payload}
	var results []*automation.EvalResult
	for _, rule := range rules {
		res, err := automation.EvalRule(rule, evalCtx)
		if err != nil {
			continue
		}
		if res != nil && res.ShouldFire {
			results = append(results, res)
		}
	}
	return results, nil
}

// RunActions executes each EvalResult via ActionRunner.
func (uc *AutomationUseCase) RunActions(ctx context.Context, results []*automation.EvalResult) error {
	if uc.actionRunner == nil {
		return nil
	}
	for _, res := range results {
		_ = uc.actionRunner.Run(ctx, res)
	}
	return nil
}

// EvalAndRun evaluates rules for the trigger and runs matching actions.
func (uc *AutomationUseCase) EvalAndRun(ctx context.Context, triggerType string, payload map[string]interface{}) error {
	results, err := uc.EvalRules(ctx, triggerType, payload)
	if err != nil {
		return err
	}
	return uc.RunActions(ctx, results)
}
