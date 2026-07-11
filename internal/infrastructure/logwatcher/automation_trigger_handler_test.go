package logwatcher

import (
	"context"
	"errors"
	"testing"

	"vrchat-tweaker/internal/domain/activity"
)

type stubFriendJoinedAutomation struct {
	called []string
	err    error
}

func (s *stubFriendJoinedAutomation) OnFriendJoined(_ context.Context, vrcUserID string) error {
	s.called = append(s.called, vrcUserID)
	return s.err
}

func TestAutomationTriggerHandler_FriendJoined(t *testing.T) {
	ctx := context.Background()
	auto := &stubFriendJoinedAutomation{}
	h := NewAutomationTriggerHandler(auto, ctx, nil)

	h.Handle(&activity.EncounterEvent{
		Action:      activity.EncounterActionJoin,
		VRCUserID:   "usr_join01",
		DisplayName: "Friend",
	})
	if len(auto.called) != 1 || auto.called[0] != "usr_join01" {
		t.Fatalf("called = %v", auto.called)
	}

	h.Handle(&activity.EncounterEvent{Action: activity.EncounterActionLeave, VRCUserID: "usr_join01"})
	h.Handle(nil)
	if len(auto.called) != 1 {
		t.Fatalf("leave/nil should not trigger, got %d calls", len(auto.called))
	}
}

func TestAutomationTriggerHandler_OnFriendJoinedErrorLogged(t *testing.T) {
	var logs []string
	auto := &stubFriendJoinedAutomation{err: errors.New("boom")}
	h := NewAutomationTriggerHandler(auto, context.Background(), func(format string, args ...any) {
		logs = append(logs, format)
	})
	h.Handle(&activity.EncounterEvent{
		Action:    activity.EncounterActionJoin,
		VRCUserID: "usr_err",
	})
	if len(logs) == 0 {
		t.Fatal("expected log on OnFriendJoined error")
	}
}

func TestNewAutomationTriggerHandler_defaultLogger(t *testing.T) {
	h := NewAutomationTriggerHandler(&stubFriendJoinedAutomation{}, context.Background(), nil)
	if h.logger == nil {
		t.Fatal("expected default logger")
	}
}

func TestMultiHandler_DelegatesToAll(t *testing.T) {
	var count int
	h1 := EventHandlerFunc(func(activity.ParsedEvent) { count++ })
	h2 := EventHandlerFunc(func(activity.ParsedEvent) { count++ })
	mh := NewMultiHandler(h1, h2)
	mh.Handle(&activity.EncounterEvent{DisplayName: "A"})
	if count != 2 {
		t.Fatalf("calls = %d, want 2", count)
	}
}

func TestEventHandlerFunc_Handle(t *testing.T) {
	var called bool
	EventHandlerFunc(func(activity.ParsedEvent) { called = true }).Handle(&activity.EncounterEvent{})
	if !called {
		t.Fatal("expected handle call")
	}
}

func TestEventHandlerFunc_Handle_nilSafe(t *testing.T) {
	var f EventHandlerFunc
	var h EventHandler = f
	h.Handle(&activity.EncounterEvent{}) // typed nil must not panic
}
