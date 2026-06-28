package usecase

import (
	"context"
	"fmt"

	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/domain/media"
	"vrchat-tweaker/internal/infrastructure/sqlite"

	"database/sql"
)

// DBMaintenanceUseCase handles DB maintenance operations (Vacuum, Clear).
type DBMaintenanceUseCase struct {
	db             *sql.DB
	encounterRepo  userEncounterRepo
	screenshotRepo media.ScreenshotRepository
	userCacheRepo  identity.UserCacheRepository
	appSettings    appSettingsRepo
}

// NewDBMaintenanceUseCase creates a new DBMaintenanceUseCase.
func NewDBMaintenanceUseCase(
	db *sql.DB,
	encounterRepo userEncounterRepo,
	screenshotRepo media.ScreenshotRepository,
	userCacheRepo identity.UserCacheRepository,
	appSettings appSettingsRepo,
) *DBMaintenanceUseCase {
	return &DBMaintenanceUseCase{
		db:             db,
		encounterRepo:  encounterRepo,
		screenshotRepo: screenshotRepo,
		userCacheRepo:  userCacheRepo,
		appSettings:    appSettings,
	}
}

// VacuumDb runs VACUUM to reclaim space and optimize the database.
func (uc *DBMaintenanceUseCase) VacuumDb(ctx context.Context) error {
	return sqlite.Vacuum(ctx, uc.db)
}

// ClearEncounters deletes all user encounters. Returns affected row count.
func (uc *DBMaintenanceUseCase) ClearEncounters(ctx context.Context) (int64, error) {
	return uc.encounterRepo.DeleteAll(ctx)
}

// ClearScreenshots deletes all screenshots. Returns affected row count.
func (uc *DBMaintenanceUseCase) ClearScreenshots(ctx context.Context) (int64, error) {
	return uc.screenshotRepo.DeleteAll(ctx)
}

// ClearFriendsCache deletes all cached users (friends, self, contacts). Returns affected row count.
func (uc *DBMaintenanceUseCase) ClearFriendsCache(ctx context.Context) (int64, error) {
	n, err := uc.userCacheRepo.DeleteAll(ctx)
	if err != nil {
		return 0, err
	}
	if uc.appSettings != nil {
		if setErr := uc.appSettings.Set(ctx, identity.SettingVRChatFriendsSyncedAt, ""); setErr != nil {
			return n, fmt.Errorf("clear friends sync timestamp: %w", setErr)
		}
	}
	return n, nil
}
