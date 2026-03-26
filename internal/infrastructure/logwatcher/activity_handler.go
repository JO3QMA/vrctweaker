package logwatcher

import (
	"context"
	"log"
	"sync"
	"sync/atomic"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/usecase"
)

// ActivityEventHandler bridges parsed log events to ActivityUseCase.
type ActivityEventHandler struct {
	uc               *usecase.ActivityUseCase
	ctx              context.Context
	logger           Logger
	onAfterEncounter func()
	// suppressEncounterNotify skips onAfterEncounter (e.g. during historical log bootstrap).
	suppressEncounterNotify atomic.Bool

	mu sync.Mutex
	// session* are the active Joining instance (SessionEventStart only). Destination does not update these.
	sessionInstanceID string
	sessionWorldID    string
	// pendingDestinationWorldID is set by Destination set; survives SessionEventEnd for RoomName before Joining.
	pendingDestinationWorldID string
	// lastLeft* snapshot at SessionEventEnd for OnPlayerLeft lines that follow OnLeftRoom (instance/world still "left" room).
	lastLeftInstanceID string
	lastLeftWorldID    string
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

// SetSuppressEncounterNotify when true skips onAfterEncounter for EncounterEvent (e.g. bulk bootstrap).
func (h *ActivityEventHandler) SetSuppressEncounterNotify(suppress bool) {
	h.suppressEncounterNotify.Store(suppress)
}

// ResetSessionContextForNewLogFile clears instance/world correlation state before reading a new
// output_log file. Each VRChat log file is an independent session; without this, RoomName and
// encounters after a file boundary can inherit the previous file's sessionWorldID.
func (h *ActivityEventHandler) ResetSessionContextForNewLogFile() {
	h.mu.Lock()
	h.sessionInstanceID = ""
	h.sessionWorldID = ""
	h.pendingDestinationWorldID = ""
	h.lastLeftInstanceID = ""
	h.lastLeftWorldID = ""
	h.mu.Unlock()
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
		h.pendingDestinationWorldID = e.WorldID
		h.mu.Unlock()
	case *activity.RoomNameEvent:
		h.mu.Lock()
		wid := h.sessionWorldID
		if wid == "" {
			wid = h.pendingDestinationWorldID
		}
		h.mu.Unlock()
		if err := h.uc.UpsertWorldRoomName(h.ctx, wid, e.RoomName, e.OccurredAt); err != nil {
			h.logger.Printf("[activity_handler] UpsertWorldRoomName: %v", err)
		}
	case *activity.EncounterEvent:
		inst := e.InstanceID
		h.mu.Lock()
		if inst == "" {
			inst = h.sessionInstanceID
		}
		if inst == "" && e.Action == activity.EncounterActionLeave {
			inst = h.lastLeftInstanceID
		}
		wid := h.sessionWorldID
		if wid == "" && e.Action == activity.EncounterActionLeave {
			wid = h.lastLeftWorldID
		}
		if wid == "" {
			wid = h.pendingDestinationWorldID
		}
		h.mu.Unlock()
		if err := h.uc.RecordEncounterAt(h.ctx, e.VRCUserID, e.DisplayName, e.Action, inst, wid, e.EncounteredAt); err != nil {
			h.logger.Printf("[activity_handler] RecordEncounter: %v", err)
		}
		if !h.suppressEncounterNotify.Load() && h.onAfterEncounter != nil {
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
			// Joining a new instance: close every still-open encounter row (all users). Correct
			// because the local user left the previous instance; co-location with prior joins ends.
			// Assumes events are processed in order (single handler goroutine).
			if err := h.uc.CloseOpenEncountersAt(h.ctx, e.OccurredAt); err != nil {
				h.logger.Printf("[activity_handler] CloseOpenEncountersAt (before new instance): %v", err)
			}
			h.mu.Lock()
			h.lastLeftInstanceID = ""
			h.lastLeftWorldID = ""
			h.sessionInstanceID = e.InstanceID
			if w := activity.WorldIDFromInstanceKey(e.InstanceID); w != "" {
				h.sessionWorldID = w
			} else {
				h.sessionWorldID = ""
			}
			h.pendingDestinationWorldID = ""
			h.mu.Unlock()
			if err := h.uc.StartPlaySession(h.ctx, e.InstanceID, e.OccurredAt); err != nil {
				h.logger.Printf("[activity_handler] StartPlaySession: %v", err)
			}
		case activity.SessionEventEnd:
			// Do not CloseOpenEncountersAt here: per-user leaves use EncounterActionLeave /
			// CloseEncounterLeave. Stragglers (crash, missing lines) are closed on the next
			// SessionEventStart above or ingestActivityLogsBootstrap's CloseOpenEncountersAtLastLogLine.
			h.mu.Lock()
			h.lastLeftInstanceID = h.sessionInstanceID
			h.lastLeftWorldID = h.sessionWorldID
			h.sessionInstanceID = ""
			h.sessionWorldID = ""
			// Keep pendingDestinationWorldID: RemainInNetworkRoom transitions emit OnLeftRoom
			// before Entering Room; room name must still map to the last Destination set world.
			h.mu.Unlock()
			if err := h.uc.EndPlaySession(h.ctx, e.OccurredAt); err != nil {
				h.logger.Printf("[activity_handler] EndPlaySession: %v", err)
			}
		}
	}
}
