package wailsapp

import (
	"fmt"

	"vrchat-tweaker/internal/domain/launcher"
)

// LaunchProfiles returns all launch profiles.
func (a *App) LaunchProfiles() ([]LaunchProfileDTO, error) {
	list, err := a.launcher.ListProfiles(a.ctx)
	if err != nil {
		return nil, err
	}
	return toLaunchProfileDTOs(list), nil
}

// LaunchVRChatWithArgs starts VRChat with the given arguments string (from GUI state).
// Use when launching with current GUI settings without saving first.
// lastLaunchProfileID updates Last launch profile on success when non-empty (Profile launch).
func (a *App) LaunchVRChatWithArgs(args string, lastLaunchProfileID string) error {
	ps, err := a.settings.GetPathSettings(a.ctx)
	if err != nil {
		return err
	}
	vrchatPath := ""
	steamPath := ""
	outputLogPath := ""
	if ps != nil {
		vrchatPath = ps.VRChatPathWindows
		steamPath = ps.SteamPathLinux
		outputLogPath = ps.OutputLogPath
	}
	launchErr := a.launcher.LaunchWithArgs(a.ctx, args, vrchatPath, steamPath, outputLogPath)
	return a.setLastLaunchProfileOnSuccess(lastLaunchProfileID, launchErr)
}

// LaunchVRChat starts VRChat with the given profile ID.
// Uses path settings (VRChat/Steam paths) from app_settings when configured.
func (a *App) LaunchVRChat(profileID string) error {
	ps, err := a.settings.GetPathSettings(a.ctx)
	if err != nil {
		return err
	}
	vrchatPath := ""
	steamPath := ""
	outputLogPath := ""
	if ps != nil {
		vrchatPath = ps.VRChatPathWindows
		steamPath = ps.SteamPathLinux
		outputLogPath = ps.OutputLogPath
	}
	launchErr := a.launcher.LaunchVRChat(a.ctx, profileID, vrchatPath, steamPath, outputLogPath)
	return a.setLastLaunchProfileOnSuccess(profileID, launchErr)
}

// JoinWorld launches VRChat into the specified world using default profile.
// Uses vrchat://launch?id=<worldID> URL scheme.
func (a *App) JoinWorld(worldID string) error {
	ps, err := a.settings.GetPathSettings(a.ctx)
	if err != nil {
		return err
	}
	vrchatPath := ""
	steamPath := ""
	outputLogPath := ""
	if ps != nil {
		vrchatPath = ps.VRChatPathWindows
		steamPath = ps.SteamPathLinux
		outputLogPath = ps.OutputLogPath
	}
	return a.launcher.LaunchToWorld(a.ctx, "", worldID, vrchatPath, steamPath, outputLogPath)
}

// JoinWorldFromScreenshot launches VRChat into the world associated with the screenshot.
// Returns error if screenshot has no world_id.
func (a *App) JoinWorldFromScreenshot(screenshotID string) error {
	s, err := a.media.GetScreenshot(a.ctx, screenshotID)
	if err != nil {
		return err
	}
	if s == nil {
		return fmt.Errorf("screenshot not found: %s", screenshotID)
	}
	if s.WorldID == "" {
		return fmt.Errorf("screenshot has no world_id")
	}
	return a.JoinWorld(s.WorldID)
}

// SaveLaunchProfile persists a launch profile.
func (a *App) SaveLaunchProfile(p LaunchProfileDTO) error {
	return a.launcher.SaveProfile(a.ctx, toLaunchProfile(p))
}

// DeleteLaunchProfile removes a launch profile by ID.
func (a *App) DeleteLaunchProfile(id string) error {
	return a.launcher.DeleteProfile(a.ctx, id)
}

// ParseLaunchArgsForGUI parses a launch arguments string into GUI fields.
func (a *App) ParseLaunchArgsForGUI(args string) launcher.LaunchArgsParsed {
	p := launcher.ParseLaunchArgsForGUI(args)
	if p == nil {
		return launcher.LaunchArgsParsed{}
	}
	return *p
}

// MergeLaunchArgsForGUI builds a launch arguments string from GUI state.
func (a *App) MergeLaunchArgsForGUI(parsed launcher.LaunchArgsParsed) string {
	return launcher.MergeLaunchArgsForGUI(&parsed)
}
