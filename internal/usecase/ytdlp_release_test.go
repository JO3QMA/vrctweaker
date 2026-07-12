package usecase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestYtdlpExeAssetFromReleaseJSON(t *testing.T) {
	t.Parallel()
	const payload = `{
  "tag_name": "2025.01.15",
  "assets": [
    {"name": "other.zip", "browser_download_url": "https://example.com/o.zip"},
    {"name": "yt-dlp.exe", "browser_download_url": "https://github.com/y/y/releases/download/2025.01.15/yt-dlp.exe"}
  ]
}`
	tag, url, err := ytdlpExeAssetFromReleaseJSON([]byte(payload))
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

func TestYtdlpExeAssetFromReleaseJSON_missingAsset(t *testing.T) {
	t.Parallel()
	const payload = `{"tag_name": "1.0.0", "assets": [{"name": "foo", "browser_download_url": "https://x"}]}`
	_, _, err := ytdlpExeAssetFromReleaseJSON([]byte(payload))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeReleaseTag(t *testing.T) {
	t.Parallel()
	if g := normalizeReleaseTag("v2024.12.01"); g != "2024.12.01" {
		t.Fatalf("got %q", g)
	}
}

func TestFetchLatestRelease_httptest(t *testing.T) {
	t.Parallel()
	const relJSON = `{
  "tag_name": "2025.02.01",
  "assets": [
    {"name": "yt-dlp.exe", "browser_download_url": "PLACEHOLDER"}
  ]
}`
	var srvURL string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/latest", func(w http.ResponseWriter, r *http.Request) {
		body := strings.ReplaceAll(relJSON, "PLACEHOLDER", srvURL+"/bin/yt-dlp.exe")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
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
}

func TestDownloadToCache_andEnsure(t *testing.T) {
	t.Parallel()
	var base string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/latest", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"2025.03.01","assets":[{"name":"yt-dlp.exe","browser_download_url":"` + base + `/bin/yt-dlp.exe"}]}`))
	})
	mux.HandleFunc("/bin/yt-dlp.exe", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("MZfake-official"))
	})
	srv := httptest.NewServer(mux)
	base = srv.URL
	defer srv.Close()

	cache := filepath.Join(t.TempDir(), "ytdlp", "yt-dlp.exe")
	u := &YTDLPUpdater{HTTPClient: srv.Client(), ReleasesLatestURL: srv.URL + "/api/latest"}
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
	if string(b) != "MZfake-official" {
		t.Fatalf("cache content %q", b)
	}

	// Second ensure should not re-download (file already present)
	if _, err := u.EnsureOfficialCache(context.Background(), cache); err != nil {
		t.Fatal(err)
	}
}
