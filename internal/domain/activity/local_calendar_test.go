package activity

import (
	"testing"
	"time"
)

func TestEndOfLocalCalendarDay(t *testing.T) {
	tm := time.Date(2024, 3, 19, 10, 30, 45, 123456789, time.Local)
	got := EndOfLocalCalendarDay(tm)
	if got.Hour() != 23 || got.Minute() != 59 || got.Second() != 59 || got.Nanosecond() != 999999999 {
		t.Errorf("EndOfLocalCalendarDay = %v, want last ns of local day", got)
	}
	y, m, d := got.In(time.Local).Date()
	if y != 2024 || m != time.March || d != 19 {
		t.Errorf("date = %v-%v-%v, want 2024-03-19 in Local", y, m, d)
	}
}

func TestStartOfNextLocalCalendarDay(t *testing.T) {
	tm := time.Date(2024, 3, 19, 23, 59, 59, 999999999, time.Local)
	got := StartOfNextLocalCalendarDay(tm)
	y, m, d := got.In(time.Local).Date()
	if y != 2024 || m != time.March || d != 20 {
		t.Errorf("StartOfNextLocalCalendarDay = %v (date %v-%v-%v), want 2024-03-20 00:00 local", got, y, m, d)
	}
	if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
		t.Errorf("StartOfNextLocalCalendarDay not midnight: %v", got)
	}
}

func TestSameLocalCalendarDay(t *testing.T) {
	a := time.Date(2024, 3, 19, 23, 0, 0, 0, time.Local)
	b := time.Date(2024, 3, 19, 1, 0, 0, 0, time.Local)
	if !SameLocalCalendarDay(a, b) {
		t.Error("same day expected")
	}
	c := time.Date(2024, 3, 20, 0, 0, 0, 0, time.Local)
	if SameLocalCalendarDay(a, c) {
		t.Error("different days must not match")
	}
}
