package wailsapp

import (
	"errors"
	"os"
	"strings"

	"vrchat-tweaker/internal/domain/vrchatconfig"
	"vrchat-tweaker/internal/usecase"
)

var errSettingsNotInitialized = errors.New("settings repository is not initialized")

// OnboardingFolderDefaultsDTO is shown on the first-launch folder step.
type OnboardingFolderDefaultsDTO struct {
	VrchatPathWindows   string `json:"vrchatPathWindows"`
	SteamPathLinux      string `json:"steamPathLinux"`
	OutputLogPath       string `json:"outputLogPath"`
	PictureOutputFolder string `json:"pictureOutputFolder"`
}

// OnboardingFolderSaveDTO persists only fields the user explicitly applied in the wizard.
type OnboardingFolderSaveDTO struct {
	VrchatPathWindows        string `json:"vrchatPathWindows"`
	SteamPathLinux           string `json:"steamPathLinux"`
	OutputLogPath            string `json:"outputLogPath"`
	PictureOutputFolder      string `json:"pictureOutputFolder"`
	ApplyVrchatPathWindows   bool   `json:"applyVrchatPathWindows"`
	ApplySteamPathLinux      bool   `json:"applySteamPathLinux"`
	ApplyOutputLogPath       bool   `json:"applyOutputLogPath"`
	ApplyPictureOutputFolder bool   `json:"applyPictureOutputFolder"`
}

// GetFirstLaunchGuideCompleted returns whether the setup wizard was completed or skipped.
func (a *App) GetFirstLaunchGuideCompleted() (bool, error) {
	if a.settings == nil {
		return false, errSettingsNotInitialized
	}
	return a.settings.GetFirstLaunchGuideCompleted(a.ctx)
}

// SetFirstLaunchGuideCompleted marks the setup wizard as done.
func (a *App) SetFirstLaunchGuideCompleted(completed bool) error {
	if a.settings == nil {
		return errSettingsNotInitialized
	}
	return a.settings.SetFirstLaunchGuideCompleted(a.ctx, completed)
}

// GetOnboardingFolderDefaults returns path settings with auto-detected values for empty fields.
func (a *App) GetOnboardingFolderDefaults() (OnboardingFolderDefaultsDTO, error) {
	if a.settings == nil {
		return OnboardingFolderDefaultsDTO{}, errSettingsNotInitialized
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

// SaveOnboardingFolderSettings persists only wizard fields marked with Apply* and non-empty values.
func (a *App) SaveOnboardingFolderSettings(save OnboardingFolderSaveDTO) error {
	if a.settings == nil {
		return errSettingsNotInitialized
	}
	current, err := a.settings.GetPathSettings(a.ctx)
	if err != nil {
		return err
	}
	if current == nil {
		current = &usecase.PathSettings{}
	}
	backup := *current
	ps := *current
	pathChanged := false

	if save.ApplyVrchatPathWindows {
		trimmed := strings.TrimSpace(save.VrchatPathWindows)
		if trimmed != "" && trimmed != ps.VRChatPathWindows {
			ps.VRChatPathWindows = trimmed
			pathChanged = true
		}
	}
	if save.ApplySteamPathLinux {
		trimmed := strings.TrimSpace(save.SteamPathLinux)
		if trimmed != "" && trimmed != ps.SteamPathLinux {
			ps.SteamPathLinux = trimmed
			pathChanged = true
		}
	}
	if save.ApplyOutputLogPath {
		trimmed := strings.TrimSpace(save.OutputLogPath)
		if trimmed != "" && trimmed != ps.OutputLogPath {
			ps.OutputLogPath = trimmed
			pathChanged = true
		}
	}

	pic := ""
	if save.ApplyPictureOutputFolder {
		pic = strings.TrimSpace(save.PictureOutputFolder)
		if pic != "" {
			info, statErr := os.Stat(pic)
			if statErr != nil || info == nil || !info.IsDir() {
				return errors.New("picture output folder is not an accessible directory")
			}
		}
	}

	if pathChanged {
		if err := a.settings.SetPathSettings(a.ctx, &ps); err != nil {
			return err
		}
	}

	if pic != "" {
		if a.vrchatConfigRepo == nil {
			return errors.New("vrchat config repository is not initialized")
		}
		if err := a.writeVRChatPictureOutputFolder(pic); err != nil {
			if pathChanged {
				_ = a.settings.SetPathSettings(a.ctx, &backup)
			}
			return err
		}
	}

	return nil
}

func (a *App) writeVRChatPictureOutputFolder(pic string) error {
	exists, err := a.vrchatConfigRepo.Exists()
	if err != nil {
		return err
	}
	var cfg *vrchatconfig.VRChatConfig
	if exists {
		cfg, err = a.vrchatConfigRepo.Read()
		if err != nil {
			return err
		}
	}
	if cfg == nil {
		cfg = &vrchatconfig.VRChatConfig{}
	}
	cfg.PictureOutputFolder = pic
	return a.vrchatConfigRepo.Write(cfg)
}
