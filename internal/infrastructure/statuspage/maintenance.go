package statuspage

import (
	"context"
	"sort"
	"strings"
	"time"
)

func (c *Client) loadMaintenances(ctx context.Context) ([]Maintenance, error) {
	var active, upcoming maintenancesResponse
	activeErr := c.fetchJSON(ctx, "scheduled-maintenances/active.json", &active)
	upcomingErr := c.fetchJSON(ctx, "scheduled-maintenances/upcoming.json", &upcoming)
	switch {
	case activeErr != nil && upcomingErr != nil:
		return nil, activeErr
	case activeErr != nil:
		return maintenancesFromAPI(upcoming.ScheduledMaintenances), nil
	case upcomingErr != nil:
		return maintenancesFromAPI(active.ScheduledMaintenances), nil
	default:
		merged := make([]apiMaintenance, 0, len(active.ScheduledMaintenances)+len(upcoming.ScheduledMaintenances))
		merged = append(merged, active.ScheduledMaintenances...)
		merged = append(merged, upcoming.ScheduledMaintenances...)
		return maintenancesFromAPI(merged), nil
	}
}

func maintenancesFromAPI(list []apiMaintenance) []Maintenance {
	out := make([]Maintenance, 0, len(list))
	index := make(map[string]int, len(list))
	for _, raw := range list {
		m, ok := maintenanceFromAPI(raw)
		if !ok {
			continue
		}
		key := maintenanceKey(m)
		if i, exists := index[key]; exists {
			if statusPreference(m.Status) < statusPreference(out[i].Status) {
				out[i] = m
			}
			continue
		}
		index[key] = len(out)
		out = append(out, m)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return maintenanceLess(out[i], out[j])
	})
	return out
}

func maintenanceFromAPI(raw apiMaintenance) (Maintenance, bool) {
	name := strings.TrimSpace(raw.Name)
	if name == "" || maintenanceFinished(raw.Status) {
		return Maintenance{}, false
	}
	return Maintenance{
		ID:             strings.TrimSpace(raw.ID),
		Name:           name,
		Status:         normalizeMaintenanceStatus(raw.Status),
		ScheduledFor:   strings.TrimSpace(raw.ScheduledFor),
		ScheduledUntil: strings.TrimSpace(raw.ScheduledUntil),
		Updates:        maintenanceUpdates(raw.IncidentUpdates),
	}, true
}

func maintenanceUpdates(list []apiIncidentUpdate) []string {
	if len(list) == 0 {
		return nil
	}
	out := make([]string, 0, len(list))
	for _, u := range list {
		body := strings.TrimSpace(u.Body)
		if body == "" {
			continue
		}
		out = append(out, body)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func maintenanceKey(m Maintenance) string {
	if m.ID != "" {
		return "id:" + m.ID
	}
	return "name:" + m.Name + "|" + m.ScheduledFor
}

func normalizeMaintenanceStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

func maintenanceFinished(status string) bool {
	switch normalizeMaintenanceStatus(status) {
	case "completed", "resolved":
		return true
	default:
		return false
	}
}

// statusPreference ranks how current a duplicate record is. Lower wins.
func statusPreference(status string) int {
	switch normalizeMaintenanceStatus(status) {
	case "in_progress":
		return 0
	case "verifying":
		return 1
	case "scheduled":
		return 2
	default:
		return 3
	}
}

func maintenanceDisplayRank(status string) int {
	switch normalizeMaintenanceStatus(status) {
	case "in_progress", "verifying":
		return 0
	default:
		return 1
	}
}

func maintenanceLess(a, b Maintenance) bool {
	ra := maintenanceDisplayRank(a.Status)
	rb := maintenanceDisplayRank(b.Status)
	if ra != rb {
		return ra < rb
	}
	ta, oka := parseMaintenanceTime(a.ScheduledFor)
	tb, okb := parseMaintenanceTime(b.ScheduledFor)
	if oka != okb {
		return oka
	}
	if !oka {
		return false
	}
	return ta.Before(tb)
}

func parseMaintenanceTime(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
