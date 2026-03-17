package automation

import (
	"encoding/json"
	"fmt"
)

// TriggerTypes and ActionTypes for validation.
const (
	TriggerAFKDetected   = "afk_detected"
	TriggerFriendJoined  = "friend_joined"
	ActionChangeStatus   = "change_status"
)

// EvalContext provides data for rule evaluation.
type EvalContext struct {
	TriggerType string
	Payload    map[string]interface{}
}

// EvalResult represents the outcome of rule evaluation.
type EvalResult struct {
	ShouldFire bool
	ActionType string
	ActionPayload map[string]interface{}
}

// EvalRule evaluates an automation rule against the given context.
func EvalRule(rule *AutomationRule, ctx *EvalContext) (*EvalResult, error) {
	if !rule.IsEnabled || rule.TriggerType != ctx.TriggerType {
		return &EvalResult{ShouldFire: false}, nil
	}

	// Parse condition and check if it matches.
	// - cond が空なら常に true
	// - cond にキーがある場合は ctx.Payload と比較して全て一致したら true
	// - 型不一致/キー不足は false 扱い（panic しない）
	var cond map[string]interface{}
	if rule.ConditionJSON != "" {
		if err := json.Unmarshal([]byte(rule.ConditionJSON), &cond); err != nil {
			return nil, err
		}
		if len(cond) > 0 && !conditionMatchesPayload(cond, ctx.Payload) {
			return &EvalResult{ShouldFire: false}, nil
		}
	}

	var actionPayload map[string]interface{}
	if rule.ActionPayload != "" {
		if err := json.Unmarshal([]byte(rule.ActionPayload), &actionPayload); err != nil {
			return nil, err
		}
	}

	return &EvalResult{
		ShouldFire:    true,
		ActionType:    rule.ActionType,
		ActionPayload: actionPayload,
	}, nil
}

// conditionMatchesPayload returns true if all keys in cond match ctx.Payload.
// Type mismatch or missing key in payload yields false (no panic).
func conditionMatchesPayload(cond map[string]interface{}, payload map[string]interface{}) bool {
	if payload == nil {
		return len(cond) == 0
	}
	for k, condVal := range cond {
		payloadVal, ok := payload[k]
		if !ok {
			return false
		}
		if !valuesEqual(condVal, payloadVal) {
			return false
		}
	}
	return true
}

func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// Normalize for comparison: use string representation for numeric/string mix
	sa := fmt.Sprintf("%v", a)
	sb := fmt.Sprintf("%v", b)
	return sa == sb
}
