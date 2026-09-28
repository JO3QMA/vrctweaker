package vrchatapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Avatar is a minimal GET /avatars/{id} payload for display name resolution.
type Avatar struct {
	ID   string `json:"id"`
	Name string `json:"displayName"`
}

// GetAvatar fetches avatar metadata by id (requires auth).
func (c *Client) GetAvatar(ctx context.Context, avatarID string) (*Avatar, error) {
	avatarID = strings.TrimSpace(avatarID)
	if avatarID == "" {
		return nil, fmt.Errorf("empty avatar id")
	}
	path := "/avatars/" + avatarID
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var av Avatar
	if err := json.Unmarshal(body, &av); err != nil {
		return nil, fmt.Errorf("decode avatar: %w", err)
	}
	return &av, nil
}
