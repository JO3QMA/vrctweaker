package appversion

import (
	"strconv"
	"strings"
)

// CompareProductVersion compares two product version strings (e.g. 1.2.3).
// Returns -1 if a < b, 0 if equal, 1 if a > b.
// Non-semver or "dev" current is treated as older than any valid release version.
func CompareProductVersion(a, b string) int {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == b {
		return 0
	}
	if a == "" || a == "dev" {
		return -1
	}
	if b == "" || b == "dev" {
		return 1
	}
	ap := parseVersionParts(a)
	bp := parseVersionParts(b)
	for i := 0; i < 3; i++ {
		if ap[i] < bp[i] {
			return -1
		}
		if ap[i] > bp[i] {
			return 1
		}
	}
	return 0
}

func parseVersionParts(v string) [3]int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if idx := strings.IndexAny(v, "-+"); idx >= 0 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	var out [3]int
	for i := 0; i < 3 && i < len(parts); i++ {
		n, err := strconv.Atoi(strings.TrimSpace(parts[i]))
		if err != nil || n < 0 {
			continue
		}
		out[i] = n
	}
	return out
}
