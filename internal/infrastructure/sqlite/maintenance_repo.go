package sqlite

import (
	"context"
	"database/sql"

	"vrchat-tweaker/internal/domain/maintenance"
)

var _ maintenance.Repository = (*MaintenanceRepository)(nil)

// MaintenanceRepository provides DB maintenance operations (VACUUM).
type MaintenanceRepository struct {
	db *sql.DB
}

// NewMaintenanceRepository creates a new MaintenanceRepository.
func NewMaintenanceRepository(db *sql.DB) *MaintenanceRepository {
	return &MaintenanceRepository{db: db}
}

// Vacuum runs SQLite VACUUM to reclaim space and optimize the database.
// Note: VACUUM cannot run inside a transaction; it runs in autocommit mode.
func (r *MaintenanceRepository) Vacuum(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `VACUUM`)
	return err
}
