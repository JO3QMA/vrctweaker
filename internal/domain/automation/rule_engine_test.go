package automation

import (
	"testing"
)

func TestEvalRule_DisabledRule(t *testing.T) {
	rule := &AutomationRule{
		ID:          "1",
		Name:        "Test",
		TriggerType: TriggerAFKDetected,
		ActionType:  ActionChangeStatus,
		IsEnabled:   false,
	}
	ctx := &EvalContext{TriggerType: TriggerAFKDetected, Payload: nil}
	res, err := EvalRule(rule, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.ShouldFire {
		t.Error("disabled rule should not fire")
	}
}

func TestEvalRule_TriggerMismatch(t *testing.T) {
	rule := &AutomationRule{
		ID:          "1",
		Name:        "Test",
		TriggerType: TriggerAFKDetected,
		ActionType:  ActionChangeStatus,
		IsEnabled:   true,
	}
	ctx := &EvalContext{TriggerType: TriggerFriendJoined, Payload: nil}
	res, err := EvalRule(rule, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.ShouldFire {
		t.Error("trigger mismatch should not fire")
	}
}

func TestEvalRule_Matches(t *testing.T) {
	rule := &AutomationRule{
		ID:            "1",
		Name:          "Test",
		TriggerType:   TriggerAFKDetected,
		ActionType:    ActionChangeStatus,
		ActionPayload: `{"status":"busy"}`,
		IsEnabled:     true,
	}
	ctx := &EvalContext{TriggerType: TriggerAFKDetected, Payload: nil}
	res, err := EvalRule(rule, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !res.ShouldFire {
		t.Error("expected rule to fire")
	}
	if res.ActionType != ActionChangeStatus {
		t.Errorf("action type = %s", res.ActionType)
	}
}

func TestEvalRule_ConditionEmpty_AlwaysTrue(t *testing.T) {
	rule := &AutomationRule{
		ID:            "1",
		Name:          "EmptyCond",
		TriggerType:   TriggerFriendJoined,
		ConditionJSON: "{}",
		ActionType:    ActionChangeStatus,
		ActionPayload: `{"status":"ask me"}`,
		IsEnabled:     true,
	}
	ctx := &EvalContext{TriggerType: TriggerFriendJoined, Payload: map[string]interface{}{"vrc_user_id": "usr_abc"}}
	res, err := EvalRule(rule, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !res.ShouldFire {
		t.Error("empty condition should always match")
	}
}

func TestEvalRule_ConditionMatchesPayload_True(t *testing.T) {
	rule := &AutomationRule{
		ID:            "1",
		Name:          "UserMatch",
		TriggerType:   TriggerFriendJoined,
		ConditionJSON: `{"vrc_user_id":"usr_123"}`,
		ActionType:    ActionChangeStatus,
		ActionPayload: `{"status":"join me"}`,
		IsEnabled:     true,
	}
	ctx := &EvalContext{
		TriggerType: TriggerFriendJoined,
		Payload:     map[string]interface{}{"vrc_user_id": "usr_123"},
	}
	res, err := EvalRule(rule, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !res.ShouldFire {
		t.Error("condition should match payload")
	}
}

func TestEvalRule_ConditionMismatch_False(t *testing.T) {
	rule := &AutomationRule{
		ID:            "1",
		Name:          "UserMismatch",
		TriggerType:   TriggerFriendJoined,
		ConditionJSON: `{"vrc_user_id":"usr_123"}`,
		ActionType:    ActionChangeStatus,
		ActionPayload: `{"status":"busy"}`,
		IsEnabled:     true,
	}
	ctx := &EvalContext{
		TriggerType: TriggerFriendJoined,
		Payload:     map[string]interface{}{"vrc_user_id": "usr_456"},
	}
	res, err := EvalRule(rule, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.ShouldFire {
		t.Error("condition mismatch should not fire")
	}
}

func TestEvalRule_ConditionKeyMissingInPayload_False(t *testing.T) {
	rule := &AutomationRule{
		ID:            "1",
		Name:          "KeyMissing",
		TriggerType:   TriggerFriendJoined,
		ConditionJSON: `{"vrc_user_id":"usr_123"}`,
		ActionType:    ActionChangeStatus,
		ActionPayload: `{"status":"busy"}`,
		IsEnabled:     true,
	}
	ctx := &EvalContext{TriggerType: TriggerFriendJoined, Payload: map[string]interface{}{}}
	res, err := EvalRule(rule, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.ShouldFire {
		t.Error("missing key in payload should not fire")
	}
}

func TestEvalRule_InvalidConditionJSON_Error(t *testing.T) {
	rule := &AutomationRule{
		ID:            "1",
		Name:          "BadJSON",
		TriggerType:   TriggerAFKDetected,
		ConditionJSON: `{invalid}`,
		ActionType:    ActionChangeStatus,
		IsEnabled:     true,
	}
	ctx := &EvalContext{TriggerType: TriggerAFKDetected, Payload: nil}
	_, err := EvalRule(rule, ctx)
	if err == nil {
		t.Error("invalid ConditionJSON should return error")
	}
}
