package logwatcher

import (
	"context"
	"log"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/automation"
	"vrchat-tweaker/internal/domain/event"
)

// EventPublishingHandler publishes automation trigger events to EventBus.
// Used by log-derived events: friend_joined (on player join), afk_detected (future).
type EventPublishingHandler struct {
	eventBus event.EventBus
	ctx      context.Context
	logger   Logger
}

// NewEventPublishingHandler creates a handler that publishes to EventBus.
func NewEventPublishingHandler(eventBus event.EventBus, ctx context.Context, logger Logger) *EventPublishingHandler {
	if logger == nil {
		logger = logWriterLogger{log.Default()}
	}
	return &EventPublishingHandler{
		eventBus: eventBus,
		ctx:      ctx,
		logger:   logger,
	}
}

// Handle implements EventHandler.
func (h *EventPublishingHandler) Handle(ev activity.ParsedEvent) {
	if ev == nil {
		return
	}
	switch e := ev.(type) {
	case *activity.EncounterEvent:
		if e.Action == activity.EncounterActionJoin && e.VRCUserID != "" {
			payload := map[string]interface{}{"vrc_user_id": e.VRCUserID}
			if err := h.eventBus.Publish(h.ctx, automation.TriggerFriendJoined, &event.Event{
				Type:    automation.TriggerFriendJoined,
				Payload: payload,
			}); err != nil {
				h.logger.Printf("[event_publishing_handler] Publish friend_joined: %v", err)
			}
		}
	}
}
