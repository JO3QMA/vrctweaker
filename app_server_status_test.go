package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"vrchat-tweaker/internal/infrastructure/statuspage"
)

func TestToServerStatusDTO_mapsSnapshot(t *testing.T) {
	got := toServerStatusDTO(statuspage.Snapshot{
		FetchState: statuspage.FetchStateOK,
		Summary: statuspage.Summary{
			Indicator:   "maintenance",
			Description: "Service Under Maintenance",
		},
		Components: []statuspage.Component{
			{Name: "Authentication / Login", Status: "under_maintenance"},
		},
		Incidents:    []statuspage.Headline{{Name: "API Degraded"}},
		Maintenances: []statuspage.Headline{{Name: "Database Maintenance"}},
	})
	if got.FetchState != statuspage.FetchStateOK {
		t.Fatalf("FetchState: got %q", got.FetchState)
	}
	if got.Summary.Indicator != "maintenance" {
		t.Fatalf("indicator: got %q", got.Summary.Indicator)
	}
	if len(got.Components) != 1 || got.Components[0].Name != "Authentication / Login" {
		t.Fatalf("components: %+v", got.Components)
	}
	if len(got.Incidents) != 1 || got.Incidents[0].Name != "API Degraded" {
		t.Fatalf("incidents: %+v", got.Incidents)
	}
	if len(got.Maintenances) != 1 {
		t.Fatalf("maintenances: %+v", got.Maintenances)
	}
}

func TestGetServerStatus_returnsDTOWithoutErrorOnFetchFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("down"))
	}))
	defer srv.Close()

	a := &App{
		ctx:          t.Context(),
		serverStatus: statuspage.NewTestClient(srv.URL + "/api/v2/"),
	}
	got, err := a.GetServerStatus()
	if err != nil {
		t.Fatalf("GetServerStatus error: %v", err)
	}
	if got.FetchState != statuspage.FetchStateUnavailable {
		t.Fatalf("FetchState: got %q want %q", got.FetchState, statuspage.FetchStateUnavailable)
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
