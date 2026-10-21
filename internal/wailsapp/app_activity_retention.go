package wailsapp

import (
	"context"
	"errors"
	"log"
	"time"
)

// activityRetentionInterval is how often a long-lived process (including close-to-tray)
// deletes Activity rows past log_retention_days. Startup and SetLogRetentionDays purge
// immediately; this covers the process staying up across the cutoff.
const activityRetentionInterval = time.Hour

// purgeActivityRetention deletes expired activity rows and records the cutoff.
// A nil activity usecase is a no-op so early startup failure does not panic.
func (a *App) purgeActivityRetention(ctx context.Context) error {
	if a == nil || a.activity == nil || ctx == nil {
		return nil
	}
	_, err := a.activity.RotateEncounters(ctx)
	return err
}

func (a *App) startActivityRetentionLoop() {
	a.startActivityRetentionLoopEvery(activityRetentionInterval)
}

func (a *App) startActivityRetentionLoopEvery(interval time.Duration) {
	if a == nil || interval <= 0 {
		return
	}
	a.activityRetentionMu.Lock()
	if a.activityRetentionCancel != nil {
		cancel := a.activityRetentionCancel
		a.activityRetentionCancel = nil
		a.activityRetentionMu.Unlock()
		cancel()
		a.activityRetentionWG.Wait()
		a.activityRetentionMu.Lock()
	}
	runCtx, cancel := context.WithCancel(context.Background())
	a.activityRetentionCancel = cancel
	a.activityRetentionWG.Add(1)
	a.activityRetentionMu.Unlock()

	go func() {
		defer a.activityRetentionWG.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				if err := a.purgeActivityRetention(runCtx); err != nil && !errors.Is(err, context.Canceled) {
					log.Printf("activity retention: %v", err)
				}
			}
		}
	}()
}

func (a *App) stopActivityRetentionLoop() {
	if a == nil {
		return
	}
	a.activityRetentionMu.Lock()
	cancel := a.activityRetentionCancel
	a.activityRetentionCancel = nil
	a.activityRetentionMu.Unlock()
	if cancel != nil {
		cancel()
		a.activityRetentionWG.Wait()
	}
}
