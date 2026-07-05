package usecase

import (
	"context"
	"os"

	"vrchat-tweaker/internal/domain/media"
)

// ListScreenshotsInGalleryScope returns screenshots for Gallery listing: limited to
// pictureFolderRoot and excluding rows whose files are missing on disk.
// When pictureFolderRoot is empty, returns an empty list.
func (uc *MediaUseCase) ListScreenshotsInGalleryScope(ctx context.Context, pictureFolderRoot string, filter *media.ScreenshotFilter) ([]*media.Screenshot, error) {
	prefix := media.PictureFolderPathPrefix(pictureFolderRoot)
	if prefix == "" {
		return nil, nil
	}
	scoped := cloneScreenshotFilter(filter)
	scoped.FilePathPrefix = prefix
	list, err := uc.repo.List(ctx, scoped)
	if err != nil {
		return nil, err
	}
	return filterScreenshotsWithExistingFiles(list), nil
}

func cloneScreenshotFilter(f *media.ScreenshotFilter) *media.ScreenshotFilter {
	if f == nil {
		return &media.ScreenshotFilter{}
	}
	cp := *f
	return &cp
}

func filterScreenshotsWithExistingFiles(list []*media.Screenshot) []*media.Screenshot {
	if len(list) == 0 {
		return list
	}
	out := make([]*media.Screenshot, 0, len(list))
	for _, s := range list {
		if s == nil {
			continue
		}
		if screenshotFileExists(s.FilePath) {
			out = append(out, s)
		}
	}
	return out
}

func screenshotFileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}
