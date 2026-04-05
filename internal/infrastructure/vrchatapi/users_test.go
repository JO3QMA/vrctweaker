package vrchatapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetUser_ok(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/1/users/usr_abc" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(Friend{
			ID:          "usr_abc",
			DisplayName: "TestUser",
			IsFriend:    true,
			Status:      "active",
		})
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	c.SetAuthToken("x")

	u, err := c.GetUser(context.Background(), "usr_abc")
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if u.ID != "usr_abc" || u.DisplayName != "TestUser" || !u.IsFriend {
		t.Fatalf("user: %+v", u)
	}
}

func TestClient_GetUser_emptyID(t *testing.T) {
	t.Parallel()
	c := NewClient("")
	_, err := c.GetUser(context.Background(), "")
	if err == nil {
		t.Fatal("want error for empty id")
	}
}
