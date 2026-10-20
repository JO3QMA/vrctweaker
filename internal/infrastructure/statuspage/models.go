package statuspage

// FetchState values returned to the frontend on Snapshot.
const (
	FetchStateOK          = "ok"
	FetchStateUnavailable = "unavailable"
	FetchStatePartial     = "partial"
)

const componentStatusOperational = "operational"

// Snapshot is the aggregated server status for Dashboard display.
// JSON names match the former Wails ServerStatusDTO.
type Snapshot struct {
	FetchState   string        `json:"fetchState"`
	Summary      Summary       `json:"summary"`
	Components   []Component   `json:"components"` // non-operational leaf components only (see filterNonOperational)
	Incidents    []Headline    `json:"incidents"`
	Maintenances []Maintenance `json:"maintenances"`
}

// Summary is the overall status indicator from summary.json.
type Summary struct {
	Indicator   string `json:"indicator"`
	Description string `json:"description"`
}

// Component is a non-group statuspage component (leaf).
type Component struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Headline is a one-line incident title.
type Headline struct {
	Name string `json:"name"`
}

// Maintenance is a scheduled or active statuspage maintenance window.
// Completed windows are omitted before this value is returned.
// Optional JSON names match the former ServerStatusMaintenanceDTO.
type Maintenance struct {
	ID             string   `json:"id,omitempty"`
	Name           string   `json:"name"`
	Status         string   `json:"status,omitempty"`
	ScheduledFor   string   `json:"scheduledFor,omitempty"`
	ScheduledUntil string   `json:"scheduledUntil,omitempty"`
	Updates        []string `json:"updates,omitempty"`
}

type summaryResponse struct {
	Status struct {
		Indicator   string `json:"indicator"`
		Description string `json:"description"`
	} `json:"status"`
}

type componentsResponse struct {
	Components []apiComponent `json:"components"`
}

type apiComponent struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Group  bool   `json:"group"`
}

type incidentsResponse struct {
	Incidents []apiIncident `json:"incidents"`
}

type apiIncident struct {
	Name string `json:"name"`
}

type maintenancesResponse struct {
	ScheduledMaintenances []apiMaintenance `json:"scheduled_maintenances"`
}

type apiMaintenance struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Status          string              `json:"status"`
	ScheduledFor    string              `json:"scheduled_for"`
	ScheduledUntil  string              `json:"scheduled_until"`
	IncidentUpdates []apiIncidentUpdate `json:"incident_updates"`
}

type apiIncidentUpdate struct {
	Body string `json:"body"`
}
