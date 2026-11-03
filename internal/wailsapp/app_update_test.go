package wailsapp

import (
	"context"
	"errors"
	"testing"

	"vrchat-tweaker/internal/usecase"
)

func TestApp_appUpdateWhenNotInitialized(t *testing.T) {
	a := NewApp()
	_, err := a.CheckForAppUpdate()
	if !errors.Is(err, errAppNotInitialized) {
		t.Fatalf("CheckForAppUpdate err = %v", err)
	}
	if err := a.SetAppUpdateAutoCheck(true); !errors.Is(err, errAppNotInitialized) {
		t.Fatalf("SetAppUpdateAutoCheck err = %v", err)
	}
}

func TestApp_appUpdateBindingsWithUseCase(t *testing.T) {
	a := NewApp()
	a.setInitReadyForTest(context.Background())
	a.appUpdate = usecase.NewAppUpdateUseCase(nil, nil, nil)
	if _, err := a.GetAppUpdateAutoCheck(); err != nil {
		t.Fatal(err)
	}
}

func TestApp_claimStartupAppUpdatePrompt(t *testing.T) {
	a := NewApp()
	a.setInitReadyForTest(context.Background())
	a.storePendingStartupAppUpdate(usecase.AppUpdateStatus{
		ShouldPrompt:  true,
		LatestVersion: "9.0.0",
	})
	st, err := a.ClaimStartupAppUpdatePrompt()
	if err != nil {
		t.Fatal(err)
	}
	if !st.ShouldPrompt || st.LatestVersion != "9.0.0" {
		t.Fatalf("status %+v", st)
	}
	st2, err := a.ClaimStartupAppUpdatePrompt()
	if err != nil {
		t.Fatal(err)
	}
	if st2.ShouldPrompt {
		t.Fatalf("expected empty second claim, got %+v", st2)
	}
}

func TestApp_clearStartupAppUpdatePrompt(t *testing.T) {
	a := NewApp()
	a.setInitReadyForTest(context.Background())
	a.storePendingStartupAppUpdate(usecase.AppUpdateStatus{ShouldPrompt: true})
	a.ClearStartupAppUpdatePrompt()
	st, err := a.ClaimStartupAppUpdatePrompt()
	if err != nil {
		t.Fatal(err)
	}
	if st.ShouldPrompt {
		t.Fatal("expected cleared pending")
	}
}
