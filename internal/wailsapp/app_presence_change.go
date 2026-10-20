package wailsapp

import (
	"fmt"

	"vrchat-tweaker/internal/usecase"
)

// GetPresenceChangeSection returns Dashboard presence change section state.
func (a *App) GetPresenceChangeSection() (*usecase.PresenceChangeSection, error) {
	if a.presenceChange == nil {
		return nil, fmt.Errorf("presence change section: not initialized")
	}
	section, err := a.presenceChange.GetSection(a.ctx)
	if err != nil {
		return nil, err
	}
	if section != nil && section.History == nil {
		// Empty history must stay [] on the wire, matching the former DTO.
		section.History = []string{}
	}
	return section, nil
}

// ApplyPresenceChange updates VRChat presence (status + description) and records history.
func (a *App) ApplyPresenceChange(status, description string) (*usecase.PresenceChangeApplyResult, error) {
	if a.presenceChange == nil {
		return nil, fmt.Errorf("presence change apply: not initialized")
	}
	return a.presenceChange.Apply(a.ctx, status, description)
}
