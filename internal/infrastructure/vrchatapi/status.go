package vrchatapi

import (
	"context"
	"net/http"
)

// UserStatus represents VRChat user status (Join Me, Ask Me, Busy, etc.).
type UserStatus string

const (
	StatusJoinMe  UserStatus = "join me"
	StatusAskMe   UserStatus = "ask me"
	StatusBusy    UserStatus = "busy"
	StatusOffline UserStatus = "offline"
)

// SetUserStatus updates the current user's status.
func (c *Client) SetUserStatus(ctx context.Context, status UserStatus) error {
	body := map[string]string{"status": string(status)}
	resp, err := c.do(ctx, http.MethodPut, "/users/me", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

// SetUserStatusDescription updates the current user's status description (e.g., world name).
func (c *Client) SetUserStatusDescription(ctx context.Context, description string) error {
	body := map[string]string{"statusDescription": description}
	resp, err := c.do(ctx, http.MethodPut, "/users/me", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}
