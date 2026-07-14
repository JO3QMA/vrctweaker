package statuspage

import (
	"context"

	"golang.org/x/sync/errgroup"
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

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		componentsErr = c.fetchJSON(gctx, "components.json", &componentsResp)
		return nil
	})
	g.Go(func() error {
		incidentsErr = c.fetchJSON(gctx, "incidents/unresolved.json", &incidentsResp)
		return nil
	})
	g.Go(func() error {
		maintenancesErr = c.fetchJSON(gctx, "scheduled-maintenances/active.json", &maintenancesResp)
		if maintenancesErr == nil && len(maintenancesResp.ScheduledMaintenances) == 0 {
			maintenancesErr = c.fetchJSON(gctx, "scheduled-maintenances/upcoming.json", &maintenancesResp)
		}
		return nil
	})
	_ = g.Wait()

	if componentsErr != nil {
		return Snapshot{
			FetchState: FetchStatePartial,
			Summary:    summary,
		}
	}

	leaves := leafComponents(componentsResp.Components)
	nonOperational := filterNonOperational(leaves)

	if len(nonOperational) > 0 && (incidentsErr != nil || maintenancesErr != nil) {
		return Snapshot{
			FetchState: FetchStatePartial,
			Summary:    summary,
		}
	}

	incidents := headlinesFromIncidents(incidentsResp.Incidents)
	maintenances := headlinesFromMaintenances(maintenancesResp.ScheduledMaintenances)

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
	if len(list) == 0 {
		return nil
	}
	return []Headline{{Name: list[0].Name}}
}

func headlinesFromMaintenances(list []apiMaintenance) []Headline {
	out := make([]Headline, 0, len(list))
	for _, m := range list {
		name := m.Name
		if name == "" {
			continue
		}
		out = append(out, Headline{Name: name})
	}
	return out
}
