package statuspage

// FetchState values returned to the frontend via ServerStatusDTO.
const (
	FetchStateOK          = "ok"
	FetchStateUnavailable = "unavailable"
	FetchStatePartial     = "partial"
)

const componentStatusOperational = "operational"

// Snapshot is the aggregated server status for Dashboard display.
type Snapshot struct {
	FetchState   string
	Summary      Summary
	Components   []Component // non-operational leaf components only (see filterNonOperational)
	Incidents    []Headline
	Maintenances []Maintenance
}

// Summary is the overall status indicator from summary.json.
type Summary struct {
	Indicator   string
	Description string
}

// Component is a non-group statuspage component (leaf).
type Component struct {
	Name   string
	Status string
}

// Headline is a one-line incident title.
type Headline struct {
	Name string
}

// Maintenance is a scheduled or active statuspage maintenance window.
// Completed windows are omitted before this value is returned.
type Maintenance struct {
	ID             string
	Name           string
	Status         string
	ScheduledFor   string
	ScheduledUntil string
	Updates        []string
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
