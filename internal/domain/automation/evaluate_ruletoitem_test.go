package automation

import "testing"

func TestRuleToItem_invalidActionPayloadUsesEmptyMap(t *testing.T) {
	item := RuleToItem(&AutomationRule{
		ID:            "1",
		Name:          "n",
		TriggerType:   TriggerFriendJoined,
		ActionType:    ActionChangeStatus,
		ActionPayload: "{not-json",
		IsEnabled:     true,
	})
	if item == nil {
		t.Fatal("nil item")
	}
	steps, err := ParseActions(item.ActionsJSON)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 {
		t.Fatalf("steps=%d", len(steps))
	}
	if steps[0].Payload == nil {
		t.Fatal("payload must be non-nil map")
	}
}

func TestEvalItem_nilContext(t *testing.T) {
	_, err := EvalItem(&AutomationItem{Kind: KindRule, IsEnabled: true, TriggerType: EventFriendJoined}, nil)
	if err == nil {
		t.Fatal("want error")
	}
}
