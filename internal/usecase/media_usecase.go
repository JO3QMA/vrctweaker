package usecase

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
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
	Phase   string
	Current int
	Total   int
	Item    string
}

// MediaUseCase handles screenshot scanning and management.
type MediaUseCase struct {
	repo      media.ScreenshotRepository
	extractor media.MetadataExtractor
}

// NewMediaUseCase creates a new MediaUseCase.
func NewMediaUseCase(repo media.ScreenshotRepository, extractor media.MetadataExtractor) *MediaUseCase {
	return &MediaUseCase{repo: repo, extractor: extractor}
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

	existing, _ := uc.repo.GetByFilePath(ctx, path)
	if existing != nil {
		return existing, false, nil
	}

	takenAt := timePtr(info.ModTime())
	worldID, worldName := "", ""
	if uc.extractor != nil {
		wid, wn, ta, _ := uc.extractor.Extract(path)
		worldID, worldName = wid, wn
		if ta != nil {
			takenAt = ta
		}
	}
	sz := info.Size()
	s := &media.Screenshot{
		ID:            uuid.New().String(),
		FilePath:      path,
		WorldID:       worldID,
		WorldName:     worldName,
		TakenAt:       takenAt,
		FileSizeBytes: &sz,
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return nil, false, err
	}
	_ = uc.EnsureScreenshotThumbnail(ctx, s.ID)
	return s, true, nil
}

// ScanDirectory scans a directory for screenshots and indexes them.
// onProgress is optional; when non-nil it receives listing/importing snapshots.
func (uc *MediaUseCase) ScanDirectory(ctx context.Context, basePath string, onProgress func(ScanProgress)) (int, error) {
	basePath = filepath.Clean(basePath)
	info, err := os.Stat(basePath)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return 0, nil
	}

	var paths []string
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
		paths = append(paths, path)
		if onProgress != nil && len(paths)%scanListingProgressEvery == 0 {
			onProgress(ScanProgress{Phase: ScanPhaseListing, Current: len(paths), Total: 0})
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	if onProgress != nil {
		onProgress(ScanProgress{Phase: ScanPhaseListing, Current: len(paths), Total: 0})
		onProgress(ScanProgress{Phase: ScanPhaseImporting, Current: 0, Total: len(paths), Item: ""})
	}

	count := 0
	for i, path := range paths {
		if ctx.Err() != nil {
			return count, ctx.Err()
		}
		_, created, ingestErr := uc.IngestScreenshotFile(ctx, path)
		if ingestErr != nil {
			// Same as previous Walk behavior: skip file on ingest error.
		} else if created {
			count++
		}
		if onProgress != nil {
			onProgress(ScanProgress{
				Phase:   ScanPhaseImporting,
				Current: i + 1,
				Total:   len(paths),
				Item:    filepath.Base(path),
			})
		}
	}
	return count, nil
}

// DeleteScreenshot removes a screenshot record.
func (uc *MediaUseCase) DeleteScreenshot(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

// ReindexScreenshots re-extracts metadata for existing screenshots under basePath and updates them.
// Returns the number of updated records.
func (uc *MediaUseCase) ReindexScreenshots(ctx context.Context, basePath string) (int, error) {
	basePath = filepath.Clean(basePath)
	prefix := basePath + string(filepath.Separator)
	filter := &media.ScreenshotFilter{FilePathPrefix: prefix}
	list, err := uc.repo.List(ctx, filter)
	if err != nil {
		return 0, err
	}
	updated := 0
	for _, s := range list {
		info, err := os.Stat(s.FilePath)
		if err != nil {
			continue
		}
		size := info.Size()
		modUnix := info.ModTime().Unix()

		thumb, errThumb := uc.repo.GetThumbnail(ctx, s.ID)
		if errThumb != nil {
			return 0, errThumb
		}
		thumbnailStale := thumb == nil
		if thumb != nil && (thumb.SourceSize != size || thumb.SourceModUnix != modUnix) {
			if err := uc.repo.DeleteThumbnail(ctx, s.ID); err != nil {
				return 0, err
			}
			thumbnailStale = true
		}

		sizeChanged := s.FileSizeBytes == nil || *s.FileSizeBytes != size
		metaChanged := false
		if uc.extractor != nil {
			worldID, worldName, takenAt, _ := uc.extractor.Extract(s.FilePath)
			metaChanged = worldID != s.WorldID || worldName != s.WorldName
			if takenAt != nil && (s.TakenAt == nil || !takenAt.Equal(*s.TakenAt)) {
				metaChanged = true
			}
			if metaChanged {
				s.WorldID = worldID
				s.WorldName = worldName
				if takenAt != nil {
					s.TakenAt = takenAt
				}
			}
		}
		if !sizeChanged && !metaChanged && !thumbnailStale {
			continue
		}
		fsz := size
		s.FileSizeBytes = &fsz
		if err := uc.repo.Save(ctx, s); err != nil {
			continue
		}
		_ = uc.EnsureScreenshotThumbnail(ctx, s.ID)
		updated++
	}
	return updated, nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
