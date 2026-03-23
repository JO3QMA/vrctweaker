package vrchatapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://api.vrchat.cloud/api/1"

// Client is the VRChat Web API client.
type Client struct {
	httpClient *http.Client
	authToken  string
}

// NewClient creates a new VRChat API client.
func NewClient(authToken string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		authToken:  authToken,
	}
}

// SetAuthToken updates the authentication token.
func (c *Client) SetAuthToken(token string) {
	c.authToken = token
}

// do performs an HTTP request with auth headers.
func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	if c.authToken != "" {
		req.Header.Set("Authorization", "Basic "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("API error: %s %s - %d", method, path, resp.StatusCode)
	}

	return resp, nil
}
