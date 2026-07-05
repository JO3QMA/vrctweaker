package vrchatapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_Login_successWithAuthToken(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/1/auth/user" || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "myuser" || pass != "mypass" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(authUserResponse{AuthToken: "direct-token"})
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	got, err := c.Login(context.Background(), "myuser", "mypass", "")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if got != "direct-token" {
		t.Fatalf("token = %q, want direct-token", got)
	}
}

func TestClient_Login_invalidCredentials(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	_, err := c.Login(context.Background(), "bad", "bad", "")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestClient_Login_apiError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	_, err := c.Login(context.Background(), "u", "p", "")
	if err == nil || !strings.Contains(err.Error(), "API error") {
		t.Fatalf("err = %v, want API error", err)
	}
}

func TestClient_Login_twoFactorRequiredWithoutCode(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/1/auth/user" {
			http.NotFound(w, r)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "auth", Value: "partial-session", Path: "/"})
		_ = json.NewEncoder(w).Encode(authUserResponse{RequiresTwoFactorAuth: []string{"totp"}})
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	_, err := c.Login(context.Background(), "u", "p", "")
	if !errors.Is(err, ErrTwoFactorRequired) {
		t.Fatalf("err = %v, want ErrTwoFactorRequired", err)
	}
}

func TestClient_Login_twoFactorSuccess(t *testing.T) {
	t.Parallel()
	var authUserCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/1/auth/user" && r.Method == http.MethodGet:
			authUserCalls++
			if authUserCalls == 1 {
				if _, _, ok := r.BasicAuth(); !ok {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				http.SetCookie(w, &http.Cookie{Name: "auth", Value: "partial-session", Path: "/"})
				_ = json.NewEncoder(w).Encode(authUserResponse{RequiresTwoFactorAuth: []string{"totp"}})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "usr_1"})
		case r.URL.Path == "/api/1/auth/twofactorauth/totp/verify" && r.Method == http.MethodPost:
			http.SetCookie(w, &http.Cookie{Name: "auth", Value: "full-session", Path: "/"})
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	got, err := c.Login(context.Background(), "u", "p", "123456")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if got != "full-session" {
		t.Fatalf("token = %q, want full-session", got)
	}
	if authUserCalls != 2 {
		t.Fatalf("auth/user calls = %d, want 2", authUserCalls)
	}
}

func TestClient_Login_twoFactorInvalidCode(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/1/auth/user":
			http.SetCookie(w, &http.Cookie{Name: "auth", Value: "partial", Path: "/"})
			_ = json.NewEncoder(w).Encode(authUserResponse{})
		case "/api/1/auth/twofactorauth/totp/verify":
			http.Error(w, "bad code", http.StatusUnauthorized)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	_, err := c.Login(context.Background(), "u", "p", "000000")
	if err == nil || !strings.Contains(err.Error(), "2FA") {
		t.Fatalf("err = %v, want 2FA error", err)
	}
}

func TestClient_Login_twoFactorMissingSessionCookie(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/1/auth/user" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(authUserResponse{})
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	_, err := c.Login(context.Background(), "u", "p", "123456")
	if err == nil || !strings.Contains(err.Error(), "no session cookie") {
		t.Fatalf("err = %v, want missing session cookie", err)
	}
}

func TestClient_Login_twoFactorVerifyServerError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/1/auth/user":
			http.SetCookie(w, &http.Cookie{Name: "auth", Value: "partial", Path: "/"})
			_ = json.NewEncoder(w).Encode(authUserResponse{})
		case "/api/1/auth/twofactorauth/totp/verify":
			http.Error(w, "server blew up", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	_, err := c.Login(context.Background(), "u", "p", "123456")
	if err == nil || !strings.Contains(err.Error(), "2FA verification failed") {
		t.Fatalf("err = %v, want verification failed", err)
	}
}

func TestClient_getAuthUserWithJar_apiError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Jar: jar}
	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	_, err = c.getAuthUserWithJar(context.Background(), client)
	if err == nil || !strings.Contains(err.Error(), "get user after 2FA") {
		t.Fatalf("err = %v, want get user after 2FA error", err)
	}
}

func TestClient_getAuthUserWithJar_returnsJSONToken(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(authUserResponse{AuthToken: "jwt-after-2fa"})
	}))
	t.Cleanup(srv.Close)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Jar: jar}
	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	got, err := c.getAuthUserWithJar(context.Background(), client)
	if err != nil {
		t.Fatalf("getAuthUserWithJar: %v", err)
	}
	if got != "jwt-after-2fa" {
		t.Fatalf("token = %q, want jwt-after-2fa", got)
	}
}

func TestClient_parseAuthUserResponse_invalidJSON(t *testing.T) {
	c := NewClient("")
	_, err := c.parseAuthUserResponse(strings.NewReader("not-json"))
	if err == nil {
		t.Fatal("expected parse error")
	}
}
