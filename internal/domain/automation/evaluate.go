package automation

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// Event is an automation bus message.
type Event struct {
	Type    string
	Payload map[string]interface{}
	// At is the evaluation wall time (schedule ticks). Zero means time.Now at eval.
	At time.Time
}

// VRChatProcessChecker reports whether VRChat is running.
type VRChatProcessChecker interface {
	VRChatRunning() (bool, error)
}

// EvalItem returns whether a rule item should run its actions.
func EvalItem(item *AutomationItem, ctx *EvalContext) (bool, error) {
	if item == nil || !item.IsEnabled || item.Kind != KindRule {
		return false, nil
	}
	if ctx == nil {
		return false, fmt.Errorf("eval context is nil")
	}
	if item.TriggerType != ctx.TriggerType {
		return false, nil
	}
	if item.TriggerType == EventScheduleTick {
		sched, err := ParseSchedule(item.ScheduleJSON)
		if err != nil {
			return false, err
		}
		now := ctx.Now
		if now.IsZero() {
			now = time.Now()
		}
		if !ScheduleMatches(sched, now) {
			return false, nil
		}
	}
	conds, err := ParseConditions(item.ConditionsJSON)
	if err != nil {
		return false, err
	}
	conds = CompatibleConditions(item.TriggerType, conds)
	return MatchConditions(conds, ctx), nil
}

// CompatibleConditions drops conditions that cannot apply to the trigger
// (e.g. leftover friend_is after switching a rule to schedule.tick).
func CompatibleConditions(trigger string, conds []Condition) []Condition {
	if len(conds) == 0 {
		return conds
	}
	out := make([]Condition, 0, len(conds))
	for _, c := range conds {
		if c.Type == "friend_is" && trigger != EventFriendJoined {
			continue
		}
		out = append(out, c)
	}
	return out
}

// NextMinuteBoundary returns the start of the next wall-clock minute after now.
func NextMinuteBoundary(now time.Time) time.Time {
	return now.Truncate(ScheduleTickResolution).Add(ScheduleTickResolution)
}

// ParseConditions decodes conditions JSON.
func ParseConditions(raw string) ([]Condition, error) {
	if raw == "" {
		return nil, nil
	}
	var conds []Condition
	if err := json.Unmarshal([]byte(raw), &conds); err != nil {
		return nil, err
	}
	return conds, nil
}

// ParseActions decodes the action sequence.
func ParseActions(raw string) ([]ActionStep, error) {
	if raw == "" {
		return nil, nil
	}
	var steps []ActionStep
	if err := json.Unmarshal([]byte(raw), &steps); err != nil {
		return nil, err
	}
	if len(steps) > MaxActionsPerItem {
		return nil, fmt.Errorf("too many actions: max %d", MaxActionsPerItem)
	}
	return steps, nil
}

// ParseSchedule decodes schedule JSON.
func ParseSchedule(raw string) (*ScheduleRule, error) {
	if raw == "" {
		return nil, fmt.Errorf("schedule required")
	}
	var s ScheduleRule
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return nil, err
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return &s, nil
}

// MatchConditions returns true when all preset conditions pass (AND).
func MatchConditions(conds []Condition, ctx *EvalContext) bool {
	for _, c := range conds {
		if !matchOneCondition(c, ctx) {
			return false
		}
	}
	return true
}

func matchOneCondition(c Condition, ctx *EvalContext) bool {
	switch c.Type {
	case "vrchat_running":
		return ctx.VRChatRunningOK && ctx.VRChatRunning
	case "friend_is":
		if c.VRCUserID == "" || ctx.Payload == nil {
			return false
		}
		got, _ := ctx.Payload["vrc_user_id"].(string)
		return got == c.VRCUserID
	default:
		// ponytail: unknown preset types fail closed.
		return false
	}
}

// ScheduleMatches reports whether t (local) matches the schedule rule.
func ScheduleMatches(s *ScheduleRule, t time.Time) bool {
	if s == nil {
		return false
	}
	loc := t.Location()
	t = t.In(loc)
	wd := int(t.Weekday())
	dayOK := false
	for _, d := range s.Weekdays {
		if d == wd {
			dayOK = true
			break
		}
	}
	if !dayOK {
		return false
	}
	return t.Hour() == s.Hour && t.Minute() == s.Minute
}

// SortItemsByID returns a copy sorted by ID ascending.
func SortItemsByID(items []*AutomationItem) []*AutomationItem {
	out := append([]*AutomationItem(nil), items...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

// LegacyEvalRule evaluates old AutomationRule shape (tests / migration).
func LegacyEvalRule(rule *AutomationRule, ctx *EvalContext) (bool, error) {
	if rule == nil || !rule.IsEnabled {
		return false, nil
	}
	item := RuleToItem(rule)
	return EvalItem(item, ctx)
}

// RuleToItem converts a legacy rule row.
func RuleToItem(rule *AutomationRule) *AutomationItem {
	if rule == nil {
		return nil
	}
	payload := map[string]interface{}{}
	if rule.ActionPayload != "" {
		var p map[string]interface{}
		if err := json.Unmarshal([]byte(rule.ActionPayload), &p); err == nil && p != nil {
			payload = p
		}
	}
	// Always emit payload key so omitempty cannot drop an empty map to nil on reload.
	actionsJSON, err := json.Marshal([]map[string]interface{}{
		{"type": rule.ActionType, "payload": payload},
	})
	if err != nil {
		actionsJSON = []byte("[]")
	}
	var conds []Condition
	brokenCond := false
	if rule.ConditionJSON != "" {
		var legacy map[string]interface{}
		if unmarshalErr := json.Unmarshal([]byte(rule.ConditionJSON), &legacy); unmarshalErr != nil {
			// Empty [] would MatchConditions as always-true; fail closed instead.
			brokenCond = true
			conds = []Condition{{Type: "migration_invalid"}}
		} else {
			for k, v := range legacy {
				if k == "vrc_user_id" {
					if s, ok := v.(string); ok {
						conds = append(conds, Condition{Type: "friend_is", VRCUserID: s})
					}
				}
			}
		}
	}
	condsJSON, err := json.Marshal(conds)
	if err != nil {
		condsJSON = []byte(`[{"type":"migration_invalid"}]`)
		brokenCond = true
	}
	return &AutomationItem{
		ID:             rule.ID,
		Name:           rule.Name,
		Kind:           KindRule,
		IsEnabled:      rule.IsEnabled && !brokenCond,
		TriggerType:    rule.TriggerType,
		ConditionsJSON: string(condsJSON),
		ActionsJSON:    string(actionsJSON),
	}
}
