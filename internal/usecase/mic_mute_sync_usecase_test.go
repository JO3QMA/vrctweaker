package usecase

import (
	"context"
	"testing"

	"vrchat-tweaker/internal/domain/micmutesync"
	"vrchat-tweaker/internal/infrastructure/vrchatosc"
)

type mapSettingsRepo map[string]string

func (m mapSettingsRepo) Get(_ context.Context, key string) (string, error) {
	return m[key], nil
}

func (m mapSettingsRepo) Set(_ context.Context, key, value string) error {
	m[key] = value
	return nil
}

func (m mapSettingsRepo) GetAll(_ context.Context) (map[string]string, error) {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out, nil
}

func newTestMicMuteSyncUC(repo mapSettingsRepo) *MicMuteSyncUseCase {
	return NewMicMuteSyncUseCase(repo, vrchatosc.NewListener(), vrchatosc.NewSender())
}

func TestMicMuteSyncUseCase_GetStatus_offByDefault(t *testing.T) {
	uc := newTestMicMuteSyncUC(mapSettingsRepo{})
	uc.goos = "windows"
	ctx := context.Background()
	st, err := uc.GetStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if st.Enabled {
		t.Fatal("expected disabled by default")
	}
	if st.SyncEngineState != "off" {
		t.Fatalf("state: %s", st.SyncEngineState)
	}
}

func TestMicMuteSyncUseCase_SaveSettings_enabled(t *testing.T) {
	repo := mapSettingsRepo{}
	uc := newTestMicMuteSyncUC(repo)
	ctx := context.Background()
	if err := uc.SaveSettings(ctx, MicMuteSyncSettings{Enabled: true, OSCEndpoint: ""}); err != nil {
		t.Fatal(err)
	}
	cfg, err := uc.GetSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled {
		t.Fatal("expected enabled")
	}
}

func TestMicMuteSyncUseCase_GetStatus_linuxUnavailable(t *testing.T) {
	uc := newTestMicMuteSyncUC(mapSettingsRepo{})
	uc.goos = "linux"
	st, err := uc.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.Available {
		t.Fatal("linux should be unavailable")
	}
	if st.SyncEngineState != "unavailable" {
		t.Fatalf("state: %s", st.SyncEngineState)
	}
}

func TestMicMuteSyncUseCase_EnsureOSCLaunchArgs(t *testing.T) {
	repo := mapSettingsRepo{keyMicMuteSyncEnabled: "true"}
	uc := newTestMicMuteSyncUC(repo)
	uc.goos = "windows"
	got, err := uc.EnsureOSCLaunchArgs(context.Background(), "-no-vr")
	if err != nil {
		t.Fatal(err)
	}
	if got == "-no-vr" {
		t.Fatal("expected osc injection")
	}
}

func TestMicMuteSyncUseCase_onVRChatMuteChanged_respectsEchoGuard(t *testing.T) {
	uc := newTestMicMuteSyncUC(mapSettingsRepo{keyMicMuteSyncEnabled: "true"})
	uc.goos = "windows"
	uc.echo.Suppress(micmutesync.SourceDiscord, echoSuppressionDuration)
	// should return early without panic
	uc.onVRChatMuteChanged(context.Background(), true)
}
