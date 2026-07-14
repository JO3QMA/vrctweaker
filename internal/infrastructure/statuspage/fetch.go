package statuspage

import (
	"context"
	"sync"
)

// Fetch loads status.vrchat.com data for Dashboard Server status.
func (c *Client) Fetch(ctx context.Context) Snapshot {
	var summaryResp summaryResponse
	if err := c.fetchJSON(ctx, "summary.json", &summaryResp); err != nil {
		return Snapshot{FetchState: FetchStateUnavailable}
	}

	summary := Summary{
		Indicator:   summaryResp.Status.Indicator,
		Description: summaryResp.Status.Description,
	}

	var (
		componentsResp   componentsResponse
		incidentsResp    incidentsResponse
		maintenancesResp maintenancesResponse
		componentsErr    error
		incidentsErr     error
		maintenancesErr  error
	)

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		componentsErr = c.fetchJSON(ctx, "components.json", &componentsResp)
	}()
	go func() {
		defer wg.Done()
		incidentsErr = c.fetchJSON(ctx, "incidents/unresolved.json", &incidentsResp)
	}()
	go func() {
		defer wg.Done()
		maintenancesErr = c.fetchJSON(ctx, "scheduled-maintenances/active.json", &maintenancesResp)
		if maintenancesErr == nil && len(maintenancesResp.ScheduledMaintenances) == 0 {
			maintenancesErr = c.fetchJSON(ctx, "scheduled-maintenances/upcoming.json", &maintenancesResp)
		}
	}()
	wg.Wait()

	incidents := headlinesFromIncidents(incidentsResp.Incidents)
	maintenances := headlinesFromMaintenances(maintenancesResp.ScheduledMaintenances)

	if componentsErr != nil {
		return Snapshot{
			FetchState:   FetchStatePartial,
			Summary:      summary,
			Incidents:    incidents,
			Maintenances: maintenances,
		}
	}

	leaves := leafComponents(componentsResp.Components)
	nonOperational := filterNonOperational(leaves)

	if len(nonOperational) > 0 && (incidentsErr != nil || maintenancesErr != nil) {
		return Snapshot{
			FetchState:   FetchStatePartial,
			Summary:      summary,
			Components:   nonOperational,
			Incidents:    incidents,
			Maintenances: maintenances,
		}
	}

	return Snapshot{
		FetchState:   FetchStateOK,
		Summary:      summary,
		Components:   nonOperational,
		Incidents:    incidents,
		Maintenances: maintenances,
	}
}

func leafComponents(list []apiComponent) []apiComponent {
	out := make([]apiComponent, 0, len(list))
	for _, c := range list {
		if !c.Group {
			out = append(out, c)
		}
	}
	return out
}

func filterNonOperational(leaves []apiComponent) []Component {
	out := make([]Component, 0)
	for _, c := range leaves {
		if c.Status == componentStatusOperational {
			continue
		}
		out = append(out, Component{Name: c.Name, Status: c.Status})
	}
	return out
}

func headlinesFromIncidents(list []apiIncident) []Headline {
	out := make([]Headline, 0, len(list))
	for _, inc := range list {
		if inc.Name == "" {
			continue
		}
		out = append(out, Headline(inc))
	}
	return out
}

func headlinesFromMaintenances(list []apiMaintenance) []Headline {
	out := make([]Headline, 0, len(list))
	for _, m := range list {
		if m.Name == "" {
			continue
		}
		out = append(out, Headline(m))
	}
	return out
}
