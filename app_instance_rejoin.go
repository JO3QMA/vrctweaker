package main

import (
	"fmt"
	"log"
	"strings"

	"vrchat-tweaker/internal/usecase"
)

func (a *App) launchPathSettings() (vrchatPath, steamPath, outputLogPath string, err error) {
	ps, err := a.settings.GetPathSettings(a.ctx)
	if err != nil {
		return "", "", "", err
	}
	if ps != nil {
		vrchatPath = ps.VRChatPathWindows
		steamPath = ps.SteamPathLinux
		outputLogPath = ps.OutputLogPath
	}
	return vrchatPath, steamPath, outputLogPath, nil
}

func (a *App) setLastLaunchProfileOnSuccess(profileID string, launchErr error) error {
	if launchErr != nil {
		return launchErr
	}
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return nil
	}
	if err := a.settings.SetLastLaunchProfileID(a.ctx, profileID); err != nil {
		// ponytail: best-effort — launch already succeeded; do not fail the caller.
		// Upgrade path: surface to user or retry persistence.
		log.Printf("instance rejoin: save last launch profile: %v", err)
	}
	return nil
}

// GetInstanceRejoinSection returns Dashboard Instance rejoin UI state, or nil when the section is hidden.
func (a *App) GetInstanceRejoinSection() (*InstanceRejoinSectionDTO, error) {
	target, err := a.activity.GetRejoinTarget(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("instance rejoin: get target: %w", err)
	}
	if target == nil {
		return nil, nil
	}
	profiles, err := a.launcher.ListProfiles(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("instance rejoin: list profiles: %w", err)
	}
	if len(profiles) == 0 {
		return nil, nil
	}
	lastID, err := a.settings.GetLastLaunchProfileID(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("instance rejoin: get last launch profile: %w", err)
	}
	selectedID := usecase.ResolveInstanceRejoinProfileID(profiles, lastID)
	return &InstanceRejoinSectionDTO{
		PlaySessionID:     target.PlaySessionID,
		WorldDisplayName:  target.WorldDisplayName,
		Profiles:          toLaunchProfileDTOs(profiles),
		SelectedProfileID: selectedID,
	}, nil
}

// InstanceRejoin launches VRChat into the Rejoin target for playSessionID using the given launch profile.
func (a *App) InstanceRejoin(profileID, playSessionID string) error {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return fmt.Errorf("profile id is required")
	}
	target, err := a.activity.ResolveRejoinTarget(a.ctx, playSessionID)
	if err != nil {
		return err
	}
	vrchatPath, steamPath, outputLogPath, err := a.launchPathSettings()
	if err != nil {
		return err
	}
	launchErr := a.launcher.LaunchToWorld(a.ctx, profileID, target.InstanceID, vrchatPath, steamPath, outputLogPath)
	return a.setLastLaunchProfileOnSuccess(profileID, launchErr)
}
