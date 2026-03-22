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
	uc               *usecase.ActivityUseCase
	ctx              context.Context
	logger           Logger
	onAfterEncounter func()

	mu                sync.Mutex
	currentInstanceID string
	currentWorldID    string
}

// NewActivityEventHandler creates a handler that persists events via ActivityUseCase.
// onAfterEncounter is optional (e.g. Wails EventsEmit after each encounter row).
func NewActivityEventHandler(uc *usecase.ActivityUseCase, ctx context.Context, logger Logger, onAfterEncounter func()) *ActivityEventHandler {
	if logger == nil {
		logger = logWriterLogger{log.Default()}
	}
	return &ActivityEventHandler{
		uc:               uc,
		ctx:              ctx,
		logger:           logger,
		onAfterEncounter: onAfterEncounter,
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
	case *activity.DestinationSetEvent:
		if err := h.uc.UpsertWorldVisit(h.ctx, e.WorldID, e.OccurredAt); err != nil {
			h.logger.Printf("[activity_handler] UpsertWorldVisit: %v", err)
		}
		h.mu.Lock()
		h.currentWorldID = e.WorldID
		if e.FullInstance != "" {
			h.currentInstanceID = e.FullInstance
		}
		h.mu.Unlock()
	case *activity.RoomNameEvent:
		h.mu.Lock()
		wid := h.currentWorldID
		h.mu.Unlock()
		if err := h.uc.UpsertWorldRoomName(h.ctx, wid, e.RoomName, e.OccurredAt); err != nil {
			h.logger.Printf("[activity_handler] UpsertWorldRoomName: %v", err)
		}
	case *activity.EncounterEvent:
		inst := e.InstanceID
		h.mu.Lock()
		if inst == "" {
			inst = h.currentInstanceID
		}
		wid := h.currentWorldID
		h.mu.Unlock()
		if err := h.uc.RecordEncounterAt(h.ctx, e.VRCUserID, e.DisplayName, e.Action, inst, wid, e.EncounteredAt); err != nil {
			h.logger.Printf("[activity_handler] RecordEncounter: %v", err)
		}
		if h.onAfterEncounter != nil {
			h.onAfterEncounter()
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
			if w := activity.WorldIDFromInstanceKey(e.InstanceID); w != "" {
				h.currentWorldID = w
			}
			h.mu.Unlock()
			if err := h.uc.StartPlaySession(h.ctx, e.InstanceID, e.OccurredAt); err != nil {
				h.logger.Printf("[activity_handler] StartPlaySession: %v", err)
			}
		case activity.SessionEventEnd:
			h.mu.Lock()
			h.currentInstanceID = ""
			h.currentWorldID = ""
			h.mu.Unlock()
			if err := h.uc.EndPlaySession(h.ctx, e.OccurredAt); err != nil {
				h.logger.Printf("[activity_handler] EndPlaySession: %v", err)
			}
		}
	}
}
