package automation

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEvalItem_schedule_staleFriendIsIgnored(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.Local) // Friday
	sched, err := json.Marshal(ScheduleRule{Weekdays: []int{5}, Hour: 12, Minute: 0})
	if err != nil {
		t.Fatal(err)
	}
	conds, err := json.Marshal([]Condition{{Type: "friend_is", VRCUserID: "usr_leftover"}})
	if err != nil {
		t.Fatal(err)
	}
	item := &AutomationItem{
		ID:             "a",
		Kind:           KindRule,
		IsEnabled:      true,
		TriggerType:    EventScheduleTick,
		ScheduleJSON:   string(sched),
		ConditionsJSON: string(conds),
		ActionsJSON:    `[{"type":"change_status","payload":{"status":"busy"}}]`,
	}
	ok, err := EvalItem(item, &EvalContext{
		TriggerType: EventScheduleTick,
		Payload:     nil,
		Now:         now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("stale friend_is on schedule.tick must not block evaluation")
	}
}

func TestEvalItem_schedule_requiresOwnScheduleMatch(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.Local) // Friday 12:00
	sched, err := json.Marshal(ScheduleRule{Weekdays: []int{5}, Hour: 18, Minute: 0})
	if err != nil {
		t.Fatal(err)
	}
	item := &AutomationItem{
		ID:           "late",
		Kind:         KindRule,
		IsEnabled:    true,
		TriggerType:  EventScheduleTick,
		ScheduleJSON: string(sched),
		ActionsJSON:  `[{"type":"change_status","payload":{"status":"busy"}}]`,
	}
	ok, err := EvalItem(item, &EvalContext{
		TriggerType: EventScheduleTick,
		Now:         now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("item scheduled for 18:00 must not match 12:00 tick")
	}
}

func TestParseSchedule_emptyWeekdaysRejected(t *testing.T) {
	_, err := ParseSchedule(`{"weekdays":[],"hour":0,"minute":0}`)
	if err == nil {
		t.Fatal("empty weekdays must be rejected")
	}
}

func TestNextMinuteBoundary(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 45, 0, time.Local)
	got := NextMinuteBoundary(now)
	want := time.Date(2026, 7, 24, 12, 1, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
