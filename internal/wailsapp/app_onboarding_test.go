package wailsapp

import (
	"context"
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
