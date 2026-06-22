package activity

import (
	"reflect"
	"testing"
	"time"
)

func localTime(y, m, d, h, min, sec int) time.Time {
	return time.Date(y, time.Month(m), d, h, min, sec, 0, time.Local)
}

func TestAggregatePlaySessions(t *testing.T) {
	from := StartOfLocalCalendarDay(localTime(2024, 1, 1, 0, 0, 0))
	to := StartOfNextLocalCalendarDay(
		StartOfLocalCalendarDay(localTime(2024, 1, 3, 0, 0, 0)),
	)

	tests := []struct {
		name         string
		sessions     []*PlaySession
		from         time.Time
		to           time.Time
		lastObserved *time.Time
		wantDaily    []DailyPlaySeconds
		wantTop      []TopWorldSummary
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
					StartTime:   localTime(2024, 1, 2, 10, 0, 0),
					EndTime:     ptrTime(localTime(2024, 1, 2, 11, 0, 0)),
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
			name: "session spanning two local days",
			sessions: []*PlaySession{
				{
					ID:          "1",
					StartTime:   localTime(2024, 1, 1, 22, 0, 0),
					EndTime:     ptrTime(localTime(2024, 1, 2, 2, 0, 0)),
					DurationSec: ptrInt(14400),
				},
			},
			from: from,
			to:   to,
			wantDaily: []DailyPlaySeconds{
				{Date: "2024-01-01", Seconds: 7200},
				{Date: "2024-01-02", Seconds: 7200},
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
					StartTime:   localTime(2024, 1, 2, 9, 0, 0),
					EndTime:     ptrTime(localTime(2024, 1, 2, 10, 0, 0)),
					DurationSec: ptrInt(3600),
				},
				{
					ID:          "2",
					StartTime:   localTime(2024, 1, 2, 14, 0, 0),
					EndTime:     ptrTime(localTime(2024, 1, 2, 15, 30, 0)),
					DurationSec: ptrInt(5400),
				},
			},
			from: from,
			to:   to,
			wantDaily: []DailyPlaySeconds{
				{Date: "2024-01-02", Seconds: 9000},
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
					StartTime:   localTime(2024, 1, 1, 22, 0, 0),
					EndTime:     ptrTime(localTime(2024, 1, 1, 22, 30, 0)),
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
			name: "open session without lastObserved is excluded",
			sessions: []*PlaySession{
				{
					ID:        "open1",
					StartTime: localTime(2024, 1, 2, 10, 0, 0),
				},
			},
			from:      from,
			to:        to,
			wantDaily: []DailyPlaySeconds{},
			wantTop:   nil,
		},
		{
			name: "open session with lastObserved is included",
			sessions: []*PlaySession{
				{
					ID:        "open1",
					StartTime: localTime(2024, 1, 2, 10, 0, 0),
				},
			},
			from:         from,
			to:           to,
			lastObserved: ptrTime(localTime(2024, 1, 2, 11, 0, 0)),
			wantDaily: []DailyPlaySeconds{
				{Date: "2024-01-02", Seconds: 3600},
			},
			wantTop: []TopWorldSummary{
				{WorldID: "_total", WorldName: "全セッション", Seconds: 3600, Sessions: 1},
			},
		},
		{
			name: "closed session counted when mixed with open session",
			sessions: []*PlaySession{
				{
					ID:        "open1",
					StartTime: localTime(2024, 1, 2, 8, 0, 0),
				},
				{
					ID:          "closed1",
					StartTime:   localTime(2024, 1, 2, 12, 0, 0),
					EndTime:     ptrTime(localTime(2024, 1, 2, 13, 0, 0)),
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
			name: "duration only without end time",
			sessions: []*PlaySession{
				{
					ID:          "dur1",
					StartTime:   localTime(2024, 1, 2, 10, 0, 0),
					EndTime:     nil,
					DurationSec: ptrInt(1800),
				},
			},
			from: from,
			to:   to,
			wantDaily: []DailyPlaySeconds{
				{Date: "2024-01-02", Seconds: 1800},
			},
			wantTop: []TopWorldSummary{
				{WorldID: "_total", WorldName: "全セッション", Seconds: 1800, Sessions: 1},
			},
		},
		{
			name: "zero duration session excluded",
			sessions: []*PlaySession{
				{
					ID:          "zero",
					StartTime:   localTime(2024, 1, 2, 10, 0, 0),
					EndTime:     ptrTime(localTime(2024, 1, 2, 10, 0, 0)),
					DurationSec: ptrInt(0),
				},
			},
			from:      from,
			to:        to,
			wantDaily: []DailyPlaySeconds{},
			wantTop:   nil,
		},
		{
			name: "session clipped at range start",
			sessions: []*PlaySession{
				{
					ID:          "clip-start",
					StartTime:   localTime(2023, 12, 31, 22, 0, 0),
					EndTime:     ptrTime(localTime(2024, 1, 1, 2, 0, 0)),
					DurationSec: ptrInt(14400),
				},
			},
			from: from,
			to:   to,
			wantDaily: []DailyPlaySeconds{
				{Date: "2024-01-01", Seconds: 7200},
			},
			wantTop: []TopWorldSummary{
				{WorldID: "_total", WorldName: "全セッション", Seconds: 7200, Sessions: 1},
			},
		},
		{
			name: "session clipped at range end",
			sessions: []*PlaySession{
				{
					ID:          "clip-end",
					StartTime:   localTime(2024, 1, 3, 22, 0, 0),
					EndTime:     ptrTime(localTime(2024, 1, 4, 2, 0, 0)),
					DurationSec: ptrInt(14400),
				},
			},
			from: from,
			to:   to,
			wantDaily: []DailyPlaySeconds{
				{Date: "2024-01-03", Seconds: 7200},
			},
			wantTop: []TopWorldSummary{
				{WorldID: "_total", WorldName: "全セッション", Seconds: 7200, Sessions: 1},
			},
		},
		{
			name: "session spanning three local days",
			sessions: []*PlaySession{
				{
					ID:          "three-day",
					StartTime:   localTime(2024, 1, 1, 22, 0, 0),
					EndTime:     ptrTime(localTime(2024, 1, 3, 2, 0, 0)),
					DurationSec: ptrInt(100800),
				},
			},
			from: from,
			to:   to,
			wantDaily: []DailyPlaySeconds{
				{Date: "2024-01-01", Seconds: 7200},
				{Date: "2024-01-02", Seconds: 86400},
				{Date: "2024-01-03", Seconds: 7200},
			},
			wantTop: []TopWorldSummary{
				{WorldID: "_total", WorldName: "全セッション", Seconds: 100800, Sessions: 1},
			},
		},
		{
			name: "zero duration sec pointer excluded",
			sessions: []*PlaySession{
				{
					ID:          "dur-zero",
					StartTime:   localTime(2024, 1, 2, 10, 0, 0),
					EndTime:     nil,
					DurationSec: ptrInt(0),
				},
			},
			from:      from,
			to:        to,
			wantDaily: []DailyPlaySeconds{},
			wantTop:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDaily, gotTop := AggregatePlaySessions(tt.sessions, tt.from, tt.to, tt.lastObserved)
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
