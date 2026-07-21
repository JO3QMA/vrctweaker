package usecase

import (
	"context"
	"strings"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/automation"
)

func TestAutomation_script_luaSandboxNoIO(t *testing.T) {
	r := newScriptRunner(func(context.Context, string, map[string]interface{}) error { return nil })
	err := r.run(context.Background(), `dofile("/etc/passwd")`, automation.Event{Type: "x"})
	if err == nil {
		t.Fatal("expected dofile to fail in sandbox")
	}
}

func TestAutomation_script_luaSandboxNoCollectgarbage(t *testing.T) {
	r := newScriptRunner(func(context.Context, string, map[string]interface{}) error { return nil })
	err := r.run(context.Background(), `collectgarbage("stop")`, automation.Event{Type: "x"})
	if err == nil {
		t.Fatal("expected collectgarbage to fail in sandbox")
	}
}

func TestAutomation_script_luaTimeout(t *testing.T) {
	r := newScriptRunner(func(context.Context, string, map[string]interface{}) error { return nil })
	r.execTimeout = 50 * time.Millisecond
	err := r.run(context.Background(), `
while true do end
`, automation.Event{Type: "x"})
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("want timeout, got %v", err)
	}
}

func TestAutomation_worker_startStopRestart(t *testing.T) {
	uc := newTestAutomationUseCase(&mockAutomationItemRepo{}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	uc.Start(ctx)
	uc.Stop()
	uc.PublishEvent(automation.Event{Type: "x"}) // must not panic after stop
	uc.Start(ctx)
	uc.PublishEvent(automation.Event{Type: automation.EventFriendJoined})
	uc.Stop()
}

func TestAutomation_delete_unknownId(t *testing.T) {
	uc := newTestAutomationUseCase(&mockAutomationItemRepo{}, nil)
	if err := uc.DeleteItem(context.Background(), "missing"); err == nil {
		t.Fatal("want not found")
	}
}
