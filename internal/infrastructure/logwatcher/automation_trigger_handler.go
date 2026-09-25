package logwatcher

import (
	"context"

	"vrchat-tweaker/internal/domain/activity"
)

// FriendEncounterAutomation runs automation for log-derived friend join/leave in the instance.
type FriendEncounterAutomation interface {
	OnFriendJoined(ctx context.Context, vrcUserID string) error
	OnFriendLeft(ctx context.Context, vrcUserID string) error
}

// AutomationTriggerHandler invokes automation for log-derived trigger events.
// Additional triggers (e.g. afk_detected) should add matching branches here.
type AutomationTriggerHandler struct {
	automation FriendEncounterAutomation
	ctx        context.Context
	logger     Logger
}

// NewAutomationTriggerHandler creates a handler that calls automation directly.
func NewAutomationTriggerHandler(automation FriendEncounterAutomation, ctx context.Context, logger Logger) *AutomationTriggerHandler {
	if logger == nil {
		logger = Std()
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
		if e.VRCUserID == "" {
			return
		}
		switch e.Action {
		case activity.EncounterActionJoin:
			if err := h.automation.OnFriendJoined(h.ctx, e.VRCUserID); err != nil {
				h.logger("[automation_trigger_handler] friend_joined: %v", err)
			}
		case activity.EncounterActionLeave:
			if err := h.automation.OnFriendLeft(h.ctx, e.VRCUserID); err != nil {
				h.logger("[automation_trigger_handler] friend_left: %v", err)
			}
		}
	}
}
