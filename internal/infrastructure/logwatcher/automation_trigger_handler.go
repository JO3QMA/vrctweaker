package logwatcher

import (
	"context"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/infrastructure/diag"
)

// FriendJoinedAutomation runs automation when a friend joins the instance (log-derived).
type FriendJoinedAutomation interface {
	OnFriendJoined(ctx context.Context, vrcUserID string) error
}

// AutomationTriggerHandler invokes automation for log-derived trigger events.
// Additional triggers (e.g. afk_detected) should add matching branches here.
type AutomationTriggerHandler struct {
	automation FriendJoinedAutomation
	ctx        context.Context
	logger     diag.Logger
}

// NewAutomationTriggerHandler creates a handler that calls automation directly.
func NewAutomationTriggerHandler(automation FriendJoinedAutomation, ctx context.Context, logger diag.Logger) *AutomationTriggerHandler {
	if logger == nil {
		logger = diag.Std()
	}
	return &AutomationTriggerHandler{
		automation: automation,
		ctx:        ctx,
		logger:     logger,
	}
}

// Handle implements EventHandler.
func (h *AutomationTriggerHandler) Handle(ev activity.ParsedEvent) {
	if ev == nil {
		return
	}
	switch e := ev.(type) {
	case *activity.EncounterEvent:
		if e.Action == activity.EncounterActionJoin && e.VRCUserID != "" {
			if err := h.automation.OnFriendJoined(h.ctx, e.VRCUserID); err != nil {
				h.logger("[automation_trigger_handler] friend_joined: %v", err)
			}
		}
	}
}
