package wailsapp

import (
	"context"
	"testing"

	"vrchat-tweaker/internal/infrastructure/tray"
	"vrchat-tweaker/internal/usecase"
)

type fakeTrayManager struct{}

func (fakeTrayManager) Supported() bool { return true }

func (fakeTrayManager) Running() bool { return false }

func (fakeTrayManager) Start(tray.Config) error { return nil }

func (fakeTrayManager) Stop() error { return nil }

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

func TestApp_closeToTrayEffective_defaultsWithoutTray(t *testing.T) {
	a := &App{}
	if a.closeToTrayEffective(t.Context()) {
		t.Fatal("want false without settings/tray")
	}
}

func TestApp_handleBeforeClose_withoutTray(t *testing.T) {
	a := &App{}
	if a.beforeClose(t.Context()) {
		t.Fatal("want false without close-to-tray")
	}
}

func TestApp_handleBeforeClose_explicitQuitBypassesCloseToTray(t *testing.T) {
	repo := &fakeAppSettingsRepo{m: make(map[string]string)}
	a := &App{
		settings: usecase.NewSettingsUseCase(repo),
		tray:     fakeTrayManager{},
	}

	a.quitPending.Store(true)
	if a.beforeClose(t.Context()) {
		t.Fatal("want false when explicit quit is requested")
	}
}

func TestApp_closeToTrayEffective_withSupportedTrayDefaultsOn(t *testing.T) {
	repo := &fakeAppSettingsRepo{m: make(map[string]string)}
	a := &App{
		settings: usecase.NewSettingsUseCase(repo),
		tray:     fakeTrayManager{},
	}
	if !a.closeToTrayEffective(t.Context()) {
		t.Fatal("want true when tray is supported and close-to-tray defaults on")
	}
}
