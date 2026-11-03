package usecase

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vrchat-tweaker/internal/appversion"
)

func TestYtdlpReleaseAssetsFromReleaseJSON(t *testing.T) {
	t.Parallel()
	const payload = `{
  "tag_name": "2025.01.15",
  "assets": [
    {"name": "SHA2-256SUMS", "browser_download_url": "https://github.com/yt-dlp/yt-dlp/releases/download/2025.01.15/SHA2-256SUMS"},
    {"name": "other.zip", "browser_download_url": "https://example.com/o.zip"},
    {"name": "yt-dlp.exe", "browser_download_url": "https://github.com/y/y/releases/download/2025.01.15/yt-dlp.exe"}
  ]
}`
	tag, url, _, err := ytdlpReleaseAssetsFromReleaseJSON([]byte(payload))
	if err != nil {
		t.Fatal(err)
	}
	if tag != "2025.01.15" {
		t.Fatalf("tag: got %q", tag)
	}
	if !strings.HasSuffix(url, "yt-dlp.exe") {
		t.Fatalf("url: got %q", url)
	}
}

func TestYtdlpReleaseAssetsFromReleaseJSON_missingAsset(t *testing.T) {
	t.Parallel()
	const payload = `{"tag_name": "1.0.0", "assets": [{"name": "foo", "browser_download_url": "https://x"}]}`
	_, _, _, err := ytdlpReleaseAssetsFromReleaseJSON([]byte(payload))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeReleaseTag(t *testing.T) {
	t.Parallel()
	if g := appversion.NormalizeProductVersion("v2024.12.01"); g != "2024.12.01" {
		t.Fatalf("got %q", g)
	}
}

func TestYtdlpReleaseTagFromDownloadURL(t *testing.T) {
	t.Parallel()
	const url = "https://github.com/yt-dlp/yt-dlp/releases/download/2026.07.04/yt-dlp.exe"
	if got := ytdlpReleaseTagFromDownloadURL(url); got != "2026.07.04" {
		t.Fatalf("got %q", got)
	}
	if got := ytdlpReleaseTagFromDownloadURL("https://evil.example/x"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestFetchLatestRelease_httptest(t *testing.T) {
	t.Parallel()
	const relJSON = `{
  "tag_name": "2025.02.01",
  "assets": [
    {"name": "SHA2-256SUMS", "browser_download_url": "SUMS_PLACEHOLDER"},
    {"name": "yt-dlp.exe", "browser_download_url": "EXE_PLACEHOLDER"}
  ]
}`
	var srvURL string
	var gotUA string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/latest", func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		body := strings.ReplaceAll(relJSON, "EXE_PLACEHOLDER", srvURL+"/bin/yt-dlp.exe")
		body = strings.ReplaceAll(body, "SUMS_PLACEHOLDER", srvURL+"/bin/SHA2-256SUMS")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	})
	mux.HandleFunc("/bin/SHA2-256SUMS", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(ytdlpTestSumsBodyForExe("ignored")))
	})
	srv := httptest.NewServer(mux)
	srvURL = srv.URL
	defer srv.Close()

	u := &YTDLPUpdater{
		HTTPClient:        srv.Client(),
		ReleasesLatestURL: srv.URL + "/api/latest",
	}
	info, err := u.FetchLatestRelease(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Tag != "2025.02.01" || info.Version != "2025.02.01" {
		t.Fatalf("info %+v", info)
	}
	if !strings.HasSuffix(info.DownloadURL, "/bin/yt-dlp.exe") {
		t.Fatalf("dl %q", info.DownloadURL)
	}
	wantUA := appversion.UserAgent()
	if gotUA != wantUA {
		t.Fatalf("User-Agent: got %q, want %q", gotUA, wantUA)
	}
}

func TestValidateYTDlpDownloadURL(t *testing.T) {
	t.Parallel()
	cdnAsset := "https://release-assets.githubusercontent.com/github-production-release-asset/307260205/f116afaa-741f-4c8c-9767-a3ae137fbae6"
	if err := validateYTDlpDownloadURL(cdnAsset); err != nil {
		t.Fatalf("release-assets CDN: %v", err)
	}
	objectsAsset := "https://objects.githubusercontent.com/github-production-release-asset/307260205/f116afaa-741f-4c8c-9767-a3ae137fbae6"
	if err := validateYTDlpDownloadURL(objectsAsset); err != nil {
		t.Fatalf("objects CDN: %v", err)
	}
	if err := validateYTDlpDownloadURL("https://release-assets.githubusercontent.com/evil/yt-dlp.exe"); err == nil {
		t.Fatal("expected reject for non-release-asset CDN path")
	}
	if err := validateYTDlpDownloadURL("https://evil.example/yt-dlp.exe"); err == nil {
		t.Fatal("expected reject for untrusted host")
	}
	if err := validateYTDlpDownloadURL("https://github.com/evil/evil/releases/download/1.0/yt-dlp.exe"); err == nil {
		t.Fatal("expected reject for non-official github release path")
	}
	if err := validateYTDlpDownloadURL("https://github.com/yt-dlp/yt-dlp/releases/download/2026.01.01/yt-dlp.exe"); err != nil {
		t.Fatalf("official URL should pass: %v", err)
	}
}

func TestDownloadToCache_followsGitHubReleaseRedirect(t *testing.T) {
	t.Parallel()
	const exeBody = "MZfake-official"
	sumsBody := ytdlpTestSumsBodyForExe(exeBody)
	release := YTDLPReleaseInfo{
		DownloadURL:  "https://github.com/yt-dlp/yt-dlp/releases/download/2026.07.04/yt-dlp.exe",
		ChecksumsURL: "https://github.com/yt-dlp/yt-dlp/releases/download/2026.07.04/SHA2-256SUMS",
	}
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Host {
			case "github.com":
				if strings.HasSuffix(req.URL.Path, "SHA2-256SUMS") {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(sumsBody)),
						Request:    req,
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusFound,
					Header:     http.Header{"Location": []string{"https://release-assets.githubusercontent.com/github-production-release-asset/307260205/f116afaa-741f-4c8c-9767-a3ae137fbae6"}},
					Body:       http.NoBody,
					Request:    req,
				}, nil
			case "release-assets.githubusercontent.com":
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(exeBody)),
					Request:    req,
				}, nil
			default:
				return nil, fmt.Errorf("unexpected host %q", req.URL.Host)
			}
		}),
	}

	cache := filepath.Join(t.TempDir(), "ytdlp", "yt-dlp.exe")
	u := &YTDLPUpdater{HTTPClient: client}
	if err := u.DownloadToCache(context.Background(), cache, release); err != nil {
		t.Fatalf("DownloadToCache: %v", err)
	}
	b, err := os.ReadFile(cache)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != exeBody {
		t.Fatalf("cache content %q", b)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDownloadToCache_andEnsure(t *testing.T) {
	t.Parallel()
	const exeBody = "MZfake-official"
	ts := startYTDlpReleaseTestServer("2025.03.01", exeBody, "", true, nil)
	defer ts.Server.Close()

	cache := filepath.Join(t.TempDir(), "ytdlp", "yt-dlp.exe")
	u := ts.Updater
	info, err := u.EnsureOfficialCache(context.Background(), cache)
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "2025.03.01" {
		t.Fatalf("version %q", info.Version)
	}
	b, err := os.ReadFile(cache)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != exeBody {
		t.Fatalf("cache content %q", b)
	}

	// Second ensure should not re-download (file already present)
	if _, err := u.EnsureOfficialCache(context.Background(), cache); err != nil {
		t.Fatal(err)
	}
}

func TestDownloadToCache_hashMatch(t *testing.T) {
	t.Parallel()
	const exeBody = "MZhash-match"
	ts := startYTDlpReleaseTestServer("2025.03.01", exeBody, "", true, nil)
	defer ts.Server.Close()
	cache := filepath.Join(t.TempDir(), "ytdlp", "yt-dlp.exe")
	info, err := ts.Updater.FetchLatestRelease(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if dlErr := ts.Updater.DownloadToCache(context.Background(), cache, info); dlErr != nil {
		t.Fatalf("DownloadToCache: %v", dlErr)
	}
	b, err := os.ReadFile(cache)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != exeBody {
		t.Fatalf("cache content %q", b)
	}
}

func TestDownloadToCache_hashMismatch_keepsExisting(t *testing.T) {
	t.Parallel()
	const exeBody = "MZactual"
	const wrongSums = "0000000000000000000000000000000000000000000000000000000000000000  yt-dlp.exe\n"
	ts := startYTDlpReleaseTestServer("2025.03.01", exeBody, wrongSums, true, nil)
	defer ts.Server.Close()
	cache := filepath.Join(t.TempDir(), "ytdlp", "yt-dlp.exe")
	const existing = "keep-me"
	if err := os.MkdirAll(filepath.Dir(cache), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cache, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := ts.Updater.FetchLatestRelease(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if dlErr := ts.Updater.DownloadToCache(context.Background(), cache, info); dlErr == nil {
		t.Fatal("expected checksum mismatch error")
	}
	b, err := os.ReadFile(cache)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != existing {
		t.Fatalf("existing cache overwritten: %q", b)
	}
}

func TestDownloadToCache_missingSUMSEntry(t *testing.T) {
	t.Parallel()
	const exeBody = "MZmissing-entry"
	ts := startYTDlpReleaseTestServer("2025.03.01", exeBody, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  other.bin\n", true, nil)
	defer ts.Server.Close()
	cache := filepath.Join(t.TempDir(), "ytdlp", "yt-dlp.exe")
	info, err := ts.Updater.FetchLatestRelease(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := ts.Updater.DownloadToCache(context.Background(), cache, info); err == nil {
		t.Fatal("expected missing checksum entry error")
	} else if !strings.Contains(err.Error(), "yt-dlp.exe") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateYTDlpDownloadURL_nonPinnedURLRejected(t *testing.T) {
	t.Parallel()
	if err := validateYTDlpDownloadURL("https://github.com/not-yt-dlp/other/releases/download/1.0/yt-dlp.exe"); err == nil {
		t.Fatal("expected reject")
	}
	if err := validateYTDlpDownloadURL("https://github.com/yt-dlp/yt-dlp/releases/download/2026.01.01/yt-dlp.exe"); err != nil {
		t.Fatalf("official URL should pass: %v", err)
	}
}
