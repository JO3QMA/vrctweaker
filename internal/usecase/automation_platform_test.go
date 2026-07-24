package usecase

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/automation"
)

func TestAutomation_eval_disabledItemSkipped(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationItemRepo{}
	repo.items = []*automation.AutomationItem{{
		ID: "a", Name: "off", Kind: automation.KindRule, IsEnabled: false,
		TriggerType: automation.EventFriendJoined,
		ActionsJSON: `[{"type":"change_status","payload":{"status":"busy"}}]`,
	}}
	setter := &mockStatusSetter{}
	uc := newTestAutomationUseCase(repo, setter)
	if err := uc.OnFriendJoined(ctx, "usr_x"); err != nil {
		t.Fatal(err)
	}
	if len(setter.getCalled()) != 0 {
		t.Fatalf("want no actions, got %v", setter.getCalled())
	}
}

func TestAutomation_condition_vrchatNotRunning(t *testing.T) {
	ctx := context.Background()
	conds, _ := json.Marshal([]automation.Condition{{Type: "vrchat_running"}})
	repo := &mockAutomationItemRepo{}
	repo.items = []*automation.AutomationItem{{
		ID: "a", Name: "vr", Kind: automation.KindRule, IsEnabled: true,
		TriggerType:    automation.EventFriendJoined,
		ConditionsJSON: string(conds),
		ActionsJSON:    `[{"type":"change_status","payload":{"status":"busy"}}]`,
	}}
	proc := &mockVRChatProcessChecker{running: false}
	uc := NewAutomationUseCase(repo, &mockStatusSetter{}, proc)
	if err := uc.OnFriendJoined(ctx, "usr_x"); err != nil {
		t.Fatal(err)
	}
	if len(uc.GetRunLog()) != 0 {
		t.Fatal("expected no run log when condition fails")
	}
}

func TestAutomation_sequence_stopsOnError(t *testing.T) {
	ctx := context.Background()
	actions := `[{"type":"change_status","payload":{"status":"busy"}},{"type":"set_power_plan","payload":{"preset":"balanced"}}]`
	repo := &mockAutomationItemRepo{}
	repo.items = []*automation.AutomationItem{{
		ID: "a", Name: "seq", Kind: automation.KindRule, IsEnabled: true,
		TriggerType: automation.EventFriendJoined,
		ActionsJSON: actions,
	}}
	setter := &mockStatusSetter{}
	uc := newTestAutomationUseCase(repo, setter)
	if err := uc.OnFriendJoined(ctx, "usr_x"); err != nil {
		t.Fatal(err)
	}
	logs := uc.GetRunLog()
	if len(logs) != 1 || logs[0].Success {
		t.Fatalf("want failure log, got %#v", logs)
	}
	if logs[0].ActionsCompleted != 1 || logs[0].ActionsTotal != 2 {
		t.Fatalf("want 1/2, got %d/%d", logs[0].ActionsCompleted, logs[0].ActionsTotal)
	}
}

func TestAutomation_toggle_unknownId(t *testing.T) {
	uc := newTestAutomationUseCase(&mockAutomationItemRepo{}, nil)
	if err := uc.ToggleItem(context.Background(), "missing", true); err == nil {
		t.Fatal("want error")
	}
}

func TestAutomation_schedule_stableItemOrder(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationItemRepo{}
	repo.items = []*automation.AutomationItem{
		{ID: "b", Name: "B", Kind: automation.KindRule, IsEnabled: true, TriggerType: automation.EventFriendJoined,
			ActionsJSON: `[{"type":"change_status","payload":{"status":"busy"}}]`},
		{ID: "a", Name: "A", Kind: automation.KindRule, IsEnabled: true, TriggerType: automation.EventFriendJoined,
			ActionsJSON: `[{"type":"change_status","payload":{"status":"ask me"}}]`},
	}
	setter := &mockStatusSetter{}
	uc := newTestAutomationUseCase(repo, setter)
	if err := uc.OnFriendJoined(ctx, "usr_x"); err != nil {
		t.Fatal(err)
	}
	called := setter.getCalled()
	if len(called) != 2 || called[0] != "ask me" || called[1] != "busy" {
		t.Fatalf("want [ask me busy], got %v", called)
	}
}

func TestAutomation_runLog_displayNameNoUserId(t *testing.T) {
	ctx := context.Background()
	repo := &mockAutomationItemRepo{}
	repo.items = []*automation.AutomationItem{{
		ID: "a", Name: "f", Kind: automation.KindRule, IsEnabled: true,
		TriggerType: automation.EventFriendJoined,
		ActionsJSON: `[{"type":"change_status","payload":{"status":"busy"}}]`,
	}}
	uc := newTestAutomationUseCase(repo, &mockStatusSetter{})
	uc.displayNamer = staticDisplayNamer("Friend A")
	if err := uc.OnFriendJoined(ctx, "usr_secret"); err != nil {
		t.Fatal(err)
	}
	logs := uc.GetRunLog()
	if len(logs) != 1 {
		t.Fatal(logs)
	}
	if logs[0].ContextLabel != "Friend A" {
		t.Fatalf("context %q", logs[0].ContextLabel)
	}
	if logs[0].ContextLabel == "usr_secret" {
		t.Fatal("user id leaked to run log label")
	}
}

type staticDisplayNamer string

func (s staticDisplayNamer) DisplayNameFor(context.Context, string) string {
	return string(s)
}

func TestAutomation_schedule_wrongWeekday(t *testing.T) {
	sched, _ := json.Marshal(automation.ScheduleRule{Weekdays: []int{0}, Hour: 12, Minute: 0})
	sundayNoon := time.Date(2026, 7, 19, 12, 0, 0, 0, time.Local) // Sunday
	if !automation.ScheduleMatches(mustParseSchedule(string(sched)), sundayNoon) {
		t.Fatal("Sunday should match")
	}
	monday := sundayNoon.Add(24 * time.Hour)
	if automation.ScheduleMatches(mustParseSchedule(string(sched)), monday) {
		t.Fatal("Monday should not match Sunday-only schedule")
	}
}

func TestAutomation_scheduleTick_runsMatchingRule(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.Local) // Friday
	sched, err := json.Marshal(automation.ScheduleRule{Weekdays: []int{5}, Hour: 12, Minute: 0})
	if err != nil {
		t.Fatal(err)
	}
	conds, err := json.Marshal([]automation.Condition{{Type: "friend_is", VRCUserID: "usr_leftover"}})
	if err != nil {
		t.Fatal(err)
	}
	repo := &mockAutomationItemRepo{}
	repo.items = []*automation.AutomationItem{{
		ID: "a", Name: "sched", Kind: automation.KindRule, IsEnabled: true,
		TriggerType:    automation.EventScheduleTick,
		ScheduleJSON:   string(sched),
		ConditionsJSON: string(conds),
		ActionsJSON:    `[{"type":"change_status","payload":{"status":"busy"}}]`,
	}}
	setter := &mockStatusSetter{}
	uc := newTestAutomationUseCase(repo, setter)
	uc.handleEvent(ctx, automation.Event{Type: automation.EventScheduleTick, At: now})
	if got := setter.getCalled(); len(got) != 1 || got[0] != "busy" {
		t.Fatalf("want [busy], got %v", got)
	}
}

func TestAutomation_scheduleTick_skipsNonMatchingItem(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.Local)
	matchSched, _ := json.Marshal(automation.ScheduleRule{Weekdays: []int{5}, Hour: 12, Minute: 0})
	otherSched, _ := json.Marshal(automation.ScheduleRule{Weekdays: []int{5}, Hour: 18, Minute: 0})
	repo := &mockAutomationItemRepo{}
	repo.items = []*automation.AutomationItem{
		{
			ID: "noon", Name: "noon", Kind: automation.KindRule, IsEnabled: true,
			TriggerType: automation.EventScheduleTick, ScheduleJSON: string(matchSched),
			ActionsJSON: `[{"type":"change_status","payload":{"status":"busy"}}]`,
		},
		{
			ID: "eve", Name: "eve", Kind: automation.KindRule, IsEnabled: true,
			TriggerType: automation.EventScheduleTick, ScheduleJSON: string(otherSched),
			ActionsJSON: `[{"type":"change_status","payload":{"status":"ask me"}}]`,
		},
	}
	setter := &mockStatusSetter{}
	uc := newTestAutomationUseCase(repo, setter)
	uc.handleEvent(ctx, automation.Event{Type: automation.EventScheduleTick, At: now})
	if got := setter.getCalled(); len(got) != 1 || got[0] != "busy" {
		t.Fatalf("want only noon rule, got %v", got)
	}
}

func mustParseSchedule(raw string) *automation.ScheduleRule {
	s, err := automation.ParseSchedule(raw)
	if err != nil {
		panic(err)
	}
	return s
}
