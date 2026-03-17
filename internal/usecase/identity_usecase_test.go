package usecase

import (
	"context"
	"errors"
	"testing"

	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/infrastructure/vrchatapi"
)

// mockFriendRepo implements identity.FriendCacheRepository for tests.
type mockFriendRepo struct {
	list          []*identity.FriendCache
	getByID       map[string]*identity.FriendCache
	listFavorites []*identity.FriendCache
	saveErr       error
	saveBatchErr  error
}

func (m *mockFriendRepo) List(_ context.Context) ([]*identity.FriendCache, error) {
	return m.list, nil
}

func (m *mockFriendRepo) GetByVRCUserID(_ context.Context, id string) (*identity.FriendCache, error) {
	if m.getByID != nil {
		return m.getByID[id], nil
	}
	return nil, nil
}

func (m *mockFriendRepo) ListFavorites(_ context.Context) ([]*identity.FriendCache, error) {
	return m.listFavorites, nil
}

func (m *mockFriendRepo) Save(_ context.Context, f *identity.FriendCache) error {
	return m.saveErr
}

func (m *mockFriendRepo) SaveBatch(_ context.Context, _ []*identity.FriendCache) error {
	return m.saveBatchErr
}

func (m *mockFriendRepo) Delete(_ context.Context, _ string) error {
	return nil
}

func (m *mockFriendRepo) DeleteAll(_ context.Context) (int64, error) {
	n := int64(len(m.list))
	m.list = nil
	m.listFavorites = nil
	if m.getByID != nil {
		m.getByID = make(map[string]*identity.FriendCache)
	}
	return n, nil
}

// mockAPIClient implements vrchatapi.VRChatAPIClient for tests.
type mockAPIClient struct {
	loginToken   string
	loginErr     error
	token        string
	getFriends   []vrchatapi.Friend
	getFriendsErr error
	setStatusErr error
}

func (m *mockAPIClient) Login(_ context.Context, _, _, _ string) (string, error) {
	return m.loginToken, m.loginErr
}

func (m *mockAPIClient) SetAuthToken(token string) {
	m.token = token
}

func (m *mockAPIClient) GetFriends(_ context.Context) ([]vrchatapi.Friend, error) {
	return m.getFriends, m.getFriendsErr
}

func (m *mockAPIClient) SetUserStatus(_ context.Context, _ vrchatapi.UserStatus) error {
	return m.setStatusErr
}

func TestIdentityUseCase_IsLoggedIn(t *testing.T) {
	ctx := context.Background()
	credStore := vrchatapi.NewStubCredentialStore()
	apiClient := &mockAPIClient{}
	friendRepo := &mockFriendRepo{}
	uc := NewIdentityUseCase(friendRepo, apiClient, credStore)

	// Empty cred store -> not logged in
	ok, err := uc.IsLoggedIn(ctx)
	if err != nil {
		t.Fatalf("IsLoggedIn: %v", err)
	}
	if ok {
		t.Error("IsLoggedIn want false when cred store empty, got true")
	}

	// Set token -> logged in
	if err := credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "token"); err != nil {
		t.Fatalf("credStore.Set: %v", err)
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
	friendRepo := &mockFriendRepo{}
	uc := NewIdentityUseCase(friendRepo, apiClient, credStore)

	if err := credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, "token"); err != nil {
		t.Fatalf("credStore.Set: %v", err)
	}

	if err := uc.Logout(ctx); err != nil {
		t.Fatalf("Logout: %v", err)
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
	friendRepo := &mockFriendRepo{}

	t.Run("empty_credentials_returns_error", func(t *testing.T) {
		apiClient := &mockAPIClient{}
		uc := NewIdentityUseCase(friendRepo, apiClient, credStore)
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
		uc := NewIdentityUseCase(friendRepo, apiClient, credStore)
		err := uc.Login(ctx, "user", "pass", "")
		if err != vrchatapi.ErrInvalidCredentials {
			t.Errorf("Login: want ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("cred_store_set_error_propagates", func(t *testing.T) {
		apiClient := &mockAPIClient{loginToken: "new-token"}
		failingCred := &failingCredStore{setErr: errors.New("keyring unavailable")}
		uc := NewIdentityUseCase(friendRepo, apiClient, failingCred)
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
		uc := NewIdentityUseCase(friendRepo, apiClient, credStore)
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
