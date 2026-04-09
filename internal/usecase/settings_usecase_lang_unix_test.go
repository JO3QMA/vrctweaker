//go:build !windows

package usecase

import (
	"context"
	"testing"
)

// LANG / LC_* は Unix 系でのみ初回ロケール解決に使われる。Windows は GetUserDefaultLocaleName のみで
// 環境変数を参照しないため、このケースは非 Windows でのみ検証する。
func TestSettingsUseCase_GetUILanguage_firstRun_LANG(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "ja_JP.UTF-8")
	repo := &fakeAppSettingsRepo{m: make(map[string]string)}
	uc := NewSettingsUseCase(repo)
	ctx := context.Background()

	got, err := uc.GetUILanguage(ctx)
	if err != nil {
		t.Fatalf("GetUILanguage: %v", err)
	}
	if got != "ja" {
		t.Fatalf("got %q, want ja", got)
	}
	if repo.m[keyUILocale] != "ja" {
		t.Fatalf("persisted %q, want ja", repo.m[keyUILocale])
	}
}
