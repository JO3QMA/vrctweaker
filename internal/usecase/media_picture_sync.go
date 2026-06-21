package usecase

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"vrchat-tweaker/internal/domain/media"
)

// SyncPictureFolder ingests new screenshots under basePath and selectively reindexes
// existing rows (metadata when world_id is empty or the file changed; thumbnails when stale).
// Returns the number of newly ingested plus updated rows.
func (uc *MediaUseCase) SyncPictureFolder(ctx context.Context, basePath string, onProgress func(ScanProgress)) (int, error) {
	ingested, created, err := uc.ingestImagePathsInDir(ctx, basePath, onProgress)
	if err != nil {
		return ingested, err
	}
	updated, err := uc.reindexScreenshotsUnderPath(ctx, basePath, created, nil)
	if err != nil {
		return ingested, err
	}
	return ingested + updated, nil
}

// ingestImagePathsInDir walks basePath for images and ingests new paths only.
func (uc *MediaUseCase) ingestImagePathsInDir(ctx context.Context, basePath string, onProgress func(ScanProgress)) (int, map[string]struct{}, error) {
	basePath = filepath.Clean(basePath)
	created := make(map[string]struct{})
	info, err := os.Stat(basePath)
	if err != nil {
		return 0, created, err
	}
	if !info.IsDir() {
		return 0, created, nil
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
		return 0, created, err
	}

	if onProgress != nil {
		onProgress(ScanProgress{Phase: ScanPhaseListing, Current: len(paths), Total: 0})
		onProgress(ScanProgress{Phase: ScanPhaseImporting, Current: 0, Total: len(paths), Item: ""})
	}

	count := 0
	for i, path := range paths {
		if ctx.Err() != nil {
			return count, created, ctx.Err()
		}
		_, wasCreated, ingestErr := uc.IngestScreenshotFile(ctx, path)
		if ingestErr != nil {
			// Same as ScanDirectory: skip file on ingest error.
		} else if wasCreated {
			count++
			created[filepath.Clean(path)] = struct{}{}
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
	return count, created, nil
}

// reindexScreenshotsUnderPath updates existing screenshots under basePath.
// Paths in skipIngested are skipped (freshly ingested in the same sync).
func (uc *MediaUseCase) reindexScreenshotsUnderPath(ctx context.Context, basePath string, skipIngested map[string]struct{}, onProgress func(ScanProgress)) (int, error) {
	basePath = filepath.Clean(basePath)
	prefix := media.PictureFolderPathPrefix(basePath)
	filter := &media.ScreenshotFilter{FilePathPrefix: prefix}
	list, err := uc.repo.List(ctx, filter)
	if err != nil {
		return 0, err
	}
	updated := 0
	for _, s := range list {
		if ctx.Err() != nil {
			return updated, ctx.Err()
		}
		if skipIngested != nil {
			if _, skip := skipIngested[filepath.Clean(s.FilePath)]; skip {
				continue
			}
		}
		didUpdate, err := uc.reindexScreenshotFile(ctx, s)
		if err != nil {
			return updated, err
		}
		if didUpdate {
			updated++
		}
	}
	return updated, nil
}

func (uc *MediaUseCase) reindexScreenshotFile(ctx context.Context, s *media.Screenshot) (bool, error) {
	info, err := os.Stat(s.FilePath)
	if err != nil {
		return false, nil
	}
	size := info.Size()
	modUnix := info.ModTime().Unix()

	thumb, errThumb := uc.repo.GetThumbnail(ctx, s.ID)
	if errThumb != nil {
		return false, errThumb
	}
	thumbnailStale := thumb == nil
	if thumb != nil && (thumb.SourceSize != size || thumb.SourceModUnix != modUnix) {
		if err := uc.repo.DeleteThumbnail(ctx, s.ID); err != nil {
			return false, err
		}
		thumbnailStale = true
	}

	sizeChanged := s.FileSizeBytes == nil || *s.FileSizeBytes != size
	needsMetaExtract := s.WorldID == "" || sizeChanged || thumbnailStale

	metaChanged := false
	var meta media.ScreenshotMetadata
	if needsMetaExtract && uc.extractor != nil {
		var exErr error
		meta, exErr = uc.extractor.Extract(s.FilePath)
		if exErr != nil {
			meta = media.ScreenshotMetadata{}
		}
		metaChanged = meta.WorldID != s.WorldID || meta.AuthorVRCUserID != s.AuthorVRCUserID
		if meta.TakenAt != nil && (s.TakenAt == nil || !meta.TakenAt.Equal(*s.TakenAt)) {
			metaChanged = true
		}
		if metaChanged {
			s.WorldID = meta.WorldID
			s.AuthorVRCUserID = meta.AuthorVRCUserID
			if meta.TakenAt != nil {
				s.TakenAt = meta.TakenAt
			}
			s.WorldName = meta.WorldDisplayName
		}
	}

	if !sizeChanged && !metaChanged && !thumbnailStale {
		return false, nil
	}
	fsz := size
	s.FileSizeBytes = &fsz
	if err := uc.repo.Save(ctx, s); err != nil {
		return false, nil
	}
	if metaChanged {
		at := info.ModTime()
		if s.TakenAt != nil {
			at = *s.TakenAt
		}
		uc.upsertWorldInfo(ctx, meta.WorldID, meta.WorldDisplayName, at)
		uc.upsertAuthorFromScreenshot(ctx, meta.AuthorVRCUserID, meta.AuthorDisplayName, at)
	}
	_ = uc.EnsureScreenshotThumbnail(ctx, s.ID)
	return true, nil
}
