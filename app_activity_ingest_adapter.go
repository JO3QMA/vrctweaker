package main

import (
	"context"

	"vrchat-tweaker/internal/infrastructure/logwatcher"
)

// activityIngestAdapterForPath returns a per–log-source ingest adapter reused across bootstrap and live tail.
// SessionCorrelator state must survive from bootstrap replay into tailOutputLogFile.
func (a *App) activityIngestAdapterForPath(ctx context.Context, logger logwatcher.Logger, emitEncounters func(), filePath string) *logwatcher.ActivityIngestAdapter {
	abs := absLogPath(filePath)
	a.activityIngestMu.Lock()
	defer a.activityIngestMu.Unlock()
	if a.activityIngestAdapters == nil {
		a.activityIngestAdapters = make(map[string]*logwatcher.ActivityIngestAdapter)
	}
	if ad, ok := a.activityIngestAdapters[abs]; ok {
		return ad
	}
	ad := logwatcher.NewActivityIngestAdapter(a.activity, ctx, logger, emitEncounters, abs)
	a.activityIngestAdapters[abs] = ad
	return ad
}

func (a *App) evictActivityIngestAdapter(filePath string) {
	if filePath == "" {
		return
	}
	abs := absLogPath(filePath)
	a.activityIngestMu.Lock()
	defer a.activityIngestMu.Unlock()
	delete(a.activityIngestAdapters, abs)
}

func (a *App) resetActivityIngestAdapterCache() {
	a.activityIngestMu.Lock()
	defer a.activityIngestMu.Unlock()
	a.activityIngestAdapters = nil
}
