package main

import (
	"fmt"
)

// GetPresenceChangeSection returns Dashboard presence change section state.
func (a *App) GetPresenceChangeSection() (*PresenceChangeSectionDTO, error) {
	if a.presenceChange == nil {
		return nil, fmt.Errorf("presence change section: not initialized")
	}
	section, err := a.presenceChange.GetSection(a.ctx)
	if err != nil {
		return nil, err
	}
	history := section.History
	if history == nil {
		history = []string{}
	}
	return &PresenceChangeSectionDTO{
		LoggedIn:          section.LoggedIn,
		Status:            section.Status,
		StatusDescription: section.StatusDescription,
		History:           history,
	}, nil
}

// ApplyPresenceChange updates VRChat presence (status + description) and records history.
func (a *App) ApplyPresenceChange(status, description string) (*PresenceChangeApplyResultDTO, error) {
	if a.presenceChange == nil {
		return nil, fmt.Errorf("presence change apply: not initialized")
	}
	res, err := a.presenceChange.Apply(a.ctx, status, description)
	if err != nil {
		return nil, err
	}
	return &PresenceChangeApplyResultDTO{
		Status:            res.Status,
		StatusDescription: res.StatusDescription,
	}, nil
}
