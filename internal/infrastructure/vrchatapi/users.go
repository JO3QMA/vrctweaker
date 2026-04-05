package vrchatapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// GetUser fetches a user profile (GET /users/{userId}). Response shape matches Friend for shared fields.
func (c *Client) GetUser(ctx context.Context, userID string) (*Friend, error) {
	if userID == "" {
		return nil, fmt.Errorf("get user: empty user id")
	}
	path := "/users/" + url.PathEscape(userID)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var u Friend
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, fmt.Errorf("decode user: %w", err)
	}
	if u.ID == "" {
		return nil, fmt.Errorf("get user: empty id in response")
	}
	return &u, nil
}
