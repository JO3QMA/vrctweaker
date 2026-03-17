package launcher

import "context"

// LaunchProfileRepository defines persistence operations for launch profiles.
type LaunchProfileRepository interface {
	// List returns all launch profiles.
	List(ctx context.Context) ([]*LaunchProfile, error)
	// GetByID returns a launch profile by ID.
	GetByID(ctx context.Context, id string) (*LaunchProfile, error)
	// GetDefault returns the default launch profile, or nil if none is set.
	GetDefault(ctx context.Context) (*LaunchProfile, error)
	// Save persists a launch profile (create or update).
	Save(ctx context.Context, p *LaunchProfile) error
	// Delete removes a launch profile by ID.
	Delete(ctx context.Context, id string) error
}
