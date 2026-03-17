package sqlite

import (
	"context"
	"database/sql"

	"vrchat-tweaker/internal/domain/media"
)

var _ media.ScreenshotRepository = (*ScreenshotRepository)(nil)

// ScreenshotRepository implements media.ScreenshotRepository.
type ScreenshotRepository struct {
	db *sql.DB
}

// NewScreenshotRepository creates a new ScreenshotRepository.
func NewScreenshotRepository(db *sql.DB) *ScreenshotRepository {
	return &ScreenshotRepository{db: db}
}

// List returns screenshots with optional filters.
func (r *ScreenshotRepository) List(ctx context.Context, filter *media.ScreenshotFilter) ([]*media.Screenshot, error) {
	query := `SELECT id, file_path, world_id, world_name, taken_at FROM screenshots WHERE 1=1`
	args := []interface{}{}
	if filter != nil {
		if filter.WorldID != "" {
			query += ` AND world_id = ?`
			args = append(args, filter.WorldID)
		}
		if filter.FromDate != nil {
			query += ` AND taken_at >= ?`
			args = append(args, filter.FromDate.Format("2006-01-02T15:04:05Z07:00"))
		}
		if filter.ToDate != nil {
			query += ` AND taken_at <= ?`
			args = append(args, filter.ToDate.Format("2006-01-02T15:04:05Z07:00"))
		}
		if filter.WorldName != "" {
			query += ` AND world_name LIKE ?`
			args = append(args, "%"+filter.WorldName+"%")
		}
		if filter.FilePathPrefix != "" {
			query += ` AND (file_path LIKE ? OR file_path = ?)`
			prefix := filter.FilePathPrefix + "%"
			args = append(args, prefix, filter.FilePathPrefix)
		}
	}
	query += ` ORDER BY taken_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var list []*media.Screenshot
	for rows.Next() {
		s, err := scanScreenshot(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

// GetByID returns a screenshot by ID.
func (r *ScreenshotRepository) GetByID(ctx context.Context, id string) (*media.Screenshot, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, file_path, world_id, world_name, taken_at FROM screenshots WHERE id = ?`, id)
	return scanScreenshotRow(row)
}

// GetByFilePath returns a screenshot by file path.
func (r *ScreenshotRepository) GetByFilePath(ctx context.Context, filePath string) (*media.Screenshot, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, file_path, world_id, world_name, taken_at FROM screenshots WHERE file_path = ?`, filePath)
	return scanScreenshotRow(row)
}

// Save persists a screenshot.
func (r *ScreenshotRepository) Save(ctx context.Context, s *media.Screenshot) error {
	takenAt := nullableTime(s.TakenAt)
	_, err := r.db.ExecContext(ctx, `INSERT INTO screenshots (id, file_path, world_id, world_name, taken_at)
		VALUES (?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET
		file_path = excluded.file_path, world_id = excluded.world_id, world_name = excluded.world_name, taken_at = excluded.taken_at`,
		s.ID, s.FilePath, s.WorldID, s.WorldName, takenAt)
	return err
}

// Delete removes a screenshot.
func (r *ScreenshotRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM screenshots WHERE id = ?`, id)
	return err
}

// DeleteAll removes all screenshots.
func (r *ScreenshotRepository) DeleteAll(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM screenshots`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func scanScreenshot(rows *sql.Rows) (*media.Screenshot, error) {
	var id, filePath, worldID, worldName string
	var takenAt sql.NullString
	if err := rows.Scan(&id, &filePath, &worldID, &worldName, &takenAt); err != nil {
		return nil, err
	}
	return &media.Screenshot{
		ID:        id,
		FilePath:  filePath,
		WorldID:   worldID,
		WorldName: worldName,
		TakenAt:   parseTime(takenAt),
	}, nil
}

func scanScreenshotRow(row *sql.Row) (*media.Screenshot, error) {
	var id, filePath, worldID, worldName string
	var takenAt sql.NullString
	err := row.Scan(&id, &filePath, &worldID, &worldName, &takenAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &media.Screenshot{
		ID:        id,
		FilePath:  filePath,
		WorldID:   worldID,
		WorldName: worldName,
		TakenAt:   parseTime(takenAt),
	}, nil
}
