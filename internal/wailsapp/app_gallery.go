package wailsapp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/domain/media"
	"vrchat-tweaker/internal/infrastructure/desktop"
	"vrchat-tweaker/internal/usecase"
)

const galleryScanProgressEvent = "gallery:scan-progress"

// galleryScanDoneEvent is emitted when ScanScreenshotDir finishes (success, error, or cancel).
const galleryScanDoneEvent = "gallery:scan-done"

// galleryScreenshotsChangedEvent is emitted when the picture folder watcher ingests a new screenshot row.
const galleryScreenshotsChangedEvent = "gallery:screenshots-changed"

const galleryScanProgressEmitMinInterval = 90 * time.Millisecond

// scanProgressEmitter throttles gallery:scan-progress EventsEmit; flush sends the latest pending payload.
type scanProgressEmitter struct {
	ctx        context.Context
	mu         sync.Mutex
	lastEmit   time.Time
	pending    usecase.ScanProgress
	hasPending bool
}

func (e *scanProgressEmitter) emit(p usecase.ScanProgress) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pending = p
	e.hasPending = true
	if time.Since(e.lastEmit) >= galleryScanProgressEmitMinInterval {
		runtime.EventsEmit(e.ctx, galleryScanProgressEvent, e.pending)
		e.lastEmit = time.Now()
		e.hasPending = false
	}
}

func (e *scanProgressEmitter) flush() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.hasPending {
		return
	}
	runtime.EventsEmit(e.ctx, galleryScanProgressEvent, e.pending)
	e.lastEmit = time.Now()
	e.hasPending = false
}

// Screenshots returns screenshots (optional worldId filter) within the current picture folder.
func (a *App) Screenshots(worldId string) ([]ScreenshotDTO, error) {
	filter := &media.ScreenshotFilter{}
	if worldId != "" {
		filter.WorldID = worldId
	}
	return a.listGalleryScreenshotDTOs(filter)
}

// SearchScreenshots returns screenshots matching the filter (worldId, worldName, dateFrom, dateTo).
func (a *App) SearchScreenshots(filter ScreenshotSearchDTO) ([]ScreenshotDTO, error) {
	f := toScreenshotFilter(filter)
	return a.listGalleryScreenshotDTOs(f)
}

func (a *App) listGalleryScreenshotDTOs(filter *media.ScreenshotFilter) ([]ScreenshotDTO, error) {
	root := a.resolveVRChatPictureWatchRoot()
	list, err := a.media.ListScreenshotsInGalleryScope(a.ctx, root, filter)
	if err != nil {
		return nil, err
	}
	return toScreenshotDTOs(list), nil
}

// GetScreenshot returns a screenshot by ID, or nil if not found.
func (a *App) GetScreenshot(id string) (*ScreenshotDTO, error) {
	s, err := a.media.GetScreenshot(a.ctx, id)
	if err != nil || s == nil {
		return nil, err
	}
	return toScreenshotDTO(s), nil
}

// GetGalleryAutoEnrichMetadata returns whether screenshot ingest auto-enrichment is enabled.
func (a *App) GetGalleryAutoEnrichMetadata() (bool, error) {
	return a.settings.GetGalleryAutoEnrichMetadata(a.ctx)
}

// SetGalleryAutoEnrichMetadata saves the screenshot ingest auto-enrichment toggle.
func (a *App) SetGalleryAutoEnrichMetadata(enabled bool) error {
	return a.settings.SetGalleryAutoEnrichMetadata(a.ctx, enabled)
}

func (a *App) retryScreenshotEnrichment() {
	if a.media == nil {
		return
	}
	ctx := a.appContext()
	if _, err := a.media.RetryPendingEnrichments(ctx); err != nil {
		runtime.LogWarning(ctx, "screenshot enrichment retry: "+err.Error())
	}
}

// ScreenshotThumbnailDataURL returns a JPEG data URL for the screenshot thumbnail (for WebView; avoids file://).
// Uses the DB cache when valid; otherwise builds and stores a thumbnail from the source file (lazy fill for legacy rows).
func (a *App) ScreenshotThumbnailDataURL(id string) (string, error) {
	return a.media.ScreenshotThumbnailDataURL(a.ctx, id)
}

// OpenScreenshotExternally opens the screenshot file with the OS default application.
func (a *App) OpenScreenshotExternally(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("screenshot id is empty")
	}
	s, err := a.media.GetScreenshot(a.ctx, id)
	if err != nil {
		return err
	}
	if s == nil {
		return fmt.Errorf("screenshot not found")
	}
	return desktop.OpenFileWithDefaultApp(s.FilePath)
}

// RevealScreenshotInFileManager opens the file manager and shows the screenshot file where supported.
func (a *App) RevealScreenshotInFileManager(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("screenshot id is empty")
	}
	s, err := a.media.GetScreenshot(a.ctx, id)
	if err != nil {
		return err
	}
	if s == nil {
		return fmt.Errorf("screenshot not found")
	}
	return desktop.RevealInFileManager(s.FilePath)
}

// ScanScreenshotDir synchronizes a picture folder (ingest new files + selective reindex).
func (a *App) ScanScreenshotDir(path string) (int, error) {
	a.galleryScanMu.Lock()
	for a.galleryScanCancel != nil || a.galleryScanDone != nil {
		cancel := a.galleryScanCancel
		done := a.galleryScanDone
		a.galleryScanMu.Unlock()
		if cancel != nil {
			cancel()
		}
		if done != nil {
			<-done
		}
		a.galleryScanMu.Lock()
	}
	done := make(chan struct{})
	scanCtx, cancel := context.WithCancel(a.ctx)
	a.galleryScanCancel = cancel
	a.galleryScanDone = done
	a.galleryScanMu.Unlock()

	defer func() {
		cancel()
		close(done)
		a.galleryScanMu.Lock()
		if a.galleryScanDone == done {
			a.galleryScanDone = nil
			a.galleryScanCancel = nil
		}
		a.galleryScanMu.Unlock()
	}()

	em := &scanProgressEmitter{ctx: a.ctx}
	var count int
	var err error
	defer func() {
		em.flush()
		dto := usecase.GalleryScanDone{Count: count}
		if err != nil {
			dto.Error = err.Error()
			if errors.Is(err, context.Canceled) {
				dto.Cancelled = true
			}
		}
		runtime.EventsEmit(a.ctx, galleryScanDoneEvent, dto)
	}()

	count, err = a.media.SyncPictureFolder(scanCtx, path, em.emit)
	if err == nil {
		a.retryScreenshotEnrichment()
	}
	return count, err
}

// IsGalleryScanning reports whether a gallery folder scan is in progress.
func (a *App) IsGalleryScanning() bool {
	a.galleryScanMu.Lock()
	defer a.galleryScanMu.Unlock()
	return a.galleryScanCancel != nil
}

func (a *App) stopGalleryScanAndWait() {
	a.galleryScanMu.Lock()
	cancel := a.galleryScanCancel
	done := a.galleryScanDone
	if cancel != nil || done != nil {
		// Keep fields until done is closed to let concurrent Scan wait.
		a.galleryScanMu.Unlock()
		if cancel != nil {
			cancel()
		}
		if done != nil {
			<-done
		}
		a.galleryScanMu.Lock()
		if a.galleryScanDone == done {
			a.galleryScanDone = nil
			a.galleryScanCancel = nil
		}
	}
	a.galleryScanMu.Unlock()
}

// ReindexScreenshotDir re-extracts metadata for existing screenshots under path.
// Returns the number of updated records.
func (a *App) ReindexScreenshotDir(path string) (int, error) {
	return a.media.ReindexScreenshots(a.ctx, path)
}

func (a *App) startupGalleryIncremental() {
	const waitTick = 50 * time.Millisecond
	const waitMaxIters = 3600 // ~3 minutes while a manual folder scan is in progress

	for i := 0; i < waitMaxIters && a.IsGalleryScanning(); i++ {
		select {
		case <-a.ctx.Done():
			return
		case <-time.After(waitTick):
		}
	}
	if a.IsGalleryScanning() {
		runtime.LogWarning(a.ctx, "startup gallery incremental: skipped (folder scan still running)")
		return
	}

	root := a.resolveVRChatPictureWatchRoot()
	if root == "" {
		return
	}
	if _, err := os.Stat(root); err != nil {
		return
	}

	since, ok := a.settings.GetGalleryLastExitAt(a.ctx)
	if !ok {
		return
	}

	count, err := a.media.IngestUnderPictureRootSince(a.ctx, root, since)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		runtime.LogWarning(a.ctx, "startup gallery incremental: "+err.Error())
		return
	}
	if count > 0 {
		runtime.EventsEmit(a.ctx, galleryScreenshotsChangedEvent, struct{}{})
	}
}
