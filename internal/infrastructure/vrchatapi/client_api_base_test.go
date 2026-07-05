package vrchatapi

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
)

func TestClient_apiBase_usesProductionWhenUnset(t *testing.T) {
	c := NewClient("")
	if got := c.apiBase(); got != baseURL {
		t.Fatalf("apiBase = %q, want %q", got, baseURL)
	}
}

func TestClient_apiBase_usesCustomRoot(t *testing.T) {
	c := NewClient("")
	c.apiRoot = "http://127.0.0.1:9999/api/1"
	if got := c.apiBase(); got != c.apiRoot {
		t.Fatalf("apiBase = %q, want %q", got, c.apiRoot)
	}
}

func TestClient_getAuthCookieFromJar_ok(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(baseURL)
	jar.SetCookies(u, []*http.Cookie{{Name: "auth", Value: "sess"}})

	c := NewClient("")
	got, err := c.getAuthCookieFromJar(jar)
	if err != nil {
		t.Fatalf("getAuthCookieFromJar: %v", err)
	}
	if got != "sess" {
		t.Fatalf("cookie = %q, want sess", got)
	}
}

func TestClient_getAuthCookieFromJar_missing(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	c := NewClient("")
	if _, err := c.getAuthCookieFromJar(jar); err == nil {
		t.Fatal("expected error for missing auth cookie")
	}
}
