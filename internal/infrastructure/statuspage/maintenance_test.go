package statuspage

import (
	"context"
	"net/http"
	"testing"
)

func TestMaintenancesFromAPI_ordersInProgressBeforeSoonerScheduled(t *testing.T) {
	got := maintenancesFromAPI([]apiMaintenance{
		{
			ID:           "later",
			Name:         "Later",
			Status:       "scheduled",
			ScheduledFor: "2026-12-25T00:00:00Z",
		},
		{
			ID:           "done",
			Name:         "Done",
			Status:       "Completed",
			ScheduledFor: "2026-12-01T00:00:00Z",
		},
		{
			ID:             "soon",
			Name:           "Soon",
			Status:         "scheduled",
			ScheduledFor:   "2026-12-24T01:00:00.000Z",
			ScheduledUntil: "2026-12-24T02:00:00.000Z",
			IncidentUpdates: []apiIncidentUpdate{
				{Body: "  Planned restart  "},
				{Body: "   "},
			},
		},
		{
			ID:             "now",
			Name:           " Now ",
			Status:         " verifying ",
			ScheduledFor:   "2026-12-23T03:00:00Z",
			ScheduledUntil: "2026-12-23T06:00:00Z",
			IncidentUpdates: []apiIncidentUpdate{
				{Body: "Checking"},
			},
		},
		{Name: "   ", Status: "scheduled"},
	})
	if len(got) != 3 {
		t.Fatalf("len: got %d %+v", len(got), got)
	}
	if got[0].Name != "Now" || got[0].Status != "verifying" || len(got[0].Updates) != 1 || got[0].Updates[0] != "Checking" {
		t.Fatalf("first: %+v", got[0])
	}
	if got[1].Name != "Soon" || got[1].ScheduledUntil != "2026-12-24T02:00:00.000Z" || len(got[1].Updates) != 1 {
		t.Fatalf("second: %+v", got[1])
	}
	if got[2].Name != "Later" {
		t.Fatalf("third: %+v", got[2])
	}
}

func TestMaintenancesFromAPI_prefersInProgressDuplicate(t *testing.T) {
	got := maintenancesFromAPI([]apiMaintenance{
		{
			ID:           "api",
			Name:         "API",
			Status:       "scheduled",
			ScheduledFor: "2026-12-23T03:34:00Z",
		},
		{
			ID:             "api",
			Name:           "API",
			Status:         "in_progress",
			ScheduledFor:   "2026-12-23T03:34:00Z",
			ScheduledUntil: "2026-12-23T14:45:00Z",
			IncidentUpdates: []apiIncidentUpdate{
				{Body: "In progress"},
			},
		},
	})
	if len(got) != 1 {
		t.Fatalf("len: got %d", len(got))
	}
	if got[0].Status != "in_progress" || len(got[0].Updates) != 1 || got[0].Updates[0] != "In progress" {
		t.Fatalf("got %+v", got[0])
	}

	reversed := maintenancesFromAPI([]apiMaintenance{
		{ID: "api", Name: "API", Status: "in_progress", ScheduledFor: "2026-12-23T03:34:00Z"},
		{ID: "api", Name: "API", Status: "scheduled", ScheduledFor: "2026-12-23T03:34:00Z"},
	})
	if len(reversed) != 1 || reversed[0].Status != "in_progress" {
		t.Fatalf("reversed: %+v", reversed)
	}
}

func TestFetch_mergesActiveAndUpcomingMaintenances(t *testing.T) {
	const activeBody = `{"scheduled_maintenances":[
		{"id":"live","name":"API","status":"in_progress","scheduled_for":"2026-12-23T03:34:00Z","scheduled_until":"2026-12-23T14:45:00Z","incident_updates":[{"body":"Working"}]},
		{"id":"old","name":"Old","status":"completed","scheduled_for":"2026-12-01T00:00:00Z","scheduled_until":"2026-12-01T01:00:00Z"}
	]}`
	const upcomingBody = `{"scheduled_maintenances":[
		{"id":"live","name":"API","status":"scheduled","scheduled_for":"2026-12-23T03:34:00Z","scheduled_until":"2026-12-23T14:45:00Z"},
		{"id":"next","name":"Database","status":"scheduled","scheduled_for":"2026-12-24T01:00:00Z","scheduled_until":"2026-12-24T04:00:00Z","incident_updates":[{"body":"Overnight"}]}
	]}`
	srv := testServer(t, map[string]http.HandlerFunc{
		"/api/v2/summary.json":                         jsonHandler(summaryOperational),
		"/api/v2/components.json":                      jsonHandler(componentsAllOperational),
		"/api/v2/incidents/unresolved.json":            jsonHandler(emptyIncidents),
		"/api/v2/scheduled-maintenances/active.json":   jsonHandler(activeBody),
		"/api/v2/scheduled-maintenances/upcoming.json": jsonHandler(upcomingBody),
	})
	defer srv.Close()

	got := clientForServer(t, srv).Fetch(context.Background())
	if got.FetchState != FetchStateOK {
		t.Fatalf("FetchState: got %q", got.FetchState)
	}
	if len(got.Maintenances) != 2 {
		t.Fatalf("maintenances: %+v", got.Maintenances)
	}
	if got.Maintenances[0].Name != "API" || got.Maintenances[0].Status != "in_progress" ||
		got.Maintenances[0].ScheduledUntil != "2026-12-23T14:45:00Z" ||
		len(got.Maintenances[0].Updates) != 1 || got.Maintenances[0].Updates[0] != "Working" {
		t.Fatalf("active: %+v", got.Maintenances[0])
	}
	if got.Maintenances[1].Name != "Database" || got.Maintenances[1].Status != "scheduled" ||
		len(got.Maintenances[1].Updates) != 1 || got.Maintenances[1].Updates[0] != "Overnight" {
		t.Fatalf("upcoming: %+v", got.Maintenances[1])
	}
}
