package wailsapp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"vrchat-tweaker/internal/infrastructure/statuspage"
)

func TestGetServerStatus_returnsDTOWithoutErrorOnFetchFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("down"))
	}))
	defer srv.Close()

	a := &App{
		serverStatus: statuspage.NewTestClient(srv.URL + "/api/v2/"),
	}
	got, err := a.GetServerStatus()
	if err != nil {
		t.Fatalf("GetServerStatus error: %v", err)
	}
	if got.FetchState != statuspage.FetchStateUnavailable {
		t.Fatalf("FetchState: got %q want %q", got.FetchState, statuspage.FetchStateUnavailable)
	}
	b, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	const want = `{"fetchState":"unavailable","summary":{"indicator":"","description":""},"components":[],"incidents":[],"maintenances":[]}`
	if string(b) != want {
		t.Fatalf("json=%s", b)
	}
}

func TestServerStatusClient_cachesDefaultClient(t *testing.T) {
	a := &App{}
	first := a.serverStatusClient()
	second := a.serverStatusClient()
	if first != second {
		t.Fatal("expected same cached client instance")
	}
	if a.serverStatus == nil {
		t.Fatal("expected serverStatus field to be set")
	}
}
