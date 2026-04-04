package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/infrastructure/vrchatapi"
)

// mockSettingsRepo implements settings.AppSettingsRepository for tests.
type mockSettingsRepo struct {
	m      map[string]string
	setErr error
}

func newMockSettingsRepo() *mockSettingsRepo {
	return &mockSettingsRepo{m: make(map[string]string)}
}

func (m *mockSettingsRepo) Get(_ context.Context, key string) (string, error) {
	return m.m[key], nil
}

func (m *mockSettingsRepo) Set(_ context.Context, key, value string) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.m[key] = value
	return nil
}

func (m *mockSettingsRepo) GetAll(context.Context) (map[string]string, error) {
	out := make(map[string]string, len(m.m))
	for k, v := range m.m {
		out[k] = v
	}
	return out, nil
}

// mockUserCacheRepo implements identity.UserCacheRepository for tests.
type mockUserCacheRepo struct {
	list            []*identity.UserCache
	getByID         map[string]*identity.UserCache
	listFavorites   []*identity.UserCache
	saveErr         error
	saveBatchErr    error
	lastSaveBatch   []*identity.UserCache
	getSelfRow      *identity.UserCache
	upsertSelfErr   error
	deleteSelfCount int
	deleteSelfErr   error
	lastUpsertSelf  *identity.UserCache
}

func (m *mockUserCacheRepo) List(_ context.Context) ([]*identity.UserCache, error) {
	return m.list, nil
}

func (m *mockUserCacheRepo) GetByVRCUserID(_ context.Context, id string) (*identity.UserCache, error) {
	if m.getByID != nil {
		return m.getByID[id], nil
	}
	return nil, nil
}

func (m *mockUserCacheRepo) ListFavorites(_ context.Context) ([]*identity.UserCache, error) {
	if m.listFavorites != nil {
		return m.listFavorites, nil
	}
	var out []*identity.UserCache
	for _, u := range m.list {
		if u.IsFavorite {
			out = append(out, u)
		}
	}
	return out, nil
}

func (m *mockUserCacheRepo) Save(_ context.Context, f *identity.UserCache) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	if m.getByID != nil && f != nil {
		cpy := *f
		m.getByID[f.VRCUserID] = &cpy
	}
	return nil
}

func (m *mockUserCacheRepo) SaveBatch(_ context.Context, users []*identity.UserCache) error {
	if m.saveBatchErr != nil {
		return m.saveBatchErr
	}
	m.lastSaveBatch = append([]*identity.UserCache(nil), users...)
	m.list = append([]*identity.UserCache(nil), users...)
	return nil
}

func (m *mockUserCacheRepo) Delete(_ context.Context, _ string) error {
	return nil
}

func (m *mockUserCacheRepo) DeleteAll(_ context.Context) (int64, error) {
	n := int64(len(m.list))
	m.list = nil
	m.listFavorites = nil
	if m.getByID != nil {
		m.getByID = make(map[string]*identity.UserCache)
	}
	return n, nil
}

func (m *mockUserCacheRepo) GetSelfBySessionFingerprint(_ context.Context, fp string) (*identity.UserCache, error) {
	if m.getSelfRow != nil && m.getSelfRow.SessionFingerprint == fp {
		return m.getSelfRow, nil
	}
	return nil, nil
}

func (m *mockUserCacheRepo) UpsertSelf(_ context.Context, u *identity.UserCache) error {
	m.lastUpsertSelf = u
	return m.upsertSelfErr
}

func (m *mockUserCacheRepo) DeleteSelfRows(context.Context) error {
	m.deleteSelfCount++
	return m.deleteSelfErr
}

// mockAPIClient implements vrchatapi.VRChatAPIClient for tests.
type mockAPIClient struct {
	loginToken          string
	loginErr            error
	token               string
	getCurrentUser      *vrchatapi.CurrentUserProfile
	getCurrentUserErr   error
	getCurrentUserCalls int
	getFriends          []vrchatapi.Friend
	getFriendsErr       error
	getFriendsCalls     int
	getUser             *vrchatapi.Friend
	getUserErr          error
	getUserCalls        int
	setStatusErr        error
}

func (m *mockAPIClient) Login(_ context.Context, _, _, _ string) (string, error) {
	return m.loginToken, m.loginErr
}

func (m *mockAPIClient) SetAuthToken(token string) {
	m.token = token
}

func (m *mockAPIClient) GetCurrentUser(_ context.Context) (*vrchatapi.CurrentUserProfile, error) {
	m.getCurrentUserCalls++
	if m.getCurrentUserErr != nil {
		return nil, m.getCurrentUserErr
	}
	return m.getCurrentUser, nil
}

func (m *mockAPIClient) GetFriends(_ context.Context) ([]vrchatapi.Friend, error) {
	m.getFriendsCalls++
	return m.getFriends, m.getFriendsErr
}

func (m *mockAPIClient) GetUser(_ context.Context, _ string) (*vrchatapi.Friend, error) {
	m.getUserCalls++
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}
	return m.getUser, nil
}

func (m *mockAPIClient) SetUserStatus(_ context.Context, _ vrchatapi.UserStatus) error {
	return m.setStatusErr
}

func TestIdentityUseCase_IsLoggedIn(t *testing.T) {
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	apiClient := &mockAPIClient{}
	userRepo := &mockUserCacheRepo{}
	settingsRepo := newMockSettingsRepo()
	uc := NewIdentityUseCase(userRepo, apiClient, credStore, settingsRepo)

	// Empty cred store -> not logged in
	ok, err := uc.IsLoggedIn(ctx)
	if err != nil {
		t.Fatalf("IsLoggedIn: %v", err)
	}
	if ok {
		t.Error("IsLoggedIn want false when cred store empty, got true")
	}

	// Set token -> logged in
	if setErr := credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "token"); setErr != nil {
		t.Fatalf("credStore.Set: %v", setErr)
	}
	ok, err = uc.IsLoggedIn(ctx)
	if err != nil {
		t.Fatalf("IsLoggedIn: %v", err)
	}
	if !ok {
		t.Error("IsLoggedIn want true when token stored, got false")
	}
}

func TestIdentityUseCase_Logout(t *testing.T) {
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	apiClient := &mockAPIClient{token: "old-token"}
	userRepo := &mockUserCacheRepo{}
	settingsRepo := newMockSettingsRepo()
	uc := NewIdentityUseCase(userRepo, apiClient, credStore, settingsRepo)

	if err := credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "token"); err != nil {
		t.Fatalf("credStore.Set: %v", err)
	}

	if err := uc.Logout(ctx); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	if userRepo.deleteSelfCount != 1 {
		t.Errorf("DeleteSelfRows want 1 call, got %d", userRepo.deleteSelfCount)
	}

	// Cred store should be empty
	_, err := credStore.Get(vrchatapi.CredentialService, vrchatapi.CredentialUser)
	if err == nil {
		t.Error("Logout: cred store should be empty after logout")
	}

	// API client token should be cleared
	if apiClient.token != "" {
		t.Errorf("Logout: apiClient token want empty, got %q", apiClient.token)
	}
}

func TestIdentityUseCase_Login(t *testing.T) {
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	userRepo := &mockUserCacheRepo{}
	settingsRepo := newMockSettingsRepo()

	t.Run("empty_credentials_returns_error", func(t *testing.T) {
		apiClient := &mockAPIClient{}
		uc := NewIdentityUseCase(userRepo, apiClient, credStore, settingsRepo)
		err := uc.Login(ctx, "", "password", "")
		if err != vrchatapi.ErrInvalidCredentials {
			t.Errorf("Login(empty user): want ErrInvalidCredentials, got %v", err)
		}
		err = uc.Login(ctx, "user", "", "")
		if err != vrchatapi.ErrInvalidCredentials {
			t.Errorf("Login(empty pass): want ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("api_login_error_propagates", func(t *testing.T) {
		apiClient := &mockAPIClient{loginErr: vrchatapi.ErrInvalidCredentials}
		uc := NewIdentityUseCase(userRepo, apiClient, credStore, settingsRepo)
		err := uc.Login(ctx, "user", "pass", "")
		if err != vrchatapi.ErrInvalidCredentials {
			t.Errorf("Login: want ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("cred_store_set_error_propagates", func(t *testing.T) {
		apiClient := &mockAPIClient{loginToken: "new-token"}
		failingCred := &failingCredStore{setErr: errors.New("keyring unavailable")}
		uc := NewIdentityUseCase(userRepo, apiClient, failingCred, settingsRepo)
		err := uc.Login(ctx, "user", "pass", "")
		if err == nil {
			t.Fatal("Login: want error when cred store fails, got nil")
		}
		if apiClient.token != "" {
			t.Error("Login: apiClient token should not be set when save fails")
		}
	})

	t.Run("success_saves_token_and_sets_client", func(t *testing.T) {
		credStore := vrchatapi.NewStubCredentialStore()
		apiClient := &mockAPIClient{loginToken: "auth-token-123"}
		uc := NewIdentityUseCase(userRepo, apiClient, credStore, settingsRepo)
		err := uc.Login(ctx, "user", "pass", "")
		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		token, err := credStore.Get(vrchatapi.CredentialService, vrchatapi.CredentialUser)
		if err != nil {
			t.Fatalf("credStore.Get: %v", err)
		}
		if token != "auth-token-123" {
			t.Errorf("credStore: want token auth-token-123, got %q", token)
		}
		if apiClient.token != "auth-token-123" {
			t.Errorf("apiClient token: want auth-token-123, got %q", apiClient.token)
		}
	})
}

func TestIdentityUseCase_GetCurrentUser(t *testing.T) {
	ctx := context.Background()
	userRepo := &mockUserCacheRepo{}
	settingsRepo := newMockSettingsRepo()

	t.Run("not_logged_in", func(t *testing.T) {
		credStore := vrchatapi.NewStubCredentialStore()
		apiClient := &mockAPIClient{}
		uc := NewIdentityUseCase(userRepo, apiClient, credStore, settingsRepo)
		_, err := uc.GetCurrentUser(ctx, false)
		if err != vrchatapi.ErrNotAuthenticated {
			t.Fatalf("err = %v, want ErrNotAuthenticated", err)
		}
	})

	t.Run("success_fetches_api_and_upserts_self", func(t *testing.T) {
		credStore := vrchatapi.NewStubCredentialStore()
		if err := credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "tok"); err != nil {
			t.Fatal(err)
		}
		prof := &vrchatapi.CurrentUserProfile{ID: "usr_x", DisplayName: "Test"}
		apiClient := &mockAPIClient{getCurrentUser: prof}
		repo := &mockUserCacheRepo{}
		uc := NewIdentityUseCase(repo, apiClient, credStore, settingsRepo)
		got, err := uc.GetCurrentUser(ctx, false)
		if err != nil {
			t.Fatalf("GetCurrentUser: %v", err)
		}
		if got != prof {
			t.Fatal("profile pointer mismatch")
		}
		if apiClient.token != "tok" {
			t.Errorf("apiClient token: want tok, got %q", apiClient.token)
		}
		if apiClient.getCurrentUserCalls != 1 {
			t.Errorf("GetCurrentUser API calls want 1, got %d", apiClient.getCurrentUserCalls)
		}
		if repo.lastUpsertSelf == nil || repo.lastUpsertSelf.VRCUserID != "usr_x" {
			t.Fatalf("UpsertSelf want usr_x, got %+v", repo.lastUpsertSelf)
		}
		if repo.lastUpsertSelf.UserKind != identity.UserKindSelf {
			t.Errorf("UpsertSelf kind want self, got %q", repo.lastUpsertSelf.UserKind)
		}
		wantFP := identity.AuthTokenFingerprint("tok")
		if repo.lastUpsertSelf.SessionFingerprint != wantFP {
			t.Errorf("session fingerprint mismatch")
		}
	})

	t.Run("cache_hit_skips_api", func(t *testing.T) {
		credStore := vrchatapi.NewStubCredentialStore()
		if err := credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "tok"); err != nil {
			t.Fatal(err)
		}
		fp := identity.AuthTokenFingerprint("tok")
		repo := &mockUserCacheRepo{
			getSelfRow: &identity.UserCache{
				VRCUserID:                   "usr_cached",
				DisplayName:                 "Cached",
				Status:                      "active",
				UserKind:                    identity.UserKindSelf,
				LastUpdated:                 time.Now(),
				SessionFingerprint:          fp,
				Username:                    "cacheduser",
				StatusDescription:           "hi",
				UserState:                   "offline",
				AvatarThumbnailURL:          "http://a",
				UserIconURL:                 "http://i",
				ProfilePicOverrideThumbnail: "http://p",
			},
		}
		apiClient := &mockAPIClient{getCurrentUser: &vrchatapi.CurrentUserProfile{ID: "wrong"}}
		uc := NewIdentityUseCase(repo, apiClient, credStore, settingsRepo)
		got, err := uc.GetCurrentUser(ctx, false)
		if err != nil {
			t.Fatalf("GetCurrentUser: %v", err)
		}
		if got.ID != "usr_cached" || got.DisplayName != "Cached" {
			t.Fatalf("got %+v", got)
		}
		if apiClient.getCurrentUserCalls != 0 {
			t.Errorf("API should not be called, got %d calls", apiClient.getCurrentUserCalls)
		}
	})

	t.Run("upsert_error_propagates", func(t *testing.T) {
		credStore := vrchatapi.NewStubCredentialStore()
		if err := credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "tok"); err != nil {
			t.Fatal(err)
		}
		repo := &mockUserCacheRepo{upsertSelfErr: errors.New("disk full")}
		apiClient := &mockAPIClient{getCurrentUser: &vrchatapi.CurrentUserProfile{ID: "u1"}}
		uc := NewIdentityUseCase(repo, apiClient, credStore, settingsRepo)
		_, err := uc.GetCurrentUser(ctx, false)
		if err == nil {
			t.Fatal("want error from UpsertSelf")
		}
	})

	t.Run("forceRefresh_bypasses_cache", func(t *testing.T) {
		credStore := vrchatapi.NewStubCredentialStore()
		if err := credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "tok"); err != nil {
			t.Fatal(err)
		}
		fp := identity.AuthTokenFingerprint("tok")
		fresh := &vrchatapi.CurrentUserProfile{ID: "usr_fresh", DisplayName: "Fresh"}
		repo := &mockUserCacheRepo{
			getSelfRow: &identity.UserCache{
				VRCUserID:          "usr_cached",
				DisplayName:        "Cached",
				UserKind:           identity.UserKindSelf,
				LastUpdated:        time.Now(),
				SessionFingerprint: fp,
			},
		}
		apiClient := &mockAPIClient{getCurrentUser: fresh}
		uc := NewIdentityUseCase(repo, apiClient, credStore, settingsRepo)
		got, err := uc.GetCurrentUser(ctx, true)
		if err != nil {
			t.Fatalf("GetCurrentUser: %v", err)
		}
		if got.ID != "usr_fresh" || got.DisplayName != "Fresh" {
			t.Fatalf("got %+v, want fresh profile from API", got)
		}
		if apiClient.getCurrentUserCalls != 1 {
			t.Errorf("API calls want 1, got %d", apiClient.getCurrentUserCalls)
		}
		if repo.lastUpsertSelf == nil || repo.lastUpsertSelf.VRCUserID != "usr_fresh" {
			t.Fatalf("UpsertSelf want usr_fresh, got %+v", repo.lastUpsertSelf)
		}
	})
}

func TestIdentityUseCase_ListFriends_refreshesWhenStale(t *testing.T) {
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	if err := credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "t"); err != nil {
		t.Fatal(err)
	}
	settingsRepo := newMockSettingsRepo()
	userRepo := &mockUserCacheRepo{}
	apiClient := &mockAPIClient{
		getFriends: []vrchatapi.Friend{{ID: "f1", DisplayName: "F1", Status: "active"}},
	}
	uc := NewIdentityUseCase(userRepo, apiClient, credStore, settingsRepo)

	list, err := uc.ListFriends(ctx)
	if err != nil {
		t.Fatalf("ListFriends: %v", err)
	}
	if len(list) != 1 || list[0].VRCUserID != "f1" {
		t.Fatalf("list = %+v", list)
	}
	if apiClient.getFriendsCalls != 1 {
		t.Errorf("GetFriends calls want 1, got %d", apiClient.getFriendsCalls)
	}
	if settingsRepo.m[identity.SettingVRChatFriendsSyncedAt] == "" {
		t.Error("expected friends sync timestamp to be set")
	}
}

func TestIdentityUseCase_ListFriends_skipsRefreshWhenFresh(t *testing.T) {
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	if err := credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "t"); err != nil {
		t.Fatal(err)
	}
	settingsRepo := newMockSettingsRepo()
	settingsRepo.m[identity.SettingVRChatFriendsSyncedAt] = time.Now().UTC().Format(time.RFC3339)
	userRepo := &mockUserCacheRepo{
		list: []*identity.UserCache{{VRCUserID: "x", DisplayName: "X", Status: "offline", UserKind: identity.UserKindFriend}},
	}
	apiClient := &mockAPIClient{
		getFriends: []vrchatapi.Friend{{ID: "new", DisplayName: "N", Status: "active"}},
	}
	uc := NewIdentityUseCase(userRepo, apiClient, credStore, settingsRepo)
	list, err := uc.ListFriends(ctx)
	if err != nil {
		t.Fatalf("ListFriends: %v", err)
	}
	if len(list) != 1 || list[0].VRCUserID != "x" {
		t.Fatalf("want cached list unchanged, got %+v", list)
	}
	if apiClient.getFriendsCalls != 0 {
		t.Errorf("GetFriends should not run when fresh, got %d calls", apiClient.getFriendsCalls)
	}
}

func TestIdentityUseCase_RefreshFriends_sets_sync_timestamp(t *testing.T) {
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	_ = credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "t")
	settingsRepo := newMockSettingsRepo()
	userRepo := &mockUserCacheRepo{}
	apiClient := &mockAPIClient{getFriends: []vrchatapi.Friend{}}
	uc := NewIdentityUseCase(userRepo, apiClient, credStore, settingsRepo)
	if err := uc.RefreshFriends(ctx); err != nil {
		t.Fatal(err)
	}
	if settingsRepo.m[identity.SettingVRChatFriendsSyncedAt] == "" {
		t.Error("sync timestamp not set")
	}
}

func TestIdentityUseCase_RefreshFriends_preservesSelfRow(t *testing.T) {
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	if err := credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "t"); err != nil {
		t.Fatal(err)
	}
	settingsRepo := newMockSettingsRepo()
	at := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	selfRow := &identity.UserCache{
		VRCUserID:          "usr_me",
		DisplayName:        "OriginalMe",
		Status:             "busy",
		UserKind:           identity.UserKindSelf,
		LastUpdated:        at,
		SessionFingerprint: "fp1",
		Username:           "meuser",
	}
	userRepo := &mockUserCacheRepo{
		getByID: map[string]*identity.UserCache{"usr_me": selfRow},
	}
	apiClient := &mockAPIClient{
		getFriends: []vrchatapi.Friend{{
			ID:          "usr_me",
			DisplayName: "FromFriendsAPI",
			Status:      "join me",
		}},
	}
	uc := NewIdentityUseCase(userRepo, apiClient, credStore, settingsRepo)
	if err := uc.RefreshFriends(ctx); err != nil {
		t.Fatal(err)
	}
	if len(userRepo.lastSaveBatch) != 1 {
		t.Fatalf("SaveBatch want 1 user, got %d", len(userRepo.lastSaveBatch))
	}
	saved := userRepo.lastSaveBatch[0]
	if saved.UserKind != identity.UserKindSelf || saved.DisplayName != "OriginalMe" || saved.Status != "busy" {
		t.Fatalf("self row overwritten by friends sync: %+v", saved)
	}
}

func TestIdentityUseCase_ResolveUserProfileForNavigation_notLoggedIn_cacheHit(t *testing.T) {
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	row := &identity.UserCache{VRCUserID: "u1", DisplayName: "A", UserKind: identity.UserKindContact}
	userRepo := &mockUserCacheRepo{getByID: map[string]*identity.UserCache{"u1": row}}
	apiClient := &mockAPIClient{}
	settingsRepo := newMockSettingsRepo()
	uc := NewIdentityUseCase(userRepo, apiClient, credStore, settingsRepo)
	u, openF, err := uc.ResolveUserProfileForNavigation(ctx, "u1")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if openF {
		t.Error("contact should not use friends view")
	}
	if u.DisplayName != "A" {
		t.Errorf("displayName %q", u.DisplayName)
	}
	if apiClient.getUserCalls != 0 {
		t.Error("GetUser should not run when not logged in")
	}
}

func TestIdentityUseCase_ResolveUserProfileForNavigation_notLoggedIn_friendOpensFriends(t *testing.T) {
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	row := &identity.UserCache{VRCUserID: "u1", DisplayName: "F", UserKind: identity.UserKindFriend}
	userRepo := &mockUserCacheRepo{getByID: map[string]*identity.UserCache{"u1": row}}
	uc := NewIdentityUseCase(userRepo, &mockAPIClient{}, credStore, newMockSettingsRepo())
	_, openF, err := uc.ResolveUserProfileForNavigation(ctx, "u1")
	if err != nil {
		t.Fatal(err)
	}
	if !openF {
		t.Error("want openInFriendsView for cached friend")
	}
}

func TestIdentityUseCase_ResolveUserProfileForNavigation_notLoggedIn_miss(t *testing.T) {
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	userRepo := &mockUserCacheRepo{getByID: map[string]*identity.UserCache{}}
	uc := NewIdentityUseCase(userRepo, &mockAPIClient{}, credStore, newMockSettingsRepo())
	_, _, err := uc.ResolveUserProfileForNavigation(ctx, "missing")
	if !errors.Is(err, ErrProfileNotInCache) {
		t.Fatalf("want ErrProfileNotInCache, got %v", err)
	}
}

func TestIdentityUseCase_ResolveUserProfileForNavigation_loggedIn_newContactFromAPI(t *testing.T) {
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	if err := credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "tok"); err != nil {
		t.Fatal(err)
	}
	userRepo := &mockUserCacheRepo{getByID: map[string]*identity.UserCache{}}
	apiClient := &mockAPIClient{
		getUser: &vrchatapi.Friend{
			ID:          "usr_new",
			DisplayName: "FromAPI",
			IsFriend:    false,
			Status:      "active",
		},
	}
	uc := NewIdentityUseCase(userRepo, apiClient, credStore, newMockSettingsRepo())
	u, openF, err := uc.ResolveUserProfileForNavigation(ctx, "usr_new")
	if err != nil {
		t.Fatal(err)
	}
	if openF {
		t.Error("non-friend API should not open friends view")
	}
	if u.UserKind != identity.UserKindContact || u.DisplayName != "FromAPI" {
		t.Fatalf("row: %+v", u)
	}
	if apiClient.getUserCalls != 1 {
		t.Errorf("GetUser calls %d", apiClient.getUserCalls)
	}
}

func TestIdentityUseCase_ResolveUserProfileForNavigation_loggedIn_friendFromAPI(t *testing.T) {
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	if err := credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "tok"); err != nil {
		t.Fatal(err)
	}
	userRepo := &mockUserCacheRepo{getByID: map[string]*identity.UserCache{}}
	apiClient := &mockAPIClient{
		getUser: &vrchatapi.Friend{
			ID:          "usr_f",
			DisplayName: "Pal",
			IsFriend:    true,
			Status:      "active",
		},
	}
	uc := NewIdentityUseCase(userRepo, apiClient, credStore, newMockSettingsRepo())
	u, openF, err := uc.ResolveUserProfileForNavigation(ctx, "usr_f")
	if err != nil {
		t.Fatal(err)
	}
	if !openF || u.UserKind != identity.UserKindFriend {
		t.Fatalf("openF=%v kind=%v", openF, u.UserKind)
	}
}

func TestIdentityUseCase_ResolveUserProfileForNavigation_loggedIn_apiErr_fallsBackToCache(t *testing.T) {
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	if err := credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "tok"); err != nil {
		t.Fatal(err)
	}
	row := &identity.UserCache{VRCUserID: "u1", DisplayName: "Cached", UserKind: identity.UserKindContact}
	userRepo := &mockUserCacheRepo{getByID: map[string]*identity.UserCache{"u1": row}}
	apiClient := &mockAPIClient{getUserErr: errors.New("network")}
	uc := NewIdentityUseCase(userRepo, apiClient, credStore, newMockSettingsRepo())
	u, openF, err := uc.ResolveUserProfileForNavigation(ctx, "u1")
	if err != nil {
		t.Fatal(err)
	}
	if openF || u.DisplayName != "Cached" {
		t.Fatalf("openF=%v u=%+v", openF, u)
	}
}

// failingCredStore implements CredentialStore with configurable errors.
type failingCredStore struct {
	getErr    error
	setErr    error
	deleteErr error
}

func (f *failingCredStore) Get(_, _ string) (string, error) {
	return "", f.getErr
}

func (f *failingCredStore) Set(_, _, _ string) error {
	return f.setErr
}

func (f *failingCredStore) Delete(_, _ string) error {
	return f.deleteErr
}
