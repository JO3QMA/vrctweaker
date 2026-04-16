package vrchatapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testStatusUserID = "usr_status_test"

func TestClient_SetUserStatus_requestBody(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/1/users/"+testStatusUserID {
			http.NotFound(w, r)
			return
		}
		b, _ := io.ReadAll(r.Body)
		var m map[string]string
		if err := json.Unmarshal(b, &m); err != nil {
			t.Errorf("body json: %v", err)
		}
		if m["status"] != "busy" {
			t.Errorf("status = %q, want busy", m["status"])
		}
		if _, ok := m["statusDescription"]; ok {
			t.Errorf("unexpected statusDescription in body: %v", m)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	c.SetAuthToken("tok")
	if err := c.SetUserStatus(context.Background(), testStatusUserID, StatusBusy); err != nil {
		t.Fatalf("SetUserStatus: %v", err)
	}
}

func TestClient_SetUserStatusDescription_requestBody(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/1/users/"+testStatusUserID {
			http.NotFound(w, r)
			return
		}
		b, _ := io.ReadAll(r.Body)
		var m map[string]string
		if err := json.Unmarshal(b, &m); err != nil {
			t.Errorf("body json: %v", err)
		}
		if m["statusDescription"] != "working" {
			t.Errorf("statusDescription = %q", m["statusDescription"])
		}
		if _, ok := m["status"]; ok {
			t.Errorf("unexpected status in body: %v", m)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	c.SetAuthToken("tok")
	if err := c.SetUserStatusDescription(context.Background(), testStatusUserID, "working"); err != nil {
		t.Fatalf("SetUserStatusDescription: %v", err)
	}
}

func TestClient_SetUserStatusAndDescription_requestBody(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/1/users/"+testStatusUserID {
			http.NotFound(w, r)
			return
		}
		b, _ := io.ReadAll(r.Body)
		var m map[string]string
		if err := json.Unmarshal(b, &m); err != nil {
			t.Errorf("body json: %v", err)
		}
		if m["status"] != "join me" {
			t.Errorf("status = %q, want join me", m["status"])
		}
		if m["statusDescription"] != "イベント中" {
			t.Errorf("statusDescription = %q", m["statusDescription"])
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := NewClient("")
	c.apiRoot = srv.URL + "/api/1"
	c.SetAuthToken("tok")
	if err := c.SetUserStatusAndDescription(context.Background(), testStatusUserID, StatusJoinMe, "イベント中"); err != nil {
		t.Fatalf("SetUserStatusAndDescription: %v", err)
	}
}
