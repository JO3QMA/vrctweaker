package wailsapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/infrastructure/logwatcher"
	"vrchat-tweaker/internal/infrastructure/picturewatcher"
)

func (a *App) resolveVRChatPictureWatchRoot() string {
	cfg, err := a.vrchatConfigRepo.Read()
	if err == nil {
		if p := strings.TrimSpace(cfg.PictureOutputFolder); p != "" {
			return filepath.Clean(p)
		}
	}
	p, err := a.DefaultVRChatPictureFolder()
	if err != nil {
		return ""
	}
	return filepath.Clean(p)
}

func (a *App) startPictureFolderWatcher(ctx context.Context) {
	root := a.resolveVRChatPictureWatchRoot()
	if root == "" {
		runtime.LogWarning(ctx, "picture folder watcher: could not resolve VRChat picture directory")
		return
	}
	info, err := os.Stat(root)
	if err != nil {
		runtime.LogWarning(ctx, fmt.Sprintf("picture folder watcher: stat %s: %v", root, err))
		return
	}
	if !info.IsDir() {
		runtime.LogWarning(ctx, "picture folder watcher: not a directory: "+root)
		return
	}
	ingest := func(c context.Context, path string) error {
		_, created, ingestErr := a.media.IngestScreenshotFile(c, path)
		if ingestErr != nil {
			return ingestErr
		}
		if created {
			runtime.EventsEmit(a.ctx, galleryScreenshotsChangedEvent, struct{}{})
		}
		return nil
	}
	log := pictureWatchLogger{ctx: ctx}
	if err := picturewatcher.Start(ctx, root, ingest, logwatcher.Logger(log.Printf)); err != nil {
		runtime.LogError(ctx, "failed to start picture folder watcher: "+err.Error())
		return
	}
	runtime.LogInfo(ctx, "picture folder watcher started for "+root)
}

type pictureWatchLogger struct {
	ctx context.Context
}

func (l pictureWatchLogger) Printf(format string, v ...any) {
	runtime.LogWarning(l.ctx, fmt.Sprintf(format, v...))
}
