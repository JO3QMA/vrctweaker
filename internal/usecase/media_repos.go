package usecase

import (
	"context"

	"vrchat-tweaker/internal/domain/media"
)

// ponytail:#129 domain ScreenshotRepository removed; boundary stays usecase-local.
type screenshotRepo interface {
	List(ctx context.Context, filter *media.ScreenshotFilter) ([]*media.Screenshot, error)
	GetByID(ctx context.Context, id string) (*media.Screenshot, error)
	GetByFilePath(ctx context.Context, filePath string) (*media.Screenshot, error)
	Save(ctx context.Context, s *media.Screenshot) error
	Delete(ctx context.Context, id string) error
	DeleteAll(ctx context.Context) (int64, error)
	GetThumbnail(ctx context.Context, screenshotID string) (*media.ScreenshotThumbnail, error)
	UpsertThumbnail(ctx context.Context, screenshotID string, thumb *media.ScreenshotThumbnail) error
	DeleteThumbnail(ctx context.Context, screenshotID string) error
}
