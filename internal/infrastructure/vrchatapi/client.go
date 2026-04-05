package vrchatapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"
)

const baseURL = "https://api.vrchat.cloud/api/1"

// Client is the VRChat Web API client.
type Client struct {
	mu         sync.RWMutex
	httpClient *http.Client
	authToken  string
	// apiRoot is the full base URL including /api/1. Empty means use package baseURL (production).
	apiRoot string
}

// NewClient creates a new VRChat API client.
func NewClient(authToken string) *Client {
	jar, _ := cookiejar.New(nil)
	c := &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
			Jar:     jar,
		},
		authToken: "",
	}
	c.SetAuthToken(authToken)
	return c
}

// SetAuthToken updates the stored session credential and syncs the `auth` cookie
// for api.vrchat.cloud. VRChat accepts either JSON authToken (often sent as
// Basic in examples) or the `auth` cookie; after 2FA we only persist the cookie
// value, which must be sent as Cookie, not as Authorization: Basic.
// SetAuthToken is safe for concurrent use.
func (c *Client) SetAuthToken(token string) {
	c.mu.Lock()
	c.authToken = token
	c.mu.Unlock()

	if c.httpClient.Jar == nil {
		return
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return
	}
	// net/http/cookiejar.Jar is safe for concurrent use.
	if token == "" {
		c.httpClient.Jar.SetCookies(u, []*http.Cookie{
			{Name: "auth", Path: "/", MaxAge: -1},
		})
		return
	}
	c.httpClient.Jar.SetCookies(u, []*http.Cookie{
		{Name: "auth", Value: token, Path: "/", Secure: true},
	})
}

// GetAuthToken returns the stored auth token. Empty string means no active session.
// Safe for concurrent use.
func (c *Client) GetAuthToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.authToken
}

// do performs an HTTP request; the CookieJar sends the `auth` cookie set by SetAuthToken.
func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	root := baseURL
	if c.apiRoot != "" {
		root = c.apiRoot
	}
	req, err := http.NewRequestWithContext(ctx, method, root+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 && c.GetAuthToken() != "" {
		// A 401 while we believed we were authenticated means the server-side session
		// has expired or been invalidated. Signal this so callers can clear state.
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%w: %s %s", ErrSessionExpired, method, path)
	}
	if resp.StatusCode >= 400 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		_ = resp.Body.Close()
		msg := string(snippet)
		if msg == "" {
			return nil, fmt.Errorf("API error: %s %s - %d", method, path, resp.StatusCode)
		}
		return nil, fmt.Errorf("API error: %s %s - %d: %s", method, path, resp.StatusCode, msg)
	}

	return resp, nil
}
