package vrchatapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// UserStatus represents VRChat user status (Join Me, Ask Me, Busy, etc.).
type UserStatus string

const (
	StatusActive  UserStatus = "active"
	StatusJoinMe  UserStatus = "join me"
	StatusAskMe   UserStatus = "ask me"
	StatusBusy    UserStatus = "busy"
	StatusOffline UserStatus = "offline"
)

func putUserPath(userID string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("update user: empty user id")
	}
	return "/users/" + url.PathEscape(userID), nil
}

// SetUserStatus updates the user's status (PUT /users/{userId}).
func (c *Client) SetUserStatus(ctx context.Context, userID string, status UserStatus) error {
	path, err := putUserPath(userID)
	if err != nil {
		return err
	}
	body := map[string]string{"status": string(status)}
	resp, err := c.do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

// SetUserStatusDescription updates the user's status description (e.g., world name).
func (c *Client) SetUserStatusDescription(ctx context.Context, userID string, description string) error {
	path, err := putUserPath(userID)
	if err != nil {
		return err
	}
	body := map[string]string{"statusDescription": description}
	resp, err := c.do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

// SetUserStatusAndDescription updates status and status description in a single request.
func (c *Client) SetUserStatusAndDescription(ctx context.Context, userID string, status UserStatus, description string) error {
	path, err := putUserPath(userID)
	if err != nil {
		return err
	}
	body := map[string]string{
		"status":            string(status),
		"statusDescription": description,
	}
	resp, err := c.do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}
