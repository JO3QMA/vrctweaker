package usecase

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
)

type ytdlpBlockingExeGate struct {
	startOnce sync.Once
	Started   chan struct{}
	Proceed   chan struct{}
}

func newYTDlpBlockingExeGate() *ytdlpBlockingExeGate {
	return &ytdlpBlockingExeGate{
		Started: make(chan struct{}),
		Proceed: make(chan struct{}),
	}
}

func (g *ytdlpBlockingExeGate) holdAndServe(w http.ResponseWriter, exeBody string) {
	g.startOnce.Do(func() { close(g.Started) })
	<-g.Proceed
	_, _ = w.Write([]byte(exeBody))
}

func startYTDlpReleaseTestServerBlocking(tag, exeBody string, gate *ytdlpBlockingExeGate) *ytdlpReleaseTestServer {
	if gate == nil {
		gate = newYTDlpBlockingExeGate()
	}
	if tag == "" {
		tag = "2025.03.01"
	}
	sumsBody := ytdlpTestSumsBodyForExe(exeBody)
	exePath := fmt.Sprintf("/yt-dlp/yt-dlp/releases/download/%s/yt-dlp.exe", tag)
	sumsPath := fmt.Sprintf("/yt-dlp/yt-dlp/releases/download/%s/SHA2-256SUMS", tag)

	mux := http.NewServeMux()
	mux.HandleFunc(exePath, func(w http.ResponseWriter, _ *http.Request) {
		gate.holdAndServe(w, exeBody)
	})
	mux.HandleFunc(sumsPath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(sumsBody))
	})

	var base string
	mux.HandleFunc("/api/latest", func(w http.ResponseWriter, _ *http.Request) {
		exeURL := "https://github.com" + exePath
		sumsURL := "https://github.com" + sumsPath
		payload := fmt.Sprintf(`{"tag_name":%q,"assets":[
  {"name":"SHA2-256SUMS","browser_download_url":%q},
  {"name":"yt-dlp.exe","browser_download_url":%q}
]}`, tag, sumsURL, exeURL)
		_, _ = w.Write([]byte(payload))
	})

	srv := httptest.NewServer(mux)
	base = srv.URL
	client := &http.Client{
		Transport: rewriteGitHubHostTransport{testBase: base, base: srv.Client().Transport},
	}
	updater := &YTDLPUpdater{
		HTTPClient:        client,
		ReleasesLatestURL: base + "/api/latest",
	}
	return &ytdlpReleaseTestServer{Server: srv, Base: base, Updater: updater}
}

type ytdlpReleaseTestServer struct {
	Server  *httptest.Server
	Base    string
	Updater *YTDLPUpdater
}

func serveYTDlpTestExe(w http.ResponseWriter, exeBody string, trackDownloads *int) {
	if trackDownloads != nil {
		*trackDownloads++
	}
	_, _ = w.Write([]byte(exeBody))
}

func ytdlpTestSumsBodyForExe(exeBody string) string {
	sum := sha256.Sum256([]byte(exeBody))
	return hex.EncodeToString(sum[:]) + "  yt-dlp.exe\n"
}

type rewriteGitHubHostTransport struct {
	testBase string
	base     http.RoundTripper
}

func (t rewriteGitHubHostTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.EqualFold(req.URL.Host, "github.com") {
		testURL, err := url.Parse(t.testBase)
		if err != nil {
			return nil, err
		}
		cloned := req.Clone(req.Context())
		cloned.URL.Scheme = testURL.Scheme
		cloned.URL.Host = testURL.Host
		req = cloned
	}
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

// startYTDlpReleaseTestServer serves release API + exe + SHA2-256SUMS.
// When officialGitHubPaths is true, asset URLs use github.com/yt-dlp/yt-dlp/releases/download/ (rewritten to srv).
func startYTDlpReleaseTestServer(tag, exeBody, sumsBody string, officialGitHubPaths bool, trackDownloads *int) *ytdlpReleaseTestServer {
	if tag == "" {
		tag = "2025.03.01"
	}
	if sumsBody == "" {
		sumsBody = ytdlpTestSumsBodyForExe(exeBody)
	}
	exePath := fmt.Sprintf("/yt-dlp/yt-dlp/releases/download/%s/yt-dlp.exe", tag)
	sumsPath := fmt.Sprintf("/yt-dlp/yt-dlp/releases/download/%s/SHA2-256SUMS", tag)

	mux := http.NewServeMux()
	mux.HandleFunc("/bin/yt-dlp.exe", func(w http.ResponseWriter, _ *http.Request) {
		serveYTDlpTestExe(w, exeBody, trackDownloads)
	})
	mux.HandleFunc("/bin/SHA2-256SUMS", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(sumsBody))
	})
	mux.HandleFunc(exePath, func(w http.ResponseWriter, _ *http.Request) {
		serveYTDlpTestExe(w, exeBody, trackDownloads)
	})
	mux.HandleFunc(sumsPath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(sumsBody))
	})

	var base string
	mux.HandleFunc("/api/latest", func(w http.ResponseWriter, _ *http.Request) {
		var exeURL, sumsURL string
		if officialGitHubPaths {
			exeURL = "https://github.com" + exePath
			sumsURL = "https://github.com" + sumsPath
		} else {
			exeURL = base + "/bin/yt-dlp.exe"
			sumsURL = base + "/bin/SHA2-256SUMS"
		}
		payload := fmt.Sprintf(`{"tag_name":%q,"assets":[
  {"name":"SHA2-256SUMS","browser_download_url":%q},
  {"name":"yt-dlp.exe","browser_download_url":%q}
]}`, tag, sumsURL, exeURL)
		_, _ = w.Write([]byte(payload))
	})

	srv := httptest.NewServer(mux)
	base = srv.URL

	client := srv.Client()
	if officialGitHubPaths {
		client = &http.Client{
			Transport: rewriteGitHubHostTransport{testBase: base, base: client.Transport},
		}
	}
	updater := &YTDLPUpdater{
		HTTPClient:        client,
		ReleasesLatestURL: base + "/api/latest",
	}
	return &ytdlpReleaseTestServer{Server: srv, Base: base, Updater: updater}
}
