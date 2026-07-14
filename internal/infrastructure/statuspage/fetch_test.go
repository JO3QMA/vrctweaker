package statuspage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func testServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	for path, h := range handlers {
		mux.HandleFunc(path, h)
	}
	return httptest.NewServer(mux)
}

func clientForServer(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	return NewTestClient(srv.URL + "/api/v2/")
}

const summaryOperational = `{"status":{"indicator":"none","description":"All Systems Operational"}}`
const summaryMaintenance = `{"status":{"indicator":"maintenance","description":"Service Under Maintenance"}}`

const componentsAllOperational = `{"components":[
{"name":"API / Website","status":"operational","group":true},
{"name":"Authentication / Login","status":"operational","group":false},
{"name":"Realtime Networking","status":"operational","group":true},
{"name":"Japan (Tokyo)","status":"operational","group":false}
]}`

const componentsOneMaintenance = `{"components":[
{"name":"Authentication / Login","status":"under_maintenance","group":false},
{"name":"Japan (Tokyo)","status":"operational","group":false}
]}`

const emptyIncidents = `{"incidents":[]}`
const emptyMaintenances = `{"scheduled_maintenances":[]}`
const oneIncident = `{"incidents":[{"name":"API Degraded"}]}`
const oneMaintenance = `{"scheduled_maintenances":[{"name":"Database Maintenance"}]}`

func jsonHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}
}

func statusHandler(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestFetch_allOperational(t *testing.T) {
	srv := testServer(t, map[string]http.HandlerFunc{
		"/api/v2/summary.json":                         jsonHandler(summaryOperational),
		"/api/v2/components.json":                      jsonHandler(componentsAllOperational),
		"/api/v2/incidents/unresolved.json":            jsonHandler(emptyIncidents),
		"/api/v2/scheduled-maintenances/active.json":   jsonHandler(emptyMaintenances),
		"/api/v2/scheduled-maintenances/upcoming.json": jsonHandler(emptyMaintenances),
	})
	defer srv.Close()

	got := clientForServer(t, srv).Fetch(context.Background())
	if got.FetchState != FetchStateOK {
		t.Fatalf("FetchState: got %q want %q", got.FetchState, FetchStateOK)
	}
	if len(got.Components) != 0 {
		t.Fatalf("components: got %d want 0", len(got.Components))
	}
}

func TestFetch_abnormalOneComponent(t *testing.T) {
	srv := testServer(t, map[string]http.HandlerFunc{
		"/api/v2/summary.json":                       jsonHandler(summaryMaintenance),
		"/api/v2/components.json":                    jsonHandler(componentsOneMaintenance),
		"/api/v2/incidents/unresolved.json":          jsonHandler(oneIncident),
		"/api/v2/scheduled-maintenances/active.json": jsonHandler(oneMaintenance),
	})
	defer srv.Close()

	got := clientForServer(t, srv).Fetch(context.Background())
	if got.FetchState != FetchStateOK {
		t.Fatalf("FetchState: got %q want %q", got.FetchState, FetchStateOK)
	}
	if len(got.Components) != 1 {
		t.Fatalf("components: got %d want 1", len(got.Components))
	}
	if got.Components[0].Name != "Authentication / Login" {
		t.Fatalf("component name: got %q", got.Components[0].Name)
	}
	if got.Incidents[0].Name != "API Degraded" {
		t.Fatalf("incident: got %q", got.Incidents[0].Name)
	}
}

func TestFetch_summaryFailure(t *testing.T) {
	srv := testServer(t, map[string]http.HandlerFunc{
		"/api/v2/summary.json": statusHandler(http.StatusServiceUnavailable, "down"),
	})
	defer srv.Close()

	got := clientForServer(t, srv).Fetch(context.Background())
	if got.FetchState != FetchStateUnavailable {
		t.Fatalf("FetchState: got %q want %q", got.FetchState, FetchStateUnavailable)
	}
}

func TestFetch_partialComponentsFailure(t *testing.T) {
	srv := testServer(t, map[string]http.HandlerFunc{
		"/api/v2/summary.json":                       jsonHandler(summaryOperational),
		"/api/v2/components.json":                    statusHandler(http.StatusInternalServerError, "fail"),
		"/api/v2/incidents/unresolved.json":          jsonHandler(oneIncident),
		"/api/v2/scheduled-maintenances/active.json": jsonHandler(oneMaintenance),
	})
	defer srv.Close()

	got := clientForServer(t, srv).Fetch(context.Background())
	if got.FetchState != FetchStatePartial {
		t.Fatalf("FetchState: got %q want %q", got.FetchState, FetchStatePartial)
	}
	if got.Summary.Indicator != "none" {
		t.Fatalf("summary indicator: got %q", got.Summary.Indicator)
	}
	if len(got.Incidents) != 1 || got.Incidents[0].Name != "API Degraded" {
		t.Fatalf("incidents: %+v", got.Incidents)
	}
	if len(got.Maintenances) != 1 {
		t.Fatalf("maintenances: %+v", got.Maintenances)
	}
}

func TestFetch_invalidJSON(t *testing.T) {
	srv := testServer(t, map[string]http.HandlerFunc{
		"/api/v2/summary.json": jsonHandler("{not-json"),
	})
	defer srv.Close()

	got := clientForServer(t, srv).Fetch(context.Background())
	if got.FetchState != FetchStateUnavailable {
		t.Fatalf("FetchState: got %q want %q", got.FetchState, FetchStateUnavailable)
	}
}

func TestFetch_bodyTooLarge(t *testing.T) {
	huge := `{"status":{"indicator":"none","description":"` + strings.Repeat("x", serverStatusMaxBodyBytes) + `"}}`
	srv := testServer(t, map[string]http.HandlerFunc{
		"/api/v2/summary.json": jsonHandler(huge),
	})
	defer srv.Close()

	got := clientForServer(t, srv).Fetch(context.Background())
	if got.FetchState != FetchStateUnavailable {
		t.Fatalf("FetchState: got %q want %q", got.FetchState, FetchStateUnavailable)
	}
}

func TestValidateHost_rejectsForeignHost(t *testing.T) {
	u, err := url.Parse("https://evil.example/api/v2/summary.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateHost(u); err == nil || !strings.Contains(err.Error(), "host not allowed") {
		t.Fatalf("expected host not allowed, got %v", err)
	}
}

func TestFetch_non200(t *testing.T) {
	srv := testServer(t, map[string]http.HandlerFunc{
		"/api/v2/summary.json": statusHandler(http.StatusServiceUnavailable, "unavailable"),
	})
	defer srv.Close()

	got := clientForServer(t, srv).Fetch(context.Background())
	if got.FetchState != FetchStateUnavailable {
		t.Fatalf("FetchState: got %q want %q", got.FetchState, FetchStateUnavailable)
	}
}

func TestValidateHost_rejectsHTTP(t *testing.T) {
	u, _ := http.NewRequest(http.MethodGet, "http://status.vrchat.com/api/v2/summary.json", nil)
	if err := validateHost(u.URL); err == nil {
		t.Fatal("expected https error")
	}
}

func TestFetch_maintenancesFallbackWhenActiveFails(t *testing.T) {
	srv := testServer(t, map[string]http.HandlerFunc{
		"/api/v2/summary.json":                         jsonHandler(summaryOperational),
		"/api/v2/components.json":                      jsonHandler(componentsAllOperational),
		"/api/v2/incidents/unresolved.json":            jsonHandler(emptyIncidents),
		"/api/v2/scheduled-maintenances/active.json":   statusHandler(http.StatusInternalServerError, "fail"),
		"/api/v2/scheduled-maintenances/upcoming.json": jsonHandler(oneMaintenance),
	})
	defer srv.Close()

	got := clientForServer(t, srv).Fetch(context.Background())
	if got.FetchState != FetchStateOK {
		t.Fatalf("FetchState: got %q want %q", got.FetchState, FetchStateOK)
	}
	if len(got.Maintenances) != 1 || got.Maintenances[0].Name != "Database Maintenance" {
		t.Fatalf("maintenances: %+v", got.Maintenances)
	}
}

func TestTruncateForErr_respectsRuneBoundary(t *testing.T) {
	got := truncateForErr("日本語テスト", 3)
	if got != "日本語…" {
		t.Fatalf("got %q", got)
	}
}

func TestFetchJSON_rejectsAbsolutePath(t *testing.T) {
	c := NewClient()
	err := c.fetchJSON(context.Background(), "/summary.json", &summaryResponse{})
	if err == nil || !strings.Contains(err.Error(), "path must be relative") {
		t.Fatalf("expected relative path error, got %v", err)
	}
}

func TestHeadlinesFromIncidents_returnsAllNonEmpty(t *testing.T) {
	got := headlinesFromIncidents([]apiIncident{
		{Name: "First"},
		{Name: ""},
		{Name: "Second"},
	})
	if len(got) != 2 || got[0].Name != "First" || got[1].Name != "Second" {
		t.Fatalf("got %+v", got)
	}
}
