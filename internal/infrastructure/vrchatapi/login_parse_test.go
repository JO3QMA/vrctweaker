package vrchatapi

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
)

func TestParseAuthUserBodyThenCookie_prefersJSONToken(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(baseURL)
	jar.SetCookies(u, []*http.Cookie{{Name: "auth", Value: "cookie-session"}})

	body := strings.NewReader(`{"authToken":"jwt-like","id":"usr_1"}`)
	got, err := parseAuthUserBodyThenCookie(body, jar)
	if err != nil {
		t.Fatal(err)
	}
	if got != "jwt-like" {
		t.Errorf("token = %q, want jwt-like", got)
	}
}

func TestParseAuthUserBodyThenCookie_fallsBackToCookie(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(baseURL)
	jar.SetCookies(u, []*http.Cookie{{Name: "auth", Value: "cookie-session"}})

	body := strings.NewReader(`{"id":"usr_1","displayName":"X"}`)
	got, err := parseAuthUserBodyThenCookie(body, jar)
	if err != nil {
		t.Fatal(err)
	}
	if got != "cookie-session" {
		t.Errorf("token = %q, want cookie-session", got)
	}
}

func TestParseAuthUserBodyThenCookie_nilJarErrors(t *testing.T) {
	body := strings.NewReader(`{"id":"usr_1"}`)
	_, err := parseAuthUserBodyThenCookie(body, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
