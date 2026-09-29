package vrchatapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Avatar is a minimal GET /avatars/{id} payload for display name resolution.
// VRChat returns the avatar title in the "name" field (not displayName).
type Avatar struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetAvatar fetches avatar metadata by id (requires auth).
func (c *Client) GetAvatar(ctx context.Context, avatarID string) (*Avatar, error) {
	avatarID = strings.TrimSpace(avatarID)
	if avatarID == "" {
		return nil, fmt.Errorf("empty avatar id")
	}
	path := "/avatars/" + url.PathEscape(avatarID)
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
	if strings.TrimSpace(av.ID) == "" {
		return nil, fmt.Errorf("avatar response missing id")
	}
	return &av, nil
}
