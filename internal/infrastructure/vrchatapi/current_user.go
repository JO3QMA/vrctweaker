package vrchatapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ErrNotAuthenticated is returned when GetCurrentUser is called without an auth token.
var ErrNotAuthenticated = errors.New("not authenticated")

// ErrSessionExpired is returned when the VRChat API responds with 401 after the client
// believed it was authenticated. This indicates the auth cookie has been invalidated
// server-side (password change, manual logout, session expiry, etc.).
var ErrSessionExpired = errors.New("session expired")

// CurrentUserProfile is a subset of GET /auth/user for display (non-sensitive fields).
type CurrentUserProfile struct {
	ID                             string `json:"id"`
	DisplayName                    string `json:"displayName"`
	Username                       string `json:"username"`
	Status                         string `json:"status"`
	StatusDescription              string `json:"statusDescription"`
	State                          string `json:"state"`
	CurrentAvatarThumbnailImageURL string `json:"currentAvatarThumbnailImageUrl"`
	UserIcon                       string `json:"userIcon"`
	ProfilePicOverrideThumbnail    string `json:"profilePicOverrideThumbnail"`
}

// GetCurrentUser fetches the logged-in user via GET /auth/user (session `auth` cookie).
func (c *Client) GetCurrentUser(ctx context.Context) (*CurrentUserProfile, error) {
	if c.GetAuthToken() == "" {
		return nil, ErrNotAuthenticated
	}
	resp, err := c.do(ctx, http.MethodGet, "/auth/user", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var u CurrentUserProfile
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, fmt.Errorf("parse current user: %w", err)
	}
	return &u, nil
}
