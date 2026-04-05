package vrchatapi

import (
	"net/url"
	"testing"
)

func TestNewClient_SetAuthToken_setsAuthCookieInJar(t *testing.T) {
	c := NewClient("session-value-xyz")
	u, err := url.Parse(baseURL)
	if err != nil {
		t.Fatal(err)
	}
	cs := c.httpClient.Jar.Cookies(u)
	var found string
	for _, co := range cs {
		if co.Name == "auth" {
			found = co.Value
			break
		}
	}
	if found != "session-value-xyz" {
		t.Fatalf("auth cookie value = %q, want session-value-xyz", found)
	}
}

func TestNewClient_SetAuthToken_empty_clearsStoredToken(t *testing.T) {
	c := NewClient("tok")
	c.SetAuthToken("")
	if c.GetAuthToken() != "" {
		t.Fatalf("authToken = %q, want empty", c.GetAuthToken())
	}
}
