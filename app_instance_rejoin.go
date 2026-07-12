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
		log.Printf("instance rejoin: save last launch profile: %v", err)
	}
	return nil
}

// GetInstanceRejoinSection returns Dashboard Instance rejoin UI state, or nil when the section is hidden.
func (a *App) GetInstanceRejoinSection() (*InstanceRejoinSectionDTO, error) {
	target, err := a.activity.GetRejoinTarget(a.ctx)
	if err != nil {
		log.Printf("instance rejoin: get target: %v", err)
		return nil, nil
	}
	if target == nil {
		return nil, nil
	}
	profiles, err := a.launcher.ListProfiles(a.ctx)
	if err != nil {
		log.Printf("instance rejoin: list profiles: %v", err)
		return nil, nil
	}
	if len(profiles) == 0 {
		return nil, nil
	}
	lastID, err := a.settings.GetLastLaunchProfileID(a.ctx)
	if err != nil {
		log.Printf("instance rejoin: get last launch profile: %v", err)
		return nil, nil
	}
	selectedID := usecase.ResolveInstanceRejoinProfileID(profiles, lastID)
	return &InstanceRejoinSectionDTO{
		WorldDisplayName:  target.WorldDisplayName,
		Profiles:          toLaunchProfileDTOs(profiles),
		SelectedProfileID: selectedID,
	}, nil
}

// InstanceRejoin launches VRChat into the current Rejoin target using the given launch profile.
func (a *App) InstanceRejoin(profileID string) error {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return fmt.Errorf("profile id is required")
	}
	target, err := a.activity.GetRejoinTarget(a.ctx)
	if err != nil {
		return err
	}
	if target == nil {
		return fmt.Errorf("no rejoin target available")
	}
	vrchatPath, steamPath, outputLogPath, err := a.launchPathSettings()
	if err != nil {
		return err
	}
	launchErr := a.launcher.LaunchToWorld(a.ctx, profileID, target.InstanceID, vrchatPath, steamPath, outputLogPath)
	return a.setLastLaunchProfileOnSuccess(profileID, launchErr)
}
