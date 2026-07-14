package main

import (
	"context"

	"vrchat-tweaker/internal/infrastructure/statuspage"
)

// GetServerStatus returns VRChat service health from status.vrchat.com for the Dashboard.
// Infrastructure failures are expressed via FetchState on the DTO, not via error.
func (a *App) GetServerStatus() (ServerStatusDTO, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	snap := a.serverStatusClient().Fetch(ctx)
	return toServerStatusDTO(snap), nil
}

func (a *App) serverStatusClient() *statuspage.Client {
	if a.serverStatus != nil {
		return a.serverStatus
	}
	return statuspage.NewClient()
}

func toServerStatusDTO(snap statuspage.Snapshot) ServerStatusDTO {
	components := make([]ServerStatusComponentDTO, len(snap.Components))
	for i, c := range snap.Components {
		components[i] = ServerStatusComponentDTO{Name: c.Name, Status: c.Status}
	}
	incidents := make([]ServerStatusHeadlineDTO, len(snap.Incidents))
	for i, h := range snap.Incidents {
		incidents[i] = ServerStatusHeadlineDTO{Name: h.Name}
	}
	maintenances := make([]ServerStatusHeadlineDTO, len(snap.Maintenances))
	for i, h := range snap.Maintenances {
		maintenances[i] = ServerStatusHeadlineDTO{Name: h.Name}
	}
	return ServerStatusDTO{
		FetchState: snap.FetchState,
		Summary: ServerStatusSummaryDTO{
			Indicator:   snap.Summary.Indicator,
			Description: snap.Summary.Description,
		},
		Components:   components,
		Incidents:    incidents,
		Maintenances: maintenances,
	}
}
