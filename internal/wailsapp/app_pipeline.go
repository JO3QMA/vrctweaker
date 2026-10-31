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
	var cancelFn context.CancelFunc
	a.pipelineMu.Lock()
	if a.pipelineCancel != nil {
		cancelFn = a.pipelineCancel
		a.pipelineCancel = nil
	}
	a.pipelineMu.Unlock()
	if cancelFn != nil {
		cancelFn()
	}
}

func (a *App) stopVRChatPipeline() {
	a.stopVRChatPipelineCancel()
	a.pipelineWG.Wait()
}

func (a *App) startVRChatPipeline() {
	a.pipelineMu.Lock()
	for a.pipelineCancel != nil {
		cancelFn := a.pipelineCancel
		a.pipelineCancel = nil
		a.pipelineMu.Unlock()
		cancelFn()
		a.pipelineWG.Wait()
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
	runCtx, cancel := context.WithCancel(a.ctx)
	a.pipelineCancel = cancel
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
	a.pipelineWG.Add(1)
	a.pipelineMu.Unlock()
	go func() {
		defer a.pipelineWG.Done()
		err := vrchatpipeline.Run(runCtx, cfg)
		if err != nil && !errors.Is(err, context.Canceled) {
			runtime.LogWarning(a.ctx, "vrchat pipeline: "+err.Error())
		}
	}()
}
