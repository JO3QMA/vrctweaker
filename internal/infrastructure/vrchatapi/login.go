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

	// Partial session must be in the jar after the first GET /auth/user.
	if _, cookieErr := c.getAuthCookieFromJar(jar); cookieErr != nil {
		return "", fmt.Errorf("2FA required but no session cookie: %w", cookieErr)
	}

	// POST /auth/twofactorauth/totp/verify (CookieJar sends auth cookie; do not override it)
	if verifyErr := c.verifyTwoFactor(ctx, loginClient, twoFactorCode); verifyErr != nil {
		return "", verifyErr
	}

	// GET /auth/user again. Use the cookie jar only — do not attach the pre-verify
	// auth cookie manually; TOTP verify may Set-Cookie a new session value.
	// If JSON omits authToken (common), use the updated auth cookie as API token.
	authToken, err = c.getAuthUserWithJar(ctx, loginClient)
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
	req.Header.Set("User-Agent", userAgent)
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

// parseAuthUserBodyThenCookie returns authToken from JSON, or if empty the auth
// cookie from jar (VRChat often omits authToken when the session is cookie-only).
func parseAuthUserBodyThenCookie(body io.Reader, jar http.CookieJar) (string, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}
	var resp authUserResponse
	if unmarshalErr := json.Unmarshal(data, &resp); unmarshalErr != nil {
		return "", fmt.Errorf("parse response: %w", unmarshalErr)
	}
	if resp.AuthToken != "" {
		return resp.AuthToken, nil
	}
	if jar == nil {
		return "", errors.New("no authToken in JSON and no cookie jar")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	for _, cookie := range jar.Cookies(u) {
		if cookie.Name == "auth" && cookie.Value != "" {
			return cookie.Value, nil
		}
	}
	return "", errors.New("auth cookie not found or empty after empty authToken JSON")
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

func (c *Client) verifyTwoFactor(ctx context.Context, client *http.Client, code string) error {
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
	req.Header.Set("User-Agent", userAgent)

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

// getAuthUserWithJar performs GET /auth/user using only the client's CookieJar
// (no manually injected cookies). Returns authToken from JSON, or if empty the
// auth cookie value (VRChat may omit authToken after cookie-based session).
func (c *Client) getAuthUserWithJar(ctx context.Context, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/auth/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("get user after 2FA: %d", resp.StatusCode)
	}
	token, err := parseAuthUserBodyThenCookie(resp.Body, client.Jar)
	if err != nil {
		return "", err
	}
	return token, nil
}
