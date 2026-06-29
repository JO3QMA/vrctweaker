package activity

import (
	"sort"
	"time"
)

// DailyPlaySeconds represents play time for a single date.
type DailyPlaySeconds struct {
	Date    string `json:"date"` // YYYY-MM-DD (local calendar)
	Seconds int    `json:"seconds"`
}

// TopWorldSummary represents aggregated stats for a world (or total when world_id is absent).
type TopWorldSummary struct {
	WorldID   string `json:"worldId"`
	WorldName string `json:"worldName,omitempty"`
	Seconds   int    `json:"seconds"`
	Sessions  int    `json:"sessions"`
}

// ActivityStats holds aggregated activity statistics.
type ActivityStats struct {
	DailyPlaySeconds []DailyPlaySeconds `json:"dailyPlaySeconds"`
	TopWorlds        []TopWorldSummary  `json:"topWorlds"`
}

// AggregatePlaySessions computes daily play seconds and top-world-like summary from sessions.
// Pure function for testability. Sessions are assumed to overlap [from, to] (caller fetches accordingly).
// lastObserved extends open sessions (no end time) through that instant for daily totals.
// For topWorlds: without world_id, returns a single "_total" entry with total seconds and session count.
func AggregatePlaySessions(sessions []*PlaySession, from, to time.Time, lastObserved *time.Time) ([]DailyPlaySeconds, []TopWorldSummary) {
	dailyMap := make(map[string]int)
	var totalSeconds int
	var includedSessions int

	fromDate := StartOfLocalCalendarDay(from)
	toDate := StartOfLocalCalendarDay(to)

	for _, s := range sessions {
		start, end := effectiveTimeRange(s, lastObserved)
		if end == nil {
			continue
		}
		if !start.Before(*end) {
			continue
		}

		// Clip to [from, to] range
		if start.Before(from) {
			start = from
		}
		endVal := *end
		if endVal.After(to) {
			endVal = to
			end = &endVal
		}
		if !start.Before(*end) {
			continue
		}
		totalSeconds += int(end.Sub(start).Seconds())
		includedSessions++

		curDay := StartOfLocalCalendarDay(start)
		for !curDay.After(toDate) {
			dayEnd := StartOfNextLocalCalendarDay(curDay)
			segStart := start
			if curDay.After(segStart) {
				segStart = curDay
			}
			segEnd := *end
			if dayEnd.Before(segEnd) {
				segEnd = dayEnd
			}
			if !curDay.Before(fromDate) && segStart.Before(segEnd) {
				daySec := int(segEnd.Sub(segStart).Seconds())
				if daySec > 0 {
					dailyMap[LocalDateISO(curDay)] += daySec
				}
			}
			if !dayEnd.Before(*end) {
				break
			}
			curDay = dayEnd
		}
	}

	// Build daily array in date order
	var dates []string
	for d := range dailyMap {
		dates = append(dates, d)
	}
	sort.Strings(dates)
	daily := make([]DailyPlaySeconds, 0, len(dates))
	for _, d := range dates {
		daily = append(daily, DailyPlaySeconds{Date: d, Seconds: dailyMap[d]})
	}

	// topWorlds: single aggregate (world_id不足のため)
	var topWorlds []TopWorldSummary
	if includedSessions > 0 || totalSeconds > 0 {
		topWorlds = []TopWorldSummary{
			{WorldID: "_total", WorldName: "全セッション", Seconds: totalSeconds, Sessions: includedSessions},
		}
	}

	return daily, topWorlds
}

// effectiveTimeRange returns start and end for aggregation. Open sessions use lastObserved when set.
func effectiveTimeRange(s *PlaySession, lastObserved *time.Time) (time.Time, *time.Time) {
	start := s.StartTime
	if s.EndTime != nil {
		return start, s.EndTime
	}
	if s.DurationSec != nil && *s.DurationSec > 0 {
		e := start.Add(time.Duration(*s.DurationSec) * time.Second)
		return start, &e
	}
	if lastObserved != nil && !lastObserved.Before(start) {
		end := *lastObserved
		return start, &end
	}
	return start, nil
}
