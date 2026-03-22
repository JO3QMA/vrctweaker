package sqlite

import (
	"context"
	"database/sql"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

var _ activity.WorldInfoRepository = (*WorldInfoRepository)(nil)

// WorldInfoRepository implements activity.WorldInfoRepository.
type WorldInfoRepository struct {
	db *sql.DB
}

// NewWorldInfoRepository creates a WorldInfoRepository.
func NewWorldInfoRepository(db *sql.DB) *WorldInfoRepository {
	return &WorldInfoRepository{db: db}
}

// UpsertVisit sets or updates last_visited_at for a world.
func (r *WorldInfoRepository) UpsertVisit(ctx context.Context, worldID string, at time.Time) error {
	if worldID == "" {
		return nil
	}
	ts := at.Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `INSERT INTO world_info (world_id, display_name, last_visited_at) VALUES (?, NULL, ?)
		ON CONFLICT(world_id) DO UPDATE SET last_visited_at = excluded.last_visited_at`,
		worldID, ts)
	return err
}

// UpsertDisplayName updates display name and last_visited_at.
func (r *WorldInfoRepository) UpsertDisplayName(ctx context.Context, worldID, displayName string, at time.Time) error {
	if worldID == "" {
		return nil
	}
	ts := at.Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `INSERT INTO world_info (world_id, display_name, last_visited_at) VALUES (?, ?, ?)
		ON CONFLICT(world_id) DO UPDATE SET
			display_name = excluded.display_name,
			last_visited_at = excluded.last_visited_at`,
		worldID, displayName, ts)
	return err
}

// GetByWorldID returns world info or nil.
func (r *WorldInfoRepository) GetByWorldID(ctx context.Context, worldID string) (*activity.WorldInfo, error) {
	row := r.db.QueryRowContext(ctx, `SELECT world_id, display_name, last_visited_at FROM world_info WHERE world_id = ?`, worldID)
	var wid, lva string
	var dn sql.NullString
	if err := row.Scan(&wid, &dn, &lva); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	t, _ := time.Parse(time.RFC3339, lva)
	name := ""
	if dn.Valid {
		name = dn.String
	}
	return &activity.WorldInfo{WorldID: wid, DisplayName: name, LastVisitedAt: t}, nil
}
