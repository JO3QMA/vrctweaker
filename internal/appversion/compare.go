package appversion

import (
	"strconv"
	"strings"
)

type versionCore struct {
	parts      [3]int
	prerelease bool
}

// IsDevelopmentVersion reports dev/empty builds that should not prompt for self-update.
func IsDevelopmentVersion(version string) bool {
	v := strings.TrimSpace(version)
	return v == "" || v == "dev"
}

// CompareProductVersion compares two product version strings (e.g. 1.2.3).
// Returns -1 if a < b, 0 if equal, 1 if a > b.
// Non-semver or "dev" current is treated as older than any valid release version.
// A prerelease (suffix after -/+) is older than the same numeric version without suffix.
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
	ac := parseVersionCore(a)
	bc := parseVersionCore(b)
	for i := 0; i < 3; i++ {
		if ac.parts[i] < bc.parts[i] {
			return -1
		}
		if ac.parts[i] > bc.parts[i] {
			return 1
		}
	}
	if ac.prerelease && !bc.prerelease {
		return -1
	}
	if !ac.prerelease && bc.prerelease {
		return 1
	}
	return 0
}

func parseVersionCore(v string) versionCore {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	prerelease := false
	if idx := strings.IndexAny(v, "-+"); idx >= 0 {
		prerelease = true
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
	return versionCore{parts: out, prerelease: prerelease}
}
