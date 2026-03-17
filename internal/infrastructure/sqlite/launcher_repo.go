package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"vrchat-tweaker/internal/domain/launcher"
)

// Ensure LauncherProfileRepository implements launcher.LaunchProfileRepository.
var _ launcher.LaunchProfileRepository = (*LauncherProfileRepository)(nil)

// LauncherProfileRepository implements launcher.LaunchProfileRepository.
type LauncherProfileRepository struct {
	db *sql.DB
}

// NewLauncherProfileRepository creates a new LauncherProfileRepository.
func NewLauncherProfileRepository(db *sql.DB) *LauncherProfileRepository {
	return &LauncherProfileRepository{db: db}
}

// List returns all launch profiles.
func (r *LauncherProfileRepository) List(ctx context.Context) ([]*launcher.LaunchProfile, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, arguments, is_default, created_at, updated_at FROM launch_profiles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var list []*launcher.LaunchProfile
	for rows.Next() {
		p, err := scanLaunchProfile(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

// GetByID returns a launch profile by ID.
func (r *LauncherProfileRepository) GetByID(ctx context.Context, id string) (*launcher.LaunchProfile, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, arguments, is_default, created_at, updated_at FROM launch_profiles WHERE id = ?`, id)
	return scanLaunchProfileRow(row)
}

// GetDefault returns the default launch profile.
func (r *LauncherProfileRepository) GetDefault(ctx context.Context) (*launcher.LaunchProfile, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, arguments, is_default, created_at, updated_at FROM launch_profiles WHERE is_default = 1 LIMIT 1`)
	return scanLaunchProfileRow(row)
}

// Save persists a launch profile.
func (r *LauncherProfileRepository) Save(ctx context.Context, p *launcher.LaunchProfile) error {
	isDefault := 0
	if p.IsDefault {
		isDefault = 1
	}
	createdAt, updatedAt := nullableTime(p.CreatedAt), nullableTime(p.UpdatedAt)

	_, err := r.db.ExecContext(ctx, `INSERT INTO launch_profiles (id, name, arguments, is_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET
		name = excluded.name, arguments = excluded.arguments, is_default = excluded.is_default, updated_at = excluded.updated_at`,
		p.ID, p.Name, p.Arguments, isDefault, createdAt, updatedAt)
	if err != nil {
		return err
	}

	if p.IsDefault {
		_, _ = r.db.ExecContext(ctx, `UPDATE launch_profiles SET is_default = 0 WHERE id != ?`, p.ID)
	}
	return nil
}

// Delete removes a launch profile.
func (r *LauncherProfileRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM launch_profiles WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("launch profile not found: %s", id)
	}
	return nil
}

func scanLaunchProfile(rows *sql.Rows) (*launcher.LaunchProfile, error) {
	var id, name, arguments string
	var isDefault int
	var createdAt, updatedAt sql.NullString
	if err := rows.Scan(&id, &name, &arguments, &isDefault, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	return &launcher.LaunchProfile{
		ID:        id,
		Name:      name,
		Arguments: arguments,
		IsDefault: isDefault == 1,
		CreatedAt: parseTime(createdAt),
		UpdatedAt: parseTime(updatedAt),
	}, nil
}

func scanLaunchProfileRow(row *sql.Row) (*launcher.LaunchProfile, error) {
	var id, name, arguments string
	var isDefault int
	var createdAt, updatedAt sql.NullString
	err := row.Scan(&id, &name, &arguments, &isDefault, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &launcher.LaunchProfile{
		ID:        id,
		Name:      name,
		Arguments: arguments,
		IsDefault: isDefault == 1,
		CreatedAt: parseTime(createdAt),
		UpdatedAt: parseTime(updatedAt),
	}, nil
}
