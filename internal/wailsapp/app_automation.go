package wailsapp

import (
	"errors"

	"vrchat-tweaker/internal/domain/automation"
)

var errAutomationNotInitialized = errors.New("automation not initialized")

func (a *App) requireAutomation() error {
	if a == nil || a.automation == nil || a.ctx == nil {
		return errAutomationNotInitialized
	}
	return nil
}

// ListAutomationItems returns all automation items.
func (a *App) ListAutomationItems() ([]*automation.AutomationItem, error) {
	if err := a.requireAutomation(); err != nil {
		return nil, err
	}
	items, err := a.automation.ListItems(a.ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []*automation.AutomationItem{}, nil
	}
	return items, nil
}

// SaveAutomationItem persists an automation item.
func (a *App) SaveAutomationItem(item automation.AutomationItem) error {
	if err := a.requireAutomation(); err != nil {
		return err
	}
	return a.automation.SaveItem(a.ctx, &item)
}

// DeleteAutomationItem removes an automation item.
func (a *App) DeleteAutomationItem(id string) error {
	if err := a.requireAutomation(); err != nil {
		return err
	}
	return a.automation.DeleteItem(a.ctx, id)
}

// ToggleAutomationItem enables or disables an item.
func (a *App) ToggleAutomationItem(id string, enabled bool) error {
	if err := a.requireAutomation(); err != nil {
		return err
	}
	return a.automation.ToggleItem(a.ctx, id, enabled)
}

// GetAutomationRunLog returns recent run log entries.
func (a *App) GetAutomationRunLog() ([]automation.RunLogEntry, error) {
	if err := a.requireAutomation(); err != nil {
		return nil, err
	}
	entries := a.automation.GetRunLog()
	if entries == nil {
		return []automation.RunLogEntry{}, nil
	}
	return entries, nil
}

// GetAutomationRuntimeStatus returns automation subsystem availability.
func (a *App) GetAutomationRuntimeStatus() (automation.RuntimeStatus, error) {
	if a == nil || a.automation == nil {
		return automation.RuntimeStatus{Available: false, ReasonKey: "not_initialized"}, nil
	}
	return a.automation.RuntimeStatus(), nil
}

// ListDetectedPowerPlans returns OS power plans (empty off Windows).
func (a *App) ListDetectedPowerPlans() ([]automation.DetectedPowerPlan, error) {
	if err := a.requireAutomation(); err != nil {
		return nil, err
	}
	plans, err := a.automation.ListDetectedPowerPlans()
	if err != nil {
		return nil, err
	}
	if plans == nil {
		return []automation.DetectedPowerPlan{}, nil
	}
	return plans, nil
}
