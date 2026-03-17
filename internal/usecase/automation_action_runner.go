package usecase

import (
	"context"

	"vrchat-tweaker/internal/domain/automation"
)

// ActionRunner executes automation actions (e.g. change_status).
type ActionRunner interface {
	Run(ctx context.Context, result *automation.EvalResult) error
}

// StatusSetter sets the user's VRChat status (for change_status action).
type StatusSetter interface {
	SetStatus(ctx context.Context, status string) error
}

// change_status で許可するステータス値
var allowedStatuses = map[string]bool{
	"busy":    true,
	"ask me":  true,
	"join me": true,
}

// DefaultActionRunner implements ActionRunner using StatusSetter for change_status.
type DefaultActionRunner struct {
	statusSetter StatusSetter
}

// NewDefaultActionRunner creates an ActionRunner that delegates change_status to the given StatusSetter.
func NewDefaultActionRunner(statusSetter StatusSetter) *DefaultActionRunner {
	return &DefaultActionRunner{statusSetter: statusSetter}
}

// Run executes the action described by result. Ignores nil or non-firing results.
func (r *DefaultActionRunner) Run(ctx context.Context, result *automation.EvalResult) error {
	if result == nil || !result.ShouldFire {
		return nil
	}
	switch result.ActionType {
	case automation.ActionChangeStatus:
		return r.runChangeStatus(ctx, result.ActionPayload)
	default:
		return nil
	}
}

func (r *DefaultActionRunner) runChangeStatus(ctx context.Context, payload map[string]interface{}) error {
	if payload == nil {
		return nil
	}
	s, _ := payload["status"].(string)
	if s == "" {
		return nil
	}
	if !allowedStatuses[s] {
		return nil
	}
	return r.statusSetter.SetStatus(ctx, s)
}
