package wailsapp

import (
	"errors"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Login authenticates with VRChat. On success the returned PlaintextToken must be
// immediately wrapped by the frontend via Web Crypto and persisted via
// PersistWrappedCredential. The token is held in Go memory for the current session.
func (a *App) Login(username, password, twoFactorCode string) LoginResultDTO {
	token, err := a.identity.Login(a.ctx, username, password, twoFactorCode)
	if err != nil {
		return LoginResultDTO{OK: false, Error: err.Error()}
	}
	a.startVRChatPipeline()
	return LoginResultDTO{OK: true, PlaintextToken: token}
}

// Logout clears stored credentials.
func (a *App) Logout() error {
	a.stopVRChatPipeline()
	return a.identity.Logout(a.ctx)
}

// IsLoggedIn returns true when the session is active (token in memory).
func (a *App) IsLoggedIn() (bool, error) {
	return a.identity.IsLoggedIn(a.ctx)
}

// HasStoredCredential returns true when the credential store has a value (blob or legacy).
// Use this on startup to decide whether to attempt unlock before showing the login form.
func (a *App) HasStoredCredential() (bool, error) {
	return a.identity.HasStoredCredential(a.ctx)
}

// GetCredentialBlob returns the raw credential value from the store.
// The frontend inspects the value to decide whether to decrypt (wrapped blob, starts with
// "VRCTWKV1:") or forward as-is (legacy plaintext) before calling UnlockVRChatSession.
func (a *App) GetCredentialBlob() (string, error) {
	return a.identity.GetCredentialBlob(a.ctx)
}

// UnlockVRChatSession sets the decrypted auth token from the frontend and loads the user profile.
// Must be called after the frontend successfully decrypts the credential blob.
func (a *App) UnlockVRChatSession(token string) error {
	if err := a.identity.UnlockSession(a.ctx, token); err != nil {
		return err
	}
	a.startVRChatPipeline()
	return nil
}

// PersistWrappedCredential saves a Web-Crypto wrapped blob to the credential store.
// The blob must start with "VRCTWKV1:" (the WrappedBlobMagic prefix).
// Call this after every successful login or key migration.
func (a *App) PersistWrappedCredential(blob string) error {
	return a.identity.PersistWrappedCredential(a.ctx, blob)
}

// ClearStoredCredential removes the credential blob from the store.
// Call when IDB key loss makes the blob unrecoverable, or for a full logout.
func (a *App) ClearStoredCredential() error {
	return a.identity.ClearStoredCredential(a.ctx)
}

// GetVRChatCurrentUser returns the logged-in user's profile from the VRChat API.
// When forceRefresh is true, bypasses the local self-profile cache and refetches from the API.
func (a *App) GetVRChatCurrentUser(forceRefresh bool) (VRChatCurrentUserDTO, error) {
	u, refreshed, err := a.identity.GetCurrentUser(a.ctx, forceRefresh)
	if err != nil {
		return VRChatCurrentUserDTO{}, err
	}
	if refreshed {
		runtime.EventsEmit(a.ctx, selfCacheChangedEvent, struct{}{})
	}
	if u == nil {
		return VRChatCurrentUserDTO{}, errors.New("empty current user")
	}
	return VRChatCurrentUserDTO{
		ID:                             u.ID,
		DisplayName:                    u.DisplayName,
		Username:                       u.Username,
		Status:                         u.Status,
		StatusDescription:              u.StatusDescription,
		State:                          u.State,
		CurrentAvatarThumbnailImageURL: u.CurrentAvatarThumbnailImageURL,
		UserIcon:                       u.UserIcon,
		ProfilePicOverrideThumbnail:    u.ProfilePicOverrideThumbnail,
	}, nil
}

// RefreshFriends fetches friends from API and updates cache.
func (a *App) RefreshFriends() error {
	return a.identity.RefreshFriends(a.ctx)
}

// ReconcileVRChatSocialCache refreshes self and friends from the VRChat REST API (e.g. after sleep resume).
func (a *App) ReconcileVRChatSocialCache() error {
	if err := a.identity.ReconcileSocialCacheFromAPIHandled(a.ctx); err != nil {
		return err
	}
	runtime.EventsEmit(a.ctx, friendsChangedEvent, struct{}{})
	return nil
}

// Friends returns cached friends.
func (a *App) Friends() ([]UserCacheDTO, error) {
	list, err := a.identity.ListFriends(a.ctx)
	if err != nil {
		return nil, err
	}
	return toUserCacheDTOs(list), nil
}

// ResolveUserProfileNavigation returns routing hints. Logged-in GET /users/{id} is skipped when
// the cached row has an API profile snapshot (username or current avatar) and LastUpdated is
// within UserCacheTTL. Log-only contacts are still fetched. forceRefresh is reserved for an explicit refresh.
func (a *App) ResolveUserProfileNavigation(vrcUserID string) (UserProfileNavigationDTO, error) {
	u, openFriends, openSelf, err := a.identity.ResolveUserProfileForNavigation(a.ctx, vrcUserID, false)
	if err != nil {
		return UserProfileNavigationDTO{}, err
	}
	return UserProfileNavigationDTO{
		User:              toUserCacheDTO(u),
		OpenInFriendsView: openFriends,
		OpenInSelfProfile: openSelf,
	}, nil
}

// GetSelfProfile returns the logged-in user's cached profile row (users_cache user_kind=self).
func (a *App) GetSelfProfile(forceRefresh bool) (UserCacheDTO, error) {
	u, err := a.identity.GetSelfProfile(a.ctx, forceRefresh)
	if err != nil {
		return UserCacheDTO{}, err
	}
	return toUserCacheDTO(u), nil
}

// SetFavorite updates a friend's favorite flag.
func (a *App) SetFavorite(vrcUserID string, favorite bool) error {
	return a.identity.SetFavorite(a.ctx, vrcUserID, favorite)
}
