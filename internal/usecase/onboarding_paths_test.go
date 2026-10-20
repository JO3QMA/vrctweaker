package usecase

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"vrchat-tweaker/internal/platform/paths"
)

func TestResolveOnboardingFolderDefaults_usesSavedPaths(t *testing.T) {
	dir := t.TempDir()
	logDir := filepath.Join(dir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	repo := &fakeAppSettingsRepo{m: map[string]string{
		keyOutputLogPath:     logDir,
		keyVRChatPathWindows: "C:\\VRChat\\launch.exe",
	}}
	uc := NewSettingsUseCase(repo)
	ctx := context.Background()

	got, err := uc.ResolveOnboardingFolderDefaults(ctx, func() (string, error) {
		return filepath.Join(dir, "Pictures", "VRChat"), nil
	})
	if err != nil {
		t.Fatalf("ResolveOnboardingFolderDefaults: %v", err)
	}
	if got.PathSettings.OutputLogPath != logDir {
		t.Errorf("OutputLogPath: got %q", got.PathSettings.OutputLogPath)
	}
	if got.PathSettings.VRChatPathWindows != "C:\\VRChat\\launch.exe" {
		t.Errorf("VRChatPathWindows: got %q", got.PathSettings.VRChatPathWindows)
	}
}

func TestResolveOnboardingFolderDefaults_fillsOutputLogFromDefaultDir(t *testing.T) {
	cfgDir, err := paths.VRChatDataDir()
	if err != nil {
		t.Skip(err)
	}
	if mkErr := os.MkdirAll(cfgDir, 0o755); mkErr != nil {
		t.Skip(mkErr)
	}
	uc := NewSettingsUseCase(&fakeAppSettingsRepo{m: map[string]string{}})
	got, err := uc.ResolveOnboardingFolderDefaults(context.Background(), func() (string, error) {
		return "", nil
	})
	if err != nil {
		t.Fatalf("ResolveOnboardingFolderDefaults: %v", err)
	}
	want := paths.DefaultVRChatOutputLogDir()
	if want == "" {
		t.Skip("no default output log dir")
	}
	if got.PathSettings.OutputLogPath != want {
		t.Errorf("OutputLogPath: got %q want %q", got.PathSettings.OutputLogPath, want)
	}
}

func TestResolveOnboardingFolderDefaults_pictureFromCallback(t *testing.T) {
	dir := t.TempDir()
	pic := filepath.Join(dir, "MyPics", "VRChat")
	if err := os.MkdirAll(pic, 0o755); err != nil {
		t.Fatal(err)
	}
	uc := NewSettingsUseCase(&fakeAppSettingsRepo{m: map[string]string{}})
	got, err := uc.ResolveOnboardingFolderDefaults(context.Background(), func() (string, error) {
		return pic, nil
	})
	if err != nil {
		t.Fatalf("ResolveOnboardingFolderDefaults: %v", err)
	}
	if got.PictureOutputFolder != pic {
		t.Errorf("PictureOutputFolder: got %q want %q", got.PictureOutputFolder, pic)
	}
}
