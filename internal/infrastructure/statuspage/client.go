package statuspage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"vrchat-tweaker/internal/infrastructure/vrchatapi"
)

const (
	DefaultBaseURL           = "https://status.vrchat.com/api/v2/"
	serverStatusHTTPTimeout  = 15 * time.Second
	serverStatusMaxBodyBytes = 1 << 20
	serverStatusAllowedHost  = "status.vrchat.com"
)

// Client fetches VRChat statuspage API responses.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	UserAgent  string

	once          sync.Once
	defaultClient *http.Client
	// insecureSkipHostCheck disables status.vrchat.com host checks (tests only).
	insecureSkipHostCheck bool
}

// NewClient returns a client with production defaults.
func NewClient() *Client {
	return &Client{
		BaseURL:   DefaultBaseURL,
		UserAgent: vrchatapi.UserAgent,
	}
}

// NewTestClient returns a client for httptest servers (skips production host checks).
func NewTestClient(baseURL string) *Client {
	return &Client{
		BaseURL:               baseURL,
		insecureSkipHostCheck: true,
	}
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	c.once.Do(func() {
		c.defaultClient = &http.Client{
			Timeout:       serverStatusHTTPTimeout,
			CheckRedirect: c.redirectPolicy,
		}
	})
	return c.defaultClient
}

func (c *Client) redirectPolicy(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("too many redirects")
	}
	if c != nil && c.insecureSkipHostCheck {
		return nil
	}
	return validateHost(req.URL)
}

func validateHost(u *url.URL) error {
	if u == nil {
		return errors.New("url is nil")
	}
	if strings.ToLower(u.Scheme) != "https" {
		return fmt.Errorf("url must use https: %s", u.String())
	}
	host := strings.ToLower(u.Hostname())
	if host != serverStatusAllowedHost {
		return fmt.Errorf("host not allowed: %s", host)
	}
	return nil
}

func (c *Client) validateRequestHost(u *url.URL) error {
	if c != nil && c.insecureSkipHostCheck {
		return nil
	}
	return validateHost(u)
}

func (c *Client) fetchJSON(ctx context.Context, path string, dest any) error {
	if strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must be relative: %q", path)
	}

	base := strings.TrimSpace(c.BaseURL)
	if base == "" {
		base = DefaultBaseURL
	}
	rawURL, err := url.Parse(base)
	if err != nil {
		return fmt.Errorf("parse base url: %w", err)
	}
	ref, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("parse path: %w", err)
	}
	full := rawURL.ResolveReference(ref)
	if hostErr := c.validateRequestHost(full); hostErr != nil {
		return hostErr
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, full.String(), nil)
	if err != nil {
		return err
	}
	ua := strings.TrimSpace(c.UserAgent)
	if ua == "" {
		ua = vrchatapi.UserAgent
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, serverStatusMaxBodyBytes+1))
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	if len(body) > serverStatusMaxBodyBytes {
		return fmt.Errorf("response body exceeds %d bytes", serverStatusMaxBodyBytes)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("statuspage api: %s: %s", resp.Status, truncateForErr(string(body), 200))
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}
	return nil
}

func truncateForErr(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
