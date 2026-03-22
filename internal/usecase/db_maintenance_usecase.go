package usecase

import (
	"context"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/domain/maintenance"
	"vrchat-tweaker/internal/domain/media"
)

// DBMaintenanceUseCase handles DB maintenance operations (Vacuum, Clear).
type DBMaintenanceUseCase struct {
	encounterRepo   activity.UserEncounterRepository
	screenshotRepo  media.ScreenshotRepository
	userCacheRepo   identity.UserCacheRepository
	maintenanceRepo maintenance.Repository
}

// NewDBMaintenanceUseCase creates a new DBMaintenanceUseCase.
func NewDBMaintenanceUseCase(
	encounterRepo activity.UserEncounterRepository,
	screenshotRepo media.ScreenshotRepository,
	userCacheRepo identity.UserCacheRepository,
	maintenanceRepo maintenance.Repository,
) *DBMaintenanceUseCase {
	return &DBMaintenanceUseCase{
		encounterRepo:   encounterRepo,
		screenshotRepo:  screenshotRepo,
		userCacheRepo:   userCacheRepo,
		maintenanceRepo: maintenanceRepo,
	}
}

// VacuumDb runs VACUUM to reclaim space and optimize the database.
func (uc *DBMaintenanceUseCase) VacuumDb(ctx context.Context) error {
	return uc.maintenanceRepo.Vacuum(ctx)
}

// ClearEncounters deletes all user encounters. Returns affected row count.
func (uc *DBMaintenanceUseCase) ClearEncounters(ctx context.Context) (int64, error) {
	return uc.encounterRepo.DeleteAll(ctx)
}

// ClearScreenshots deletes all screenshots. Returns affected row count.
func (uc *DBMaintenanceUseCase) ClearScreenshots(ctx context.Context) (int64, error) {
	return uc.screenshotRepo.DeleteAll(ctx)
}

// ClearFriendsCache deletes all cached friends. Returns affected row count.
func (uc *DBMaintenanceUseCase) ClearFriendsCache(ctx context.Context) (int64, error) {
	return uc.userCacheRepo.DeleteAll(ctx)
}
