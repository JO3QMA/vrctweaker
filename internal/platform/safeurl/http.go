package safeurl

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrUnsafeURL is returned when a URL is not an allowed http(s) URL with a host.
var ErrUnsafeURL = errors.New("unsafe url")

// ParseHTTPHTTPS returns a normalized absolute URL string when raw is http or https.
func ParseHTTPHTTPS(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("%w: empty", ErrUnsafeURL)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnsafeURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("%w: unsupported scheme", ErrUnsafeURL)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%w: missing host", ErrUnsafeURL)
	}
	return u.String(), nil
}
