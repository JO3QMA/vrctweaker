package wailsapp

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	goruntime "runtime"
	"time"

	"vrchat-tweaker/internal/infrastructure/desktop"
	"vrchat-tweaker/internal/infrastructure/sleepsuppress"
	"vrchat-tweaker/internal/infrastructure/ytdlpmaintain"
	"vrchat-tweaker/internal/usecase"
)

type ytdlpMaintainSettingGetter struct {
	a *App
}

func (g ytdlpMaintainSettingGetter) YTDLPToolsReplaceMaintain(ctx context.Context) (bool, error) {
	return g.a.settings.GetYTDLPToolsReplaceMaintain(ctx)
}

type ytdlpMaintainReapplier struct {
	a *App
}

func (r ytdlpMaintainReapplier) ReapplyIfNeeded(ctx context.Context) error {
	return r.a.ytdlp.ReapplyIfNeeded(ctx)
}

type ytdlpToolsDirProvider struct {
	a *App
}

func (p ytdlpToolsDirProvider) ToolsDir() (string, error) {
	if p.a.ytdlp != nil {
		return p.a.ytdlp.ToolsDir()
	}
	tools, err := usecase.VRChatYTDLPToolsPath()
	if err != nil {
		return "", err
	}
	return filepath.Dir(tools), nil
}

func (a *App) startYTDLPMaintainLoop() {
	a.ytdlpMaintainMu.Lock()
	if a.ytdlp == nil || a.settings == nil {
		a.ytdlpMaintainMu.Unlock()
		return
	}
	if a.ytdlpMaintainCancel != nil {
		cancel := a.ytdlpMaintainCancel
		a.ytdlpMaintainCancel = nil
		a.ytdlpMaintainMu.Unlock()
		cancel()
		a.ytdlpMaintainWG.Wait()
		a.ytdlpMaintainMu.Lock()
	}
	runCtx, cancel := context.WithCancel(context.Background())
	a.ytdlpMaintainCancel = cancel
	a.ytdlpMaintainWG.Add(1)
	a.ytdlpMaintainMu.Unlock()

	go func() {
		defer a.ytdlpMaintainWG.Done()
		checker := sleepsuppress.NewVRChatProcessChecker()
		if err := ytdlpmaintain.Run(
			runCtx,
			2*time.Second,
			ytdlpMaintainSettingGetter{a: a},
			checker,
			ytdlpMaintainReapplier{a: a},
			ytdlpToolsDirProvider{a: a},
		); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("ytdlp maintain loop: %v", err)
		}
	}()
}

func (a *App) stopYTDLPMaintainLoop() {
	var cancel context.CancelFunc
	a.ytdlpMaintainMu.Lock()
	cancel = a.ytdlpMaintainCancel
	a.ytdlpMaintainCancel = nil
	a.ytdlpMaintainMu.Unlock()
	if cancel != nil {
		cancel()
		a.ytdlpMaintainWG.Wait()
	}
}

// RuntimeIsWindows reports whether the backend process is running on Windows.
func (a *App) RuntimeIsWindows() bool {
	return goruntime.GOOS == "windows"
}

// GetYTDLPMaintainStatus returns desired/effective Tools replace maintain state (no GitHub call).
func (a *App) GetYTDLPMaintainStatus() (usecase.YTDLPMaintainStatus, error) {
	if a.ytdlp == nil {
		return usecase.YTDLPMaintainStatus{
			Supported:         false,
			UnsupportedReason: "notInitialized",
		}, nil
	}
	return a.ytdlp.GetStatus(a.ctx)
}

// AcknowledgeYTDLPToolsReplaceRisk records first-enable risk acknowledgment.
func (a *App) AcknowledgeYTDLPToolsReplaceRisk() error {
	if a.ytdlp == nil {
		return errors.New("notInitialized")
	}
	return a.ytdlp.AcknowledgeRisk(a.ctx)
}

// SetYTDLPToolsReplaceMaintain enables or disables maintain (enable requires risk ack).
func (a *App) SetYTDLPToolsReplaceMaintain(on bool) error {
	if a.ytdlp == nil {
		return errors.New("notInitialized")
	}
	return usecase.WrapMaintainAPIError(a.ytdlp.SetMaintainDesired(a.ctx, on))
}

// CheckYTDLPLatestRelease queries GitHub for the latest official yt-dlp.exe.
func (a *App) CheckYTDLPLatestRelease() (usecase.YTDLPMaintainStatus, error) {
	if a.ytdlp == nil {
		return usecase.YTDLPMaintainStatus{
			Supported:         false,
			UnsupportedReason: "notInitialized",
		}, nil
	}
	return a.ytdlp.CheckLatest(a.ctx)
}

// UpdateOfficialYTDLPCache downloads latest (or given URL) into Official cache; re-links when maintain is on.
func (a *App) UpdateOfficialYTDLPCache(downloadURL, latestTag string) (usecase.YTDLPMaintainStatus, error) {
	if a.ytdlp == nil {
		return usecase.YTDLPMaintainStatus{
			Supported:         false,
			UnsupportedReason: "notInitialized",
		}, nil
	}
	return a.ytdlp.UpdateOfficialCache(a.ctx, downloadURL, latestTag)
}

// OpenYTDLPCacheFolder opens the Official yt-dlp cache directory in the file manager.
func (a *App) OpenYTDLPCacheFolder() error {
	return openYTDLPFolderInFileManager(usecase.OfficialYTDLPCachePath)
}

// OpenYTDLPToolsFolder opens VRChat's Tools directory (parent of yt-dlp.exe) in the file manager.
func (a *App) OpenYTDLPToolsFolder() error {
	return openYTDLPFolderInFileManager(usecase.VRChatYTDLPToolsPath)
}

// GetYTDLPCookieLinkageStatus returns Cookie linkage effective state (yt-dlp user config).
func (a *App) GetYTDLPCookieLinkageStatus() (usecase.CookieLinkageStatus, error) {
	if a.cookieLinkage == nil {
		return usecase.CookieLinkageStatus{
			Supported:         false,
			UnsupportedReason: "notInitialized",
			SourceKind:        usecase.CookieSourceNone,
		}, nil
	}
	return a.cookieLinkage.GetStatus(a.ctx)
}

// AcknowledgeYTDLPCookieLinkageRisk records Cookie linkage risk acknowledgment.
func (a *App) AcknowledgeYTDLPCookieLinkageRisk() error {
	if a.cookieLinkage == nil {
		return errors.New("notInitialized")
	}
	return a.cookieLinkage.AcknowledgeRisk(a.ctx)
}

// SetYTDLPCookieLinkageBrowser upserts Browser cookie source (chrome/edge/firefox).
func (a *App) SetYTDLPCookieLinkageBrowser(browser string) error {
	if a.cookieLinkage == nil {
		return errors.New("notInitialized")
	}
	return usecase.WrapCookieLinkageAPIError(a.cookieLinkage.SetBrowserSource(a.ctx, browser))
}

// SetYTDLPCookieLinkageCookiesFile upserts Cookies file source.
func (a *App) SetYTDLPCookieLinkageCookiesFile(path string) error {
	if a.cookieLinkage == nil {
		return errors.New("notInitialized")
	}
	return usecase.WrapCookieLinkageAPIError(a.cookieLinkage.SetCookiesFileSource(a.ctx, path))
}

// DisableYTDLPCookieLinkage removes Managed cookie options from yt-dlp user config.
func (a *App) DisableYTDLPCookieLinkage() error {
	if a.cookieLinkage == nil {
		return errors.New("notInitialized")
	}
	return usecase.WrapCookieLinkageAPIError(a.cookieLinkage.Disable(a.ctx))
}

func openYTDLPFolderInFileManager(pathResolver func() (string, error)) error {
	path, err := pathResolver()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
		return mkErr
	}
	return desktop.OpenFolderInFileManager(dir)
}
