package wailsapp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"vrchat-tweaker/internal/usecase"
)

func TestApp_GetFirstLaunchGuideCompleted_defaultFalse(t *testing.T) {
	a := &App{
		ctx:      context.Background(),
		settings: usecase.NewSettingsUseCase(&memSettingsRepo{m: map[string]string{}}),
	}
	done, err := a.GetFirstLaunchGuideCompleted()
	if err != nil {
		t.Fatalf("GetFirstLaunchGuideCompleted: %v", err)
	}
	if done {
		t.Fatal("want false")
	}
}

func TestApp_GetFirstLaunchGuideCompleted_nilSettings(t *testing.T) {
	a := &App{ctx: context.Background()}
	_, err := a.GetFirstLaunchGuideCompleted()
	if !errors.Is(err, errSettingsNotInitialized) {
		t.Fatalf("err = %v", err)
	}
}

func TestApp_SetFirstLaunchGuideCompleted(t *testing.T) {
	repo := &memSettingsRepo{m: map[string]string{}}
	a := &App{ctx: context.Background(), settings: usecase.NewSettingsUseCase(repo)}
	if err := a.SetFirstLaunchGuideCompleted(true); err != nil {
		t.Fatalf("SetFirstLaunchGuideCompleted: %v", err)
	}
	done, err := a.GetFirstLaunchGuideCompleted()
	if err != nil || !done {
		t.Fatalf("want true got %v err=%v", done, err)
	}
}

func TestApp_SaveOnboardingFolderSettings_onlyTouchedFields(t *testing.T) {
	repo := &memSettingsRepo{m: map[string]string{
		"vrchat_path_windows": "C:\\saved\\launch.exe",
	}}
	a := &App{ctx: context.Background(), settings: usecase.NewSettingsUseCase(repo)}
	logDir := filepath.Join(t.TempDir(), "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	err := a.SaveOnboardingFolderSettings(OnboardingFolderSaveDTO{
		OutputLogPath:      logDir,
		ApplyOutputLogPath: true,
	})
	if err != nil {
		t.Fatalf("SaveOnboardingFolderSettings: %v", err)
	}
	ps, err := a.settings.GetPathSettings(a.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ps.OutputLogPath != logDir {
		t.Fatalf("output log: %q", ps.OutputLogPath)
	}
	if ps.VRChatPathWindows != "C:\\saved\\launch.exe" {
		t.Fatalf("vrchat should stay saved: %q", ps.VRChatPathWindows)
	}
}
