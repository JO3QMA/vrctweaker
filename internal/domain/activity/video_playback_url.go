package activity

import (
	"net/url"
	"strings"
)

// StripYouTubeTrackingParams removes tracking query parameters from YouTube attempt URLs.
// Non-YouTube URLs and unparseable URLs are returned unchanged.
func StripYouTubeTrackingParams(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return raw
	}
	if !isYouTubeVideoHost(u.Hostname()) {
		return raw
	}
	if u.RawQuery == "" {
		return raw
	}

	kept := make([]string, 0, strings.Count(u.RawQuery, "&")+1)
	for _, part := range strings.Split(u.RawQuery, "&") {
		if part == "" {
			continue
		}
		key := part
		if i := strings.Index(part, "="); i >= 0 {
			key = part[:i]
		}
		if isYouTubeTrackingQueryParam(key) {
			continue
		}
		kept = append(kept, part)
	}
	u.RawQuery = strings.Join(kept, "&")
	return u.String()
}

func isYouTubeVideoHost(host string) bool {
	if host == "youtu.be" {
		return true
	}
	return host == "youtube.com" || strings.HasSuffix(host, ".youtube.com")
}

func isYouTubeTrackingQueryParam(name string) bool {
	lower := strings.ToLower(name)
	switch lower {
	case "si", "feature", "pp", "igsh", "fbclid", "gclid":
		return true
	default:
		return strings.HasPrefix(lower, "utm_")
	}
}
