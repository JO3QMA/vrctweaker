package usecase

import (
	"context"
	"time"

	"vrchat-tweaker/internal/appversion"
)

const ytdlpLatestReleaseCacheTTL = 6 * time.Hour

// YTDLPMaintainActionResult is returned by CheckAndUpdateOfficialCache (Video tab primary action).
type YTDLPMaintainActionResult struct {
	Status          YTDLPMaintainStatus `json:"status"`
	Outcome         string              `json:"outcome"` // firstDownload, updated, unchanged, error
	PreviousVersion string              `json:"previousVersion"`
	NewVersion      string              `json:"newVersion"`
	DetailError     string              `json:"detailError,omitempty"`
}

// RefreshLatestRelease returns GitHub latest metadata, using a persisted cache unless force is true.
func (uc *YTDLPMaintainUseCase) RefreshLatestRelease(ctx context.Context, force bool) (YTDLPMaintainStatus, error) {
	st, err := uc.GetStatus(ctx)
	if err != nil {
		return st, err
	}
	if !st.Supported {
		return st, nil
	}
	if !force && uc.latestReleaseCacheFresh(ctx, st) {
		return st, nil
	}
	return uc.CheckLatest(ctx)
}

func (uc *YTDLPMaintainUseCase) latestReleaseCacheFresh(ctx context.Context, st YTDLPMaintainStatus) bool {
	if appversion.NormalizeProductVersion(st.LatestVersion) == "" {
		return false
	}
	if uc.settings == nil {
		return false
	}
	checked, ok := uc.settings.GetYTDLPKnownLatestCheckedAt(ctx)
	if !ok {
		return false
	}
	return time.Since(checked) < ytdlpLatestReleaseCacheTTL
}

func isYTDLPCacheOlderThan(current, latest string) bool {
	current = appversion.NormalizeProductVersion(current)
	latest = appversion.NormalizeProductVersion(latest)
	if latest == "" {
		return false
	}
	if current == "" {
		return true
	}
	return appversion.CompareProductVersion(current, latest) < 0
}

// CheckAndUpdateOfficialCache checks GitHub for the latest release and downloads when newer (or cache missing).
func (uc *YTDLPMaintainUseCase) CheckAndUpdateOfficialCache(ctx context.Context) (YTDLPMaintainActionResult, error) {
	res := YTDLPMaintainActionResult{}
	st, err := uc.GetStatus(ctx)
	if err != nil {
		return res, err
	}
	if !st.Supported {
		return res, ErrYTDLPUnsupportedPlatform
	}
	prev := appversion.NormalizeProductVersion(st.CacheVersion)
	res.PreviousVersion = prev

	info, ferr := uc.updater.FetchLatestRelease(ctx)
	if ferr != nil {
		st.LatestError = ferr.Error()
		res.Status = st
		res.Outcome = "error"
		res.DetailError = ferr.Error()
		return res, nil
	}

	st = enrichYTDLPMaintainRelease(ctx, st, "", info.Tag, info.DownloadURL, info.Version)
	st.LatestError = ""
	uc.mu.Lock()
	if err := uc.persistKnownLatest(ctx, st); err != nil {
		uc.mu.Unlock()
		return res, err
	}
	if uc.settings != nil {
		if err := uc.settings.SetYTDLPKnownLatestCheckedAt(ctx, time.Now()); err != nil {
			uc.mu.Unlock()
			return res, err
		}
	}
	uc.mu.Unlock()

	needDownload := !st.CachePresent || isYTDLPCacheOlderThan(st.CacheVersion, st.LatestVersion)
	if !needDownload {
		fresh, gerr := uc.GetStatus(ctx)
		if gerr != nil {
			return res, gerr
		}
		fresh.LatestVersion = st.LatestVersion
		fresh.LatestTag = st.LatestTag
		fresh.LatestDownloadURL = st.LatestDownloadURL
		fresh.LatestError = ""
		res.Status = fresh
		res.Outcome = "unchanged"
		res.NewVersion = st.LatestVersion
		return res, nil
	}

	updated, uerr := uc.UpdateOfficialCache(ctx, info.DownloadURL, info.Tag)
	res.Status = updated
	if uerr != nil {
		res.Outcome = "error"
		res.DetailError = uerr.Error()
		if updated.LatestError != "" {
			res.DetailError = updated.LatestError
		}
		return res, nil
	}
	res.NewVersion = appversion.NormalizeProductVersion(updated.CacheVersion)
	if prev == "" {
		res.Outcome = "firstDownload"
	} else {
		res.Outcome = "updated"
	}
	return res, nil
}
