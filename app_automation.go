package main

import (
	"vrchat-tweaker/internal/domain/automation"
)

// AutomationItemDTO is a Wails-facing automation item.
type AutomationItemDTO struct {
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

// AutomationRunLogEntryDTO is one run log row for the UI.
type AutomationRunLogEntryDTO struct {
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

// AutomationRuntimeStatusDTO is automation subsystem health.
type AutomationRuntimeStatusDTO struct {
	Available bool   `json:"available"`
	ReasonKey string `json:"reasonKey,omitempty"`
}

// DetectedPowerPlanDTO is an OS power plan.
type DetectedPowerPlanDTO struct {
	GUID string `json:"guid"`
	Name string `json:"name"`
}

func toAutomationItemDTO(item *automation.AutomationItem) AutomationItemDTO {
	if item == nil {
		return AutomationItemDTO{}
	}
	return AutomationItemDTO{
		ID:             item.ID,
		Name:           item.Name,
		Kind:           item.Kind,
		IsEnabled:      item.IsEnabled,
		TriggerType:    item.TriggerType,
		ScheduleJSON:   item.ScheduleJSON,
		ConditionsJSON: item.ConditionsJSON,
		ActionsJSON:    item.ActionsJSON,
		ScriptSource:   item.ScriptSource,
	}
}

func toAutomationItemDTOs(items []*automation.AutomationItem) []AutomationItemDTO {
	out := make([]AutomationItemDTO, len(items))
	for i, item := range items {
		out[i] = toAutomationItemDTO(item)
	}
	return out
}

func toRunLogDTOs(entries []automation.RunLogEntry) []AutomationRunLogEntryDTO {
	out := make([]AutomationRunLogEntryDTO, len(entries))
	for i, e := range entries {
		out[i] = AutomationRunLogEntryDTO{
			At:               e.At,
			ItemID:           e.ItemID,
			ItemName:         e.ItemName,
			EventType:        e.EventType,
			Success:          e.Success,
			ActionsCompleted: e.ActionsCompleted,
			ActionsTotal:     e.ActionsTotal,
			ContextLabel:     e.ContextLabel,
			ErrorSummary:     e.ErrorSummary,
		}
	}
	return out
}

// ListAutomationItems returns all automation items.
func (a *App) ListAutomationItems() ([]AutomationItemDTO, error) {
	items, err := a.automation.ListItems(a.ctx)
	if err != nil {
		return nil, err
	}
	return toAutomationItemDTOs(items), nil
}

// SaveAutomationItem persists an automation item.
func (a *App) SaveAutomationItem(item AutomationItemDTO) error {
	return a.automation.SaveItem(a.ctx, &automation.AutomationItem{
		ID:             item.ID,
		Name:           item.Name,
		Kind:           item.Kind,
		IsEnabled:      item.IsEnabled,
		TriggerType:    item.TriggerType,
		ScheduleJSON:   item.ScheduleJSON,
		ConditionsJSON: item.ConditionsJSON,
		ActionsJSON:    item.ActionsJSON,
		ScriptSource:   item.ScriptSource,
	})
}

// DeleteAutomationItem removes an automation item.
func (a *App) DeleteAutomationItem(id string) error {
	return a.automation.DeleteItem(a.ctx, id)
}

// ToggleAutomationItem enables or disables an item.
func (a *App) ToggleAutomationItem(id string, enabled bool) error {
	return a.automation.ToggleItem(a.ctx, id, enabled)
}

// GetAutomationRunLog returns recent run log entries.
func (a *App) GetAutomationRunLog() ([]AutomationRunLogEntryDTO, error) {
	return toRunLogDTOs(a.automation.GetRunLog()), nil
}

// GetAutomationRuntimeStatus returns automation subsystem availability.
func (a *App) GetAutomationRuntimeStatus() (AutomationRuntimeStatusDTO, error) {
	st := a.automation.RuntimeStatus()
	return AutomationRuntimeStatusDTO{Available: st.Available, ReasonKey: st.ReasonKey}, nil
}

// ListDetectedPowerPlans returns OS power plans (empty off Windows).
func (a *App) ListDetectedPowerPlans() ([]DetectedPowerPlanDTO, error) {
	plans, err := a.automation.ListDetectedPowerPlans()
	if err != nil {
		return nil, err
	}
	out := make([]DetectedPowerPlanDTO, len(plans))
	for i, p := range plans {
		out[i] = DetectedPowerPlanDTO{GUID: p.GUID, Name: p.Name}
	}
	return out, nil
}
