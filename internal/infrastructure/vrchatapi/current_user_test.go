package vrchatapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCurrentUserProfile_JSON(t *testing.T) {
	const fixture = `{
		"id": "usr_c1644b5b-3ca4-45b4-97c6-a2a0de70d469",
		"displayName": "Display One",
		"username": "user_one",
		"status": "active",
		"statusDescription": "In a world",
		"state": "offline",
		"currentAvatarThumbnailImageUrl": "https://api.vrchat.cloud/api/1/image/file_x/1/256",
		"userIcon": "",
		"profilePicOverrideThumbnail": "",
		"authToken": "must-not-map-to-exported",
		"friends": ["usr_other"]
	}`

	var got CurrentUserProfile
	if err := json.Unmarshal([]byte(fixture), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ID != "usr_c1644b5b-3ca4-45b4-97c6-a2a0de70d469" {
		t.Errorf("ID = %q", got.ID)
	}
	if got.DisplayName != "Display One" || got.Username != "user_one" {
		t.Errorf("DisplayName/Username = %q / %q", got.DisplayName, got.Username)
	}
	if got.Status != "active" || got.StatusDescription != "In a world" || got.State != "offline" {
		t.Errorf("status fields = %q %q %q", got.Status, got.StatusDescription, got.State)
	}
	if got.CurrentAvatarThumbnailImageURL == "" {
		t.Error("expected avatar thumbnail URL")
	}
}

func TestClient_GetCurrentUser_notAuthenticated(t *testing.T) {
	c := NewClient("")
	_, err := c.GetCurrentUser(context.Background())
	if err != ErrNotAuthenticated {
		t.Fatalf("err = %v, want ErrNotAuthenticated", err)
	}
}

func TestClient_GetCurrentUser_ok(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/1/auth/user" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(CurrentUserProfile{
			ID:          "usr_me",
			DisplayName: "Me",
			Username:    "me_user",
			Status:      "active",
		})
	}))
	t.Cleanup(srv.Close)

	c := NewClient("tok")
	c.apiRoot = srv.URL + "/api/1"
	u, err := c.GetCurrentUser(context.Background())
	if err != nil {
		t.Fatalf("GetCurrentUser: %v", err)
	}
	if u.ID != "usr_me" || u.DisplayName != "Me" {
		t.Fatalf("user: %+v", u)
	}
}

func TestClient_GetCurrentUser_sessionExpired(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	c := NewClient("expired")
	c.apiRoot = srv.URL + "/api/1"
	_, err := c.GetCurrentUser(context.Background())
	if !errors.Is(err, ErrSessionExpired) {
		t.Fatalf("err = %v, want ErrSessionExpired", err)
	}
}

func TestClient_GetCurrentUser_invalidJSON(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	}))
	t.Cleanup(srv.Close)

	c := NewClient("tok")
	c.apiRoot = srv.URL + "/api/1"
	_, err := c.GetCurrentUser(context.Background())
	if err == nil || !strings.Contains(err.Error(), "parse current user") {
		t.Fatalf("err = %v, want parse error", err)
	}
}
