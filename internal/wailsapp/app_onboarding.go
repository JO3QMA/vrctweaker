package wailsapp

import (
	"strings"

	"vrchat-tweaker/internal/domain/vrchatconfig"
	"vrchat-tweaker/internal/usecase"
)

// OnboardingFolderDefaultsDTO is shown on the first-launch folder step.
type OnboardingFolderDefaultsDTO struct {
	VrchatPathWindows   string `json:"vrchatPathWindows"`
	SteamPathLinux      string `json:"steamPathLinux"`
	OutputLogPath       string `json:"outputLogPath"`
	PictureOutputFolder string `json:"pictureOutputFolder"`
}

// GetFirstLaunchGuideCompleted returns whether the setup wizard was completed or skipped.
func (a *App) GetFirstLaunchGuideCompleted() (bool, error) {
	if a.settings == nil {
		return false, nil
	}
	return a.settings.GetFirstLaunchGuideCompleted(a.ctx)
}

// SetFirstLaunchGuideCompleted marks the setup wizard as done.
func (a *App) SetFirstLaunchGuideCompleted(completed bool) error {
	if a.settings == nil {
		return nil
	}
	return a.settings.SetFirstLaunchGuideCompleted(a.ctx, completed)
}

// GetOnboardingFolderDefaults returns path settings with auto-detected values for empty fields.
func (a *App) GetOnboardingFolderDefaults() (OnboardingFolderDefaultsDTO, error) {
	if a.settings == nil {
		return OnboardingFolderDefaultsDTO{}, nil
	}
	defs, err := a.settings.ResolveOnboardingFolderDefaults(a.ctx, a.resolveOnboardingPictureFolder)
	if err != nil {
		return OnboardingFolderDefaultsDTO{}, err
	}
	if defs == nil {
		return OnboardingFolderDefaultsDTO{}, nil
	}
	return OnboardingFolderDefaultsDTO{
		VrchatPathWindows:   defs.PathSettings.VRChatPathWindows,
		SteamPathLinux:      defs.PathSettings.SteamPathLinux,
		OutputLogPath:       defs.PathSettings.OutputLogPath,
		PictureOutputFolder: defs.PictureOutputFolder,
	}, nil
}

func (a *App) resolveOnboardingPictureFolder() (string, error) {
	if a.vrchatConfigRepo != nil {
		cfg, err := a.vrchatConfigRepo.Read()
		if err == nil && cfg != nil {
			if p := strings.TrimSpace(cfg.PictureOutputFolder); p != "" {
				return p, nil
			}
		}
	}
	return a.DefaultVRChatPictureFolder()
}

// SaveOnboardingFolderSettings persists path settings and optional picture output folder from the wizard.
func (a *App) SaveOnboardingFolderSettings(defs OnboardingFolderDefaultsDTO) error {
	if a.settings == nil {
		return nil
	}
	ps := usecase.PathSettings{
		VRChatPathWindows: defs.VrchatPathWindows,
		SteamPathLinux:    defs.SteamPathLinux,
		OutputLogPath:     defs.OutputLogPath,
	}
	if err := a.settings.SetPathSettings(a.ctx, &ps); err != nil {
		return err
	}
	pic := strings.TrimSpace(defs.PictureOutputFolder)
	if pic == "" || a.vrchatConfigRepo == nil {
		return nil
	}
	cfg, err := a.vrchatConfigRepo.Read()
	if err != nil {
		cfg = &vrchatconfig.VRChatConfig{}
	}
	cfg.PictureOutputFolder = pic
	return a.vrchatConfigRepo.Write(cfg)
}
