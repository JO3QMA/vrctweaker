package sqlite

import (
	"context"
	"database/sql"

	"vrchat-tweaker/internal/domain/media"
)

var _ media.ScreenshotRepository = (*ScreenshotRepository)(nil)

const screenshotSelectBase = `SELECT s.id, s.file_path, COALESCE(s.world_id, ''),
	COALESCE(w.display_name, '') AS world_name_resolved,
	COALESCE(s.author_vrc_user_id, ''),
	COALESCE(u.display_name, '') AS author_display_name,
	s.taken_at, s.file_size_bytes
	FROM screenshots s
	LEFT JOIN world_info w ON w.world_id = s.world_id
	LEFT JOIN users_cache u ON u.vrc_user_id = s.author_vrc_user_id`

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
	query := screenshotSelectBase + ` WHERE 1=1`
	args := []interface{}{}
	if filter != nil {
		if filter.WorldID != "" {
			query += ` AND s.world_id = ?`
			args = append(args, filter.WorldID)
		}
		if filter.FromDate != nil {
			query += ` AND s.taken_at >= ?`
			args = append(args, filter.FromDate.Format("2006-01-02T15:04:05Z07:00"))
		}
		if filter.ToDate != nil {
			query += ` AND s.taken_at <= ?`
			args = append(args, filter.ToDate.Format("2006-01-02T15:04:05Z07:00"))
		}
		if filter.WorldName != "" {
			pat := "%" + filter.WorldName + "%"
			query += ` AND COALESCE(w.display_name, '') LIKE ?`
			args = append(args, pat)
		}
		if filter.FilePathPrefix != "" {
			query += ` AND (s.file_path LIKE ? OR s.file_path = ?)`
			prefix := filter.FilePathPrefix + "%"
			args = append(args, prefix, filter.FilePathPrefix)
		}
	}
	query += ` ORDER BY s.taken_at DESC`

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
	row := r.db.QueryRowContext(ctx, screenshotSelectBase+` WHERE s.id = ?`, id)
	return scanScreenshotRow(row)
}

// GetByFilePath returns a screenshot by file path.
func (r *ScreenshotRepository) GetByFilePath(ctx context.Context, filePath string) (*media.Screenshot, error) {
	row := r.db.QueryRowContext(ctx, screenshotSelectBase+` WHERE s.file_path = ?`, filePath)
	return scanScreenshotRow(row)
}

// Save persists a screenshot.
func (r *ScreenshotRepository) Save(ctx context.Context, s *media.Screenshot) error {
	takenAt := nullableTime(s.TakenAt)
	_, err := r.db.ExecContext(ctx, `INSERT INTO screenshots (id, file_path, world_id, author_vrc_user_id, taken_at, file_size_bytes)
		VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET
		file_path = excluded.file_path, world_id = excluded.world_id,
		author_vrc_user_id = excluded.author_vrc_user_id, taken_at = excluded.taken_at, file_size_bytes = excluded.file_size_bytes`,
		s.ID, s.FilePath, s.WorldID, nullString(s.AuthorVRCUserID), takenAt, nullableInt64(s.FileSizeBytes))
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

// GetThumbnail returns cached thumbnail bytes (JPEG) or nil if none.
func (r *ScreenshotRepository) GetThumbnail(ctx context.Context, screenshotID string) (*media.ScreenshotThumbnail, error) {
	row := r.db.QueryRowContext(ctx, `SELECT jpeg_blob, source_size, source_mod_unix FROM screenshot_thumbnails WHERE screenshot_id = ?`, screenshotID)
	var blob []byte
	var size, modUnix int64
	err := row.Scan(&blob, &size, &modUnix)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &media.ScreenshotThumbnail{
		JpegBlob:      blob,
		SourceSize:    size,
		SourceModUnix: modUnix,
	}, nil
}

// UpsertThumbnail stores or replaces the thumbnail for a screenshot.
func (r *ScreenshotRepository) UpsertThumbnail(ctx context.Context, screenshotID string, thumb *media.ScreenshotThumbnail) error {
	if thumb == nil {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO screenshot_thumbnails (screenshot_id, jpeg_blob, source_size, source_mod_unix)
		VALUES (?, ?, ?, ?) ON CONFLICT(screenshot_id) DO UPDATE SET
		jpeg_blob = excluded.jpeg_blob, source_size = excluded.source_size, source_mod_unix = excluded.source_mod_unix`,
		screenshotID, thumb.JpegBlob, thumb.SourceSize, thumb.SourceModUnix)
	return err
}

// DeleteThumbnail removes the cached thumbnail for a screenshot.
func (r *ScreenshotRepository) DeleteThumbnail(ctx context.Context, screenshotID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM screenshot_thumbnails WHERE screenshot_id = ?`, screenshotID)
	return err
}

func scanScreenshot(rows *sql.Rows) (*media.Screenshot, error) {
	var id, filePath, worldID, worldNameResolved, authorID, authorName string
	var takenAt sql.NullString
	var fileSize sql.NullInt64
	if err := rows.Scan(&id, &filePath, &worldID, &worldNameResolved, &authorID, &authorName, &takenAt, &fileSize); err != nil {
		return nil, err
	}
	return &media.Screenshot{
		ID:                id,
		FilePath:          filePath,
		WorldID:           worldID,
		WorldName:         worldNameResolved,
		AuthorVRCUserID:   authorID,
		AuthorDisplayName: authorName,
		TakenAt:           parseTime(takenAt),
		FileSizeBytes:     parseInt64Ptr(fileSize),
	}, nil
}

func scanScreenshotRow(row *sql.Row) (*media.Screenshot, error) {
	var id, filePath, worldID, worldNameResolved, authorID, authorName string
	var takenAt sql.NullString
	var fileSize sql.NullInt64
	err := row.Scan(&id, &filePath, &worldID, &worldNameResolved, &authorID, &authorName, &takenAt, &fileSize)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &media.Screenshot{
		ID:                id,
		FilePath:          filePath,
		WorldID:           worldID,
		WorldName:         worldNameResolved,
		AuthorVRCUserID:   authorID,
		AuthorDisplayName: authorName,
		TakenAt:           parseTime(takenAt),
		FileSizeBytes:     parseInt64Ptr(fileSize),
	}, nil
}
