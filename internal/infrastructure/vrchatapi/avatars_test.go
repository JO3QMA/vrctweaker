package vrchatapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetAvatar_ok(t *testing.T) {
	t.Parallel()
	const avatarID = "avtr_11111111-2222-3333-4444-555555555555"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/1/avatars/"+avatarID {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"id":   avatarID,
			"name": "Test Avatar",
		})
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	c.SetAuthToken("x")

	av, err := c.GetAvatar(context.Background(), avatarID)
	if err != nil {
		t.Fatalf("GetAvatar: %v", err)
	}
	if av.ID != avatarID || av.Name != "Test Avatar" {
		t.Fatalf("avatar: %+v", av)
	}
}

func TestClient_GetAvatar_emptyIDInResponse(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"name": "No ID",
		})
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	c.SetAuthToken("x")

	_, err := c.GetAvatar(context.Background(), "avtr_11111111-2222-3333-4444-555555555555")
	if err == nil {
		t.Fatal("expected error for empty id in response")
	}
}

func TestClient_GetAvatar_pathEscape(t *testing.T) {
	t.Parallel()
	const avatarID = "avtr_11111111-2222-3333-4444-555555555555"
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]string{
			"id":   avatarID,
			"name": "Test",
		})
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	c.SetAuthToken("x")

	_, err := c.GetAvatar(context.Background(), avatarID)
	if err != nil {
		t.Fatalf("GetAvatar: %v", err)
	}
	if gotPath != "/api/1/avatars/"+avatarID {
		t.Fatalf("path %q", gotPath)
	}
}
