package wailsapp

import (
	"testing"

	"vrchat-tweaker/internal/usecase"
)

func TestApp_appUpdateWhenNotInitialized(t *testing.T) {
	a := NewApp()
	_, err := a.CheckForAppUpdate()
	if err == nil {
		t.Fatal("expected error")
	}
	if err := a.SetAppUpdateAutoCheck(true); err == nil {
		t.Fatal("expected error")
	}
}

func TestApp_appUpdateBindingsWithUseCase(t *testing.T) {
	a := NewApp()
	a.appUpdate = usecase.NewAppUpdateUseCase(nil, nil, nil)
	if _, err := a.GetAppUpdateAutoCheck(); err != nil {
		t.Fatal(err)
	}
}
