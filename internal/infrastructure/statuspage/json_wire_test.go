package statuspage

import (
	"encoding/json"
	"testing"
)

func TestSnapshotJSON_matchesFormerServerStatusDTO(t *testing.T) {
	snap := Snapshot{
		FetchState: FetchStateOK,
		Summary: Summary{
			Indicator:   "maintenance",
			Description: "Service Under Maintenance",
		},
		Components: []Component{
			{Name: "Authentication / Login", Status: "under_maintenance"},
		},
		Incidents: []Headline{{Name: "API Degraded"}},
		Maintenances: []Maintenance{{
			ID:             "mnt-1",
			Name:           "Database Maintenance",
			Status:         "scheduled",
			ScheduledFor:   "2026-12-23T03:34:00Z",
			ScheduledUntil: "2026-12-23T14:45:00Z",
			Updates:        []string{"Database window"},
		}},
	}
	b, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	const want = `{"fetchState":"ok","summary":{"indicator":"maintenance","description":"Service Under Maintenance"},"components":[{"name":"Authentication / Login","status":"under_maintenance"}],"incidents":[{"name":"API Degraded"}],"maintenances":[{"id":"mnt-1","name":"Database Maintenance","status":"scheduled","scheduledFor":"2026-12-23T03:34:00Z","scheduledUntil":"2026-12-23T14:45:00Z","updates":["Database window"]}]}`
	if string(b) != want {
		t.Fatalf("json=%s", b)
	}
}

func TestMaintenanceJSON_omitsEmptyOptionalFields(t *testing.T) {
	b, err := json.Marshal(Maintenance{Name: "Window"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(b) != `{"name":"Window"}` {
		t.Fatalf("json=%s", b)
	}
}
