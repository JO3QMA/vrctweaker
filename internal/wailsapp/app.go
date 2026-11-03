package wailsapp

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/infrastructure/filesystem"
	"vrchat-tweaker/internal/infrastructure/logwatcher"
	"vrchat-tweaker/internal/infrastructure/statuspage"
	"vrchat-tweaker/internal/infrastructure/tray"
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

	initMu            sync.RWMutex
	initReady         bool
	initLoading       bool
	initPhase         string
	initSummaryKey    string
	initDetail        string
	initErr           error
	canOpenDataFolder bool
	dataDir           string
	db                *sql.DB
	wailsLifecycleCtx context.Context
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts.
const selfCacheChangedEvent = "identity:self-cache-changed"

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
	a.closeDB()
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
