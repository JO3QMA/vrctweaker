package logwatcher

import (
	"context"
	"log"
	"sync"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/usecase"
)

// ActivityEventHandler bridges parsed log events to ActivityUseCase.
type ActivityEventHandler struct {
	uc     *usecase.ActivityUseCase
	ctx    context.Context
	logger Logger

	mu                sync.Mutex
	currentInstanceID string
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
		inst := e.InstanceID
		if inst == "" {
			h.mu.Lock()
			inst = h.currentInstanceID
			h.mu.Unlock()
		}
		if err := h.uc.RecordEncounterAt(h.ctx, e.VRCUserID, e.DisplayName, e.Action, inst, e.EncounteredAt); err != nil {
			h.logger.Printf("[activity_handler] RecordEncounter: %v", err)
		}
	case *activity.SessionEvent:
		switch e.Type {
		case activity.SessionEventStart:
			if e.InstanceID == "" {
				return
			}
			if err := h.uc.EndPlaySession(h.ctx, e.OccurredAt); err != nil {
				h.logger.Printf("[activity_handler] EndPlaySession (before new instance): %v", err)
			}
			h.mu.Lock()
			h.currentInstanceID = e.InstanceID
			h.mu.Unlock()
			if err := h.uc.StartPlaySession(h.ctx, e.InstanceID, e.OccurredAt); err != nil {
				h.logger.Printf("[activity_handler] StartPlaySession: %v", err)
			}
		case activity.SessionEventEnd:
			h.mu.Lock()
			h.currentInstanceID = ""
			h.mu.Unlock()
			if err := h.uc.EndPlaySession(h.ctx, e.OccurredAt); err != nil {
				h.logger.Printf("[activity_handler] EndPlaySession: %v", err)
			}
		}
	}
}
