package github

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// APIJSONKind selects size limits and error text for a GitHub API GET.
// yt-dlp and the app updater keep their historical limits and messages.
type APIJSONKind int

const (
	// APIJSONYTDLP is the yt-dlp latest-release fetch: 8 MiB cap with truncation,
	// errors prefixed "github api".
	APIJSONYTDLP APIJSONKind = iota + 1
	// APIJSONAppRelease is the app updater fetch: 1 MiB hard cap,
	// errors prefixed "github releases".
	APIJSONAppRelease
)

const githubAPIAccept = "application/vnd.github+json"

const (
	ytdlpReleaseMaxBytes   int64 = 8 << 20
	appReleaseMaxBytes     int64 = 1 << 20
	appReleaseErrBodyBytes       = 4096
	ytdlpErrBodyRunes            = 200
)

// TruncateErrBody trims space and caps s at max runes, appending an ellipsis.
func TruncateErrBody(s string, max int) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…"
}

// GetJSON GETs a GitHub API URL and returns the body on HTTP 200.
// Accept is application/vnd.github+json. User-Agent is the provided value.
func GetJSON(ctx context.Context, client *http.Client, rawURL, userAgent string, kind APIJSONKind) ([]byte, error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", githubAPIAccept)
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		return nil, wrapGitHubRequestErr(kind, err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch kind {
	case APIJSONYTDLP:
		return readYTDLPReleaseBody(resp, ytdlpReleaseMaxBytes)
	case APIJSONAppRelease:
		return readAppReleaseBody(resp, appReleaseMaxBytes, appReleaseErrBodyBytes)
	default:
		return nil, errors.New("github: unknown api json kind")
	}
}

func wrapGitHubRequestErr(kind APIJSONKind, err error) error {
	if kind == APIJSONAppRelease {
		return fmt.Errorf("github releases: %w", err)
	}
	return fmt.Errorf("github api request: %w", err)
}

func readYTDLPReleaseBody(resp *http.Response, maxBytes int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
	if err != nil {
		return nil, fmt.Errorf("read release response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api: %s: %s", resp.Status, TruncateErrBody(string(body), ytdlpErrBodyRunes))
	}
	return body, nil
}

func readAppReleaseBody(resp *http.Response, maxBytes int64, errBodyBytes int) ([]byte, error) {
	if resp.StatusCode != http.StatusOK {
		slurp, _ := io.ReadAll(io.LimitReader(resp.Body, int64(errBodyBytes)))
		return nil, fmt.Errorf("github releases: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(slurp)))
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("github releases: read body: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, errors.New("github releases: response exceeds size limit")
	}
	return data, nil
}
