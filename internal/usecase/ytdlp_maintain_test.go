package usecase

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestYTDLPMaintain_SetMaintain_requiresRiskAck(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("maintain enable path is Windows-gated")
	}
	repo := &fakeAppSettingsRepo{m: map[string]string{}}
	settings := NewSettingsUseCase(repo)
	uc := NewYTDLPMaintainUseCase(settings, NewYTDLPUpdater())
	dir := t.TempDir()
	uc.ToolsPathOverride = filepath.Join(dir, "Tools", "yt-dlp.exe")
	uc.CachePathOverride = filepath.Join(dir, "cache", "yt-dlp.exe")
	uc.UnlockTimeout = time.Second

	err := uc.SetMaintainDesired(context.Background(), true)
	if !errors.Is(err, ErrYTDLPRiskAckRequired) {
		t.Fatalf("got %v", err)
	}
}

func TestYTDLPMaintain_enableLinksTools(t *testing.T) {
	if runtime.GOOS != "windows" {
		// Symlink + GOOS gate: still test core via EnsureAndLink after forcing Supported? GetStatus is GOOS-gated.
		// On Linux we unit-test LinkToolsToCache separately; skip full maintain here.
		t.Skip("full maintain flow is Windows-only")
	}
	var base string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/latest", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"2025.04.01","assets":[{"name":"yt-dlp.exe","browser_download_url":"` + base + `/bin/yt-dlp.exe"}]}`))
	})
	mux.HandleFunc("/bin/yt-dlp.exe", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("MZcache"))
	})
	srv := httptest.NewServer(mux)
	base = srv.URL
	defer srv.Close()

	repo := &fakeAppSettingsRepo{m: map[string]string{}}
	settings := NewSettingsUseCase(repo)
	updater := &YTDLPUpdater{HTTPClient: srv.Client(), ReleasesLatestURL: srv.URL + "/api/latest"}
	uc := NewYTDLPMaintainUseCase(settings, updater)
	dir := t.TempDir()
	uc.ToolsPathOverride = filepath.Join(dir, "Tools", "yt-dlp.exe")
	uc.CachePathOverride = filepath.Join(dir, "cache", "yt-dlp.exe")
	uc.UnlockTimeout = time.Second

	ctx := context.Background()
	if err := uc.AcknowledgeRisk(ctx); err != nil {
		t.Fatal(err)
	}
	if err := uc.SetMaintainDesired(ctx, true); err != nil {
		t.Fatal(err)
	}
	st, err := uc.GetStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !st.MaintainDesired || !st.EffectiveOfficial || !st.CachePresent {
		t.Fatalf("status %+v", st)
	}

	if err := uc.SetMaintainDesired(ctx, false); err != nil {
		t.Fatal(err)
	}
	// Disable must not remove Tools link
	if _, err := os.Lstat(uc.ToolsPathOverride); err != nil {
		t.Fatalf("tools should remain: %v", err)
	}
	st, _ = uc.GetStatus(ctx)
	if st.MaintainDesired {
		t.Fatal("desired should be off")
	}
}

func TestYTDLPMaintain_ReapplyIfNeeded(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cache := filepath.Join(dir, "cache", "yt-dlp.exe")
	tools := filepath.Join(dir, "Tools", "yt-dlp.exe")
	if err := os.MkdirAll(filepath.Dir(cache), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cache, []byte("official"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(tools), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tools, []byte("bundled"), 0o755); err != nil {
		t.Fatal(err)
	}

	repo := &fakeAppSettingsRepo{m: map[string]string{
		keyYTDLPToolsReplaceMaintain: "true",
	}}
	settings := NewSettingsUseCase(repo)
	uc := NewYTDLPMaintainUseCase(settings, NewYTDLPUpdater())
	uc.ToolsPathOverride = tools
	uc.CachePathOverride = cache
	uc.UnlockTimeout = time.Second

	if err := uc.ReapplyIfNeeded(context.Background()); err != nil {
		t.Fatal(err)
	}
	eff, err := EffectiveOfficialLink(tools, cache)
	if err != nil || !eff {
		t.Fatalf("effective=%v err=%v", eff, err)
	}
}
