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
	dailySegments := make(map[string][]playTimeInterval)
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
				dayKey := LocalDateISO(curDay)
				dailySegments[dayKey] = append(dailySegments[dayKey], playTimeInterval{
					start: segStart,
					end:   segEnd,
				})
			}
			if !dayEnd.Before(*end) {
				break
			}
			curDay = dayEnd
		}
	}

	dailyMap := make(map[string]int, len(dailySegments))
	for day, segments := range dailySegments {
		dailyMap[day] = mergePlayTimeIntervalSeconds(segments)
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

type playTimeInterval struct {
	start time.Time
	end   time.Time
}

// mergePlayTimeIntervalSeconds returns the total seconds covered by intervals after merging overlaps.
func mergePlayTimeIntervalSeconds(intervals []playTimeInterval) int {
	if len(intervals) == 0 {
		return 0
	}
	sorted := append([]playTimeInterval(nil), intervals...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].start.Equal(sorted[j].start) {
			return sorted[i].end.Before(sorted[j].end)
		}
		return sorted[i].start.Before(sorted[j].start)
	})

	curStart := sorted[0].start
	curEnd := sorted[0].end
	var total int
	for _, iv := range sorted[1:] {
		if !iv.start.After(curEnd) {
			if iv.end.After(curEnd) {
				curEnd = iv.end
			}
			continue
		}
		total += int(curEnd.Sub(curStart).Seconds())
		curStart, curEnd = iv.start, iv.end
	}
	total += int(curEnd.Sub(curStart).Seconds())
	return total
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
