// Package errtext provides helpers for building error messages.
package errtext

import "strings"

// TruncateForErr trims s and caps it at max runes, appending "…" when truncated.
// A non-positive max returns the trimmed string unchanged.
func TruncateForErr(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…"
}
