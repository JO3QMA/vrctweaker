package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/infrastructure/vrchatapi"
)

func TestNormalizePresenceStatus(t *testing.T) {
	tests := map[string]string{
		"join me": "join me",
		"JOIN ME": "join me",
		"busy":    "busy",
		"offline": "active",
		"":        "active",
	}
	for in, want := range tests {
		if got := NormalizePresenceStatus(in); got != want {
			t.Errorf("NormalizePresenceStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPrependUniqueHistory(t *testing.T) {
	got := prependUniqueHistory([]string{"b", "a"}, "c")
	if len(got) != 3 || got[0] != "c" || got[1] != "b" || got[2] != "a" {
		t.Fatalf("prepend new: %v", got)
	}
	got = prependUniqueHistory([]string{"b", "a"}, "a")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("dedupe move: %v", got)
	}
	long := make([]string, 25)
	for i := range long {
		long[i] = "x"
	}
	got = prependUniqueHistory(long, "new")
	if len(got) != maxPresenceDescriptionHistory || got[0] != "new" {
		t.Fatalf("truncate: len=%d head=%q", len(got), got[0])
	}
}

func TestPresenceChangeUseCase_GetSection_notLoggedIn(t *testing.T) {
	ctx := context.Background()
	uc := NewPresenceChangeUseCase(newPresenceTestIdentity(t, false).uc, newMockSettingsRepo())
	got, err := uc.GetSection(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.LoggedIn || got.Status != "" {
		t.Fatalf("got=%+v", got)
	}
}

func TestPresenceChangeUseCase_GetSection_loggedIn(t *testing.T) {
	ctx := context.Background()
	settings := newMockSettingsRepo()
	settings.m[PresenceDescriptionHistoryKey] = `["older","working"]`
	idUC := newPresenceTestIdentity(t, true)
	idUC.selfRow.Status = "busy"
	idUC.selfRow.StatusDescription = "working"
	idUC.userRepo.getSelfRow = idUC.selfRow
	uc := NewPresenceChangeUseCase(idUC.uc, settings)
	got, err := uc.GetSection(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !got.LoggedIn || got.Status != "busy" || got.StatusDescription != "working" {
		t.Fatalf("got=%+v", got)
	}
	if len(got.History) != 2 || got.History[0] != "older" {
		t.Fatalf("history=%v", got.History)
	}
}

func TestPresenceChangeUseCase_GetSection_selfError(t *testing.T) {
	ctx := context.Background()
	idUC := newPresenceTestIdentity(t, true)
	idUC.userRepo.getSelfErr = errors.New("db down")
	uc := NewPresenceChangeUseCase(idUC.uc, newMockSettingsRepo())
	_, err := uc.GetSection(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPresenceChangeUseCase_Apply_recordsHistoryAndEmits(t *testing.T) {
	ctx := context.Background()
	settings := newMockSettingsRepo()
	idUC := newPresenceTestIdentity(t, true)
	var emitted bool
	idUC.uc.SetSelfCacheChangedHook(func() { emitted = true })
	uc := NewPresenceChangeUseCase(idUC.uc, settings)
	res, err := uc.Apply(ctx, "busy", "  focus  ")
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "busy" || res.StatusDescription != "focus" {
		t.Fatalf("res=%+v api last=%q", res, idUC.api.lastBothDescription)
	}
	if idUC.api.setBothCalls != 1 || idUC.api.lastBothStatus != vrchatapi.UserStatus("busy") {
		t.Fatalf("api calls=%d status=%q", idUC.api.setBothCalls, idUC.api.lastBothStatus)
	}
	raw, _ := settings.Get(ctx, PresenceDescriptionHistoryKey)
	var hist []string
	if err := json.Unmarshal([]byte(raw), &hist); err != nil || len(hist) != 1 || hist[0] != "focus" {
		t.Fatalf("history raw=%q", raw)
	}
	if !emitted {
		t.Fatal("expected self cache changed hook")
	}
}

func TestPresenceChangeUseCase_Apply_emptyDescriptionNoHistory(t *testing.T) {
	ctx := context.Background()
	settings := newMockSettingsRepo()
	uc := NewPresenceChangeUseCase(newPresenceTestIdentity(t, true).uc, settings)
	_, err := uc.Apply(ctx, "active", "   ")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := settings.Get(ctx, PresenceDescriptionHistoryKey)
	if strings.TrimSpace(raw) != "" {
		t.Fatalf("history=%q", raw)
	}
}

func TestPresenceChangeUseCase_Apply_invalidStatus(t *testing.T) {
	ctx := context.Background()
	uc := NewPresenceChangeUseCase(newPresenceTestIdentity(t, true).uc, newMockSettingsRepo())
	_, err := uc.Apply(ctx, "offline", "hi")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPresenceChangeUseCase_Apply_descriptionTooLong(t *testing.T) {
	ctx := context.Background()
	uc := NewPresenceChangeUseCase(newPresenceTestIdentity(t, true).uc, newMockSettingsRepo())
	long := strings.Repeat("あ", maxVRChatStatusDescriptionRunes+1)
	_, err := uc.Apply(ctx, "busy", long)
	if err == nil {
		t.Fatal("expected error")
	}
}

type presenceTestIdentity struct {
	uc       *IdentityUseCase
	api      *mockAPIClient
	userRepo *mockUserCacheRepo
	selfRow  *identity.UserCache
}

func newPresenceTestIdentity(t *testing.T, loggedIn bool) *presenceTestIdentity {
	t.Helper()
	api := &mockAPIClient{}
	userRepo := &mockUserCacheRepo{}
	settingsRepo := newMockSettingsRepo()
	uc := NewIdentityUseCase(userRepo, api, vrchatapi.NewStubCredentialStore(), settingsRepo, nil)
	h := &presenceTestIdentity{uc: uc, api: api, userRepo: userRepo}
	if !loggedIn {
		return h
	}
	api.SetAuthToken("presence-test-token")
	fp := identity.AuthTokenFingerprint("presence-test-token")
	h.selfRow = &identity.UserCache{
		VRCUserID:          "usr_self",
		Status:             "active",
		StatusDescription:  "",
		UserKind:           identity.UserKindSelf,
		SessionFingerprint: fp,
		LastUpdated:        time.Now(),
	}
	userRepo.getSelfRow = h.selfRow
	api.getCurrentUser = &vrchatapi.CurrentUserProfile{
		ID:                "usr_self",
		Status:            "busy",
		StatusDescription: "focus",
	}
	return h
}
