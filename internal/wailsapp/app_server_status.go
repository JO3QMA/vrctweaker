package wailsapp

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/infrastructure/statuspage"
)

// GetServerStatus returns VRChat service health from status.vrchat.com for the Dashboard.
// Infrastructure failures are expressed via FetchState on the snapshot, not via error.
func (a *App) GetServerStatus() (statuspage.Snapshot, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	snap := a.serverStatusClient().Fetch(ctx)
	if snap.FetchState != statuspage.FetchStateOK && a.ctx != nil {
		runtime.LogWarning(a.ctx, "server status fetch: state="+snap.FetchState)
	}
	return normalizeServerStatus(snap), nil
}

func (a *App) serverStatusClient() *statuspage.Client {
	if a.serverStatus != nil {
		return a.serverStatus
	}
	a.serverStatus = statuspage.NewClient()
	return a.serverStatus
}

// normalizeServerStatus keeps empty collections as [] rather than null.
func normalizeServerStatus(snap statuspage.Snapshot) statuspage.Snapshot {
	if snap.Components == nil {
		snap.Components = []statuspage.Component{}
	}
	if snap.Incidents == nil {
		snap.Incidents = []statuspage.Headline{}
	}
	if snap.Maintenances == nil {
		snap.Maintenances = []statuspage.Maintenance{}
	}
	return snap
}
