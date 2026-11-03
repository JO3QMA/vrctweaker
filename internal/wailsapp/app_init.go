package wailsapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/gen2brain/beeep"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/domain/vrchatconfig"
	"vrchat-tweaker/internal/infrastructure/desktop"
	"vrchat-tweaker/internal/infrastructure/filesystem"
	"vrchat-tweaker/internal/infrastructure/sleepsuppress"
	"vrchat-tweaker/internal/infrastructure/sqlite"
	"vrchat-tweaker/internal/infrastructure/statuspage"
	"vrchat-tweaker/internal/infrastructure/vrchatapi"
	"vrchat-tweaker/internal/platform/paths"
	"vrchat-tweaker/internal/usecase"
)

const startupStatusChangedEvent = "app:startup-status-changed"

var (
	errAppNotInitialized = errors.New("app not initialized")
	errAppStillStarting  = errors.New("app is still starting")
)

type initFailureKind int

const (
	initFailDataDirResolve initFailureKind = iota
	initFailDataDirCreate
	initFailDatabaseOpen
)

// StartupInitStatus describes application core initialization for the frontend.
type StartupInitStatus struct {
	Ready             bool   `json:"ready"`
	Loading           bool   `json:"loading"`
	Phase             string `json:"phase"`
	SummaryKey        string `json:"summaryKey"`
	Detail            string `json:"detail"`
	DataDir           string `json:"dataDir"`
	CanOpenDataFolder bool   `json:"canOpenDataFolder"`
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.wailsLifecycleCtx = ctx
	initCtx, cancel := context.WithCancel(ctx)
	a.initMu.Lock()
	a.initCancel = cancel
	a.initDone = make(chan struct{})
	a.initLoading = true
	a.initPhase = "opening"
	a.initMu.Unlock()
	a.emitStartupStatusChanged(ctx)
	go func() {
		defer close(a.initDone)
		a.runStartupSequence(initCtx)
	}()
}

func (a *App) stopStartupSequence() {
	a.initMu.Lock()
	cancel := a.initCancel
	done := a.initDone
	a.initMu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}

func (a *App) runStartupSequence(ctx context.Context) {
	defer func() {
		a.initMu.Lock()
		a.initLoading = false
		if !a.initReady {
			a.initPhase = ""
		}
		a.initMu.Unlock()
		a.emitStartupStatusChanged(ctx)
	}()

	if err := a.initializeCore(ctx); err != nil {
		if a.wailsLifecycleCtx != nil {
			runtime.LogError(a.wailsLifecycleCtx, err.Error())
		}
		return
	}
	if ctx.Err() != nil {
		a.teardownCore()
		return
	}

	a.initMu.Lock()
	a.initReady = true
	a.initPhase = ""
	a.initMu.Unlock()
	a.emitStartupStatusChanged(ctx)

	if ctx.Err() != nil {
		a.teardownCore()
		return
	}
	a.startupPostInit(ctx)
}

func (a *App) initializeCore(ctx context.Context) error {
	a.teardownCoreLocked()

	dataDir, err := paths.AppDataDir()
	if err != nil {
		a.recordInitFailure(initFailDataDirResolve, err, "")
		return err
	}
	if mkdirErr := os.MkdirAll(dataDir, 0o700); mkdirErr != nil {
		a.recordInitFailure(initFailDataDirCreate, mkdirErr, dataDir)
		return mkdirErr
	}

	a.initMu.Lock()
	a.initPhase = "opening"
	a.initMu.Unlock()
	a.emitStartupStatusChanged(ctx)

	db, err := sqlite.Open(dataDir, &sqlite.OpenOptions{
		OnSchemaPhase: func(phase string) {
			a.initMu.Lock()
			a.initPhase = phase
			a.initMu.Unlock()
			a.emitStartupStatusChanged(ctx)
		},
	})
	if err != nil {
		a.recordInitFailure(initFailDatabaseOpen, err, dataDir)
		return err
	}

	a.initMu.Lock()
	a.dataDir = dataDir
	a.db = db
	a.initErr = nil
	a.initSummaryKey = ""
	a.initDetail = ""
	a.initMu.Unlock()

	return a.wireUseCases(ctx, db, dataDir)
}

func (a *App) wireUseCases(ctx context.Context, db *sql.DB, dataDir string) error {
	launcherRepo := sqlite.NewLauncherProfileRepository(db)
	mediaRepo := sqlite.NewScreenshotRepository(db)
	playRepo := sqlite.NewPlaySessionRepository(db)
	encounterRepo := sqlite.NewUserEncounterRepository(db)
	videoPlaybackRepo := sqlite.NewVideoPlaybackRepository(db)
	friendAvatarRepo := sqlite.NewFriendAvatarObservationRepository(db)
	userCacheRepo := sqlite.NewUserCacheRepository(db)
	worldRepo := sqlite.NewWorldInfoRepository(db)
	automationRepo := sqlite.NewAutomationItemRepository(db)
	settingsRepo := sqlite.NewAppSettingsRepository(db)

	credStore := vrchatapi.NewAutoCredentialStore(dataDir, func(msg string) {
		a.logWarning(msg)
	})
	apiClient := vrchatapi.NewClient("")

	const defaultNotificationTitle = "VRChat Tweaker"
	notify := func(title, message string) error {
		if title == "" {
			title = defaultNotificationTitle
		}
		return beeep.Notify(title, message, "")
	}
	a.launcher = usecase.NewLauncherUseCase(launcherRepo)
	a.media = usecase.NewMediaUseCase(mediaRepo, worldRepo, userCacheRepo)
	enrichmentRepo := sqlite.NewScreenshotEnrichmentRepository(db)
	a.media.SetEnrichmentDeps(usecase.MediaEnrichmentDeps{
		PlaySessions: playRepo,
		Encounters:   encounterRepo,
		Enrichment:   enrichmentRepo,
		Settings:     settingsRepo,
	})
	a.activity = usecase.NewActivityUseCase(playRepo, encounterRepo, settingsRepo, userCacheRepo, worldRepo).
		WithVideoPlaybackRepo(videoPlaybackRepo).
		WithFriendAvatarObservationRepo(friendAvatarRepo)
	a.identity = usecase.NewIdentityUseCase(userCacheRepo, apiClient, credStore, settingsRepo, notify)
	a.identity.SetAvatarNameCacheRepo(sqlite.NewAvatarNameCacheRepository(db))
	a.identity.SetSelfCacheChangedHook(func() {
		a.eventsEmit(selfCacheChangedEvent, struct{}{})
	})
	a.presenceChange = usecase.NewPresenceChangeUseCase(a.identity, settingsRepo)
	a.automation = usecase.NewAutomationUseCase(automationRepo, a.identity, sleepsuppress.NewVRChatProcessChecker())
	a.automation.SetRunLogChangedHook(func() {
		a.eventsEmit("automation:run-log-changed", nil)
	})
	a.settings = usecase.NewSettingsUseCase(settingsRepo)
	a.dbMaintenance = usecase.NewDBMaintenanceUseCase(db, encounterRepo, friendAvatarRepo, mediaRepo, userCacheRepo, settingsRepo)
	a.ytdlp = usecase.NewYTDLPMaintainUseCase(a.settings, usecase.NewYTDLPUpdater())
	a.appUpdate = usecase.NewAppUpdateUseCase(a.settings, nil, nil)
	a.cookieLinkage = usecase.NewCookieLinkageUseCase(a.settings)
	a.serverStatus = statuspage.NewClient()

	configPath := paths.VRChatConfigPathOrFallback()
	configRepo := filesystem.NewVRChatConfigFileRepository(configPath)
	a.vrchatConfigRepo = configRepo
	a.assetCache = usecase.NewVRChatAssetCacheUseCase(
		func() (*vrchatconfig.VRChatConfig, error) {
			return a.vrchatConfigRepo.Read()
		},
		sleepsuppress.NewVRChatProcessChecker(),
		usecase.DefaultVRChatAssetCacheFolder,
		func() (string, error) { return a.DefaultVRChatPictureFolder() },
		paths.VRChatDataDir,
	)
	return nil
}

func (a *App) startupPostInit(ctx context.Context) {
	if a.automation != nil {
		a.automation.Start(ctx)
	}
	if err := a.purgeActivityRetention(ctx); err != nil && !errors.Is(err, context.Canceled) {
		runtime.LogWarning(ctx, "activity retention: "+err.Error())
	}
	a.startOutputLogWatcher(ctx)
	a.startPictureFolderWatcher(ctx)
	go a.startupGalleryIncremental()
	a.startSleepSuppressLoop()
	a.startYTDLPMaintainLoop()
	a.startActivityRetentionLoop()
	a.syncTrayFromSettings(ctx)
	a.startAppUpdateCheckOnStartup()
}

func (a *App) recordInitFailure(kind initFailureKind, err error, dataDir string) {
	a.initMu.Lock()
	defer a.initMu.Unlock()
	a.initReady = false
	a.initErr = err
	a.dataDir = dataDir
	a.initDetail = err.Error()
	switch kind {
	case initFailDataDirResolve:
		a.initSummaryKey = "app.startup.failed.dataDirResolve"
		a.canOpenDataFolder = false
	case initFailDataDirCreate:
		a.initSummaryKey = "app.startup.failed.dataDirCreate"
		a.canOpenDataFolder = dataDirExistsAsFolder(dataDir)
	case initFailDatabaseOpen:
		a.initSummaryKey = "app.startup.failed.databaseOpen"
		a.canOpenDataFolder = dataDirExistsAsFolder(dataDir)
	default:
		a.initSummaryKey = "app.startup.failed.databaseOpen"
		a.canOpenDataFolder = dataDirExistsAsFolder(dataDir)
	}
	a.teardownCoreLocked()
}

func dataDirExistsAsFolder(dir string) bool {
	if dir == "" {
		return false
	}
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

func (a *App) teardownCoreLocked() {
	if a.db != nil {
		_ = a.db.Close()
		a.db = nil
	}
	a.launcher = nil
	a.media = nil
	a.activity = nil
	a.identity = nil
	a.automation = nil
	a.settings = nil
	a.presenceChange = nil
	a.dbMaintenance = nil
	a.assetCache = nil
	a.ytdlp = nil
	a.appUpdate = nil
	a.cookieLinkage = nil
	a.serverStatus = nil
	a.vrchatConfigRepo = nil
}

func (a *App) teardownCore() {
	a.initMu.Lock()
	defer a.initMu.Unlock()
	a.teardownCoreLocked()
}

func (a *App) logWarning(msg string) {
	if a.wailsLifecycleCtx == nil {
		return
	}
	runtime.LogWarning(a.wailsLifecycleCtx, msg)
}

func (a *App) eventsEmit(event string, data ...any) {
	if a.wailsLifecycleCtx == nil {
		return
	}
	runtime.EventsEmit(a.wailsLifecycleCtx, event, data...)
}

func (a *App) emitStartupStatusChanged(ctx context.Context) {
	if ctx == nil || a.wailsLifecycleCtx == nil || ctx != a.wailsLifecycleCtx {
		return
	}
	runtime.EventsEmit(ctx, startupStatusChangedEvent, a.getStartupInitStatus())
}

// GetStartupInitStatus returns current core initialization state for the UI.
func (a *App) GetStartupInitStatus() StartupInitStatus {
	a.initMu.RLock()
	defer a.initMu.RUnlock()
	return a.startupInitStatusLocked()
}

func (a *App) getStartupInitStatus() StartupInitStatus {
	a.initMu.RLock()
	defer a.initMu.RUnlock()
	return a.startupInitStatusLocked()
}

func (a *App) startupInitStatusLocked() StartupInitStatus {
	return StartupInitStatus{
		Ready:             a.initReady,
		Loading:           a.initLoading,
		Phase:             a.initPhase,
		SummaryKey:        a.initSummaryKey,
		Detail:            a.initDetail,
		DataDir:           a.dataDir,
		CanOpenDataFolder: a.canOpenDataFolder,
	}
}

// RetryStartupInit re-runs core initialization after a failure.
func (a *App) RetryStartupInit() (StartupInitStatus, error) {
	ctx := a.appContext()
	a.initMu.Lock()
	if a.initLoading {
		st := a.startupInitStatusLocked()
		a.initMu.Unlock()
		return st, errAppStillStarting
	}
	if a.initReady {
		st := a.startupInitStatusLocked()
		a.initMu.Unlock()
		return st, nil
	}
	a.initLoading = true
	a.initPhase = "opening"
	a.initSummaryKey = ""
	a.initDetail = ""
	a.initErr = nil
	a.initMu.Unlock()
	a.emitStartupStatusChanged(ctx)

	if err := a.initializeCore(ctx); err != nil {
		a.initMu.Lock()
		a.initLoading = false
		a.initPhase = ""
		st := a.startupInitStatusLocked()
		a.initMu.Unlock()
		a.emitStartupStatusChanged(ctx)
		return st, nil
	}

	if ctx.Err() != nil {
		a.teardownCore()
		a.initMu.Lock()
		a.initLoading = false
		a.initPhase = ""
		st := a.startupInitStatusLocked()
		a.initMu.Unlock()
		a.emitStartupStatusChanged(ctx)
		return st, nil
	}

	a.initMu.Lock()
	a.initReady = true
	a.initLoading = false
	a.initPhase = ""
	st := a.startupInitStatusLocked()
	a.initMu.Unlock()
	a.emitStartupStatusChanged(ctx)
	if ctx.Err() != nil {
		a.teardownCore()
		return st, nil
	}
	a.startupPostInit(ctx)
	return st, nil
}

// OpenAppDataFolder opens the application data directory in the file manager.
func (a *App) OpenAppDataFolder() error {
	a.initMu.RLock()
	dir := a.dataDir
	canOpen := a.canOpenDataFolder
	a.initMu.RUnlock()
	if !canOpen || dir == "" {
		return fmt.Errorf("application data folder is unavailable")
	}
	return desktop.OpenFolderInFileManager(dir)
}

func (a *App) requireReady() error {
	if a == nil {
		return errAppNotInitialized
	}
	a.initMu.RLock()
	defer a.initMu.RUnlock()
	if a.initReady {
		return nil
	}
	if a.initLoading {
		return errAppStillStarting
	}
	// Unit tests construct App with use cases directly without running startup.
	if a.ctx != nil && a.settings != nil {
		return nil
	}
	return errAppNotInitialized
}

func (a *App) closeDB() {
	a.initMu.Lock()
	defer a.initMu.Unlock()
	if a.db != nil {
		_ = a.db.Close()
		a.db = nil
	}
}

// setInitReadyForTest marks startup complete without full wiring (unit tests only).
func (a *App) setInitReadyForTest(ctx context.Context) {
	a.initMu.Lock()
	defer a.initMu.Unlock()
	if ctx != nil {
		a.ctx = ctx
	}
	a.initReady = true
	a.initLoading = false
	a.initPhase = ""
}
