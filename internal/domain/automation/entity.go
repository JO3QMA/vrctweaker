package automation

// AutomationRule represents an IF-THEN automation rule.
type AutomationRule struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	TriggerType   string `json:"triggerType"` // e.g., afk_detected, friend_joined
	ConditionJSON string `json:"conditionJson"`
	ActionType    string `json:"actionType"` // e.g., change_status
	ActionPayload string `json:"actionPayload"`
	IsEnabled     bool   `json:"isEnabled"`
}
