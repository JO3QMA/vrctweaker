package automation

// AutomationRule represents an IF-THEN automation rule.
type AutomationRule struct {
	ID            string
	Name          string
	TriggerType   string // e.g., afk_detected, friend_joined
	ConditionJSON string // JSON parameters for the trigger
	ActionType    string // e.g., change_status
	ActionPayload string // JSON parameters for the action
	IsEnabled    bool
}
