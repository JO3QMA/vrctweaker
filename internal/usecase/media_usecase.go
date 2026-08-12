package usecase

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/domain/media"
)

const (
	// ScanPhaseListing is reported while collecting image paths under the tree.
	ScanPhaseListing = "listing"
	// ScanPhaseImporting is reported while ingesting collected paths.
	ScanPhaseImporting = "importing"
)

// How often to emit listing progress (every N images found) during filepath.Walk.
const scanListingProgressEvery = 50

// ScanProgress is a snapshot for UI progress (optional callback from ScanDirectory).
type ScanProgress struct {
	Phase   string `json:"phase"`
	Current int    `json:"current"`
	Total   int    `json:"total"`
	Item    string `json:"item,omitempty"`
}

// GalleryScanDone is emitted on Wails event gallery:scan-done when a folder scan finishes.
type GalleryScanDone struct {
	Count     int    `json:"count"`
	Error     string `json:"error,omitempty"`
	Cancelled bool   `json:"cancelled"`
}

// extractScreenshotMetadata reads embedded screenshot metadata.
// Tests may replace this.
var extractScreenshotMetadata = media.Extract

// MediaUseCase handles screenshot scanning and management.
type MediaUseCase struct {
	repo          screenshotRepo
	worldRepo     worldInfoRepo
	userCacheRepo userCacheRepo
	fileExists    *galleryFileExistsCache
}

// NewMediaUseCase creates a new MediaUseCase.
// worldRepo and userCacheRepo may be nil; when set, extracted metadata is upserted into world_info and users_cache.
func NewMediaUseCase(repo screenshotRepo, worldRepo worldInfoRepo, userCacheRepo userCacheRepo) *MediaUseCase {
	return &MediaUseCase{
		repo:          repo,
		worldRepo:     worldRepo,
		userCacheRepo: userCacheRepo,
		fileExists:    newGalleryFileExistsCache(),
	}
}

func (uc *MediaUseCase) upsertWorldInfo(ctx context.Context, worldID, worldName string, at time.Time) {
	if uc.worldRepo == nil || worldID == "" {
		return
	}
	if worldName != "" {
		_ = uc.worldRepo.UpsertDisplayName(ctx, worldID, worldName, at)
		return
	}
	_ = uc.worldRepo.UpsertVisit(ctx, worldID, at)
}

func (uc *MediaUseCase) upsertAuthorFromScreenshot(ctx context.Context, vrcUserID, displayName string, at time.Time) {
	if uc.userCacheRepo == nil || vrcUserID == "" || displayName == "" {
		return
	}
	existing, err := uc.userCacheRepo.GetByVRCUserID(ctx, vrcUserID)
	if err != nil {
		return
	}
	if existing == nil {
		existing = &identity.UserCache{VRCUserID: vrcUserID, UserKind: identity.UserKindContact}
	}
	existing.MergeFromLog(displayName, at)
	_ = uc.userCacheRepo.Save(ctx, existing)
}

// ListScreenshots returns screenshots with optional filters.
func (uc *MediaUseCase) ListScreenshots(ctx context.Context, filter *media.ScreenshotFilter) ([]*media.Screenshot, error) {
	return uc.repo.List(ctx, filter)
}

// GetScreenshot returns a screenshot by ID.
func (uc *MediaUseCase) GetScreenshot(ctx context.Context, id string) (*media.Screenshot, error) {
	return uc.repo.GetByID(ctx, id)
}

// IngestScreenshotFile registers a single image file if it is new (by path).
// Returns the screenshot row, whether it was newly created, and an error only for
// persistence/stat failures. Thumbnail generation errors are ignored so the row stays saved.
func (uc *MediaUseCase) IngestScreenshotFile(ctx context.Context, path string) (*media.Screenshot, bool, error) {
	return uc.ingestScreenshotFile(ctx, path, true)
}

// ingestScreenshotFile is the shared implementation. invalidateCache is true for
// single-file ingestion (picture-folder watcher), so a restored/new file is
// reflected by the next listing immediately. Bulk flows (ScanDirectory,
// IngestUnderPictureRootSince, SyncPictureFolder) pass false and invalidate the
// whole cache once at the end: a per-file generation bump during a long sync
// would otherwise discard every concurrent listing's cache fill.
func (uc *MediaUseCase) ingestScreenshotFile(ctx context.Context, path string, invalidateCache bool) (*media.Screenshot, bool, error) {
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return nil, false, err
	}
	if !info.Mode().IsRegular() {
		return nil, false, nil
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg":
	default:
		return nil, false, nil
	}

	if invalidateCache {
		// The file is confirmed on disk; drop any cached missing/exists result so
		// the next Gallery listing re-checks it (covers restored and new files).
		uc.fileExists.invalidatePath(path)
	}

	existing, _ := uc.repo.GetByFilePath(ctx, path)
	if existing != nil {
		return existing, false, nil
	}

	takenAt := timePtr(info.ModTime())
	meta, exErr := extractScreenshotMetadata(path)
	if exErr != nil {
		meta = media.ScreenshotMetadata{}
	}
	if meta.TakenAt != nil {
		takenAt = meta.TakenAt
	}
	sz := info.Size()
	s := &media.Screenshot{
		ID:              uuid.New().String(),
		FilePath:        path,
		WorldID:         meta.WorldID,
		AuthorVRCUserID: meta.AuthorVRCUserID,
		WorldName:       meta.WorldDisplayName,
		TakenAt:         takenAt,
		FileSizeBytes:   &sz,
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return nil, false, err
	}
	at := info.ModTime()
	if s.TakenAt != nil {
		at = *s.TakenAt
	}
	uc.upsertWorldInfo(ctx, meta.WorldID, meta.WorldDisplayName, at)
	uc.upsertAuthorFromScreenshot(ctx, meta.AuthorVRCUserID, meta.AuthorDisplayName, at)
	_ = uc.EnsureScreenshotThumbnail(ctx, s.ID)
	return s, true, nil
}

// ScanDirectory scans a directory for screenshots and indexes them.
// onProgress is optional; when non-nil it receives listing/importing snapshots.
func (uc *MediaUseCase) ScanDirectory(ctx context.Context, basePath string, onProgress func(ScanProgress)) (int, error) {
	count, _, err := uc.ingestImagePathsInDir(ctx, basePath, onProgress)
	if err != nil {
		return count, err
	}
	uc.fileExists.invalidateAll()
	return count, nil
}

// IngestUnderPictureRootSince walks basePath for image files whose ModTime is strictly after since
// and ingests new paths only (see IngestScreenshotFile). Skips files on ingest error, like ScanDirectory.
func (uc *MediaUseCase) IngestUnderPictureRootSince(ctx context.Context, basePath string, since time.Time) (int, error) {
	basePath = filepath.Clean(basePath)
	info, err := os.Stat(basePath)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return 0, nil
	}

	createdCount := 0
	err = filepath.Walk(basePath, func(path string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if fi.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".png", ".jpg", ".jpeg":
		default:
			return nil
		}
		if !fi.ModTime().After(since) {
			return nil
		}
		_, created, ingestErr := uc.ingestScreenshotFile(ctx, path, false)
		if ingestErr != nil {
			return nil
		}
		if created {
			createdCount++
		}
		return nil
	})
	if err != nil {
		return createdCount, err
	}
	uc.fileExists.invalidateAll()
	return createdCount, nil
}

// DeleteScreenshot removes a screenshot record.
func (uc *MediaUseCase) DeleteScreenshot(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

// ReindexScreenshots re-extracts metadata for existing screenshots under basePath and updates them.
// Returns the number of updated records.
func (uc *MediaUseCase) ReindexScreenshots(ctx context.Context, basePath string) (int, error) {
	return uc.reindexScreenshotsUnderPath(ctx, basePath, nil, nil)
}

func timePtr(t time.Time) *time.Time {
	return &t
}
