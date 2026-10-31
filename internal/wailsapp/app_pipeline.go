package wailsapp

import (
	"context"
	"errors"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/infrastructure/vrchatapi"
	"vrchat-tweaker/internal/infrastructure/vrchatpipeline"
)

// friendsChangedEvent is emitted when the friends cache may have changed (Pipeline or REST reconcile).
const friendsChangedEvent = "vrchat:friends-changed"

func pipelineEventUpdatesFriends(typ string) bool {
	switch typ {
	case "friend-delete", "friend-offline", "friend-active", "friend-online",
		"friend-location", "friend-update", "friend-add":
		return true
	default:
		return false
	}
}

// stopVRChatPipelineCancel stops the pipeline without waiting. Do not call stopVRChatPipeline
// (which waits) from callbacks invoked synchronously inside vrchatpipeline.Run — that deadlocks.
func (a *App) stopVRChatPipelineCancel() {
	a.pipelineMu.Lock()
	cancel := a.pipelineCancel
	a.pipelineCancel = nil
	// Keep pipelineDone until the goroutine closes it; startVRChatPipeline will wait on it.
	a.pipelineMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (a *App) stopVRChatPipeline() {
	a.pipelineMu.Lock()
	cancel := a.pipelineCancel
	done := a.pipelineDone
	if cancel != nil {
		a.pipelineCancel = nil
	}
	a.pipelineMu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
	// Goroutine's deferred cleanup also clears done/cancel if still same instance;
	// ensure cleared if the goroutine already exited and left them.
	a.pipelineMu.Lock()
	if a.pipelineDone == done {
		a.pipelineDone = nil
		a.pipelineCancel = nil
	}
	a.pipelineMu.Unlock()
}

func (a *App) startVRChatPipeline() {
	a.pipelineMu.Lock()
	for a.pipelineDone != nil {
		done := a.pipelineDone
		cancel := a.pipelineCancel
		a.pipelineMu.Unlock()
		if cancel != nil {
			cancel()
		}
		<-done
		a.pipelineMu.Lock()
	}
	if a.identity == nil {
		a.pipelineMu.Unlock()
		return
	}
	token := a.identity.CurrentAuthToken()
	if token == "" {
		a.pipelineMu.Unlock()
		return
	}
	done := make(chan struct{})
	runCtx, cancel := context.WithCancel(a.ctx)
	a.pipelineCancel = cancel
	a.pipelineDone = done
	cfg := vrchatpipeline.Config{
		AuthToken: token,
		UserAgent: vrchatapi.UserAgent(),
		OnReconnect: func(ctx context.Context) error {
			err := a.identity.PipelineReconnectRestSync(ctx)
			if err != nil {
				if errors.Is(err, vrchatapi.ErrSessionExpired) || errors.Is(err, vrchatapi.ErrNotAuthenticated) {
					a.stopVRChatPipelineCancel()
				}
				return err
			}
			runtime.EventsEmit(a.ctx, friendsChangedEvent, struct{}{})
			return nil
		},
		OnEvent: func(ctx context.Context, typ string, payload []byte) error {
			err := a.identity.HandleVRChatPipelineEvent(ctx, typ, payload)
			if err == nil && pipelineEventUpdatesFriends(typ) {
				runtime.EventsEmit(a.ctx, friendsChangedEvent, struct{}{})
			}
			return err
		},
	}
	a.pipelineMu.Unlock()
	go func() {
		defer close(done)
		defer func() {
			a.pipelineMu.Lock()
			if a.pipelineDone == done {
				a.pipelineDone = nil
				a.pipelineCancel = nil
			}
			a.pipelineMu.Unlock()
		}()
		err := vrchatpipeline.Run(runCtx, cfg)
		if err != nil && !errors.Is(err, context.Canceled) {
			runtime.LogWarning(a.ctx, "vrchat pipeline: "+err.Error())
		}
	}()
}
