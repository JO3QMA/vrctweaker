package usecase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
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
		if r.Method != http.MethodGet {
			t.Errorf("method %s", r.Method)
		}
		body := strings.ReplaceAll(relJSON, "PLACEHOLDER", srvURL+"/bin/yt-dlp.exe")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	})
	mux.HandleFunc("/bin/yt-dlp.exe", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("MZfake"))
	})
	srv := httptest.NewServer(mux)
	srvURL = srv.URL
	defer srv.Close()

	u := &YTDLPUpdater{
		HTTPClient:        srv.Client(),
		ReleasesLatestURL: srv.URL + "/api/latest",
	}
	ctx := context.Background()
	tag, dl, err := u.FetchLatestRelease(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if tag != "2025.02.01" {
		t.Fatalf("tag %q", tag)
	}
	if !strings.HasSuffix(dl, "/bin/yt-dlp.exe") {
		t.Fatalf("dl %q", dl)
	}
}

func TestDownloadToFile(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello-ytdlp"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "out.part")
	ctx := context.Background()
	if err := downloadToFile(ctx, srv.Client(), srv.URL, dest); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "hello-ytdlp" {
		t.Fatalf("content %q", b)
	}
}

func TestFinishYTDLPInstall_backupAndReplace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	exe := filepath.Join(dir, "yt-dlp.exe")
	part := filepath.Join(dir, "yt-dlp.exe.part")
	bak := filepath.Join(dir, "yt-dlp.exe.bak")

	if err := os.WriteFile(exe, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(part, []byte("newbinary"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := finishYTDLPInstall(exe, part); err != nil {
		t.Fatal(err)
	}
	gotExe, _ := os.ReadFile(exe)
	if string(gotExe) != "newbinary" {
		t.Fatalf("exe content %q", gotExe)
	}
	gotBak, _ := os.ReadFile(bak)
	if string(gotBak) != "old" {
		t.Fatalf("bak content %q", gotBak)
	}
	if _, err := os.Stat(part); !os.IsNotExist(err) {
		t.Fatal("part should be gone")
	}
}

func TestFinishYTDLPInstall_freshInstall(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	exe := filepath.Join(dir, "yt-dlp.exe")
	part := filepath.Join(dir, "yt-dlp.exe.part")
	if err := os.WriteFile(part, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := finishYTDLPInstall(exe, part); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(exe)
	if string(b) != "x" {
		t.Fatal(string(b))
	}
}

func TestLocalYTDLPVersion_script(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("shell script not used on windows")
	}
	dir := t.TempDir()
	exe := filepath.Join(dir, "yt-dlp-exe-mock")
	script := "#!/bin/sh\necho 2099.01.01\n"
	if err := os.WriteFile(exe, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	v := LocalYTDLPVersion(ctx, exe)
	if v != "2099.01.01" {
		t.Fatalf("got %q", v)
	}
}

func TestGetUpdateStatus_unsupportedOS(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("windows is supported")
	}
	u := NewYTDLPUpdater()
	st := u.GetUpdateStatus(context.Background(), filepath.Join("/tmp", "VRChat", "VRChat", "config.json"))
	if st.Supported {
		t.Fatal("expected unsupported")
	}
	if st.UnsupportedReason == "" {
		t.Fatal("expected reason")
	}
}

func TestVRChatYTDLPExePath(t *testing.T) {
	t.Parallel()
	cfg := filepath.Join("/home", "u", ".local", "share", "VRChat", "VRChat", "config.json")
	want := filepath.Join("/home", "u", ".local", "share", "VRChat", "VRChat", "Tools", ytdlpReleaseAssetName)
	if g := VRChatYTDLPExePath(cfg); g != want {
		t.Fatalf("got %q want %q", g, want)
	}
}
