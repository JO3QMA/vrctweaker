package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"vrchat-tweaker/internal/domain/automation"
	"vrchat-tweaker/internal/infrastructure/vrchatwindow"
)

type mockWindowResizer struct {
	lastW, lastH int
	err          error
	calls        int
}

func (m *mockWindowResizer) Resize(width, height int) error {
	m.calls++
	m.lastW, m.lastH = width, height
	return m.err
}

func TestParseWindowSizePayload(t *testing.T) {
	w, h, err := parseWindowSizePayload(map[string]interface{}{"width": float64(1280), "height": float64(720)})
	if err != nil || w != 1280 || h != 720 {
		t.Fatalf("got %d×%d err=%v", w, h, err)
	}
	_, _, err = parseWindowSizePayload(map[string]interface{}{"width": 0, "height": 720})
	if !errors.Is(err, vrchatwindow.ErrInvalidSize) {
		t.Fatalf("want ErrInvalidSize, got %v", err)
	}
	_, _, err = parseWindowSizePayload(nil)
	if err == nil {
		t.Fatal("want error")
	}
}

func TestAutomation_setVRChatWindowSize(t *testing.T) {
	ctx := context.Background()
	actions, _ := json.Marshal([]automation.ActionStep{{
		Type:    automation.ActionSetVRChatWindowSize,
		Payload: map[string]interface{}{"width": 1280, "height": 720},
	}})
	repo := &mockAutomationItemRepo{}
	repo.items = []*automation.AutomationItem{{
		ID: "a", Name: "win", Kind: automation.KindRule, IsEnabled: true,
		TriggerType: automation.EventFriendJoined,
		ActionsJSON: string(actions),
	}}
	resizer := &mockWindowResizer{}
	uc := newTestAutomationUseCase(repo, &mockStatusSetter{})
	uc.windowResizer = resizer
	if err := uc.OnFriendJoined(ctx, "usr_x"); err != nil {
		t.Fatal(err)
	}
	if resizer.calls != 1 || resizer.lastW != 1280 || resizer.lastH != 720 {
		t.Fatalf("resizer %#v", resizer)
	}
	logs := uc.GetRunLog()
	if len(logs) != 1 || !logs[0].Success {
		t.Fatalf("want success log, got %#v", logs)
	}
}

func TestAutomation_setVRChatWindowSize_notRunning(t *testing.T) {
	ctx := context.Background()
	actions, _ := json.Marshal([]automation.ActionStep{{
		Type:    automation.ActionSetVRChatWindowSize,
		Payload: map[string]interface{}{"width": 800, "height": 600},
	}})
	repo := &mockAutomationItemRepo{}
	repo.items = []*automation.AutomationItem{{
		ID: "a", Name: "win", Kind: automation.KindRule, IsEnabled: true,
		TriggerType: automation.EventFriendJoined,
		ActionsJSON: string(actions),
	}}
	resizer := &mockWindowResizer{err: vrchatwindow.ErrNotRunning}
	uc := newTestAutomationUseCase(repo, nil)
	uc.windowResizer = resizer
	if err := uc.OnFriendJoined(ctx, "usr_x"); err != nil {
		t.Fatal(err)
	}
	logs := uc.GetRunLog()
	if len(logs) != 1 || logs[0].Success {
		t.Fatalf("want failure, got %#v", logs)
	}
}

func TestAutomation_setVRChatWindowSize_exclusiveSkip(t *testing.T) {
	// Exclusive fullscreen is implemented as resizer returning nil (no-op success).
	ctx := context.Background()
	actions, _ := json.Marshal([]automation.ActionStep{{
		Type:    automation.ActionSetVRChatWindowSize,
		Payload: map[string]interface{}{"width": 1280, "height": 720},
	}})
	repo := &mockAutomationItemRepo{}
	repo.items = []*automation.AutomationItem{{
		ID: "a", Name: "win", Kind: automation.KindRule, IsEnabled: true,
		TriggerType: automation.EventFriendJoined,
		ActionsJSON: string(actions),
	}}
	resizer := &mockWindowResizer{} // nil err = success/skip
	uc := newTestAutomationUseCase(repo, nil)
	uc.windowResizer = resizer
	if err := uc.OnFriendJoined(ctx, "usr_x"); err != nil {
		t.Fatal(err)
	}
	if resizer.calls != 1 {
		t.Fatalf("calls=%d", resizer.calls)
	}
	logs := uc.GetRunLog()
	if len(logs) != 1 || !logs[0].Success {
		t.Fatalf("want success (skip), got %#v", logs)
	}
}
