package usecase

import (
	"context"
	"errors"
	"fmt"
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
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if runtime.GOOS != "windows" {
		return ErrYTDLPUnsupportedPlatform
	}
	if uc.settings == nil {
		return errors.New("settings not configured")
	}
	if on {
		ack, err := uc.settings.GetYTDLPToolsReplaceRiskAck(ctx)
		if err != nil {
			return err
		}
		if !ack {
			return ErrYTDLPRiskAckRequired
		}
		if st, linkErr := uc.ensureAndLinkLocked(ctx); linkErr != nil {
			return linkErr
		} else if st.PendingError != "" {
			return errors.New(st.PendingError)
		}
		return uc.settings.SetYTDLPToolsReplaceMaintain(ctx, true)
	}
	return uc.settings.SetYTDLPToolsReplaceMaintain(ctx, false)
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
	defer uc.mu.Unlock()
	st, err := uc.getStatusLocked(ctx)
	if err != nil {
		return st, err
	}
	if !st.Supported {
		return st, ErrYTDLPUnsupportedPlatform
	}
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
	if err := uc.updater.DownloadToCache(ctx, st.CachePath, url); err != nil {
		return st, err
	}
	st.CachePresent = true
	st.CacheVersion = normalizeReleaseTag(tag)
	if st.CacheVersion == "" {
		st.CacheVersion = LocalYTDLPVersion(ctx, st.CachePath)
	}
	if st.MaintainDesired {
		if linkErr := uc.linkIfNeeded(ctx, st.ToolsPath, st.CachePath); linkErr != nil {
			st.PendingError = linkErr.Error()
			return st, nil
		}
		st.PendingError = ""
		st.EffectiveOfficial = true
	}
	return st, nil
}

// EnsureAndLink ensures cache exists and links Tools when needed. Records pending errors.
// If Tools is already a symlink to cache, it does not remove/recreate (re-enable after stop
// must not fail just because the existing correct link is briefly locked).
func (uc *YTDLPMaintainUseCase) EnsureAndLink(ctx context.Context) (YTDLPMaintainStatus, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.ensureAndLinkLocked(ctx)
}

func (uc *YTDLPMaintainUseCase) ensureAndLinkLocked(ctx context.Context) (YTDLPMaintainStatus, error) {
	st, err := uc.getStatusLocked(ctx)
	if err != nil {
		return st, err
	}
	if !st.Supported {
		return st, ErrYTDLPUnsupportedPlatform
	}
	if _, err := uc.updater.EnsureOfficialCache(ctx, st.CachePath); err != nil {
		return st, err
	}
	st.CachePresent = true
	st.CacheVersion = LocalYTDLPVersion(ctx, st.CachePath)
	if err := uc.linkIfNeeded(ctx, st.ToolsPath, st.CachePath); err != nil {
		st.PendingError = err.Error()
		return st, nil
	}
	st.PendingError = ""
	st.EffectiveOfficial = true
	return st, nil
}

// ReapplyIfNeeded links Tools when maintain is desired and the link is not effective.
func (uc *YTDLPMaintainUseCase) ReapplyIfNeeded(ctx context.Context) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if uc.settings == nil {
		return nil
	}
	on, err := uc.settings.GetYTDLPToolsReplaceMaintain(ctx)
	if err != nil || !on {
		return nil
	}
	tools, err := uc.toolsPath()
	if err != nil {
		return err
	}
	cache, err := uc.cachePath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(cache); err != nil {
		if _, e2 := uc.updater.EnsureOfficialCache(ctx, cache); e2 != nil {
			return e2
		}
	}
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
		st.MaintainDesired, _ = uc.settings.GetYTDLPToolsReplaceMaintain(ctx)
		st.RiskAcknowledged, _ = uc.settings.GetYTDLPToolsReplaceRiskAck(ctx)
		st.PendingError, _ = uc.settings.GetYTDLPToolsReplacePendingError(ctx)
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
// Frontend should resolve the key via i18n; fallback to error string for unknown errors.
func FormatMaintainError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, ErrYTDLPRiskAckRequired) {
		return "error_risk_ack_required"
	}
	if errors.Is(err, ErrYTDLPUnsupportedPlatform) {
		return "error_unsupported_platform"
	}
	return fmt.Sprintf("%v", err)
}
