package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
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
		st.UnsupportedReason = "この機能は Windows 版のみ利用できます。"
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
	if on {
		ack, err := uc.settings.GetYTDLPToolsReplaceRiskAck(ctx)
		if err != nil {
			return err
		}
		if !ack {
			return ErrYTDLPRiskAckRequired
		}
		if setErr := uc.settings.SetYTDLPToolsReplaceMaintain(ctx, true); setErr != nil {
			return setErr
		}
		_, err = uc.EnsureAndLink(ctx)
		return err
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
	st, err := uc.GetStatus(ctx)
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
		if linkErr := uc.linkAndRecord(ctx, st.ToolsPath, st.CachePath); linkErr != nil {
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
	st, err := uc.GetStatus(ctx)
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
	if err := uc.linkAndRecord(ctx, st.ToolsPath, st.CachePath); err != nil {
		st.PendingError = err.Error()
		return st, nil
	}
	st.PendingError = ""
	st.EffectiveOfficial = true
	return st, nil
}

// ReapplyIfNeeded links Tools when maintain is desired and the link is not effective.
func (uc *YTDLPMaintainUseCase) ReapplyIfNeeded(ctx context.Context) error {
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
	need, err := NeedsOfficialLink(tools, cache)
	if err != nil {
		return err
	}
	if !need {
		_ = uc.settings.SetYTDLPToolsReplacePendingError(ctx, "")
		return nil
	}
	if _, err := os.Stat(cache); err != nil {
		if _, e2 := uc.updater.EnsureOfficialCache(ctx, cache); e2 != nil {
			return e2
		}
	}
	return uc.linkAndRecord(ctx, tools, cache)
}

func (uc *YTDLPMaintainUseCase) linkAndRecord(ctx context.Context, tools, cache string) error {
	err := LinkToolsToCache(tools, cache, uc.unlockTimeout())
	if uc.settings == nil {
		return err
	}
	if err != nil {
		_ = uc.settings.SetYTDLPToolsReplacePendingError(ctx, err.Error())
		return err
	}
	return uc.settings.SetYTDLPToolsReplacePendingError(ctx, "")
}

// FormatMaintainError returns a user-facing message for maintain API errors.
func FormatMaintainError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, ErrYTDLPRiskAckRequired) {
		return "初回有効化の前にリスク確認が必要です。"
	}
	if errors.Is(err, ErrYTDLPUnsupportedPlatform) {
		return "この機能は Windows 版のみ利用できます。"
	}
	return fmt.Sprintf("%v", err)
}
