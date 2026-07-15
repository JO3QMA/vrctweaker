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

// GetDashboardLaunchBlock returns Dashboard launch block UI state (always shown when load succeeds).
func (a *App) GetDashboardLaunchBlock() (*DashboardLaunchBlockDTO, error) {
	profiles, err := a.launcher.ListProfiles(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("dashboard launch block: list profiles: %w", err)
	}
	lastID, err := a.settings.GetLastLaunchProfileID(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("dashboard launch block: get last launch profile: %w", err)
	}
	selectedID := usecase.ResolveInstanceRejoinProfileID(profiles, lastID)

	var rejoin *DashboardRejoinDTO
	target, err := a.activity.GetRejoinTarget(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("dashboard launch block: get rejoin target: %w", err)
	}
	if target != nil {
		rejoin = &DashboardRejoinDTO{
			PlaySessionID:    target.PlaySessionID,
			WorldDisplayName: target.WorldDisplayName,
		}
	}
	return &DashboardLaunchBlockDTO{
		Profiles:          toLaunchProfileDTOs(profiles),
		SelectedProfileID: selectedID,
		Rejoin:            rejoin,
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
