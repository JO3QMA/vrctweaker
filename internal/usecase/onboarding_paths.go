package usecase

import (
	"context"
	"os"
	"strings"

	"vrchat-tweaker/internal/platform/paths"
)

// OnboardingFolderDefaults holds folder paths shown in the first-launch wizard.
type OnboardingFolderDefaults struct {
	PathSettings        PathSettings
	PictureOutputFolder string
}

// ResolveOnboardingFolderDefaults returns saved path settings with detected defaults
// filled in for display when fields are empty and the default path exists.
func (uc *SettingsUseCase) ResolveOnboardingFolderDefaults(
	ctx context.Context,
	defaultPictureFolder func() (string, error),
) (*OnboardingFolderDefaults, error) {
	ps, err := uc.GetPathSettings(ctx)
	if err != nil {
		return nil, err
	}
	if ps == nil {
		ps = &PathSettings{}
	}
	out := OnboardingFolderDefaults{
		PathSettings: PathSettings{
			VRChatPathWindows: ps.VRChatPathWindows,
			SteamPathLinux:    ps.SteamPathLinux,
			OutputLogPath:     ps.OutputLogPath,
		},
	}
	if strings.TrimSpace(out.PathSettings.VRChatPathWindows) == "" {
		def := defaultVRChatLaunchPathSuggestion()
		if uc.ValidatePath(def) {
			out.PathSettings.VRChatPathWindows = def
		}
	}
	if strings.TrimSpace(out.PathSettings.SteamPathLinux) == "" {
		def := defaultSteamLinuxPathSuggestion()
		if uc.ValidatePath(def) {
			out.PathSettings.SteamPathLinux = def
		}
	}
	if strings.TrimSpace(out.PathSettings.OutputLogPath) == "" {
		def := paths.DefaultVRChatOutputLogDir()
		if def != "" {
			if info, statErr := os.Stat(def); statErr == nil && info != nil && info.IsDir() {
				out.PathSettings.OutputLogPath = def
			}
		}
	}
	if defaultPictureFolder != nil {
		pic, picErr := defaultPictureFolder()
		if picErr == nil {
			out.PictureOutputFolder = strings.TrimSpace(pic)
		}
	}
	return &out, nil
}

func defaultVRChatLaunchPathSuggestion() string {
	return "C:\\Program Files (x86)\\Steam\\steamapps\\common\\VRChat\\launch.exe"
}

func defaultSteamLinuxPathSuggestion() string {
	return "/usr/bin/steam"
}
