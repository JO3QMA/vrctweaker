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

func TestAutomation_script_luaSandboxNoRequireOrPrint(t *testing.T) {
	r := newScriptRunner(func(context.Context, string, map[string]interface{}) error { return nil })
	for _, src := range []string{`require("os")`, `print("leak")`, `module("x")`} {
		err := r.run(context.Background(), src, automation.Event{Type: "x"})
		if err == nil {
			t.Fatalf("expected %q to fail in sandbox", src)
		}
	}
}

func TestAutomation_script_luaCyclicActionsPayload(t *testing.T) {
	called := false
	r := newScriptRunner(func(context.Context, string, map[string]interface{}) error {
		called = true
		return nil
	})
	err := r.run(context.Background(), `
local t = {}
t.self = t
function on_event(ev, payload)
  tweaker.actions.run("change_status", t)
end
`, automation.Event{Type: "x"})
	if err == nil {
		t.Fatal("expected cyclic table to fail")
	}
	if called {
		t.Fatal("runAction must not be called for cyclic payload")
	}
}

func TestAutomation_script_actionsRunWaitsOnCancel(t *testing.T) {
	started := make(chan struct{})
	finished := make(chan struct{})
	r := newScriptRunner(func(ctx context.Context, _ string, _ map[string]interface{}) error {
		close(started)
		<-ctx.Done()
		close(finished)
		return ctx.Err()
	})
	r.execTimeout = 40 * time.Millisecond
	err := r.run(context.Background(), `
function on_event(ev, payload)
  tweaker.actions.run("change_status", {status="ask me"})
end
`, automation.Event{Type: "x"})
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("want timeout, got %v", err)
	}
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("actions.run leaked goroutine after timeout")
	}
}

func TestAutomation_script_luaStringRepLimited(t *testing.T) {
	r := newScriptRunner(func(context.Context, string, map[string]interface{}) error { return nil })
	err := r.run(context.Background(), `local x = ("a"):rep(100000000)`, automation.Event{Type: "x"})
	if err == nil {
		t.Fatal("expected oversized string.rep to fail")
	}
	err = r.run(context.Background(), `
function on_event(ev, payload)
  if ("ab"):rep(3) ~= "ababab" then error("rep broken") end
end
`, automation.Event{Type: "x"})
	if err != nil {
		t.Fatalf("small string.rep should work: %v", err)
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
