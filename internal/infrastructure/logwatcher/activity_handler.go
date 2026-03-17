package logwatcher

import (
	"context"
	"log"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/usecase"
)

// ActivityEventHandler bridges parsed log events to ActivityUseCase.
type ActivityEventHandler struct {
	uc     *usecase.ActivityUseCase
	ctx    context.Context
	logger Logger
}

// NewActivityEventHandler creates a handler that persists events via ActivityUseCase.
func NewActivityEventHandler(uc *usecase.ActivityUseCase, ctx context.Context, logger Logger) *ActivityEventHandler {
	if logger == nil {
		logger = logWriterLogger{log.Default()}
	}
	return &ActivityEventHandler{
		uc:     uc,
		ctx:    ctx,
		logger: logger,
	}
}

type logWriterLogger struct {
	*log.Logger
}

func (l logWriterLogger) Printf(format string, args ...interface{}) {
	l.Logger.Printf(format, args...)
}

// Handle implements EventHandler.
func (h *ActivityEventHandler) Handle(event activity.ParsedEvent) {
	if event == nil {
		return
	}
	switch e := event.(type) {
	case *activity.EncounterEvent:
		if err := h.uc.RecordEncounterAt(h.ctx, e.VRCUserID, e.DisplayName, e.Action, e.InstanceID, e.EncounteredAt); err != nil {
			h.logger.Printf("[activity_handler] RecordEncounter: %v", err)
		}
	case *activity.SessionEvent:
		switch e.Type {
		case activity.SessionEventStart:
			if err := h.uc.StartPlaySession(h.ctx, e.InstanceID, e.OccurredAt); err != nil {
				h.logger.Printf("[activity_handler] StartPlaySession: %v", err)
			}
		case activity.SessionEventEnd:
			if err := h.uc.EndPlaySession(h.ctx, e.OccurredAt); err != nil {
				h.logger.Printf("[activity_handler] EndPlaySession: %v", err)
			}
		}
	}
}
