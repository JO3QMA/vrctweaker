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
	Components   []Component
	Incidents    []Headline
	Maintenances []Headline
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

// Headline is a one-line incident or maintenance title.
type Headline struct {
	Name string
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
	Name string `json:"name"`
}
