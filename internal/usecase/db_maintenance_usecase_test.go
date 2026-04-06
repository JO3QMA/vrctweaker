package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/domain/media"
)

type maintEncounterRepo struct {
	deleteAllN   int64
	deleteAllErr error
}

func (m *maintEncounterRepo) List(context.Context, *activity.EncounterFilter) ([]*activity.UserEncounter, error) {
	return nil, nil
}

func (m *maintEncounterRepo) ListWithContext(context.Context, *activity.EncounterFilter) ([]*activity.EncounterWithContext, error) {
	return nil, nil
}

func (m *maintEncounterRepo) Save(context.Context, *activity.UserEncounter) error { return nil }

func (m *maintEncounterRepo) CloseEncounterLeave(context.Context, string, time.Time) (int64, error) {
	return 0, nil
}

func (m *maintEncounterRepo) CloseOpenEncountersAt(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func (m *maintEncounterRepo) DeleteOlderThan(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func (m *maintEncounterRepo) DeleteAll(context.Context) (int64, error) {
	return m.deleteAllN, m.deleteAllErr
}

func (m *maintEncounterRepo) Count(context.Context) (int64, error) { return 0, nil }

func (m *maintEncounterRepo) BackfillMissingWorldContext(context.Context) (int64, error) {
	return 0, nil
}

type maintScreenshotRepo struct {
	deleteAllN   int64
	deleteAllErr error
}

func (m *maintScreenshotRepo) List(context.Context, *media.ScreenshotFilter) ([]*media.Screenshot, error) {
	return nil, nil
}

func (m *maintScreenshotRepo) GetByID(context.Context, string) (*media.Screenshot, error) {
	return nil, nil
}

func (m *maintScreenshotRepo) GetByFilePath(context.Context, string) (*media.Screenshot, error) {
	return nil, nil
}

func (m *maintScreenshotRepo) Save(context.Context, *media.Screenshot) error { return nil }

func (m *maintScreenshotRepo) Delete(context.Context, string) error { return nil }

func (m *maintScreenshotRepo) DeleteAll(context.Context) (int64, error) {
	return m.deleteAllN, m.deleteAllErr
}

func (m *maintScreenshotRepo) GetThumbnail(context.Context, string) (*media.ScreenshotThumbnail, error) {
	return nil, nil
}

func (m *maintScreenshotRepo) UpsertThumbnail(context.Context, string, *media.ScreenshotThumbnail) error {
	return nil
}

func (m *maintScreenshotRepo) DeleteThumbnail(context.Context, string) error { return nil }

type maintUserCacheRepo struct {
	deleteAllN   int64
	deleteAllErr error
}

func (m *maintUserCacheRepo) List(context.Context) ([]*identity.UserCache, error) { return nil, nil }

func (m *maintUserCacheRepo) GetByVRCUserID(context.Context, string) (*identity.UserCache, error) {
	return nil, nil
}

func (m *maintUserCacheRepo) ListFavorites(context.Context) ([]*identity.UserCache, error) {
	return nil, nil
}

func (m *maintUserCacheRepo) Save(context.Context, *identity.UserCache) error { return nil }

func (m *maintUserCacheRepo) SaveBatch(context.Context, []*identity.UserCache) error { return nil }

func (m *maintUserCacheRepo) Delete(context.Context, string) error { return nil }

func (m *maintUserCacheRepo) DeleteAll(context.Context) (int64, error) {
	return m.deleteAllN, m.deleteAllErr
}

func (m *maintUserCacheRepo) GetSelfBySessionFingerprint(context.Context, string) (*identity.UserCache, error) {
	return nil, nil
}

func (m *maintUserCacheRepo) UpsertSelf(context.Context, *identity.UserCache) error { return nil }

func (m *maintUserCacheRepo) DeleteSelfRows(context.Context) error { return nil }

type maintMaintenanceRepo struct {
	vacuumErr error
}

func (m *maintMaintenanceRepo) Vacuum(context.Context) error { return m.vacuumErr }

func TestDBMaintenanceUseCase_VacuumDb(t *testing.T) {
	ctx := context.Background()
	mr := &maintMaintenanceRepo{}
	uc := NewDBMaintenanceUseCase(
		&maintEncounterRepo{},
		&maintScreenshotRepo{},
		&maintUserCacheRepo{},
		mr,
		newMockSettingsRepo(),
	)
	if err := uc.VacuumDb(ctx); err != nil {
		t.Fatal(err)
	}
	wantErr := errors.New("vacuum fail")
	mr.vacuumErr = wantErr
	if err := uc.VacuumDb(ctx); !errors.Is(err, wantErr) {
		t.Fatalf("VacuumDb: got %v, want %v", err, wantErr)
	}
}

func TestDBMaintenanceUseCase_ClearEncounters(t *testing.T) {
	ctx := context.Background()
	er := &maintEncounterRepo{deleteAllN: 7}
	uc := NewDBMaintenanceUseCase(er, &maintScreenshotRepo{}, &maintUserCacheRepo{}, &maintMaintenanceRepo{}, newMockSettingsRepo())
	n, err := uc.ClearEncounters(ctx)
	if err != nil || n != 7 {
		t.Fatalf("ClearEncounters: n=%d err=%v", n, err)
	}
	er.deleteAllErr = errors.New("del")
	_, err = uc.ClearEncounters(ctx)
	if err == nil {
		t.Fatal("want error")
	}
}

func TestDBMaintenanceUseCase_ClearScreenshots(t *testing.T) {
	ctx := context.Background()
	sr := &maintScreenshotRepo{deleteAllN: 3}
	uc := NewDBMaintenanceUseCase(&maintEncounterRepo{}, sr, &maintUserCacheRepo{}, &maintMaintenanceRepo{}, newMockSettingsRepo())
	n, err := uc.ClearScreenshots(ctx)
	if err != nil || n != 3 {
		t.Fatalf("ClearScreenshots: n=%d err=%v", n, err)
	}
	sr.deleteAllErr = errors.New("del")
	_, err = uc.ClearScreenshots(ctx)
	if err == nil {
		t.Fatal("want error")
	}
}

func TestDBMaintenanceUseCase_ClearFriendsCache_noAppSettings(t *testing.T) {
	ctx := context.Background()
	ur := &maintUserCacheRepo{deleteAllN: 5}
	uc := NewDBMaintenanceUseCase(&maintEncounterRepo{}, &maintScreenshotRepo{}, ur, &maintMaintenanceRepo{}, nil)
	n, err := uc.ClearFriendsCache(ctx)
	if err != nil || n != 5 {
		t.Fatalf("ClearFriendsCache: n=%d err=%v", n, err)
	}
}

func TestDBMaintenanceUseCase_ClearFriendsCache_clearsSyncKey(t *testing.T) {
	ctx := context.Background()
	settingsRepo := newMockSettingsRepo()
	settingsRepo.m[identity.SettingVRChatFriendsSyncedAt] = "2024-01-01T00:00:00Z"
	ur := &maintUserCacheRepo{deleteAllN: 2}
	uc := NewDBMaintenanceUseCase(&maintEncounterRepo{}, &maintScreenshotRepo{}, ur, &maintMaintenanceRepo{}, settingsRepo)
	n, err := uc.ClearFriendsCache(ctx)
	if err != nil || n != 2 {
		t.Fatalf("ClearFriendsCache: n=%d err=%v", n, err)
	}
	if settingsRepo.m[identity.SettingVRChatFriendsSyncedAt] != "" {
		t.Fatalf("sync key not cleared: %q", settingsRepo.m[identity.SettingVRChatFriendsSyncedAt])
	}
}

func TestDBMaintenanceUseCase_ClearFriendsCache_deleteAllErr(t *testing.T) {
	ctx := context.Background()
	ur := &maintUserCacheRepo{deleteAllErr: errors.New("cache")}
	uc := NewDBMaintenanceUseCase(&maintEncounterRepo{}, &maintScreenshotRepo{}, ur, &maintMaintenanceRepo{}, newMockSettingsRepo())
	_, err := uc.ClearFriendsCache(ctx)
	if err == nil {
		t.Fatal("want error")
	}
}

func TestDBMaintenanceUseCase_ClearFriendsCache_setSyncKeyErr(t *testing.T) {
	ctx := context.Background()
	settingsRepo := newMockSettingsRepo()
	settingsRepo.setErr = errors.New("set failed")
	ur := &maintUserCacheRepo{deleteAllN: 1}
	uc := NewDBMaintenanceUseCase(&maintEncounterRepo{}, &maintScreenshotRepo{}, ur, &maintMaintenanceRepo{}, settingsRepo)
	n, err := uc.ClearFriendsCache(ctx)
	if err == nil {
		t.Fatal("want error")
	}
	if n != 1 {
		t.Fatalf("want row count 1 preserved, got %d", n)
	}
}
