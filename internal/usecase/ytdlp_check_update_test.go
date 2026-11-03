package usecase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestRefreshLatestRelease_usesCacheWithinTTL(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("yt-dlp maintain status is Windows-gated")
	}
	t.Parallel()
	var apiCalls int
	mux := http.NewServeMux()
	mux.HandleFunc("/api/latest", func(w http.ResponseWriter, _ *http.Request) {
		apiCalls++
		_, _ = w.Write([]byte(`{"tag_name":"2026.01.01","assets":[
  {"name":"SHA2-256SUMS","browser_download_url":"https://example.com/sums"},
  {"name":"yt-dlp.exe","browser_download_url":"https://example.com/yt-dlp.exe"}
]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	repo := &fakeAppSettingsRepo{m: map[string]string{
		keyYTDLPKnownLatestVersion:     "2026.01.01",
		keyYTDLPKnownLatestTag:         "2026.01.01",
		keyYTDLPKnownLatestDownloadURL: "https://example.com/yt-dlp.exe",
		keyYTDLPKnownLatestCheckedAt:   time.Now().UTC().Format(time.RFC3339),
	}}
	settings := NewSettingsUseCase(repo)
	updater := &YTDLPUpdater{
		HTTPClient:        srv.Client(),
		ReleasesLatestURL: srv.URL + "/api/latest",
	}
	uc := NewYTDLPMaintainUseCase(settings, updater)
	dir := t.TempDir()
	uc.CachePathOverride = filepath.Join(dir, "yt-dlp.exe")
	uc.ToolsPathOverride = filepath.Join(dir, "Tools", "yt-dlp.exe")

	ctx := context.Background()
	st, err := uc.RefreshLatestRelease(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if st.LatestVersion != "2026.01.01" {
		t.Fatalf("latest version: got %q", st.LatestVersion)
	}
	if apiCalls != 0 {
		t.Fatalf("expected cached refresh, api calls = %d", apiCalls)
	}

	if _, err := uc.RefreshLatestRelease(ctx, true); err != nil {
		t.Fatal(err)
	}
	if apiCalls != 1 {
		t.Fatalf("force refresh: api calls = %d", apiCalls)
	}
}

func TestCheckAndUpdateOfficialCache_unchangedSkipsDownload(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("yt-dlp maintain status is Windows-gated")
	}
	t.Parallel()
	var downloads int
	ts := startYTDlpReleaseTestServer("2026.02.02", "MZ", "", true, &downloads)
	defer ts.Server.Close()

	dir := t.TempDir()
	cache := filepath.Join(dir, "yt-dlp.exe")
	if err := os.WriteFile(cache, []byte("cached"), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := &fakeAppSettingsRepo{m: map[string]string{
		keyYTDLPOfficialCacheTag: "2026.02.02",
	}}
	settings := NewSettingsUseCase(repo)
	uc := NewYTDLPMaintainUseCase(settings, ts.Updater)
	uc.CachePathOverride = cache
	uc.ToolsPathOverride = filepath.Join(dir, "Tools", "yt-dlp.exe")

	ctx := context.Background()
	res, err := uc.CheckAndUpdateOfficialCache(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.Outcome != "unchanged" {
		t.Fatalf("outcome %q", res.Outcome)
	}
	if downloads != 0 {
		t.Fatalf("downloads = %d", downloads)
	}
}

func TestCheckAndUpdateOfficialCache_downloadsWhenNewer(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("yt-dlp maintain status is Windows-gated")
	}
	t.Parallel()
	ts := startYTDlpReleaseTestServer("2026.03.03", "MZnew", "", true, nil)
	defer ts.Server.Close()

	dir := t.TempDir()
	cache := filepath.Join(dir, "yt-dlp.exe")
	if err := os.WriteFile(cache, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := &fakeAppSettingsRepo{m: map[string]string{
		keyYTDLPOfficialCacheTag: "2026.01.01",
	}}
	settings := NewSettingsUseCase(repo)
	uc := NewYTDLPMaintainUseCase(settings, ts.Updater)
	uc.CachePathOverride = cache

	ctx := context.Background()
	res, err := uc.CheckAndUpdateOfficialCache(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.Outcome != "updated" {
		t.Fatalf("outcome %q", res.Outcome)
	}
	if res.PreviousVersion != "2026.01.01" {
		t.Fatalf("previous %q", res.PreviousVersion)
	}
	st, _ := uc.GetStatus(ctx)
	if !st.CachePresent {
		t.Fatal("cache should be present")
	}
}

func TestCheckAndUpdateOfficialCache_firstDownload(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("yt-dlp maintain status is Windows-gated")
	}
	t.Parallel()
	ts := startYTDlpReleaseTestServer("2026.04.04", "MZfirst", "", true, nil)
	defer ts.Server.Close()

	dir := t.TempDir()
	cache := filepath.Join(dir, "yt-dlp.exe")

	repo := &fakeAppSettingsRepo{m: map[string]string{}}
	settings := NewSettingsUseCase(repo)
	uc := NewYTDLPMaintainUseCase(settings, ts.Updater)
	uc.CachePathOverride = cache

	ctx := context.Background()
	res, err := uc.CheckAndUpdateOfficialCache(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.Outcome != "firstDownload" {
		t.Fatalf("outcome %q", res.Outcome)
	}
	if res.PreviousVersion != "" {
		t.Fatalf("previous %q", res.PreviousVersion)
	}
}

func TestIsYTDLPCacheOlderThan(t *testing.T) {
	t.Parallel()
	if !isYTDLPCacheOlderThan("2026.01.01", "2026.02.02") {
		t.Fatal("expected older")
	}
	if isYTDLPCacheOlderThan("2026.02.02", "2026.02.02") {
		t.Fatal("same version should not download")
	}
	if !isYTDLPCacheOlderThan("", "2026.02.02") {
		t.Fatal("empty cache should download")
	}
}

func TestApplyOfficialCacheUpdate_rereadsMaintainAfterDownload(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Tools symlink maintain path is Windows-only")
	}
	gate := newYTDlpBlockingExeGate()
	ts := startYTDlpReleaseTestServerBlocking("2026.06.06", "MZblock", gate)
	defer ts.Server.Close()

	dir := t.TempDir()
	cache := filepath.Join(dir, "yt-dlp.exe")
	tools := filepath.Join(dir, "Tools", "yt-dlp.exe")
	if err := os.MkdirAll(filepath.Dir(tools), 0o755); err != nil {
		t.Fatal(err)
	}

	repo := &fakeAppSettingsRepo{m: map[string]string{}}
	settings := NewSettingsUseCase(repo)
	uc := NewYTDLPMaintainUseCase(settings, ts.Updater)
	uc.CachePathOverride = cache
	uc.ToolsPathOverride = tools
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
	info, err := ts.Updater.FetchLatestRelease(ctx)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	var updated YTDLPMaintainStatus
	var applyErr error
	go func() {
		updated, applyErr = uc.applyOfficialCacheUpdate(ctx, st, cache, tools, info)
		close(done)
	}()

	select {
	case <-gate.Started:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for blocked exe download")
	}
	if err := uc.SetMaintainDesired(ctx, false); err != nil {
		t.Fatal(err)
	}
	close(gate.Proceed)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for applyOfficialCacheUpdate")
	}
	if applyErr != nil {
		t.Fatalf("applyOfficialCacheUpdate: %v", applyErr)
	}
	if updated.MaintainDesired {
		t.Fatal("expected maintainDesired false after toggle")
	}
	eff, err := EffectiveOfficialLink(tools, cache)
	if err != nil {
		t.Fatal(err)
	}
	if eff {
		t.Fatal("expected Tools not linked when maintain turned off before link step")
	}
}

func TestApplyOfficialCacheUpdate_linksWhenMaintainEnabledDuringDownload(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Tools symlink maintain path is Windows-only")
	}
	gate := newYTDlpBlockingExeGate()
	ts := startYTDlpReleaseTestServerBlocking("2026.06.07", "MZblock2", gate)
	defer ts.Server.Close()

	dir := t.TempDir()
	cache := filepath.Join(dir, "yt-dlp.exe")
	tools := filepath.Join(dir, "Tools", "yt-dlp.exe")
	if err := os.MkdirAll(filepath.Dir(tools), 0o755); err != nil {
		t.Fatal(err)
	}

	repo := &fakeAppSettingsRepo{m: map[string]string{}}
	settings := NewSettingsUseCase(repo)
	uc := NewYTDLPMaintainUseCase(settings, ts.Updater)
	uc.CachePathOverride = cache
	uc.ToolsPathOverride = tools
	uc.UnlockTimeout = time.Second

	ctx := context.Background()
	if err := uc.AcknowledgeRisk(ctx); err != nil {
		t.Fatal(err)
	}
	st, err := uc.GetStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	info, err := ts.Updater.FetchLatestRelease(ctx)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	var updated YTDLPMaintainStatus
	var applyErr error
	go func() {
		updated, applyErr = uc.applyOfficialCacheUpdate(ctx, st, cache, tools, info)
		close(done)
	}()

	select {
	case <-gate.Started:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for blocked exe download")
	}
	if err := uc.AcknowledgeRisk(ctx); err != nil {
		t.Fatal(err)
	}
	if err := uc.SetMaintainDesired(ctx, true); err != nil {
		t.Fatal(err)
	}
	close(gate.Proceed)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for applyOfficialCacheUpdate")
	}
	if applyErr != nil {
		t.Fatalf("applyOfficialCacheUpdate: %v", applyErr)
	}
	if !updated.MaintainDesired || !updated.EffectiveOfficial {
		t.Fatalf("expected maintain on and linked, got %+v", updated)
	}
	eff, err := EffectiveOfficialLink(tools, cache)
	if err != nil {
		t.Fatal(err)
	}
	if !eff {
		t.Fatal("expected effective official link")
	}
}
