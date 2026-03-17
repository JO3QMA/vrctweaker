package vrchatapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Friend represents a VRChat API friend response.
type Friend struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Status      string `json:"status"`
}

// GetFriends fetches the current user's friend list.
func (c *Client) GetFriends(ctx context.Context) ([]Friend, error) {
	resp, err := c.do(ctx, http.MethodGet, "/auth/user/friends", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var friends []Friend
	if err := json.Unmarshal(body, &friends); err != nil {
		return nil, err
	}
	return friends, nil
}

// SetFavorite sets or unsets a friend as favorite.
func (c *Client) SetFavorite(ctx context.Context, userID string, favorite bool) error {
	method := http.MethodDelete
	path := fmt.Sprintf("/users/%s/favorite", userID)
	if favorite {
		method = http.MethodPut
	}
	resp, err := c.do(ctx, method, path, nil)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}
