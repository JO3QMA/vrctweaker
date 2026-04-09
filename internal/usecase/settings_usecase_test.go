package usecase

import (
	"context"
	"testing"
	"time"
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

func TestSettingsUseCase_GalleryLastExitAt_roundtrip(t *testing.T) {
	repo := &fakeAppSettingsRepo{m: make(map[string]string)}
	uc := NewSettingsUseCase(repo)
	ctx := context.Background()

	want := time.Date(2025, 3, 21, 12, 30, 45, 123456789, time.FixedZone("JST", 9*3600))
	if err := uc.SetGalleryLastExitAt(ctx, want); err != nil {
		t.Fatalf("SetGalleryLastExitAt: %v", err)
	}
	got, ok := uc.GetGalleryLastExitAt(ctx)
	if !ok {
		t.Fatal("GetGalleryLastExitAt: want ok true")
	}
	if !got.Equal(want.UTC()) {
		t.Errorf("time: got %v, want %v", got, want.UTC())
	}
}

func TestSettingsUseCase_GetGalleryLastExitAt_empty(t *testing.T) {
	repo := &fakeAppSettingsRepo{m: make(map[string]string)}
	uc := NewSettingsUseCase(repo)
	ctx := context.Background()

	_, ok := uc.GetGalleryLastExitAt(ctx)
	if ok {
		t.Error("GetGalleryLastExitAt: want ok false for empty")
	}
}

func TestSettingsUseCase_GetGalleryLastExitAt_invalid(t *testing.T) {
	repo := &fakeAppSettingsRepo{m: map[string]string{keyGalleryLastExitAt: "not-a-time"}}
	uc := NewSettingsUseCase(repo)
	ctx := context.Background()

	_, ok := uc.GetGalleryLastExitAt(ctx)
	if ok {
		t.Error("GetGalleryLastExitAt: want ok false for invalid string")
	}
}

func TestSettingsUseCase_SuppressSleepWhileVRChat_roundtrip(t *testing.T) {
	repo := &fakeAppSettingsRepo{m: make(map[string]string)}
	uc := NewSettingsUseCase(repo)
	ctx := context.Background()

	if err := uc.SetSuppressSleepWhileVRChat(ctx, true); err != nil {
		t.Fatalf("SetSuppressSleepWhileVRChat: %v", err)
	}
	on, err := uc.GetSuppressSleepWhileVRChat(ctx)
	if err != nil {
		t.Fatalf("GetSuppressSleepWhileVRChat: %v", err)
	}
	if !on {
		t.Fatal("want true")
	}
	if repo.m[keySuppressSleepWhileVRChat] != "true" {
		t.Fatalf("stored value: got %q", repo.m[keySuppressSleepWhileVRChat])
	}
	if err2 := uc.SetSuppressSleepWhileVRChat(ctx, false); err2 != nil {
		t.Fatalf("SetSuppressSleepWhileVRChat false: %v", err2)
	}
	off, err := uc.GetSuppressSleepWhileVRChat(ctx)
	if err != nil {
		t.Fatalf("GetSuppressSleepWhileVRChat: %v", err)
	}
	if off {
		t.Fatal("want false")
	}
}

func TestSettingsUseCase_GetSuppressSleepWhileVRChat_defaultFalse(t *testing.T) {
	repo := &fakeAppSettingsRepo{m: make(map[string]string)}
	uc := NewSettingsUseCase(repo)
	ctx := context.Background()

	v, err := uc.GetSuppressSleepWhileVRChat(ctx)
	if err != nil {
		t.Fatalf("GetSuppressSleepWhileVRChat: %v", err)
	}
	if v {
		t.Fatal("want default false")
	}
}

func TestSettingsUseCase_Language_roundtripAndValidation(t *testing.T) {
	repo := &fakeAppSettingsRepo{m: make(map[string]string)}
	uc := NewSettingsUseCase(repo)
	ctx := context.Background()

	v, err := uc.GetLanguage(ctx)
	if err != nil {
		t.Fatalf("GetLanguage: %v", err)
	}
	if v != "" {
		t.Fatalf("want empty unset, got %q", v)
	}
	if err = uc.SetLanguage(ctx, "ja"); err != nil {
		t.Fatalf("SetLanguage: %v", err)
	}
	v, err = uc.GetLanguage(ctx)
	if err != nil {
		t.Fatalf("GetLanguage: %v", err)
	}
	if v != "ja" {
		t.Fatalf("got %q, want ja", v)
	}
	if err := uc.SetLanguage(ctx, "xx"); err == nil {
		t.Fatal("expected error for unsupported language")
	}
}
