package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseLatestReleaseJSON(t *testing.T) {
	const sha = "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	body := fmt.Sprintf(`{
		"tag_name": "v1.2.0",
		"html_url": "https://github.com/JO3QMA/vrctweaker/releases/tag/v1.2.0",
		"prerelease": false,
		"draft": false,
		"body": "Windows zip.\n\n**SHA256 (zip):** `+"`%s`"+`\n",
		"assets": [
			{"name": "vrchat-tweaker-v1.2.0-windows-amd64.zip", "browser_download_url": "https://github.com/JO3QMA/vrctweaker/releases/download/v1.2.0/vrchat-tweaker-v1.2.0-windows-amd64.zip"}
		]
	}`, sha)
	rel, err := parseLatestReleaseJSON([]byte(body), true)
	if err != nil {
		t.Fatal(err)
	}
	if rel.Version != "1.2.0" || rel.ZipSHA256 != sha {
		t.Fatalf("unexpected rel: %+v", rel)
	}
}

func TestParseLatestReleaseJSON_rejectsPrerelease(t *testing.T) {
	_, err := parseLatestReleaseJSON([]byte(`{"tag_name":"v1.0.0","prerelease":true}`), true)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReleaseClient_FetchLatestStable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := fmt.Sprintf(`{
			"tag_name": "v0.3.0",
			"html_url": "https://example.com/r",
			"prerelease": false,
			"draft": false,
			"body": "**SHA256 (zip):** `+"`%s`"+`",
			"assets": [{"name":"vrchat-tweaker-v0.3.0-windows-amd64.zip","browser_download_url":"https://github.com/x/y.zip"}]
		}`, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	c := &ReleaseClient{
		HTTPClient:        srv.Client(),
		ReleasesLatestURL: srv.URL,
		SkipURLValidation: true,
	}
	rel, err := c.FetchLatestStable(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rel.Version != "0.3.0" {
		t.Fatalf("version %q", rel.Version)
	}
}

func TestTruncateReleaseNotes(t *testing.T) {
	if TruncateReleaseNotes("short", 10) != "short" {
		t.Fatal("expected unchanged")
	}
	got := TruncateReleaseNotes("12345678901", 10)
	if got != "1234567890…" {
		t.Fatalf("got %q", got)
	}
}
