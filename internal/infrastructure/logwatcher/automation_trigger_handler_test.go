package logwatcher

import (
	"context"
	"errors"
	"testing"

	"vrchat-tweaker/internal/domain/activity"
)

type stubFriendEncounterAutomation struct {
	joined []string
	left   []string
	err    error
}

func (s *stubFriendEncounterAutomation) OnFriendJoined(_ context.Context, vrcUserID string) error {
	s.joined = append(s.joined, vrcUserID)
	return s.err
}

func (s *stubFriendEncounterAutomation) OnFriendLeft(_ context.Context, vrcUserID string) error {
	s.left = append(s.left, vrcUserID)
	return s.err
}

func TestAutomationTriggerHandler_FriendJoined(t *testing.T) {
	ctx := context.Background()
	auto := &stubFriendEncounterAutomation{}
	h := NewAutomationTriggerHandler(auto, ctx, nil)

	h.Handle(&activity.EncounterEvent{
		Action:      activity.EncounterActionJoin,
		VRCUserID:   "usr_join01",
		DisplayName: "Friend",
	})
	if len(auto.joined) != 1 || auto.joined[0] != "usr_join01" {
		t.Fatalf("joined = %v", auto.joined)
	}

	h.Handle(nil)
	if len(auto.joined) != 1 {
		t.Fatalf("nil should not trigger, got %d join calls", len(auto.joined))
	}
}

func TestAutomationTriggerHandler_FriendLeft(t *testing.T) {
	ctx := context.Background()
	auto := &stubFriendEncounterAutomation{}
	h := NewAutomationTriggerHandler(auto, ctx, nil)

	h.Handle(&activity.EncounterEvent{
		Action:    activity.EncounterActionLeave,
		VRCUserID: "usr_leave01",
	})
	if len(auto.left) != 1 || auto.left[0] != "usr_leave01" {
		t.Fatalf("left = %v", auto.left)
	}
	if len(auto.joined) != 0 {
		t.Fatalf("leave should not call join, got %v", auto.joined)
	}
}

func TestAutomationTriggerHandler_OnFriendJoinedErrorLogged(t *testing.T) {
	var logs []string
	auto := &stubFriendEncounterAutomation{err: errors.New("boom")}
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

func TestAutomationTriggerHandler_OnFriendLeftErrorLogged(t *testing.T) {
	var logs []string
	auto := &stubFriendEncounterAutomation{err: errors.New("boom")}
	h := NewAutomationTriggerHandler(auto, context.Background(), func(format string, args ...any) {
		logs = append(logs, format)
	})
	h.Handle(&activity.EncounterEvent{
		Action:    activity.EncounterActionLeave,
		VRCUserID: "usr_err",
	})
	if len(logs) == 0 {
		t.Fatal("expected log on OnFriendLeft error")
	}
}

func TestNewAutomationTriggerHandler_defaultLogger(t *testing.T) {
	h := NewAutomationTriggerHandler(&stubFriendEncounterAutomation{}, context.Background(), nil)
	if h.logger == nil {
		t.Fatal("expected default logger")
	}
}
