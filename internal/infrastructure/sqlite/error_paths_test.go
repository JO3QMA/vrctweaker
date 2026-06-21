package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/automation"
	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/domain/launcher"
	"vrchat-tweaker/internal/domain/media"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	if schemaErr := applySchema(db); schemaErr != nil {
		_ = db.Close()
		t.Fatal(schemaErr)
	}
	return db
}

func TestRepositories_returnErrorWhenDBClosed(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	t.Run("UserEncounter", func(t *testing.T) {
		db := openTestDB(t)
		repo := NewUserEncounterRepository(db)
		_ = db.Close()
		if _, err := repo.List(ctx, nil); err == nil {
			t.Fatal("List expected error")
		}
		if _, err := repo.ListWithContext(ctx, nil); err == nil {
			t.Fatal("ListWithContext expected error")
		}
		if _, err := repo.BackfillMissingWorldContext(ctx); err == nil {
			t.Fatal("Backfill expected error")
		}
		if _, err := repo.Count(ctx); err == nil {
			t.Fatal("Count expected error")
		}
		if _, err := repo.DeleteAll(ctx); err == nil {
			t.Fatal("DeleteAll expected error")
		}
		if _, err := repo.DeleteOlderThan(ctx, now); err == nil {
			t.Fatal("DeleteOlderThan expected error")
		}
		if _, err := repo.CloseEncounterLeave(ctx, "u", now); err == nil {
			t.Fatal("CloseEncounterLeave expected error")
		}
		if _, err := repo.CloseOpenEncountersAt(ctx, now); err == nil {
			t.Fatal("CloseOpenEncountersAt expected error")
		}
	})

	t.Run("PlaySession", func(t *testing.T) {
		db := openTestDB(t)
		repo := NewPlaySessionRepository(db)
		_ = db.Close()
		if _, err := repo.List(ctx, now, now); err == nil {
			t.Fatal("List expected error")
		}
	})

	t.Run("UserCache", func(t *testing.T) {
		db := openTestDB(t)
		repo := NewUserCacheRepository(db)
		_ = db.Close()
		if _, err := repo.List(ctx); err == nil {
			t.Fatal("List expected error")
		}
		if _, err := repo.ListFavorites(ctx); err == nil {
			t.Fatal("ListFavorites expected error")
		}
		if err := repo.SaveBatch(ctx, []*identity.UserCache{{VRCUserID: "x", DisplayName: "X", LastUpdated: now}}); err == nil {
			t.Fatal("SaveBatch expected error")
		}
		if _, err := repo.DeleteAll(ctx); err == nil {
			t.Fatal("DeleteAll expected error")
		}
		if err := repo.UpsertSelf(ctx, &identity.UserCache{VRCUserID: "x", DisplayName: "X", LastUpdated: now}); err == nil {
			t.Fatal("UpsertSelf expected error")
		}
	})

	t.Run("Screenshot", func(t *testing.T) {
		db := openTestDB(t)
		repo := NewScreenshotRepository(db)
		_ = db.Close()
		if _, err := repo.List(ctx, nil); err == nil {
			t.Fatal("List expected error")
		}
		if _, err := repo.DeleteAll(ctx); err == nil {
			t.Fatal("DeleteAll expected error")
		}
	})

	t.Run("Launcher", func(t *testing.T) {
		db := openTestDB(t)
		repo := NewLauncherProfileRepository(db)
		_ = db.Close()
		if _, err := repo.List(ctx); err == nil {
			t.Fatal("List expected error")
		}
		if err := repo.Save(ctx, &launcher.LaunchProfile{ID: "x", Name: "X"}); err == nil {
			t.Fatal("Save expected error")
		}
	})

	t.Run("Automation", func(t *testing.T) {
		db := openTestDB(t)
		repo := NewAutomationRuleRepository(db)
		_ = db.Close()
		if _, err := repo.List(ctx); err == nil {
			t.Fatal("List expected error")
		}
		if _, err := repo.ListEnabled(ctx); err == nil {
			t.Fatal("ListEnabled expected error")
		}
	})

	t.Run("AppSettings", func(t *testing.T) {
		db := openTestDB(t)
		repo := NewAppSettingsRepository(db)
		_ = db.Close()
		if _, err := repo.GetAll(ctx); err == nil {
			t.Fatal("GetAll expected error")
		}
	})

	t.Run("SaveAfterClose", func(t *testing.T) {
		db := openTestDB(t)
		encRepo := NewUserEncounterRepository(db)
		mediaRepo := NewScreenshotRepository(db)
		_ = db.Close()
		if err := encRepo.Save(ctx, &activity.UserEncounter{
			ID: "e", VRCUserID: "u", DisplayName: "U", JoinedAt: now,
		}); err == nil {
			t.Fatal("encounter Save expected error")
		}
		if err := mediaRepo.Save(ctx, &media.Screenshot{ID: "s", FilePath: "/x"}); err == nil {
			t.Fatal("screenshot Save expected error")
		}
		if err := NewAutomationRuleRepository(db).Save(ctx, &automation.AutomationRule{ID: "r", Name: "R"}); err == nil {
			t.Fatal("automation Save expected error")
		}
	})
}
