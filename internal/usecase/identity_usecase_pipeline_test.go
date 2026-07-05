package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/infrastructure/vrchatapi"
)

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendOnline_envelopeLocationWinsOverUserSnapshot(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	settingsRepo := newMockSettingsRepo()
	apiClient := &mockAPIClient{token: "tok"}
	repo := &mockUserCacheRepo{
		getByID: map[string]*identity.UserCache{
			"usr_f1": {
				VRCUserID: "usr_f1", UserKind: identity.UserKindFriend,
				Status: "offline", Location: "",
			},
		},
	}
	uc := NewIdentityUseCase(repo, apiClient, vrchatapi.NewStubCredentialStore(), settingsRepo, nil)
	payload, err := json.Marshal(map[string]any{
		"userId":   "usr_f1",
		"platform": "standalonewindows",
		"location": "wrld_envelope:1~inst",
		"user": map[string]any{
			"id":       "usr_f1",
			"location": "",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := uc.HandleVRChatPipelineEvent(ctx, "friend-online", payload); err != nil {
		t.Fatal(err)
	}
	saved := repo.getByID["usr_f1"]
	if saved.Location != "wrld_envelope:1~inst" {
		t.Fatalf("envelope location must win over empty user snapshot, got %q", saved.Location)
	}
	if saved.Platform != "standalonewindows" {
		t.Fatalf("platform: got %q", saved.Platform)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendActive_preservesLocation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	settingsRepo := newMockSettingsRepo()
	apiClient := &mockAPIClient{token: "tok"}
	repo := &mockUserCacheRepo{
		getByID: map[string]*identity.UserCache{
			"usr_f1": {
				VRCUserID: "usr_f1", UserKind: identity.UserKindFriend,
				Status: "offline", Location: "wrld_keep:1~abc",
			},
		},
	}
	uc := NewIdentityUseCase(repo, apiClient, vrchatapi.NewStubCredentialStore(), settingsRepo, nil)
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
	if saved.Status != "active" {
		t.Fatalf("friend-active should clear offline status, got %q", saved.Status)
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
	uc := NewIdentityUseCase(repo, apiClient, vrchatapi.NewStubCredentialStore(), settingsRepo, nil)
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
	uc := NewIdentityUseCase(repo, apiClient, vrchatapi.NewStubCredentialStore(), settingsRepo, nil)
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
	uc := NewIdentityUseCase(repo, apiClient, vrchatapi.NewStubCredentialStore(), settingsRepo, nil)
	if err := uc.PipelineReconnectRestSync(ctx); err != nil {
		t.Fatalf("transient error should not be returned: %v", err)
	}
}

func TestIdentityUseCase_CurrentAuthToken(t *testing.T) {
	t.Parallel()
	apiClient := &mockAPIClient{token: "secret-token"}
	uc := NewIdentityUseCase(&mockUserCacheRepo{}, apiClient, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	if got := uc.CurrentAuthToken(); got != "secret-token" {
		t.Fatalf("token = %q", got)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_noTokenNoOp(t *testing.T) {
	t.Parallel()
	uc := NewIdentityUseCase(&mockUserCacheRepo{}, &mockAPIClient{}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	if err := uc.HandleVRChatPipelineEvent(context.Background(), "friend-online", []byte(`{"userId":"u"}`)); err != nil {
		t.Fatal(err)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendLocation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := &mockUserCacheRepo{
		getByID: map[string]*identity.UserCache{
			"usr_f1": {VRCUserID: "usr_f1", UserKind: identity.UserKindFriend, Location: "old"},
		},
	}
	uc := NewIdentityUseCase(repo, &mockAPIClient{token: "tok"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	payload := []byte(`{"userId":"usr_f1","location":"wrld_new:1~x","travelingToLocation":"wrld_dest","worldId":"wrld_new"}`)
	if err := uc.HandleVRChatPipelineEvent(ctx, "friend-location", payload); err != nil {
		t.Fatal(err)
	}
	if repo.getByID["usr_f1"].Location != "wrld_new:1~x" {
		t.Fatalf("location = %q", repo.getByID["usr_f1"].Location)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendAdd(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := &mockUserCacheRepo{getByID: make(map[string]*identity.UserCache)}
	uc := NewIdentityUseCase(repo, &mockAPIClient{token: "tok"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	payload := []byte(`{"userId":"usr_new","user":{"id":"usr_new","displayName":"New Friend","status":"active"}}`)
	if err := uc.HandleVRChatPipelineEvent(ctx, "friend-add", payload); err != nil {
		t.Fatal(err)
	}
	saved := repo.getByID["usr_new"]
	if saved == nil || saved.DisplayName != "New Friend" {
		t.Fatalf("saved = %+v", saved)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_userUpdate_selfOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fp := identity.AuthTokenFingerprint("tok-self")
	repo := &mockUserCacheRepo{
		getSelfRow: &identity.UserCache{
			VRCUserID:          "usr_self",
			UserKind:           identity.UserKindSelf,
			SessionFingerprint: fp,
			DisplayName:        "Old",
		},
	}
	uc := NewIdentityUseCase(repo, &mockAPIClient{token: "tok-self"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	payload := []byte(`{"userId":"usr_self","user":{"displayName":"New Name","status":"join me","statusDescription":"hi","username":"selfuser"}}`)
	if err := uc.HandleVRChatPipelineEvent(ctx, "user-update", payload); err != nil {
		t.Fatal(err)
	}
	if repo.lastUpsertSelf == nil || repo.lastUpsertSelf.DisplayName != "New Name" {
		t.Fatalf("upsert self = %+v", repo.lastUpsertSelf)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_userLocation_selfOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fp := identity.AuthTokenFingerprint("tok-loc")
	repo := &mockUserCacheRepo{
		getSelfRow: &identity.UserCache{
			VRCUserID:          "usr_self",
			UserKind:           identity.UserKindSelf,
			SessionFingerprint: fp,
		},
	}
	uc := NewIdentityUseCase(repo, &mockAPIClient{token: "tok-loc"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	payload := []byte(`{"userId":"usr_self","location":"wrld_here:1","travelingToLocation":"wrld_there"}`)
	if err := uc.HandleVRChatPipelineEvent(ctx, "user-location", payload); err != nil {
		t.Fatal(err)
	}
	if repo.lastUpsertSelf == nil || repo.lastUpsertSelf.Location != "wrld_here:1" {
		t.Fatalf("upsert self = %+v", repo.lastUpsertSelf)
	}
}

func TestIdentityUseCase_ReconcileSocialCacheFromAPI_successAndFriendErrorLogged(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	settingsRepo := newMockSettingsRepo()
	apiClient := &mockAPIClient{
		token: "tok",
		getCurrentUser: &vrchatapi.CurrentUserProfile{
			ID: "usr_self", DisplayName: "Self",
		},
		getFriendsErr: errors.New("friends API down"),
	}
	repo := &mockUserCacheRepo{}
	uc := NewIdentityUseCase(repo, apiClient, vrchatapi.NewStubCredentialStore(), settingsRepo, nil)
	if err := uc.ReconcileSocialCacheFromAPI(ctx); err != nil {
		t.Fatalf("self fetch ok should not fail: %v", err)
	}
	if repo.lastUpsertSelf == nil || repo.lastUpsertSelf.VRCUserID != "usr_self" {
		t.Fatalf("self upsert = %+v", repo.lastUpsertSelf)
	}
}

func TestIdentityUseCase_ReconcileSocialCacheFromAPIHandled_sessionExpired(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	_ = credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "blob")
	apiClient := &mockAPIClient{
		token:             "tok",
		getCurrentUserErr: vrchatapi.ErrSessionExpired,
	}
	uc := NewIdentityUseCase(&mockUserCacheRepo{}, apiClient, credStore, newMockSettingsRepo(), nil)
	err := uc.ReconcileSocialCacheFromAPIHandled(ctx)
	if !errors.Is(err, vrchatapi.ErrSessionExpired) {
		t.Fatalf("err = %v", err)
	}
	if apiClient.GetAuthToken() != "" {
		t.Fatal("token should be cleared on session expiry")
	}
}

func TestIdentityUseCase_PipelineReconnectRestSync_authErrorPropagates(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	apiClient := &mockAPIClient{
		token:             "tok",
		getCurrentUserErr: vrchatapi.ErrNotAuthenticated,
	}
	uc := NewIdentityUseCase(&mockUserCacheRepo{}, apiClient, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	if err := uc.PipelineReconnectRestSync(ctx); !errors.Is(err, vrchatapi.ErrNotAuthenticated) {
		t.Fatalf("err = %v", err)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_malformedPayloadNoError(t *testing.T) {
	t.Parallel()
	uc := NewIdentityUseCase(&mockUserCacheRepo{}, &mockAPIClient{token: "tok"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	for _, typ := range []string{"friend-delete", "friend-offline", "friend-active", "friend-location", "user-update"} {
		if err := uc.HandleVRChatPipelineEvent(context.Background(), typ, []byte(`{`)); err != nil {
			t.Fatalf("%s: %v", typ, err)
		}
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_userUpdate_ignoresOtherUser(t *testing.T) {
	t.Parallel()
	fp := identity.AuthTokenFingerprint("tok-self")
	repo := &mockUserCacheRepo{
		getSelfRow: &identity.UserCache{
			VRCUserID: "usr_self", UserKind: identity.UserKindSelf, SessionFingerprint: fp,
		},
	}
	uc := NewIdentityUseCase(repo, &mockAPIClient{token: "tok-self"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	payload := []byte(`{"userId":"someone_else","user":{"displayName":"X"}}`)
	if err := uc.HandleVRChatPipelineEvent(context.Background(), "user-update", payload); err != nil {
		t.Fatal(err)
	}
	if repo.lastUpsertSelf != nil {
		t.Fatal("should not upsert self for other user id")
	}
}

func TestIdentityUseCase_ReconcileSocialCacheFromAPI_notAuthenticatedSkips(t *testing.T) {
	t.Parallel()
	uc := NewIdentityUseCase(&mockUserCacheRepo{}, &mockAPIClient{}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	if err := uc.ReconcileSocialCacheFromAPI(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendUpdate_badUserJSON(t *testing.T) {
	t.Parallel()
	uc := NewIdentityUseCase(&mockUserCacheRepo{}, &mockAPIClient{token: "tok"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	err := uc.HandleVRChatPipelineEvent(context.Background(), "friend-update", []byte(`{"userId":"u1","user":"not-json-object"}`))
	if err != nil {
		t.Fatal(err)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendActive_createsUnresolvedContact(t *testing.T) {
	t.Parallel()
	repo := &mockUserCacheRepo{getByID: make(map[string]*identity.UserCache)}
	uc := NewIdentityUseCase(repo, &mockAPIClient{token: "tok"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	payload := []byte(`{"userId":"usr_new","platform":"web"}`)
	if err := uc.HandleVRChatPipelineEvent(context.Background(), "friend-active", payload); err != nil {
		t.Fatal(err)
	}
	saved := repo.getByID["usr_new"]
	if saved == nil {
		t.Fatal("expected new row")
		return
	}
	if saved.UserKind != identity.UserKindContact {
		t.Fatalf("user_kind = %q, want contact", saved.UserKind)
	}
	if saved.DisplayName != "" {
		t.Fatalf("display_name = %q, want empty", saved.DisplayName)
	}
	if saved.Status != "active" || saved.Platform != "web" {
		t.Fatalf("presence: status=%q platform=%q", saved.Status, saved.Platform)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendActive_resolvesListableFriend(t *testing.T) {
	t.Parallel()
	repo := &mockUserCacheRepo{getByID: make(map[string]*identity.UserCache)}
	api := &mockAPIClient{
		token: "tok",
		getUser: &vrchatapi.Friend{
			ID: "usr_new", DisplayName: "Resolved", Status: "active", IsFriend: true,
		},
	}
	uc := NewIdentityUseCase(repo, api, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	payload := []byte(`{"userId":"usr_new","platform":"web"}`)
	if err := uc.HandleVRChatPipelineEvent(context.Background(), "friend-active", payload); err != nil {
		t.Fatal(err)
	}
	saved := repo.getByID["usr_new"]
	if saved == nil {
		t.Fatal("expected row")
		return
	}
	if saved.UserKind != identity.UserKindFriend || saved.DisplayName != "Resolved" {
		t.Fatalf("saved = %+v", saved)
	}
	if api.getUserCalls != 1 {
		t.Fatalf("GetUser calls = %d, want 1", api.getUserCalls)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendActive_skipsSelf(t *testing.T) {
	t.Parallel()
	repo := &mockUserCacheRepo{
		getByID: map[string]*identity.UserCache{
			"usr_self": {VRCUserID: "usr_self", UserKind: identity.UserKindSelf},
		},
	}
	uc := NewIdentityUseCase(repo, &mockAPIClient{token: "tok"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	payload := []byte(`{"userId":"usr_self","platform":"web"}`)
	if err := uc.HandleVRChatPipelineEvent(context.Background(), "friend-active", payload); err != nil {
		t.Fatal(err)
	}
	if repo.getByID["usr_self"].Platform == "web" {
		t.Fatal("self row should not be updated via friend-active merge")
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendUpdate_emptyUserBody(t *testing.T) {
	t.Parallel()
	uc := NewIdentityUseCase(&mockUserCacheRepo{}, &mockAPIClient{token: "tok"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	if err := uc.HandleVRChatPipelineEvent(context.Background(), "friend-update", []byte(`{"userId":"u1","user":""}`)); err != nil {
		t.Fatal(err)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_userLocation_ignoresOtherUser(t *testing.T) {
	t.Parallel()
	fp := identity.AuthTokenFingerprint("tok-loc2")
	repo := &mockUserCacheRepo{
		getSelfRow: &identity.UserCache{
			VRCUserID: "usr_self", UserKind: identity.UserKindSelf, SessionFingerprint: fp,
		},
	}
	uc := NewIdentityUseCase(repo, &mockAPIClient{token: "tok-loc2"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	payload := []byte(`{"userId":"other","location":"wrld_x"}`)
	if err := uc.HandleVRChatPipelineEvent(context.Background(), "user-location", payload); err != nil {
		t.Fatal(err)
	}
	if repo.lastUpsertSelf != nil {
		t.Fatal("should ignore other user")
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendOnline_invalidEmbeddedUser(t *testing.T) {
	t.Parallel()
	repo := &mockUserCacheRepo{getByID: make(map[string]*identity.UserCache)}
	uc := NewIdentityUseCase(repo, &mockAPIClient{token: "tok"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	payload := []byte(`{"userId":"usr_f3","location":"wrld:1","user":"not-json"}`)
	if err := uc.HandleVRChatPipelineEvent(context.Background(), "friend-online", payload); err != nil {
		t.Fatal(err)
	}
	if repo.getByID["usr_f3"] == nil || repo.getByID["usr_f3"].Location != "wrld:1" {
		t.Fatalf("saved = %+v", repo.getByID["usr_f3"])
	}
	if repo.getByID["usr_f3"].UserKind != identity.UserKindContact {
		t.Fatalf("user_kind = %q, want contact without profile", repo.getByID["usr_f3"].UserKind)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_unknownTypeNoOp(t *testing.T) {
	t.Parallel()
	uc := NewIdentityUseCase(&mockUserCacheRepo{}, &mockAPIClient{token: "tok"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	if err := uc.HandleVRChatPipelineEvent(context.Background(), "ping", []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendUpdate_userIDFromEnvelope(t *testing.T) {
	t.Parallel()
	repo := &mockUserCacheRepo{getByID: make(map[string]*identity.UserCache)}
	uc := NewIdentityUseCase(repo, &mockAPIClient{token: "tok"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	payload := []byte(`{"userId":"usr_env","user":{"displayName":"From Env","status":"active"}}`)
	if err := uc.HandleVRChatPipelineEvent(context.Background(), "friend-update", payload); err != nil {
		t.Fatal(err)
	}
	if repo.getByID["usr_env"] == nil {
		t.Fatal("expected row keyed by envelope userId")
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_saveErrorPropagates(t *testing.T) {
	t.Parallel()
	repo := &mockUserCacheRepo{getByID: make(map[string]*identity.UserCache), saveErr: errors.New("save failed")}
	uc := NewIdentityUseCase(repo, &mockAPIClient{token: "tok"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	payload := []byte(`{"userId":"usr_fail","platform":"web"}`)
	err := uc.HandleVRChatPipelineEvent(context.Background(), "friend-active", payload)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_friendDelete_missingUserNoError(t *testing.T) {
	t.Parallel()
	uc := NewIdentityUseCase(&mockUserCacheRepo{}, &mockAPIClient{token: "tok"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	if err := uc.HandleVRChatPipelineEvent(context.Background(), "friend-delete", []byte(`{"userId":"missing"}`)); err != nil {
		t.Fatal(err)
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_userUpdate_selfLookupError(t *testing.T) {
	t.Parallel()
	repo := &mockUserCacheRepo{getSelfErr: errors.New("self lookup failed")}
	uc := NewIdentityUseCase(repo, &mockAPIClient{token: "tok"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	err := uc.HandleVRChatPipelineEvent(context.Background(), "user-update", []byte(`{"userId":"usr_self","user":{"displayName":"X"}}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIdentityUseCase_HandleVRChatPipelineEvent_userLocation_upsertError(t *testing.T) {
	t.Parallel()
	fp := identity.AuthTokenFingerprint("tok-up")
	repo := &mockUserCacheRepo{
		getSelfRow:    &identity.UserCache{VRCUserID: "usr_self", UserKind: identity.UserKindSelf, SessionFingerprint: fp},
		upsertSelfErr: errors.New("upsert failed"),
	}
	uc := NewIdentityUseCase(repo, &mockAPIClient{token: "tok-up"}, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	err := uc.HandleVRChatPipelineEvent(context.Background(), "user-location", []byte(`{"userId":"usr_self","location":"wrld:1"}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIdentityUseCase_backfillUnresolvedFriendPresence_promotesListableFriend(t *testing.T) {
	t.Parallel()
	pending := &identity.UserCache{
		VRCUserID: "usr_u", UserKind: identity.UserKindContact,
		Status: "active", Platform: "web",
	}
	repo := &mockUserCacheRepo{
		contactsNeedingProfileResolution: []*identity.UserCache{pending},
		getByID:                          map[string]*identity.UserCache{"usr_u": pending},
	}
	api := &mockAPIClient{
		token: "tok",
		getUser: &vrchatapi.Friend{
			ID: "usr_u", DisplayName: "Backfilled", Status: "active", IsFriend: true,
		},
	}
	uc := NewIdentityUseCase(repo, api, vrchatapi.NewStubCredentialStore(), newMockSettingsRepo(), nil)
	if err := uc.backfillUnresolvedFriendPresence(context.Background()); err != nil {
		t.Fatal(err)
	}
	saved := repo.getByID["usr_u"]
	if saved == nil || saved.UserKind != identity.UserKindFriend || saved.DisplayName != "Backfilled" {
		t.Fatalf("saved = %+v", saved)
	}
}
