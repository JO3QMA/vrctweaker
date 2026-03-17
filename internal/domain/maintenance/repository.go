package maintenance

import "context"

// Repository defines DB maintenance operations.
type Repository interface {
	// Vacuum runs VACUUM to reclaim space and optimize the database.
	Vacuum(ctx context.Context) error
}
