package automation

import "testing"

func TestEvalItem_friendLeft_friendIsCondition(t *testing.T) {
	condsJSON := `[{"type":"friend_is","vrcUserId":"usr_target"}]`
	item := &AutomationItem{
		Kind:           KindRule,
		IsEnabled:      true,
		TriggerType:    EventFriendLeft,
		ConditionsJSON: condsJSON,
		ActionsJSON:    `[{"type":"change_status","payload":{"status":"busy"}}]`,
	}
	ok, err := EvalItem(item, &EvalContext{
		TriggerType: EventFriendLeft,
		Payload:     map[string]interface{}{"vrc_user_id": "usr_target"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected match")
	}
	ok, err = EvalItem(item, &EvalContext{
		TriggerType: EventFriendLeft,
		Payload:     map[string]interface{}{"vrc_user_id": "usr_other"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected no match for other friend")
	}
}

func TestCompatibleConditions_friendLeft_keepsFriendIs(t *testing.T) {
	conds := []Condition{{Type: "friend_is", VRCUserID: "usr_x"}}
	got := CompatibleConditions(EventFriendLeft, conds)
	if len(got) != 1 || got[0].VRCUserID != "usr_x" {
		t.Fatalf("got %#v", got)
	}
}
