package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/infrastructure/vrchatapi"
)

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendActive_preservesLocation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	settingsRepo := newMockSettingsRepo()
	apiClient := &mockAPIClient{token: "tok"}
	repo := &mockUserCacheRepo{
		getByID: map[string]*identity.UserCache{
			"usr_f1": {
				VRCUserID: "usr_f1", UserKind: identity.UserKindFriend,
				Status: "active", Location: "wrld_keep:1~abc",
			},
		},
	}
	uc := NewIdentityUseCase(repo, apiClient, vrchatapi.NewStubCredentialStore(), settingsRepo)
	payload, err := json.Marshal(map[string]string{
		"userId":   "usr_f1",
		"platform": "web",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := uc.HandleVRChatPipelineEvent(ctx, "friend-active", payload); err != nil {
		t.Fatal(err)
	}
	saved := repo.getByID["usr_f1"]
	if saved.Location != "wrld_keep:1~abc" {
		t.Fatalf("friend-active must not clobber location, got %q", saved.Location)
	}
	if saved.Platform != "web" {
		t.Fatalf("platform: got %q", saved.Platform)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendOffline(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	settingsRepo := newMockSettingsRepo()
	apiClient := &mockAPIClient{token: "tok"}
	repo := &mockUserCacheRepo{
		getByID: map[string]*identity.UserCache{
			"usr_f1": {
				VRCUserID: "usr_f1", UserKind: identity.UserKindFriend,
				Status: "active", Location: "wrld_x:1",
			},
		},
	}
	uc := NewIdentityUseCase(repo, apiClient, vrchatapi.NewStubCredentialStore(), settingsRepo)
	payload, err := json.Marshal(map[string]string{"userId": "usr_f1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := uc.HandleVRChatPipelineEvent(ctx, "friend-offline", payload); err != nil {
		t.Fatal(err)
	}
	saved := repo.getByID["usr_f1"]
	if saved.Status != "offline" || saved.Location != "" {
		t.Fatalf("want offline+clear location, got %+v", saved)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendDelete_demotes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	settingsRepo := newMockSettingsRepo()
	apiClient := &mockAPIClient{token: "tok"}
	repo := &mockUserCacheRepo{
		getByID: map[string]*identity.UserCache{
			"usr_f2": {VRCUserID: "usr_f2", UserKind: identity.UserKindFriend, IsFavorite: true},
		},
	}
	uc := NewIdentityUseCase(repo, apiClient, vrchatapi.NewStubCredentialStore(), settingsRepo)
	payload, err := json.Marshal(map[string]string{"userId": "usr_f2"})
	if err != nil {
		t.Fatal(err)
	}
	if err := uc.HandleVRChatPipelineEvent(ctx, "friend-delete", payload); err != nil {
		t.Fatal(err)
	}
	saved := repo.getByID["usr_f2"]
	if saved.UserKind != identity.UserKindContact || saved.IsFavorite {
		t.Fatalf("want demoted contact, got %+v", saved)
	}
}

func TestIdentityUseCase_PipelineReconnectRestSync_transientErrorReturnsNil(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	settingsRepo := newMockSettingsRepo()
	apiClient := &mockAPIClient{
		token:             "tok",
		getCurrentUserErr: errors.New("temporary API failure"),
	}
	repo := &mockUserCacheRepo{}
	uc := NewIdentityUseCase(repo, apiClient, vrchatapi.NewStubCredentialStore(), settingsRepo)
	if err := uc.PipelineReconnectRestSync(ctx); err != nil {
		t.Fatalf("transient error should not be returned: %v", err)
	}
}
