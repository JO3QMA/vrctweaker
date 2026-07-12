package usecase

import (
	"context"
	"errors"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

const defaultYTDLPUnlockTimeout = 60 * time.Second

var (
	// ErrYTDLPRiskAckRequired is returned when enabling maintain without risk acknowledgment.
	ErrYTDLPRiskAckRequired = errors.New("tools replace risk acknowledgment required")
	// ErrYTDLPUnsupportedPlatform is returned on non-Windows.
	ErrYTDLPUnsupportedPlatform = errors.New("yt-dlp Tools replace maintain is Windows only")
	// ErrYTDLPCacheMissing is returned when Official cache is absent before linking.
	ErrYTDLPCacheMissing = errors.New("official yt-dlp cache missing")
)

// YTDLPMaintainStatus is desired + effective state for the Video tab.
type YTDLPMaintainStatus struct {
	Supported         bool   `json:"supported"`
	UnsupportedReason string `json:"unsupportedReason,omitempty"`
	MaintainDesired   bool   `json:"maintainDesired"`
	RiskAcknowledged  bool   `json:"riskAcknowledged"`
	EffectiveOfficial bool   `json:"effectiveOfficial"`
	CachePresent      bool   `json:"cachePresent"`
	CacheVersion      string `json:"cacheVersion"`
	ToolsPath         string `json:"toolsPath"`
	CachePath         string `json:"cachePath"`
	PendingError      string `json:"pendingError"`
	LatestVersion     string `json:"latestVersion"`
	LatestTag         string `json:"latestTag"`
	LatestDownloadURL string `json:"latestDownloadUrl"`
	LatestError       string `json:"latestError"`
}

// YTDLPMaintainUseCase orchestrates Official cache, Tools symlink, and maintain settings.
type YTDLPMaintainUseCase struct {
	mu            sync.Mutex
	settings      *SettingsUseCase
	updater       *YTDLPUpdater
	UnlockTimeout time.Duration
	// Path overrides for tests (empty → env-derived defaults).
	ToolsPathOverride string
	CachePathOverride string
}

// NewYTDLPMaintainUseCase wires settings and the GitHub updater.
func NewYTDLPMaintainUseCase(settings *SettingsUseCase, updater *YTDLPUpdater) *YTDLPMaintainUseCase {
	if updater == nil {
		updater = NewYTDLPUpdater()
	}
	return &YTDLPMaintainUseCase{
		settings:      settings,
		updater:       updater,
		UnlockTimeout: defaultYTDLPUnlockTimeout,
	}
}

func (uc *YTDLPMaintainUseCase) toolsPath() (string, error) {
	if uc.ToolsPathOverride != "" {
		return uc.ToolsPathOverride, nil
	}
	return VRChatYTDLPToolsPath()
}

func (uc *YTDLPMaintainUseCase) cachePath() (string, error) {
	if uc.CachePathOverride != "" {
		return uc.CachePathOverride, nil
	}
	return OfficialYTDLPCachePath()
}

func (uc *YTDLPMaintainUseCase) unlockTimeout() time.Duration {
	if uc.UnlockTimeout > 0 {
		return uc.UnlockTimeout
	}
	return defaultYTDLPUnlockTimeout
}

// GetStatus returns desired/effective state without calling GitHub.
func (uc *YTDLPMaintainUseCase) GetStatus(ctx context.Context) (YTDLPMaintainStatus, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.getStatusLocked(ctx)
}

// AcknowledgeRisk records Tools replace risk acknowledgment.
func (uc *YTDLPMaintainUseCase) AcknowledgeRisk(ctx context.Context) error {
	if uc.settings == nil {
		return errors.New("settings not configured")
	}
	return uc.settings.SetYTDLPToolsReplaceRiskAck(ctx, true)
}

// SetMaintainDesired turns maintain on/off. Enabling requires prior risk acknowledgment.
// On enable, ensures Official cache exists and attempts one Tools symlink (best-effort).
// On disable, only clears desired — Tools file is left untouched.
func (uc *YTDLPMaintainUseCase) SetMaintainDesired(ctx context.Context, on bool) error {
	if runtime.GOOS != "windows" {
		return ErrYTDLPUnsupportedPlatform
	}
	if uc.settings == nil {
		return errors.New("settings not configured")
	}
	if !on {
		uc.mu.Lock()
		defer uc.mu.Unlock()
		return uc.settings.SetYTDLPToolsReplaceMaintain(ctx, false)
	}

	uc.mu.Lock()
	ack, err := uc.settings.GetYTDLPToolsReplaceRiskAck(ctx)
	if err != nil {
		uc.mu.Unlock()
		return err
	}
	if !ack {
		uc.mu.Unlock()
		return ErrYTDLPRiskAckRequired
	}
	st, err := uc.getStatusLocked(ctx)
	if err != nil {
		uc.mu.Unlock()
		return err
	}
	if !st.Supported {
		uc.mu.Unlock()
		return ErrYTDLPUnsupportedPlatform
	}
	tools, cache := st.ToolsPath, st.CachePath
	uc.mu.Unlock()

	if err := uc.ensureCacheAndLink(ctx, tools, cache); err != nil {
		return err
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.settings.SetYTDLPToolsReplaceMaintain(ctx, true)
}

func (uc *YTDLPMaintainUseCase) ensureCacheAndLink(ctx context.Context, tools, cache string) error {
	if _, err := uc.updater.EnsureOfficialCache(ctx, cache); err != nil {
		return err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.linkIfNeeded(ctx, tools, cache)
}

// CheckLatest fills Latest* fields on a fresh status (GitHub API).
func (uc *YTDLPMaintainUseCase) CheckLatest(ctx context.Context) (YTDLPMaintainStatus, error) {
	st, err := uc.GetStatus(ctx)
	if err != nil {
		return st, err
	}
	if !st.Supported {
		return st, nil
	}
	info, ferr := uc.updater.FetchLatestRelease(ctx)
	if ferr != nil {
		st.LatestError = ferr.Error()
		return st, nil
	}
	st.LatestVersion = info.Version
	st.LatestTag = info.Tag
	st.LatestDownloadURL = info.DownloadURL
	return st, nil
}

// UpdateOfficialCache downloads latest into cache. When maintain is desired, re-links Tools.
func (uc *YTDLPMaintainUseCase) UpdateOfficialCache(ctx context.Context, downloadURL, expectedTag string) (YTDLPMaintainStatus, error) {
	uc.mu.Lock()
	st, err := uc.getStatusLocked(ctx)
	if err != nil {
		uc.mu.Unlock()
		return st, err
	}
	if !st.Supported {
		uc.mu.Unlock()
		return st, ErrYTDLPUnsupportedPlatform
	}
	cachePath, toolsPath := st.CachePath, st.ToolsPath
	maintainDesired := st.MaintainDesired
	uc.mu.Unlock()

	url := downloadURL
	tag := expectedTag
	if url == "" {
		info, ferr := uc.updater.FetchLatestRelease(ctx)
		if ferr != nil {
			st.LatestError = ferr.Error()
			return st, ferr
		}
		url = info.DownloadURL
		tag = info.Tag
		st.LatestVersion = info.Version
		st.LatestTag = info.Tag
		st.LatestDownloadURL = info.DownloadURL
	}
	if dlErr := uc.updater.DownloadToCache(ctx, cachePath, url); dlErr != nil {
		return st, dlErr
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()
	st, err = uc.getStatusLocked(ctx)
	if err != nil {
		return st, err
	}
	st.CachePresent = true
	st.CacheVersion = normalizeReleaseTag(tag)
	if st.CacheVersion == "" {
		st.CacheVersion = LocalYTDLPVersion(ctx, cachePath)
	}
	if maintainDesired {
		if linkErr := uc.linkIfNeeded(ctx, toolsPath, cachePath); linkErr != nil {
			st.PendingError = linkErr.Error()
			return st, nil
		}
		st.PendingError = ""
		st.EffectiveOfficial = true
	}
	return st, nil
}

// EnsureAndLink ensures cache exists and links Tools when needed. Records pending errors.
func (uc *YTDLPMaintainUseCase) EnsureAndLink(ctx context.Context) (YTDLPMaintainStatus, error) {
	uc.mu.Lock()
	st, err := uc.getStatusLocked(ctx)
	if err != nil {
		uc.mu.Unlock()
		return st, err
	}
	if !st.Supported {
		uc.mu.Unlock()
		return st, ErrYTDLPUnsupportedPlatform
	}
	tools, cache := st.ToolsPath, st.CachePath
	uc.mu.Unlock()

	if _, cacheErr := uc.updater.EnsureOfficialCache(ctx, cache); cacheErr != nil {
		return st, cacheErr
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()
	st, err = uc.getStatusLocked(ctx)
	if err != nil {
		return st, err
	}
	if linkErr := uc.linkIfNeeded(ctx, tools, cache); linkErr != nil {
		st.PendingError = linkErr.Error()
		return st, nil
	}
	st.PendingError = ""
	st.CachePresent = true
	st.CacheVersion = LocalYTDLPVersion(ctx, cache)
	st.EffectiveOfficial = true
	return st, nil
}

// ReapplyIfNeeded links Tools when maintain is desired and the link is not effective.
func (uc *YTDLPMaintainUseCase) ReapplyIfNeeded(ctx context.Context) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	uc.mu.Lock()
	if uc.settings == nil {
		uc.mu.Unlock()
		return nil
	}
	on, err := uc.settings.GetYTDLPToolsReplaceMaintain(ctx)
	if err != nil || !on {
		uc.mu.Unlock()
		return err
	}
	tools, err := uc.toolsPath()
	if err != nil {
		uc.mu.Unlock()
		return err
	}
	cache, err := uc.cachePath()
	if err != nil {
		uc.mu.Unlock()
		return err
	}
	uc.mu.Unlock()

	if _, err := os.Stat(cache); err != nil {
		if _, e2 := uc.updater.EnsureOfficialCache(ctx, cache); e2 != nil {
			return e2
		}
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.linkIfNeeded(ctx, tools, cache)
}

// linkIfNeeded creates/repairs the Tools symlink only when NeedsOfficialLink is true.
func (uc *YTDLPMaintainUseCase) linkIfNeeded(ctx context.Context, tools, cache string) error {
	need, err := NeedsOfficialLink(tools, cache)
	if err != nil {
		return err
	}
	if !need {
		if uc.settings != nil {
			if err := uc.settings.SetYTDLPToolsReplacePendingError(ctx, ""); err != nil {
				log.Printf("ytdlp maintain: clear pending error: %v", err)
			}
		}
		return nil
	}
	return uc.linkAndRecord(ctx, tools, cache)
}

func (uc *YTDLPMaintainUseCase) linkAndRecord(ctx context.Context, tools, cache string) error {
	err := LinkToolsToCache(tools, cache, uc.unlockTimeout())
	if uc.settings == nil {
		return err
	}
	if err != nil {
		if setErr := uc.settings.SetYTDLPToolsReplacePendingError(ctx, err.Error()); setErr != nil {
			log.Printf("ytdlp maintain: set pending error: %v", setErr)
		}
		return err
	}
	return uc.settings.SetYTDLPToolsReplacePendingError(ctx, "")
}

// getStatusLocked is GetStatus without acquiring uc.mu (caller must hold the lock).
func (uc *YTDLPMaintainUseCase) getStatusLocked(ctx context.Context) (YTDLPMaintainStatus, error) {
	st := YTDLPMaintainStatus{}
	tools, err := uc.toolsPath()
	if err != nil {
		st.Supported = false
		st.UnsupportedReason = err.Error()
		return st, nil
	}
	cache, err := uc.cachePath()
	if err != nil {
		st.Supported = false
		st.UnsupportedReason = err.Error()
		return st, nil
	}
	st.ToolsPath = tools
	st.CachePath = cache

	if runtime.GOOS != "windows" {
		st.Supported = false
		st.UnsupportedReason = "unsupported_platform"
		return st, nil
	}
	st.Supported = true

	if uc.settings != nil {
		var getErr error
		st.MaintainDesired, getErr = uc.settings.GetYTDLPToolsReplaceMaintain(ctx)
		if getErr != nil {
			log.Printf("ytdlp maintain: GetYTDLPToolsReplaceMaintain: %v", getErr)
		}
		st.RiskAcknowledged, getErr = uc.settings.GetYTDLPToolsReplaceRiskAck(ctx)
		if getErr != nil {
			log.Printf("ytdlp maintain: GetYTDLPToolsReplaceRiskAck: %v", getErr)
		}
		st.PendingError, getErr = uc.settings.GetYTDLPToolsReplacePendingError(ctx)
		if getErr != nil {
			log.Printf("ytdlp maintain: GetYTDLPToolsReplacePendingError: %v", getErr)
		}
	}

	if fi, stErr := os.Stat(cache); stErr == nil && !fi.IsDir() {
		st.CachePresent = true
		st.CacheVersion = LocalYTDLPVersion(ctx, cache)
	}
	eff, err := EffectiveOfficialLink(tools, cache)
	if err != nil {
		return st, err
	}
	st.EffectiveOfficial = eff
	return st, nil
}

// FormatMaintainError returns an i18n key for maintain API errors.
func FormatMaintainError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, ErrYTDLPRiskAckRequired) {
		return "errorRiskAckRequired"
	}
	if errors.Is(err, ErrYTDLPUnsupportedPlatform) {
		return "errorUnsupportedPlatform"
	}
	log.Printf("ytdlp maintain: %v", err)
	return "errorMaintainFailed"
}
