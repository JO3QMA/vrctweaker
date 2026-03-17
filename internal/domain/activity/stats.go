package activity

import (
	"sort"
	"time"
)

// DailyPlaySeconds represents play time for a single date.
type DailyPlaySeconds struct {
	Date   string // YYYY-MM-DD
	Seconds int
}

// TopWorldSummary represents aggregated stats for a world (or total when world_id is absent).
type TopWorldSummary struct {
	WorldID   string
	WorldName string
	Seconds   int
	Sessions  int
}

// ActivityStats holds aggregated activity statistics.
type ActivityStats struct {
	DailyPlaySeconds []DailyPlaySeconds
	TopWorlds        []TopWorldSummary
}

// AggregatePlaySessions computes daily play seconds and top-world-like summary from sessions.
// Pure function for testability. Sessions are assumed to overlap [from, to] (caller fetches accordingly).
// For topWorlds: without world_id, returns a single "_total" entry with total seconds and session count.
func AggregatePlaySessions(sessions []*PlaySession, from, to time.Time) ([]DailyPlaySeconds, []TopWorldSummary) {
	dailyMap := make(map[string]int)
	var totalSeconds int
	sessionCount := len(sessions)

	fromDate := truncateToDate(from)
	toDate := truncateToDate(to)

	for _, s := range sessions {
		start, end := effectiveTimeRange(s)
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

		cur := truncateToDate(start)
		curTime := start
		for !curTime.After(*end) && (cur.Before(toDate) || cur.Equal(toDate)) {
			if cur.Before(fromDate) {
				cur = nextDate(cur)
				curTime = time.Date(cur.Year(), cur.Month(), cur.Day(), 0, 0, 0, 0, time.UTC)
				continue
			}
			dayEnd := truncateToDate(cur).Add(24 * time.Hour) // midnight next day
			if dayEnd.After(*end) {
				dayEnd = *end
			}
			daySec := int(dayEnd.Sub(curTime).Seconds())
			if daySec > 0 {
				dailyMap[cur.Format("2006-01-02")] += daySec
			}
			cur = nextDate(cur)
			curTime = time.Date(cur.Year(), cur.Month(), cur.Day(), 0, 0, 0, 0, time.UTC)
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
	if sessionCount > 0 || totalSeconds > 0 {
		topWorlds = []TopWorldSummary{
			{WorldID: "_total", WorldName: "全セッション", Seconds: totalSeconds, Sessions: sessionCount},
		}
	}

	return daily, topWorlds
}

func truncateToDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func nextDate(t time.Time) time.Time {
	return t.AddDate(0, 0, 1)
}

// effectiveTimeRange returns start and end for a session. Returns nil end if session has no duration.
func effectiveTimeRange(s *PlaySession) (time.Time, *time.Time) {
	start := s.StartTime
	var end *time.Time
	if s.EndTime != nil {
		end = s.EndTime
	} else if s.DurationSec != nil && *s.DurationSec > 0 {
		e := start.Add(time.Duration(*s.DurationSec) * time.Second)
		end = &e
	} else {
		now := time.Now().UTC()
		end = &now
	}
	return start, end
}
