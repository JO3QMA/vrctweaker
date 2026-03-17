package media

import (
	"context"
	"time"
)

// ScreenshotRepository defines persistence operations for screenshots.
type ScreenshotRepository interface {
	// List returns screenshots with optional filters.
	List(ctx context.Context, filter *ScreenshotFilter) ([]*Screenshot, error)
	// GetByID returns a screenshot by ID.
	GetByID(ctx context.Context, id string) (*Screenshot, error)
	// GetByFilePath returns a screenshot by file path.
	GetByFilePath(ctx context.Context, filePath string) (*Screenshot, error)
	// Save persists a screenshot (create or update).
	Save(ctx context.Context, s *Screenshot) error
	// Delete removes a screenshot by ID.
	Delete(ctx context.Context, id string) error
	// DeleteAll removes all screenshots. Returns affected row count.
	DeleteAll(ctx context.Context) (int64, error)
}

// ScreenshotFilter provides optional filtering for List.
type ScreenshotFilter struct {
	WorldID        string
	FromDate       *time.Time
	ToDate         *time.Time
	WorldName      string
	FilePathPrefix string // filters file_path starting with this prefix (for directory-scoped queries)
}
