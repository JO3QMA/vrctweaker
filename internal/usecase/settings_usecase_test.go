package usecase

import (
	"context"
	"testing"
)

// fakeAppSettingsRepo implements settings.AppSettingsRepository for tests.
type fakeAppSettingsRepo struct {
	m map[string]string
}

func (f *fakeAppSettingsRepo) Get(_ context.Context, key string) (string, error) {
	return f.m[key], nil
}

func (f *fakeAppSettingsRepo) Set(_ context.Context, key, value string) error {
	if f.m == nil {
		f.m = make(map[string]string)
	}
	f.m[key] = value
	return nil
}

func (f *fakeAppSettingsRepo) GetAll(_ context.Context) (map[string]string, error) {
	result := make(map[string]string, len(f.m))
	for k, v := range f.m {
		result[k] = v
	}
	return result, nil
}

func TestSettingsUseCase_GetPathSettings_SetPathSettings_roundtrip(t *testing.T) {
	repo := &fakeAppSettingsRepo{m: make(map[string]string)}
	uc := NewSettingsUseCase(repo)
	ctx := context.Background()

	ps := &PathSettings{
		VRChatPathWindows: `C:\Program Files (x86)\Steam\steamapps\common\VRChat\launch.exe`,
		SteamPathLinux:    "/usr/bin/steam",
		OutputLogPath:     "/home/user/.local/share/Steam/logs/output_log.txt",
	}

	if err := uc.SetPathSettings(ctx, ps); err != nil {
		t.Fatalf("SetPathSettings: %v", err)
	}

	got, err := uc.GetPathSettings(ctx)
	if err != nil {
		t.Fatalf("GetPathSettings: %v", err)
	}
	if got == nil {
		t.Fatal("GetPathSettings returned nil")
	}
	if got.VRChatPathWindows != ps.VRChatPathWindows {
		t.Errorf("VRChatPathWindows: got %q, want %q", got.VRChatPathWindows, ps.VRChatPathWindows)
	}
	if got.SteamPathLinux != ps.SteamPathLinux {
		t.Errorf("SteamPathLinux: got %q, want %q", got.SteamPathLinux, ps.SteamPathLinux)
	}
	if got.OutputLogPath != ps.OutputLogPath {
		t.Errorf("OutputLogPath: got %q, want %q", got.OutputLogPath, ps.OutputLogPath)
	}
}

func TestSettingsUseCase_SetPathSettings_nil(t *testing.T) {
	repo := &fakeAppSettingsRepo{m: make(map[string]string)}
	uc := NewSettingsUseCase(repo)
	ctx := context.Background()

	if err := uc.SetPathSettings(ctx, nil); err != nil {
		t.Fatalf("SetPathSettings(nil) should not error: %v", err)
	}
}

func TestSettingsUseCase_ValidatePath_empty(t *testing.T) {
	repo := &fakeAppSettingsRepo{}
	uc := NewSettingsUseCase(repo)

	if uc.ValidatePath("") {
		t.Error("ValidatePath(\"\") should return false")
	}
}
