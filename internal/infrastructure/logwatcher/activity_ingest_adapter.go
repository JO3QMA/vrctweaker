package logwatcher

import (
	"context"
	"sync/atomic"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/infrastructure/diag"
	"vrchat-tweaker/internal/usecase"
)

// ActivityIngestAdapter bridges parsed log events to SessionCorrelator and ActivityUseCase.
type ActivityIngestAdapter struct {
	uc               *usecase.ActivityUseCase
	ctx              context.Context
	logger           diag.Logger
	logSourcePath    string
	correlator       activity.SessionCorrelator
	onAfterEncounter func()
	// suppressEncounterNotify skips onAfterEncounter (e.g. during historical log bootstrap).
	suppressEncounterNotify atomic.Bool
}

// NewActivityIngestAdapter creates an adapter that correlates and persists log-derived activity.
// logSourcePath is the normalized absolute path of the output_log file (empty for legacy single-file tests).
// onAfterEncounter is optional (e.g. Wails EventsEmit after each encounter row).
func NewActivityIngestAdapter(uc *usecase.ActivityUseCase, ctx context.Context, logger diag.Logger, onAfterEncounter func(), logSourcePath string) *ActivityIngestAdapter {
	if logger == nil {
		logger = diag.Std()
	}
	return &ActivityIngestAdapter{
		uc:               uc,
		ctx:              ctx,
		logger:           logger,
		logSourcePath:    logSourcePath,
		onAfterEncounter: onAfterEncounter,
	}
}

// LogSourcePath returns the bound output_log absolute path.
func (a *ActivityIngestAdapter) LogSourcePath() string {
	return a.logSourcePath
}

// SetSuppressEncounterNotify when true skips onAfterEncounter for encounter commands (e.g. bulk bootstrap).
func (a *ActivityIngestAdapter) SetSuppressEncounterNotify(suppress bool) {
	a.suppressEncounterNotify.Store(suppress)
}

// ResetSessionContextForNewLogFile clears correlator state before reading a new output_log file.
func (a *ActivityIngestAdapter) ResetSessionContextForNewLogFile() {
	a.correlator.Reset()
}

// WarmFromParsedEvent implements SessionCorrelatorWarmer for log replay before checkpoint resume.
func (a *ActivityIngestAdapter) WarmFromParsedEvent(event activity.ParsedEvent) {
	_ = a.correlator.Apply(event)
}

// Handle implements EventHandler.
func (a *ActivityIngestAdapter) Handle(event activity.ParsedEvent) {
	for _, cmd := range a.correlator.Apply(event) {
		if err := a.uc.ApplyCommand(a.ctx, a.logSourcePath, cmd); err != nil {
			a.logger("[activity_ingest] %T: %v", cmd, err)
		}
		switch cmd.(type) {
		case activity.RecordEncounterJoinCmd, activity.RecordEncounterLeaveCmd:
			if !a.suppressEncounterNotify.Load() && a.onAfterEncounter != nil {
				a.onAfterEncounter()
			}
		}
	}
}
