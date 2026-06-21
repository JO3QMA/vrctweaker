package discordrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// exchangeCode exchanges an AUTHORIZE code for an access token.
func exchangeCode(code string) (string, error) {
	secret := ClientSecret()
	if secret == "" {
		return "", fmt.Errorf("discord_client_secret_missing")
	}
	form := url.Values{}
	form.Set("client_id", ClientID)
	form.Set("client_secret", secret)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", "http://127.0.0.1")
	req, err := http.NewRequest(http.MethodPost, "https://discord.com/api/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("discord_token_exchange_failed: %s", bytes.TrimSpace(body))
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("discord_token_empty")
	}
	return out.AccessToken, nil
}
