package logwatcher

import (
	"context"
	"log"
	"sync/atomic"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/usecase"
)

// ActivityIngestAdapter bridges parsed log events to SessionCorrelator and ActivityUseCase.
type ActivityIngestAdapter struct {
	uc               *usecase.ActivityUseCase
	ctx              context.Context
	logger           Logger
	correlator       activity.SessionCorrelator
	onAfterEncounter func()
	// suppressEncounterNotify skips onAfterEncounter (e.g. during historical log bootstrap).
	suppressEncounterNotify atomic.Bool
}

// NewActivityIngestAdapter creates an adapter that correlates and persists log-derived activity.
// onAfterEncounter is optional (e.g. Wails EventsEmit after each encounter row).
func NewActivityIngestAdapter(uc *usecase.ActivityUseCase, ctx context.Context, logger Logger, onAfterEncounter func()) *ActivityIngestAdapter {
	if logger == nil {
		logger = logWriterLogger{log.Default()}
	}
	return &ActivityIngestAdapter{
		uc:               uc,
		ctx:              ctx,
		logger:           logger,
		onAfterEncounter: onAfterEncounter,
	}
}

// SetSuppressEncounterNotify when true skips onAfterEncounter for encounter commands (e.g. bulk bootstrap).
func (a *ActivityIngestAdapter) SetSuppressEncounterNotify(suppress bool) {
	a.suppressEncounterNotify.Store(suppress)
}

// ResetSessionContextForNewLogFile clears correlator state before reading a new output_log file.
func (a *ActivityIngestAdapter) ResetSessionContextForNewLogFile() {
	a.correlator.Reset()
}

// Handle implements EventHandler.
func (a *ActivityIngestAdapter) Handle(event activity.ParsedEvent) {
	for _, cmd := range a.correlator.Apply(event) {
		if err := a.uc.ApplyCommand(a.ctx, cmd); err != nil {
			a.logger.Printf("[activity_ingest] %T: %v", cmd, err)
		}
		switch cmd.(type) {
		case activity.RecordEncounterJoinCmd, activity.RecordEncounterLeaveCmd:
			if !a.suppressEncounterNotify.Load() && a.onAfterEncounter != nil {
				a.onAfterEncounter()
			}
		}
	}
}
