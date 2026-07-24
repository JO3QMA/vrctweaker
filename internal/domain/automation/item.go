package automation

import "fmt"

// Item kinds.
const (
	KindRule   = "rule"
	KindScript = "script"
)

// Event names (catalog v1).
const (
	EventFriendJoined  = "friend_joined"
	EventScheduleTick  = "schedule.tick"
	EventVRChatProcess = "vrchat.process"
)

// Action names (catalog).
const (
	ActionChangeStatus        = "change_status"
	ActionSetPowerPlan        = "set_power_plan"
	ActionSetVRChatWindowSize = "set_vrchat_window_size"
)

// Legacy trigger constants (friend_joined matches EventFriendJoined).
const (
	TriggerFriendJoined = EventFriendJoined
	TriggerAFKDetected  = "afk_detected"
)

// AutomationItem is a rule or script in the automation list.
type AutomationItem struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Kind           string `json:"kind"`
	IsEnabled      bool   `json:"isEnabled"`
	TriggerType    string `json:"triggerType,omitempty"`
	ScheduleJSON   string `json:"scheduleJson,omitempty"`
	ConditionsJSON string `json:"conditionsJson,omitempty"`
	ActionsJSON    string `json:"actionsJson,omitempty"`
	ScriptSource   string `json:"scriptSource,omitempty"`
}

// ActionStep is one step in an action sequence.
type ActionStep struct {
	Type            string                 `json:"type"`
	Payload         map[string]interface{} `json:"payload,omitempty"`
	ContinueOnError bool                   `json:"continueOnError,omitempty"`
}

// Condition is a preset automation condition.
type Condition struct {
	Type      string `json:"type"`
	VRCUserID string `json:"vrcUserId,omitempty"`
}

// ScheduleRule is weekday + time in local TZ.
type ScheduleRule struct {
	Weekdays []int `json:"weekdays"` // 0=Sunday .. 6=Saturday
	Hour     int   `json:"hour"`
	Minute   int   `json:"minute"`
}

// Validate checks schedule field ranges.
func (s ScheduleRule) Validate() error {
	if s.Hour < 0 || s.Hour > 23 {
		return fmt.Errorf("hour must be 0-23")
	}
	if s.Minute < 0 || s.Minute > 59 {
		return fmt.Errorf("minute must be 0-59")
	}
	if len(s.Weekdays) == 0 {
		return fmt.Errorf("weekdays required")
	}
	seen := make(map[int]struct{}, len(s.Weekdays))
	for _, d := range s.Weekdays {
		if d < 0 || d > 6 {
			return fmt.Errorf("weekdays must be 0-6")
		}
		if _, ok := seen[d]; ok {
			return fmt.Errorf("weekdays must be unique")
		}
		seen[d] = struct{}{}
	}
	return nil
}

// RunLogEntry is one automation execution record for the UI.
type RunLogEntry struct {
	At               string `json:"at"`
	ItemID           string `json:"itemId"`
	ItemName         string `json:"itemName"`
	EventType        string `json:"eventType"`
	Success          bool   `json:"success"`
	ActionsCompleted int    `json:"actionsCompleted"`
	ActionsTotal     int    `json:"actionsTotal"`
	ContextLabel     string `json:"contextLabel,omitempty"`
	ErrorSummary     string `json:"errorSummary,omitempty"`
}

// RuntimeStatus is automation subsystem health for the UI.
type RuntimeStatus struct {
	Available bool   `json:"available"`
	ReasonKey string `json:"reasonKey,omitempty"`
}

// DetectedPowerPlan is an OS power scheme.
type DetectedPowerPlan struct {
	GUID string `json:"guid"`
	Name string `json:"name"`
}
