package usecase

import (
	"context"
	"testing"
)

func TestSettingsUseCase_FirstLaunchGuideCompleted_roundtrip(t *testing.T) {
	repo := &fakeAppSettingsRepo{m: make(map[string]string)}
	uc := NewSettingsUseCase(repo)
	ctx := context.Background()

	done, err := uc.GetFirstLaunchGuideCompleted(ctx)
	if err != nil {
		t.Fatalf("GetFirstLaunchGuideCompleted: %v", err)
	}
	if done {
		t.Fatal("want false by default")
	}
	if err := uc.SetFirstLaunchGuideCompleted(ctx, true); err != nil {
		t.Fatalf("SetFirstLaunchGuideCompleted: %v", err)
	}
	done, err = uc.GetFirstLaunchGuideCompleted(ctx)
	if err != nil || !done {
		t.Fatalf("want true, got %v err=%v", done, err)
	}
}
