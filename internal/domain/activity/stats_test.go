package activity

import (
	"reflect"
	"testing"
	"time"
)

func mustParse(layout, value string) time.Time {
	t, err := time.ParseInLocation(layout, value, time.UTC)
	if err != nil {
		panic(err)
	}
	return t
}

func TestAggregatePlaySessions(t *testing.T) {
	from := mustParse("2006-01-02", "2024-01-01")
	to := mustParse("2006-01-02", "2024-01-03")
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, time.UTC)

	tests := []struct {
		name      string
		sessions  []*PlaySession
		from      time.Time
		to        time.Time
		wantDaily []DailyPlaySeconds
		wantTop   []TopWorldSummary
	}{
		{
			name:      "empty sessions returns empty arrays",
			sessions:  nil,
			from:      from,
			to:        to,
			wantDaily: []DailyPlaySeconds{},
			wantTop:   nil,
		},
		{
			name: "single session same day",
			sessions: []*PlaySession{
				{
					ID:          "1",
					StartTime:   mustParse(time.RFC3339, "2024-01-02T10:00:00Z"),
					EndTime:     ptrTime(mustParse(time.RFC3339, "2024-01-02T11:00:00Z")),
					DurationSec: ptrInt(3600),
				},
			},
			from: from,
			to:   to,
			wantDaily: []DailyPlaySeconds{
				{Date: "2024-01-02", Seconds: 3600},
			},
			wantTop: []TopWorldSummary{
				{WorldID: "_total", WorldName: "全セッション", Seconds: 3600, Sessions: 1},
			},
		},
		{
			name: "session spanning two days",
			sessions: []*PlaySession{
				{
					ID:          "1",
					StartTime:   mustParse(time.RFC3339, "2024-01-01T22:00:00Z"),
					EndTime:     ptrTime(mustParse(time.RFC3339, "2024-01-02T02:00:00Z")),
					DurationSec: ptrInt(14400), // 4 hours
				},
			},
			from: from,
			to:   to,
			wantDaily: []DailyPlaySeconds{
				{Date: "2024-01-01", Seconds: 7200}, // 22:00-24:00
				{Date: "2024-01-02", Seconds: 7200}, // 00:00-02:00
			},
			wantTop: []TopWorldSummary{
				{WorldID: "_total", WorldName: "全セッション", Seconds: 14400, Sessions: 1},
			},
		},
		{
			name: "multiple sessions same day",
			sessions: []*PlaySession{
				{
					ID:          "1",
					StartTime:   mustParse(time.RFC3339, "2024-01-02T09:00:00Z"),
					EndTime:     ptrTime(mustParse(time.RFC3339, "2024-01-02T10:00:00Z")),
					DurationSec: ptrInt(3600),
				},
				{
					ID:          "2",
					StartTime:   mustParse(time.RFC3339, "2024-01-02T14:00:00Z"),
					EndTime:     ptrTime(mustParse(time.RFC3339, "2024-01-02T15:30:00Z")),
					DurationSec: ptrInt(5400),
				},
			},
			from: from,
			to:   to,
			wantDaily: []DailyPlaySeconds{
				{Date: "2024-01-02", Seconds: 9000}, // 3600 + 5400
			},
			wantTop: []TopWorldSummary{
				{WorldID: "_total", WorldName: "全セッション", Seconds: 9000, Sessions: 2},
			},
		},
		{
			name: "session outside range is excluded by clipping",
			sessions: []*PlaySession{
				{
					ID:          "1",
					StartTime:   mustParse(time.RFC3339, "2024-01-01T22:00:00Z"),
					EndTime:     ptrTime(mustParse(time.RFC3339, "2024-01-01T22:30:00Z")),
					DurationSec: ptrInt(1800),
				},
			},
			from: from,
			to:   to,
			wantDaily: []DailyPlaySeconds{
				{Date: "2024-01-01", Seconds: 1800},
			},
			wantTop: []TopWorldSummary{
				{WorldID: "_total", WorldName: "全セッション", Seconds: 1800, Sessions: 1},
			},
		},
		{
			name: "open session without end is excluded from daily and totals",
			sessions: []*PlaySession{
				{
					ID:          "open1",
					StartTime:   mustParse(time.RFC3339, "2024-01-02T10:00:00Z"),
					EndTime:     nil,
					DurationSec: nil,
				},
			},
			from:      from,
			to:        to,
			wantDaily: []DailyPlaySeconds{},
			wantTop:   nil,
		},
		{
			name: "closed session counted when mixed with open session",
			sessions: []*PlaySession{
				{
					ID:          "open1",
					StartTime:   mustParse(time.RFC3339, "2024-01-02T08:00:00Z"),
					EndTime:     nil,
					DurationSec: nil,
				},
				{
					ID:          "closed1",
					StartTime:   mustParse(time.RFC3339, "2024-01-02T12:00:00Z"),
					EndTime:     ptrTime(mustParse(time.RFC3339, "2024-01-02T13:00:00Z")),
					DurationSec: ptrInt(3600),
				},
			},
			from: from,
			to:   to,
			wantDaily: []DailyPlaySeconds{
				{Date: "2024-01-02", Seconds: 3600},
			},
			wantTop: []TopWorldSummary{
				{WorldID: "_total", WorldName: "全セッション", Seconds: 3600, Sessions: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDaily, gotTop := AggregatePlaySessions(tt.sessions, tt.from, tt.to)
			if !reflect.DeepEqual(gotDaily, tt.wantDaily) {
				t.Errorf("AggregatePlaySessions() dailyPlaySeconds = %v, want %v", gotDaily, tt.wantDaily)
			}
			if !reflect.DeepEqual(gotTop, tt.wantTop) {
				t.Errorf("AggregatePlaySessions() topWorlds = %v, want %v", gotTop, tt.wantTop)
			}
		})
	}
}

func ptrTime(t time.Time) *time.Time { return &t }
func ptrInt(i int) *int              { return &i }
