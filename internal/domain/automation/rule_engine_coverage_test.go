package automation

import "testing"

func TestEvalRule_InvalidActionPayload_Error(t *testing.T) {
	rule := &AutomationRule{
		ID:            "1",
		Name:          "BadAction",
		TriggerType:   TriggerAFKDetected,
		ActionType:    ActionChangeStatus,
		ActionPayload: `{invalid}`,
		IsEnabled:     true,
	}
	ctx := &EvalContext{TriggerType: TriggerAFKDetected, Payload: nil}
	_, err := EvalRule(rule, ctx)
	if err == nil {
		t.Fatal("expected error for invalid ActionPayload")
	}
}

func TestEvalRule_ConditionNilPayload(t *testing.T) {
	rule := &AutomationRule{
		ID:            "1",
		Name:          "NeedsPayload",
		TriggerType:   TriggerFriendJoined,
		ConditionJSON: `{"vrc_user_id":"usr_1"}`,
		ActionType:    ActionChangeStatus,
		ActionPayload: `{"status":"busy"}`,
		IsEnabled:     true,
	}
	res, err := EvalRule(rule, &EvalContext{TriggerType: TriggerFriendJoined, Payload: nil})
	if err != nil {
		t.Fatal(err)
	}
	if res.ShouldFire {
		t.Fatal("nil payload should not match condition")
	}
}

func TestValuesEqual_edgeCases(t *testing.T) {
	if !valuesEqual(nil, nil) {
		t.Fatal("nil,nil should equal")
	}
	if valuesEqual("1", 1) == false {
		t.Fatal("fmt-normalized numeric/string mix should equal")
	}
	if !valuesEqual(1, 1) {
		t.Fatal("same ints should equal")
	}
	if valuesEqual(nil, "x") {
		t.Fatal("nil vs value should not equal")
	}
}

func TestConditionMatchesPayload_emptyCond(t *testing.T) {
	if !conditionMatchesPayload(map[string]interface{}{}, map[string]interface{}{"a": 1}) {
		t.Fatal("empty cond should match any payload")
	}
}
