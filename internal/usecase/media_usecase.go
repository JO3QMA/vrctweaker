package usecase

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"vrchat-tweaker/internal/domain/media"
)

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

// ScanDirectory scans a directory for screenshots and indexes them.
func (uc *MediaUseCase) ScanDirectory(ctx context.Context, basePath string) (int, error) {
	basePath = filepath.Clean(basePath)
	info, err := os.Stat(basePath)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return 0, nil
	}

	count := 0
	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		switch ext {
		case ".png", ".jpg", ".jpeg":
			existing, _ := uc.repo.GetByFilePath(ctx, path)
			if existing != nil {
				return nil
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
			s := &media.Screenshot{
				ID:        uuid.New().String(),
				FilePath:  path,
				WorldID:   worldID,
				WorldName: worldName,
				TakenAt:   takenAt,
			}
			if err := uc.repo.Save(ctx, s); err != nil {
				return nil
			}
			count++
		}
		return nil
	})
	return count, err
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
	if uc.extractor == nil {
		return 0, nil
	}
	updated := 0
	for _, s := range list {
		worldID, worldName, takenAt, _ := uc.extractor.Extract(s.FilePath)
		changed := worldID != s.WorldID || worldName != s.WorldName
		if takenAt != nil && (s.TakenAt == nil || !takenAt.Equal(*s.TakenAt)) {
			changed = true
		}
		if !changed {
			continue
		}
		s.WorldID = worldID
		s.WorldName = worldName
		if takenAt != nil {
			s.TakenAt = takenAt
		}
		if err := uc.repo.Save(ctx, s); err != nil {
			continue
		}
		updated++
	}
	return updated, nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
