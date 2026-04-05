package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/domain/settings"
	"vrchat-tweaker/internal/infrastructure/vrchatapi"
)

// IdentityUseCase handles VRChat auth, friends, and status.
type IdentityUseCase struct {
	userCacheRepo identity.UserCacheRepository
	apiClient     vrchatapi.VRChatAPIClient
	credStore     vrchatapi.CredentialStore
	settingsRepo  settings.AppSettingsRepository
	notifier      identity.Notifier // optional; nil skips online notifications
}

// NewIdentityUseCase creates a new IdentityUseCase.
func NewIdentityUseCase(
	userCacheRepo identity.UserCacheRepository,
	apiClient vrchatapi.VRChatAPIClient,
	credStore vrchatapi.CredentialStore,
	settingsRepo settings.AppSettingsRepository,
) *IdentityUseCase {
	return NewIdentityUseCaseWithNotifier(userCacheRepo, apiClient, credStore, settingsRepo, nil)
}

// NewIdentityUseCaseWithNotifier creates a new IdentityUseCase with optional Notifier for favorite-online notifications.
func NewIdentityUseCaseWithNotifier(
	userCacheRepo identity.UserCacheRepository,
	apiClient vrchatapi.VRChatAPIClient,
	credStore vrchatapi.CredentialStore,
	settingsRepo settings.AppSettingsRepository,
	notifier identity.Notifier,
) *IdentityUseCase {
	return &IdentityUseCase{
		userCacheRepo: userCacheRepo,
		apiClient:     apiClient,
		credStore:     credStore,
		settingsRepo:  settingsRepo,
		notifier:      notifier,
	}
}

// IsLoggedIn returns true if we have stored credentials.
func (uc *IdentityUseCase) IsLoggedIn(ctx context.Context) (bool, error) {
	_, err := uc.credStore.Get(vrchatapi.CredentialService, vrchatapi.CredentialUser)
	return err == nil, nil
}

func (uc *IdentityUseCase) friendsSyncStale(ctx context.Context) (bool, error) {
	if uc.settingsRepo == nil {
		return true, nil
	}
	v, err := uc.settingsRepo.Get(ctx, identity.SettingVRChatFriendsSyncedAt)
	if err != nil {
		return true, err
	}
	if v == "" {
		return true, nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return true, nil
	}
	return time.Since(t) > identity.UserCacheTTL, nil
}

func userCacheToCurrentProfile(u *identity.UserCache) *vrchatapi.CurrentUserProfile {
	return &vrchatapi.CurrentUserProfile{
		ID:                             u.VRCUserID,
		DisplayName:                    u.DisplayName,
		Username:                       u.Username,
		Status:                         u.Status,
		StatusDescription:              u.StatusDescription,
		State:                          u.UserState,
		CurrentAvatarThumbnailImageURL: u.AvatarThumbnailURL,
		UserIcon:                       u.UserIconURL,
		ProfilePicOverrideThumbnail:    u.ProfilePicOverrideThumbnail,
	}
}

func currentUserProfileToSelfCache(p *vrchatapi.CurrentUserProfile, fingerprint string, at time.Time) *identity.UserCache {
	return &identity.UserCache{
		VRCUserID:                   p.ID,
		DisplayName:                 p.DisplayName,
		Status:                      p.Status,
		UserKind:                    identity.UserKindSelf,
		LastUpdated:                 at,
		SessionFingerprint:          fingerprint,
		Username:                    p.Username,
		StatusDescription:           p.StatusDescription,
		UserState:                   p.State,
		AvatarThumbnailURL:          p.CurrentAvatarThumbnailImageURL,
		UserIconURL:                 p.UserIcon,
		ProfilePicOverrideThumbnail: p.ProfilePicOverrideThumbnail,
	}
}

// GetCurrentUser returns the logged-in VRChat user profile (cached up to UserCacheTTL).
// When forceRefresh is true, the API is always called and the cache is updated (e.g. user-triggered re-fetch).
func (uc *IdentityUseCase) GetCurrentUser(ctx context.Context, forceRefresh bool) (*vrchatapi.CurrentUserProfile, error) {
	token, err := uc.credStore.Get(vrchatapi.CredentialService, vrchatapi.CredentialUser)
	if err != nil || token == "" {
		return nil, vrchatapi.ErrNotAuthenticated
	}
	uc.apiClient.SetAuthToken(token)
	fp := identity.AuthTokenFingerprint(token)
	if !forceRefresh && fp != "" {
		row, gerr := uc.userCacheRepo.GetSelfBySessionFingerprint(ctx, fp)
		if gerr != nil {
			return nil, gerr
		}
		if row != nil && time.Since(row.LastUpdated) < identity.UserCacheTTL {
			return userCacheToCurrentProfile(row), nil
		}
	}
	u, err := uc.apiClient.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	cache := currentUserProfileToSelfCache(u, fp, time.Now())
	if err := uc.userCacheRepo.UpsertSelf(ctx, cache); err != nil {
		return nil, fmt.Errorf("cache current user: %w", err)
	}
	return u, nil
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
	if _, err := uc.GetCurrentUser(ctx, true); err != nil {
		log.Printf("identity: current user after login: %v", err)
	}
	return nil
}

// Logout clears stored credentials and removes cached self profile rows.
func (uc *IdentityUseCase) Logout(ctx context.Context) error {
	uc.apiClient.SetAuthToken("")
	var selfErr error
	if err := uc.userCacheRepo.DeleteSelfRows(ctx); err != nil {
		selfErr = fmt.Errorf("clear self profile cache: %w", err)
	}
	if uc.settingsRepo != nil {
		_ = uc.settingsRepo.Set(ctx, identity.SettingVRChatFriendsSyncedAt, "")
	}
	if err := uc.credStore.Delete(vrchatapi.CredentialService, vrchatapi.CredentialUser); err != nil {
		return err
	}
	return selfErr
}

// ListFriends returns cached friends, refreshing from the API when the friends sync is stale and the user is logged in.
func (uc *IdentityUseCase) ListFriends(ctx context.Context) ([]*identity.UserCache, error) {
	loggedIn, _ := uc.IsLoggedIn(ctx)
	if loggedIn && uc.settingsRepo != nil {
		stale, err := uc.friendsSyncStale(ctx)
		if err == nil && stale {
			if rerr := uc.RefreshFriends(ctx); rerr != nil {
				log.Printf("identity: ListFriends: background RefreshFriends failed: %v", rerr)
			}
		}
	}
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
		f = &identity.UserCache{
			VRCUserID:   vrcUserID,
			DisplayName: vrcUserID,
			UserKind:    identity.UserKindFriend,
			LastUpdated: time.Now(),
		}
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
	now := time.Now()
	cached := make([]*identity.UserCache, len(friends))
	for i, f := range friends {
		existing, err := uc.userCacheRepo.GetByVRCUserID(ctx, f.ID)
		if err != nil {
			return err
		}
		if existing == nil {
			existing = &identity.UserCache{VRCUserID: f.ID}
		}
		isFav := existing.IsFavorite
		apiSnap := userCacheFromFriend(f, isFav, now)
		existing.MergeFromAPIFriend(apiSnap)
		cached[i] = existing
	}
	if err := uc.userCacheRepo.SaveBatch(ctx, cached); err != nil {
		return err
	}

	if uc.settingsRepo != nil {
		if err := uc.settingsRepo.Set(ctx, identity.SettingVRChatFriendsSyncedAt, time.Now().UTC().Format(time.RFC3339)); err != nil {
			return err
		}
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

// ErrProfileNotInCache is returned when the user is not in users_cache and the client is not logged in.
var ErrProfileNotInCache = errors.New("user not in cache; log in to load profile from VRChat")

// ResolveUserProfileForNavigation loads or refreshes a user row via cache and GET /users/{id} when logged in.
// The second return value is true when the row should be shown in the friends list view (user_kind friend).
func (uc *IdentityUseCase) ResolveUserProfileForNavigation(ctx context.Context, vrcUserID string) (*identity.UserCache, bool, error) {
	id := strings.TrimSpace(vrcUserID)
	if id == "" {
		return nil, false, fmt.Errorf("empty vrc user id")
	}
	row, err := uc.userCacheRepo.GetByVRCUserID(ctx, id)
	if err != nil {
		return nil, false, err
	}
	loggedIn, err := uc.IsLoggedIn(ctx)
	if err != nil {
		return nil, false, err
	}
	if !loggedIn {
		if row == nil {
			return nil, false, ErrProfileNotInCache
		}
		return row, row.UserKind == identity.UserKindFriend, nil
	}
	token, terr := uc.credStore.Get(vrchatapi.CredentialService, vrchatapi.CredentialUser)
	if terr != nil || token == "" {
		return nil, false, vrchatapi.ErrNotAuthenticated
	}
	uc.apiClient.SetAuthToken(token)
	f, err := uc.apiClient.GetUser(ctx, id)
	if err != nil {
		if row == nil {
			return nil, false, err
		}
		return row, row.UserKind == identity.UserKindFriend, nil
	}
	now := time.Now()
	if row == nil {
		row = &identity.UserCache{VRCUserID: id}
	}
	snap := userCacheFromFriend(*f, row.IsFavorite, now)
	row.MergeFromGetUserAPI(f.IsFriend, snap, now)
	if err := uc.userCacheRepo.Save(ctx, row); err != nil {
		return nil, false, err
	}
	return row, row.UserKind == identity.UserKindFriend, nil
}

func jsonStringSlice(s []string) string {
	if len(s) == 0 {
		return ""
	}
	b, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	return string(b)
}

func userCacheFromFriend(f vrchatapi.Friend, isFavorite bool, now time.Time) *identity.UserCache {
	return &identity.UserCache{
		VRCUserID:                   f.ID,
		DisplayName:                 f.DisplayName,
		Status:                      f.Status,
		IsFavorite:                  isFavorite,
		LastUpdated:                 now,
		UserKind:                    identity.UserKindFriend,
		Username:                    f.Username,
		StatusDescription:           f.StatusDescription,
		UserState:                   f.UserState,
		AvatarThumbnailURL:          f.CurrentAvatarThumbnailImageURL,
		UserIconURL:                 f.UserIcon,
		ProfilePicOverrideThumbnail: f.ProfilePicOverrideThumbnail,
		Bio:                         f.Bio,
		BioLinksJSON:                jsonStringSlice(f.BioLinks),
		CurrentAvatarImageURL:       f.CurrentAvatarImageURL,
		CurrentAvatarTagsJSON:       jsonStringSlice(f.CurrentAvatarTags),
		DeveloperType:               f.DeveloperType,
		FriendKey:                   f.FriendKey,
		ImageURL:                    f.ImageURL,
		LastPlatform:                f.LastPlatform,
		Location:                    f.Location,
		LastLogin:                   f.LastLogin,
		LastActivity:                f.LastActivity,
		LastMobile:                  f.LastMobile,
		Platform:                    f.Platform,
		ProfilePicOverride:          f.ProfilePicOverride,
		TagsJSON:                    jsonStringSlice(f.Tags),
	}
}
