package usecase

import (
	"context"

	"github.com/google/uuid"
	"vrchat-tweaker/internal/domain/automation"
)

// StatusSetter sets the user's VRChat status (for change_status action).
type StatusSetter interface {
	SetStatus(ctx context.Context, status string) error
}

// ponytail:#129 domain AutomationRuleRepository removed; boundary stays usecase-local.
type automationRuleRepo interface {
	List(ctx context.Context) ([]*automation.AutomationRule, error)
	ListEnabled(ctx context.Context) ([]*automation.AutomationRule, error)
	GetByID(ctx context.Context, id string) (*automation.AutomationRule, error)
	Save(ctx context.Context, r *automation.AutomationRule) error
	Delete(ctx context.Context, id string) error
}

// change_status で許可するステータス値
var allowedStatuses = map[string]bool{
	"busy":    true,
	"ask me":  true,
	"join me": true,
}

// AutomationUseCase handles automation rules and rule engine execution.
type AutomationUseCase struct {
	repo         automationRuleRepo
	statusSetter StatusSetter
}

// NewAutomationUseCase creates a new AutomationUseCase.
func NewAutomationUseCase(repo automationRuleRepo, statusSetter StatusSetter) *AutomationUseCase {
	return &AutomationUseCase{
		repo:         repo,
		statusSetter: statusSetter,
	}
}

// OnFriendJoined evaluates and runs automation rules for a friend join log event.
func (uc *AutomationUseCase) OnFriendJoined(ctx context.Context, vrcUserID string) error {
	if vrcUserID == "" {
		return nil
	}
	return uc.EvalAndRun(ctx, automation.TriggerFriendJoined, map[string]interface{}{
		"vrc_user_id": vrcUserID,
	})
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

// RunActions executes each EvalResult.
func (uc *AutomationUseCase) RunActions(ctx context.Context, results []*automation.EvalResult) error {
	for _, res := range results {
		_ = uc.runAction(ctx, res)
	}
	return nil
}

func (uc *AutomationUseCase) runAction(ctx context.Context, result *automation.EvalResult) error {
	if result == nil || !result.ShouldFire {
		return nil
	}
	switch result.ActionType {
	case automation.ActionChangeStatus:
		return uc.runChangeStatus(ctx, result.ActionPayload)
	default:
		return nil
	}
}

func (uc *AutomationUseCase) runChangeStatus(ctx context.Context, payload map[string]interface{}) error {
	if uc.statusSetter == nil || payload == nil {
		return nil
	}
	s, _ := payload["status"].(string)
	if s == "" {
		return nil
	}
	if !allowedStatuses[s] {
		return nil
	}
	return uc.statusSetter.SetStatus(ctx, s)
}

// EvalAndRun evaluates rules for the trigger and runs matching actions.
func (uc *AutomationUseCase) EvalAndRun(ctx context.Context, triggerType string, payload map[string]interface{}) error {
	results, err := uc.EvalRules(ctx, triggerType, payload)
	if err != nil {
		return err
	}
	return uc.RunActions(ctx, results)
}
