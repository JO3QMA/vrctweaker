package usecase

import (
	"context"
	"fmt"
	"time"

	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/infrastructure/vrchatapi"
)

// IdentityUseCase handles VRChat auth, friends, and status.
type IdentityUseCase struct {
	userCacheRepo identity.UserCacheRepository
	apiClient     vrchatapi.VRChatAPIClient
	credStore     vrchatapi.CredentialStore
	notifier      identity.Notifier // optional; nil skips online notifications
}

// NewIdentityUseCase creates a new IdentityUseCase.
func NewIdentityUseCase(
	userCacheRepo identity.UserCacheRepository,
	apiClient vrchatapi.VRChatAPIClient,
	credStore vrchatapi.CredentialStore,
) *IdentityUseCase {
	return NewIdentityUseCaseWithNotifier(userCacheRepo, apiClient, credStore, nil)
}

// NewIdentityUseCaseWithNotifier creates a new IdentityUseCase with optional Notifier for favorite-online notifications.
func NewIdentityUseCaseWithNotifier(
	userCacheRepo identity.UserCacheRepository,
	apiClient vrchatapi.VRChatAPIClient,
	credStore vrchatapi.CredentialStore,
	notifier identity.Notifier,
) *IdentityUseCase {
	return &IdentityUseCase{
		userCacheRepo: userCacheRepo,
		apiClient:     apiClient,
		credStore:     credStore,
		notifier:      notifier,
	}
}

// IsLoggedIn returns true if we have stored credentials.
func (uc *IdentityUseCase) IsLoggedIn(ctx context.Context) (bool, error) {
	_, err := uc.credStore.Get(vrchatapi.CredentialService, vrchatapi.CredentialUser)
	return err == nil, nil
}

// Login authenticates with VRChat and persists the auth token to CredentialStore.
func (uc *IdentityUseCase) Login(ctx context.Context, username, password, twoFactorCode string) error {
	if username == "" || password == "" {
		return vrchatapi.ErrInvalidCredentials
	}
	authToken, err := uc.apiClient.Login(ctx, username, password, twoFactorCode)
	if err != nil {
		return err
	}
	if err := uc.credStore.Set(vrchatapi.CredentialService, vrchatapi.CredentialUser, authToken); err != nil {
		return fmt.Errorf("認証情報の保存に失敗しました: %w", err)
	}
	uc.apiClient.SetAuthToken(authToken)
	return nil
}

// Logout clears stored credentials.
func (uc *IdentityUseCase) Logout(ctx context.Context) error {
	uc.apiClient.SetAuthToken("")
	return uc.credStore.Delete(vrchatapi.CredentialService, vrchatapi.CredentialUser)
}

// ListFriends returns cached friends.
func (uc *IdentityUseCase) ListFriends(ctx context.Context) ([]*identity.UserCache, error) {
	return uc.userCacheRepo.List(ctx)
}

// ListFavorites returns cached favorite friends.
func (uc *IdentityUseCase) ListFavorites(ctx context.Context) ([]*identity.UserCache, error) {
	return uc.userCacheRepo.ListFavorites(ctx)
}

// SetFavorite updates a friend's favorite flag.
func (uc *IdentityUseCase) SetFavorite(ctx context.Context, vrcUserID string, favorite bool) error {
	f, err := uc.userCacheRepo.GetByVRCUserID(ctx, vrcUserID)
	if err != nil {
		return err
	}
	if f == nil {
		f = &identity.UserCache{VRCUserID: vrcUserID, DisplayName: vrcUserID, LastUpdated: time.Now()}
	}
	f.IsFavorite = favorite
	f.LastUpdated = time.Now()
	return uc.userCacheRepo.Save(ctx, f)
}

// RefreshFriends fetches from API, updates cache, and notifies when favorites come online.
func (uc *IdentityUseCase) RefreshFriends(ctx context.Context) error {
	// Capture state before refresh for offline→online diff detection
	beforeFavorites, _ := uc.userCacheRepo.ListFavorites(ctx)
	beforeMap := make(map[string]string)
	for _, f := range beforeFavorites {
		beforeMap[f.VRCUserID] = f.Status
	}

	friends, err := uc.apiClient.GetFriends(ctx)
	if err != nil {
		return err
	}
	cached := make([]*identity.UserCache, len(friends))
	for i, f := range friends {
		existing, _ := uc.userCacheRepo.GetByVRCUserID(ctx, f.ID)
		isFav := false
		if existing != nil {
			isFav = existing.IsFavorite
		}
		cached[i] = &identity.UserCache{
			VRCUserID:   f.ID,
			DisplayName: f.DisplayName,
			Status:      f.Status,
			IsFavorite:  isFav,
			LastUpdated: time.Now(),
		}
	}
	if err := uc.userCacheRepo.SaveBatch(ctx, cached); err != nil {
		return err
	}

	// Detect favorite offline→online transitions and notify
	afterMap := make(map[string]*identity.UserCache)
	for _, f := range cached {
		if f.IsFavorite {
			afterMap[f.VRCUserID] = f
		}
	}
	online := identity.DetectFavoriteOnlineTransitions(beforeMap, afterMap)
	if uc.notifier != nil {
		for _, fc := range online {
			_ = uc.notifier.NotifyFavoriteOnline("VRChat Tweaker", fc.DisplayName+" がオンラインになりました")
		}
	}

	return nil
}

// SetStatus changes the current user's status via API.
func (uc *IdentityUseCase) SetStatus(ctx context.Context, status string) error {
	return uc.apiClient.SetUserStatus(ctx, vrchatapi.UserStatus(status))
}
