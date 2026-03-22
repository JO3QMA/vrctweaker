package activity

import "time"

// EndOfLocalCalendarDay returns the last instant of t's local calendar day (23:59:59.999999999 in time.Local).
func EndOfLocalCalendarDay(t time.Time) time.Time {
	loc := time.Local
	lt := t.In(loc)
	y, m, d := lt.Date()
	return time.Date(y, m, d, 23, 59, 59, 999999999, loc)
}

// StartOfNextLocalCalendarDay returns midnight at the start of the local calendar day after t.
func StartOfNextLocalCalendarDay(t time.Time) time.Time {
	loc := time.Local
	lt := t.In(loc)
	y, m, d := lt.Date()
	dayStart := time.Date(y, m, d, 0, 0, 0, 0, loc)
	return dayStart.AddDate(0, 0, 1)
}

// SameLocalCalendarDay reports whether a and b fall on the same calendar date in time.Local.
func SameLocalCalendarDay(a, b time.Time) bool {
	ay, am, ad := a.In(time.Local).Date()
	by, bm, bd := b.In(time.Local).Date()
	return ay == by && am == bm && ad == bd
}
