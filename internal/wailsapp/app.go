package wailsapp

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/domain/vrchatconfig"
	"vrchat-tweaker/internal/infrastructure/filesystem"
	"vrchat-tweaker/internal/infrastructure/logwatcher"
	"vrchat-tweaker/internal/infrastructure/sleepsuppress"
	"vrchat-tweaker/internal/infrastructure/sqlite"
	"vrchat-tweaker/internal/infrastructure/statuspage"
	"vrchat-tweaker/internal/infrastructure/tray"
	"vrchat-tweaker/internal/infrastructure/vrchatapi"
	"vrchat-tweaker/internal/platform/paths"
	"vrchat-tweaker/internal/usecase"
)

// App struct holds the application state and use cases.
type App struct {
	ctx              context.Context
	launcher         *usecase.LauncherUseCase
	media            *usecase.MediaUseCase
	activity         *usecase.ActivityUseCase
	identity         *usecase.IdentityUseCase
	automation       *usecase.AutomationUseCase
	settings         *usecase.SettingsUseCase
	presenceChange   *usecase.PresenceChangeUseCase
	dbMaintenance    *usecase.DBMaintenanceUseCase
	assetCache       *usecase.VRChatAssetCacheUseCase
	ytdlp            *usecase.YTDLPMaintainUseCase
	appUpdate        *usecase.AppUpdateUseCase
	cookieLinkage    *usecase.CookieLinkageUseCase
	serverStatus     *statuspage.Client
	vrchatConfigRepo *filesystem.VRChatConfigFileRepository

	galleryScanMu     sync.Mutex
	galleryScanCancel context.CancelFunc
	galleryScanDone   chan struct{}

	pipelineMu     sync.Mutex
	pipelineCancel context.CancelFunc
	pipelineDone   chan struct{}

	sleepSuppressMu     sync.Mutex
	sleepSuppressCancel context.CancelFunc
	sleepSuppressWG     sync.WaitGroup

	ytdlpMaintainMu     sync.Mutex
	ytdlpMaintainCancel context.CancelFunc
	ytdlpMaintainWG     sync.WaitGroup

	activityWatchMu     sync.Mutex
	activityWatchCancel context.CancelFunc
	activityWatchWG     sync.WaitGroup

	activityRetentionMu     sync.Mutex
	activityRetentionCancel context.CancelFunc
	activityRetentionWG     sync.WaitGroup

	activityIngestMu       sync.Mutex
	activityIngestAdapters map[string]*logwatcher.ActivityIngestAdapter

	tray        tray.Manager
	quitPending atomic.Bool

	appUpdatePendingMu      sync.Mutex
	pendingStartupAppUpdate *usecase.AppUpdateStatus
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts.
const selfCacheChangedEvent = "identity:self-cache-changed"

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	dataDir, err := paths.AppDataDir()
	if err != nil {
		runtime.LogError(ctx, "failed to get data dir: "+err.Error())
		return
	}
	if mkdirErr := os.MkdirAll(dataDir, 0700); mkdirErr != nil {
		runtime.LogError(ctx, "failed to create data dir: "+mkdirErr.Error())
		return
	}

	db, err := sqlite.Open(dataDir)
	if err != nil {
		runtime.LogError(ctx, "failed to open DB: "+err.Error())
		return
	}

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
		runtime.LogWarning(ctx, msg)
	})
	// Do NOT set the auth token here. The frontend must call GetCredentialBlob,
	// decrypt the wrapped blob with Web Crypto, and then call UnlockVRChatSession.
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
		runtime.EventsEmit(a.ctx, selfCacheChangedEvent, struct{}{})
	})
	a.presenceChange = usecase.NewPresenceChangeUseCase(a.identity, settingsRepo)
	a.automation = usecase.NewAutomationUseCase(automationRepo, a.identity, sleepsuppress.NewVRChatProcessChecker())
	a.automation.SetRunLogChangedHook(func() {
		runtime.EventsEmit(a.ctx, "automation:run-log-changed", nil)
	})
	a.automation.Start(ctx)
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

	// Drop rows past Activity retention before historical logs are read, and
	// record the cutoff so that replay does not insert them again.
	if err := a.purgeActivityRetention(ctx); err != nil && !errors.Is(err, context.Canceled) {
		runtime.LogWarning(ctx, "activity retention: "+err.Error())
	}

	// Start output_log watcher if path is configured
	a.startOutputLogWatcher(ctx)

	a.startPictureFolderWatcher(ctx)
	go a.startupGalleryIncremental()
	a.startSleepSuppressLoop()
	a.startYTDLPMaintainLoop()
	a.startActivityRetentionLoop()
	a.syncTrayFromSettings(ctx)
	a.startAppUpdateCheckOnStartup()
}

// shutdown persists state before the process exits (Wails lifecycle).
func (a *App) shutdown(ctx context.Context) {
	a.stopTray()
	a.stopActivityRetentionLoop()
	a.stopVRChatActivityMonitor()
	if a.automation != nil {
		a.automation.Stop()
	}
	a.stopSleepSuppressLoop()
	a.stopYTDLPMaintainLoop()
	if a.settings == nil {
		return
	}
	if err := a.settings.SetGalleryLastExitAt(ctx, time.Now().UTC()); err != nil {
		runtime.LogWarning(ctx, "gallery last exit at: "+err.Error())
	}
}

func (a *App) appContext() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

type logLogger struct{}

func (logLogger) Printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func appDiagLogger() logwatcher.Logger {
	ll := logLogger{}
	return logwatcher.Logger(ll.Printf)
}
