package vrchatapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_do_sessionExpired(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	c := NewClient("active-token")
	c.apiRoot = srv.URL + "/api/1"
	_, err := c.do(context.Background(), http.MethodGet, "/auth/user", nil)
	if !errors.Is(err, ErrSessionExpired) {
		t.Fatalf("err = %v, want ErrSessionExpired", err)
	}
}

func TestClient_do_apiErrorWithBody(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request detail", http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	_, err := c.do(context.Background(), http.MethodGet, "/users/x", nil)
	if err == nil || !strings.Contains(err.Error(), "bad request detail") {
		t.Fatalf("err = %v, want API error with body", err)
	}
}

func TestClient_do_apiErrorEmptyBody(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	_, err := c.do(context.Background(), http.MethodGet, "/users/x", nil)
	if err == nil || !strings.Contains(err.Error(), "API error") {
		t.Fatalf("err = %v, want API error", err)
	}
}

func TestClient_do_401WithoutAuthTokenNotSessionExpired(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	_, err := c.do(context.Background(), http.MethodGet, "/auth/user", nil)
	if err == nil || strings.Contains(err.Error(), ErrSessionExpired.Error()) {
		t.Fatalf("err = %v, want generic API error not session expired", err)
	}
}

func TestClient_do_jsonMarshalError(t *testing.T) {
	c := NewClient("")
	c.apiRoot = "http://example.invalid/api/1"
	_, err := c.do(context.Background(), http.MethodPost, "/users/x", make(chan int))
	if err == nil {
		t.Fatal("want json marshal error")
	}
}

func TestClient_SetAuthToken_nilJarNoPanic(t *testing.T) {
	c := &Client{httpClient: &http.Client{}}
	c.SetAuthToken("tok")
	if c.GetAuthToken() != "tok" {
		t.Fatalf("authToken = %q, want tok", c.GetAuthToken())
	}
}
