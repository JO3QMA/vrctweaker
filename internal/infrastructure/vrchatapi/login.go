package vrchatapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// authUserResponse is the response from GET /auth/user.
type authUserResponse struct {
	AuthToken             string   `json:"authToken"`
	RequiresTwoFactorAuth []string `json:"requiresTwoFactorAuth"`
}

// twoFactorVerifyRequest is the request body for POST /auth/twofactorauth/totp/verify.
type twoFactorVerifyRequest struct {
	Code string `json:"code"`
}

// ErrInvalidCredentials is returned when login fails (401).
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrTwoFactorRequired is returned when 2FA code is required but missing or wrong.
var ErrTwoFactorRequired = errors.New("two-factor authentication required")

// Login authenticates with VRChat API and returns the auth token.
// For 2FA-enabled accounts, pass the TOTP code in twoFactorCode.
func (c *Client) Login(ctx context.Context, username, password, twoFactorCode string) (string, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", fmt.Errorf("create cookiejar: %w", err)
	}
	loginClient := &http.Client{
		Timeout: 15 * time.Second,
		Jar:     jar,
	}

	// GET /auth/user with Basic auth (username:password)
	authToken, err := c.loginWithBasicAuth(ctx, loginClient, username, password)
	if err != nil {
		return "", err
	}
	if authToken != "" {
		return authToken, nil
	}

	// authToken empty: 2FA required
	if twoFactorCode == "" {
		return "", ErrTwoFactorRequired
	}

	// Get auth cookie for verify
	authCookie, err := c.getAuthCookieFromJar(jar)
	if err != nil {
		return "", fmt.Errorf("2FA required but no session cookie: %w", err)
	}

	// POST /auth/twofactorauth/totp/verify
	if err := c.verifyTwoFactor(ctx, loginClient, authCookie, twoFactorCode); err != nil {
		return "", err
	}

	// GET /auth/user again to get authToken
	authToken, err = c.getAuthUserWithCookie(ctx, loginClient, authCookie)
	if err != nil {
		return "", err
	}
	if authToken == "" {
		return "", errors.New("2FA verification succeeded but no auth token received")
	}
	return authToken, nil
}

func (c *Client) loginWithBasicAuth(ctx context.Context, client *http.Client, username, password string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/auth/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "VRChat Tweaker/1.0")
	req.SetBasicAuth(url.QueryEscape(username), url.QueryEscape(password))

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 401 {
		return "", ErrInvalidCredentials
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error: %d", resp.StatusCode)
	}

	return c.parseAuthUserResponse(resp.Body)
}

func (c *Client) parseAuthUserResponse(body io.Reader) (string, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}
	var resp authUserResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	return resp.AuthToken, nil
}

func (c *Client) getAuthCookieFromJar(jar http.CookieJar) (string, error) {
	u, _ := url.Parse(baseURL)
	cookies := jar.Cookies(u)
	for _, cookie := range cookies {
		if cookie.Name == "auth" {
			return cookie.Value, nil
		}
	}
	return "", errors.New("auth cookie not found")
}

func (c *Client) verifyTwoFactor(ctx context.Context, client *http.Client, authCookie, code string) error {
	body := twoFactorVerifyRequest{Code: code}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/auth/twofactorauth/totp/verify", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "VRChat Tweaker/1.0")
	req.AddCookie(&http.Cookie{Name: "auth", Value: authCookie})

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("2FA verify request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 401 {
		return errors.New("2FAコードが正しくありません")
	}
	if resp.StatusCode >= 400 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("2FA verification failed: %d %s", resp.StatusCode, string(msg))
	}
	return nil
}

func (c *Client) getAuthUserWithCookie(ctx context.Context, client *http.Client, authCookie string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/auth/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "VRChat Tweaker/1.0")
	req.AddCookie(&http.Cookie{Name: "auth", Value: authCookie})

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("get user after 2FA: %d", resp.StatusCode)
	}
	return c.parseAuthUserResponse(resp.Body)
}
