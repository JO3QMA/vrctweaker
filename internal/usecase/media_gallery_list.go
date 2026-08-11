package usecase

import (
	"context"
	"os"

	"vrchat-tweaker/internal/domain/media"
)

// ListScreenshotsInGalleryScope returns screenshots for Gallery listing: limited to
// pictureFolderRoot and excluding rows whose files are missing on disk.
// When pictureFolderRoot is empty, returns an empty list.
//
// Missing-file exclusion is memoized with a short TTL (galleryFileStatCacheTTL),
// so external deletions/restorations are reflected only after an invalidation
// event (sync or ingest) or once the TTL expires; within the TTL a listing may
// briefly include a just-deleted file or exclude a just-restored one.
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
	return uc.filterScreenshotsWithExistingFiles(list), nil
}

func cloneScreenshotFilter(f *media.ScreenshotFilter) *media.ScreenshotFilter {
	if f == nil {
		return &media.ScreenshotFilter{}
	}
	cp := *f
	return &cp
}

// filterScreenshotsWithExistingFiles excludes rows whose file is missing on disk.
// Existence checks are memoized for a short TTL (see galleryFileExistsCache) so
// repeated listings with many screenshots avoid a stat per row on every call.
func (uc *MediaUseCase) filterScreenshotsWithExistingFiles(list []*media.Screenshot) []*media.Screenshot {
	if len(list) == 0 {
		return list
	}
	out := make([]*media.Screenshot, 0, len(list))
	for _, s := range list {
		if s == nil {
			continue
		}
		if uc.fileExistsCheck(s.FilePath) {
			out = append(out, s)
		}
	}
	return out
}

func (uc *MediaUseCase) fileExistsCheck(path string) bool {
	return uc.fileExists.check(path)
}

// statScreenshotFile reports whether path is a regular file on disk, and whether
// the result is safe to cache. Only a definitively-missing file (IsNotExist) is
// cached as missing; other stat failures (permission, transient I/O) are retried
// on the next listing so a real file is not hidden for the whole TTL.
func statScreenshotFile(path string) (exists bool, cacheable bool) {
	info, err := os.Stat(path)
	if err != nil {
		return false, os.IsNotExist(err)
	}
	return info.Mode().IsRegular(), true
}
