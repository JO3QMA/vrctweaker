package wailsapp

import (
	"context"
	"time"

	"vrchat-tweaker/internal/infrastructure/sleepsuppress"
)

type sleepSuppressSettingGetter struct {
	a *App
}

func (g sleepSuppressSettingGetter) SuppressSleepWhileVRChat(ctx context.Context) (bool, error) {
	if g.a == nil || g.a.settings == nil {
		return false, nil
	}
	return g.a.settings.GetSuppressSleepWhileVRChat(ctx)
}

func (a *App) startSleepSuppressLoop() {
	a.sleepSuppressMu.Lock()
	defer a.sleepSuppressMu.Unlock()
	if a.settings == nil {
		return
	}
	if a.sleepSuppressCancel != nil {
		a.sleepSuppressCancel()
		a.sleepSuppressWG.Wait()
		a.sleepSuppressCancel = nil
	}
	runCtx, cancel := context.WithCancel(context.Background())
	a.sleepSuppressCancel = cancel
	a.sleepSuppressWG.Add(1)
	go func() {
		defer a.sleepSuppressWG.Done()
		checker := sleepsuppress.NewVRChatProcessChecker()
		exec := sleepsuppress.NewExecutionState()
		sg := sleepSuppressSettingGetter{a: a}
		_ = sleepsuppress.Run(runCtx, 8*time.Second, sg, checker, exec)
	}()
}

func (a *App) stopSleepSuppressLoop() {
	var cancel context.CancelFunc
	a.sleepSuppressMu.Lock()
	cancel = a.sleepSuppressCancel
	a.sleepSuppressCancel = nil
	a.sleepSuppressMu.Unlock()
	if cancel != nil {
		cancel()
		a.sleepSuppressWG.Wait()
	}
}
