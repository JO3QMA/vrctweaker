package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/automation"
	"vrchat-tweaker/internal/domain/event"
	"vrchat-tweaker/internal/domain/launcher"
	"vrchat-tweaker/internal/domain/media"
	"vrchat-tweaker/internal/infrastructure/desktop"
	"vrchat-tweaker/internal/infrastructure/filesystem"
	"vrchat-tweaker/internal/infrastructure/logwatcher"
	"vrchat-tweaker/internal/infrastructure/picturewatcher"
	"vrchat-tweaker/internal/infrastructure/sqlite"
	"vrchat-tweaker/internal/infrastructure/vrchatapi"
	"vrchat-tweaker/internal/usecase"
)

// App struct holds the application state and use cases.
type App struct {
	ctx           context.Context
	launcher      *usecase.LauncherUseCase
	media         *usecase.MediaUseCase
	activity      *usecase.ActivityUseCase
	identity      *usecase.IdentityUseCase
	automation    *usecase.AutomationUseCase
	settings      *usecase.SettingsUseCase
	dbMaintenance *usecase.DBMaintenanceUseCase
	vrchatConfig  *usecase.VRChatConfigUseCase

	galleryScanMu     sync.Mutex
	galleryScanCancel context.CancelFunc
	galleryScanWG     sync.WaitGroup
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	dataDir, err := getDataDir()
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

	eventBus := event.NewChannelEventBus()

	launcherRepo := sqlite.NewLauncherProfileRepository(db)
	mediaRepo := sqlite.NewScreenshotRepository(db)
	playRepo := sqlite.NewPlaySessionRepository(db)
	encounterRepo := sqlite.NewUserEncounterRepository(db)
	userCacheRepo := sqlite.NewUserCacheRepository(db)
	worldRepo := sqlite.NewWorldInfoRepository(db)
	automationRepo := sqlite.NewAutomationRuleRepository(db)
	settingsRepo := sqlite.NewAppSettingsRepository(db)

	credStore := vrchatapi.NewKeyringCredentialStore()
	apiClient := vrchatapi.NewClient("")
	if token, err := credStore.Get(vrchatapi.CredentialService, vrchatapi.CredentialUser); err == nil && token != "" {
		apiClient.SetAuthToken(token)
	}

	extractor := media.NewDefaultMetadataExtractor()
	maintenanceRepo := sqlite.NewMaintenanceRepository(db)
	notifier := desktop.NewBeeepNotifier("VRChat Tweaker")
	a.launcher = usecase.NewLauncherUseCase(launcherRepo)
	a.media = usecase.NewMediaUseCase(mediaRepo, extractor, worldRepo, userCacheRepo)
	a.activity = usecase.NewActivityUseCase(playRepo, encounterRepo, settingsRepo, userCacheRepo, worldRepo)
	a.identity = usecase.NewIdentityUseCaseWithNotifier(userCacheRepo, apiClient, credStore, settingsRepo, notifier)
	actionRunner := usecase.NewDefaultActionRunner(a.identity)
	a.automation = usecase.NewAutomationUseCase(automationRepo, eventBus, actionRunner)
	a.settings = usecase.NewSettingsUseCase(settingsRepo)
	a.dbMaintenance = usecase.NewDBMaintenanceUseCase(encounterRepo, mediaRepo, userCacheRepo, maintenanceRepo, settingsRepo)

	configPath := getVRChatConfigPath()
	configRepo := filesystem.NewVRChatConfigFileRepository(configPath)
	a.vrchatConfig = usecase.NewVRChatConfigUseCase(configRepo)

	a.subscribeAutomationEvents(ctx, eventBus)

	// Start output_log watcher if path is configured
	a.startOutputLogWatcher(ctx, eventBus)

	a.startPictureFolderWatcher(ctx)
	go a.startupGalleryIncremental()
}

// onShutdown persists state before the process exits (Wails lifecycle).
func (a *App) onShutdown(ctx context.Context) {
	if a.settings == nil {
		return
	}
	if err := a.settings.SetGalleryLastExitAt(ctx, time.Now().UTC()); err != nil {
		runtime.LogWarning(ctx, "gallery last exit at: "+err.Error())
	}
}

func (a *App) startupGalleryIncremental() {
	const waitTick = 50 * time.Millisecond
	const waitMaxIters = 3600 // ~3 minutes while a manual folder scan is in progress

	for i := 0; i < waitMaxIters && a.IsGalleryScanning(); i++ {
		select {
		case <-a.ctx.Done():
			return
		case <-time.After(waitTick):
		}
	}
	if a.IsGalleryScanning() {
		runtime.LogWarning(a.ctx, "startup gallery incremental: skipped (folder scan still running)")
		return
	}

	root := a.resolveVRChatPictureWatchRoot()
	if root == "" {
		return
	}
	if _, err := os.Stat(root); err != nil {
		return
	}

	since, ok := a.settings.GetGalleryLastExitAt(a.ctx)
	if !ok {
		return
	}

	count, err := a.media.IngestUnderPictureRootSince(a.ctx, root, since)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		runtime.LogWarning(a.ctx, "startup gallery incremental: "+err.Error())
		return
	}
	if count > 0 {
		runtime.EventsEmit(a.ctx, galleryScreenshotsChangedEvent, struct{}{})
	}
}

func (a *App) subscribeAutomationEvents(ctx context.Context, eventBus event.EventBus) {
	handler := func(topic string) func(context.Context, *event.Event) error {
		return func(c context.Context, ev *event.Event) error {
			payload, _ := ev.Payload.(map[string]interface{})
			return a.automation.EvalAndRun(c, topic, payload)
		}
	}
	eventBus.Subscribe(automation.TriggerAFKDetected, handler(automation.TriggerAFKDetected))
	eventBus.Subscribe(automation.TriggerFriendJoined, handler(automation.TriggerFriendJoined))
}

func defaultVRChatOutputLogDir() string {
	return filepath.Dir(getVRChatConfigPath())
}

func matchAbsPaths(a, b string) bool {
	aa, e1 := filepath.Abs(filepath.Clean(a))
	bb, e2 := filepath.Abs(filepath.Clean(b))
	if e1 != nil || e2 != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return filepath.Clean(aa) == filepath.Clean(bb)
}

func (a *App) resolveEffectiveOutputLogWatchPath(ctx context.Context) (string, error) {
	p, err := a.settings.GetOutputLogPath(ctx)
	if err != nil {
		return "", err
	}
	p = strings.TrimSpace(p)
	if p != "" {
		absPath, absErr := filepath.Abs(filepath.Clean(p))
		if absErr != nil {
			return "", absErr
		}
		if _, statErr := os.Stat(absPath); statErr != nil {
			return "", statErr
		}
		return absPath, nil
	}
	dir := defaultVRChatOutputLogDir()
	if dir == "" {
		return "", os.ErrNotExist
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(absDir); err != nil {
		return "", err
	}
	return absDir, nil
}

func (a *App) ingestActivityLogsBootstrap(ctx context.Context, absWatch string, parser *activity.LogParser, activityHandler *logwatcher.ActivityEventHandler, logger logwatcher.Logger) {
	info, err := os.Stat(absWatch)
	if err != nil {
		return
	}
	cp, _ := a.activity.GetActivityLogCheckpoint(ctx)

	if info.IsDir() {
		files, listErr := logwatcher.ListOutputLogFiles(absWatch)
		if listErr != nil {
			return
		}
		startIdx := 0
		var startOff int64
		if cp != nil && matchAbsPaths(cp.WatchPath, absWatch) {
			found := false
			for i, f := range files {
				if matchAbsPaths(f, cp.File) {
					found = true
					startIdx = i
					st, statErr := os.Stat(f)
					sz := int64(0)
					if statErr == nil && st != nil {
						sz = st.Size()
					}
					if sz > 0 && cp.ByteOffset >= sz {
						startIdx = i + 1
						startOff = 0
					} else {
						startOff = cp.ByteOffset
					}
					break
				}
			}
			if !found {
				startIdx = len(files) - 1
				startOff = 0
			}
		}
		if startIdx >= len(files) {
			return
		}
		for i := startIdx; i < len(files); i++ {
			fp := files[i]
			off := int64(0)
			if i == startIdx {
				off = startOff
			}
			pathCopy := fp
			checkpointLines := 0
			var lastVRLineTime time.Time
			if off == 0 {
				activityHandler.ResetSessionContextForNewLogFile()
			}
			_, procErr := logwatcher.ProcessOutputLogFileFromOffset(ctx, pathCopy, off, parser, activityHandler, logger, func(pos int64, line string) {
				if ts := activity.ParseVRChatTimestamp(line, time.Time{}); !ts.IsZero() {
					lastVRLineTime = ts
				}
				checkpointLines++
				if checkpointLines != 1 && checkpointLines%32 != 0 {
					return
				}
				vrTime := ""
				if ts := activity.ParseVRChatTimestamp(line, time.Time{}); !ts.IsZero() {
					vrTime = ts.Format(time.RFC3339)
				}
				_ = a.activity.SetActivityLogCheckpoint(ctx, &usecase.ActivityLogCheckpoint{
					WatchPath:      absWatch,
					File:           pathCopy,
					ByteOffset:     pos,
					VRChatLineTime: vrTime,
				})
			})
			if procErr != nil {
				if errors.Is(procErr, context.Canceled) {
					return
				}
				runtime.LogWarning(ctx, "activity log ingest: "+procErr.Error())
				return
			}
			_ = a.activity.CloseOpenPlaySessionAtLastLogLine(ctx, lastVRLineTime)
			_ = a.activity.CloseOpenEncountersAtLastLogLine(ctx, lastVRLineTime)
			st, statErr := os.Stat(pathCopy)
			endOff := int64(0)
			if statErr == nil && st != nil {
				endOff = st.Size()
			}
			_ = a.activity.SetActivityLogCheckpoint(ctx, &usecase.ActivityLogCheckpoint{
				WatchPath:  absWatch,
				File:       pathCopy,
				ByteOffset: endOff,
			})
		}
		return
	}

	off := int64(0)
	if cp != nil && matchAbsPaths(cp.WatchPath, absWatch) && matchAbsPaths(cp.File, absWatch) {
		off = cp.ByteOffset
	}
	pathCopy := absWatch
	checkpointLines := 0
	var lastVRLineTime time.Time
	if off == 0 {
		activityHandler.ResetSessionContextForNewLogFile()
	}
	_, fileProcErr := logwatcher.ProcessOutputLogFileFromOffset(ctx, pathCopy, off, parser, activityHandler, logger, func(pos int64, line string) {
		if ts := activity.ParseVRChatTimestamp(line, time.Time{}); !ts.IsZero() {
			lastVRLineTime = ts
		}
		checkpointLines++
		if checkpointLines != 1 && checkpointLines%32 != 0 {
			return
		}
		vrTime := ""
		if ts := activity.ParseVRChatTimestamp(line, time.Time{}); !ts.IsZero() {
			vrTime = ts.Format(time.RFC3339)
		}
		_ = a.activity.SetActivityLogCheckpoint(ctx, &usecase.ActivityLogCheckpoint{
			WatchPath:      absWatch,
			File:           pathCopy,
			ByteOffset:     pos,
			VRChatLineTime: vrTime,
		})
	})
	if fileProcErr != nil && !errors.Is(fileProcErr, context.Canceled) {
		runtime.LogWarning(ctx, "activity log ingest: "+fileProcErr.Error())
		return
	}
	if fileProcErr != nil {
		// context.Canceled: preserve last progress-callback checkpoint (same as directory mode)
		return
	}
	_ = a.activity.CloseOpenPlaySessionAtLastLogLine(ctx, lastVRLineTime)
	_ = a.activity.CloseOpenEncountersAtLastLogLine(ctx, lastVRLineTime)
	st, statErr := os.Stat(pathCopy)
	endOff := int64(0)
	if statErr == nil && st != nil {
		endOff = st.Size()
	}
	_ = a.activity.SetActivityLogCheckpoint(ctx, &usecase.ActivityLogCheckpoint{
		WatchPath:  absWatch,
		File:       pathCopy,
		ByteOffset: endOff,
	})
}

func (a *App) startOutputLogWatcher(ctx context.Context, eventBus event.EventBus) {
	watchPath, err := a.resolveEffectiveOutputLogWatchPath(ctx)
	if err != nil || watchPath == "" {
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			runtime.LogWarning(ctx, "output_log watch path: "+err.Error())
		}
		return
	}
	info, err := os.Stat(watchPath)
	if err != nil || info == nil {
		runtime.LogWarning(ctx, "output_log path not accessible, skipping log watcher")
		return
	}
	if !info.Mode().IsRegular() && !info.IsDir() {
		runtime.LogWarning(ctx, "output_log path must be a file or directory, skipping log watcher")
		return
	}

	parser := activity.NewLogParser()
	logger := &logLogger{}
	emitEncounters := func() {
		runtime.EventsEmit(a.ctx, activityEncountersChangedEvent, struct{}{})
	}
	activityHandler := logwatcher.NewActivityEventHandler(a.activity, ctx, logger, emitEncounters)
	publishHandler := logwatcher.NewEventPublishingHandler(eventBus, ctx, logger)
	handler := logwatcher.NewMultiHandler(activityHandler, publishHandler)

	activityHandler.SetSuppressEncounterNotify(true)
	a.ingestActivityLogsBootstrap(ctx, watchPath, parser, activityHandler, logger)
	activityHandler.SetSuppressEncounterNotify(false)

	watcher := logwatcher.NewOutputLogWatcher(watchPath, parser, handler, logger)
	watcher.SetOnActiveLogPathChange(activityHandler.ResetSessionContextForNewLogFile)
	if startErr := watcher.Start(ctx); startErr != nil {
		runtime.LogError(ctx, "failed to start output_log watcher: "+startErr.Error())
		return
	}
	runtime.LogInfo(ctx, "output_log watcher started for "+watchPath)
}

// ValidateOutputLogPath checks if path is a readable log file or a directory containing output_log*.txt.
func (a *App) ValidateOutputLogPath(path string) bool {
	return logwatcher.OutputLogPathValid(path)
}

func (a *App) resolveVRChatPictureWatchRoot() string {
	cfg, err := a.vrchatConfig.Get()
	if err == nil {
		if p := strings.TrimSpace(cfg.PictureOutputFolder); p != "" {
			return filepath.Clean(p)
		}
	}
	p, err := a.DefaultVRChatPictureFolder()
	if err != nil {
		return ""
	}
	return filepath.Clean(p)
}

func (a *App) startPictureFolderWatcher(ctx context.Context) {
	root := a.resolveVRChatPictureWatchRoot()
	if root == "" {
		runtime.LogWarning(ctx, "picture folder watcher: could not resolve VRChat picture directory")
		return
	}
	info, err := os.Stat(root)
	if err != nil {
		runtime.LogWarning(ctx, fmt.Sprintf("picture folder watcher: stat %s: %v", root, err))
		return
	}
	if !info.IsDir() {
		runtime.LogWarning(ctx, "picture folder watcher: not a directory: "+root)
		return
	}
	ingest := func(c context.Context, path string) error {
		_, created, ingestErr := a.media.IngestScreenshotFile(c, path)
		if ingestErr != nil {
			return ingestErr
		}
		if created {
			runtime.EventsEmit(a.ctx, galleryScreenshotsChangedEvent, struct{}{})
		}
		return nil
	}
	log := pictureWatchLogger{ctx: ctx}
	if err := picturewatcher.Start(ctx, root, ingest, log); err != nil {
		runtime.LogError(ctx, "failed to start picture folder watcher: "+err.Error())
		return
	}
	runtime.LogInfo(ctx, "picture folder watcher started for "+root)
}

type pictureWatchLogger struct {
	ctx context.Context
}

func (l pictureWatchLogger) Printf(format string, v ...any) {
	runtime.LogWarning(l.ctx, fmt.Sprintf(format, v...))
}

type logLogger struct{}

func (logLogger) Printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func getDataDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "vrchat-tweaker"), nil
}

// Greet returns a greeting (sample binding).
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, Welcome to VRChat Tweaker!", name)
}

// --- Launcher bindings ---

// LaunchProfiles returns all launch profiles.
func (a *App) LaunchProfiles() ([]LaunchProfileDTO, error) {
	list, err := a.launcher.ListProfiles(a.ctx)
	if err != nil {
		return nil, err
	}
	return toLaunchProfileDTOs(list), nil
}

// LaunchVRChatWithArgs starts VRChat with the given arguments string (from GUI state).
// Use when launching with current GUI settings without saving first.
func (a *App) LaunchVRChatWithArgs(args string) error {
	ps, err := a.settings.GetPathSettings(a.ctx)
	if err != nil {
		return err
	}
	vrchatPath := ""
	steamPath := ""
	outputLogPath := ""
	if ps != nil {
		vrchatPath = ps.VRChatPathWindows
		steamPath = ps.SteamPathLinux
		outputLogPath = ps.OutputLogPath
	}
	return a.launcher.LaunchWithArgs(a.ctx, args, vrchatPath, steamPath, outputLogPath)
}

// LaunchVRChat starts VRChat with the given profile ID.
// Uses path settings (VRChat/Steam paths) from app_settings when configured.
func (a *App) LaunchVRChat(profileID string) error {
	ps, err := a.settings.GetPathSettings(a.ctx)
	if err != nil {
		return err
	}
	vrchatPath := ""
	steamPath := ""
	outputLogPath := ""
	if ps != nil {
		vrchatPath = ps.VRChatPathWindows
		steamPath = ps.SteamPathLinux
		outputLogPath = ps.OutputLogPath
	}
	return a.launcher.LaunchVRChat(a.ctx, profileID, vrchatPath, steamPath, outputLogPath)
}

// JoinWorld launches VRChat into the specified world using default profile.
// Uses vrchat://launch?id=<worldID> URL scheme.
func (a *App) JoinWorld(worldID string) error {
	ps, err := a.settings.GetPathSettings(a.ctx)
	if err != nil {
		return err
	}
	vrchatPath := ""
	steamPath := ""
	outputLogPath := ""
	if ps != nil {
		vrchatPath = ps.VRChatPathWindows
		steamPath = ps.SteamPathLinux
		outputLogPath = ps.OutputLogPath
	}
	return a.launcher.LaunchToWorld(a.ctx, "", worldID, vrchatPath, steamPath, outputLogPath)
}

// JoinWorldFromScreenshot launches VRChat into the world associated with the screenshot.
// Returns error if screenshot has no world_id.
func (a *App) JoinWorldFromScreenshot(screenshotID string) error {
	s, err := a.media.GetScreenshot(a.ctx, screenshotID)
	if err != nil {
		return err
	}
	if s == nil {
		return fmt.Errorf("screenshot not found: %s", screenshotID)
	}
	if s.WorldID == "" {
		return fmt.Errorf("screenshot has no world_id")
	}
	return a.JoinWorld(s.WorldID)
}

// SaveLaunchProfile persists a launch profile.
func (a *App) SaveLaunchProfile(p LaunchProfileDTO) error {
	return a.launcher.SaveProfile(a.ctx, toLaunchProfile(p))
}

// DeleteLaunchProfile removes a launch profile by ID.
func (a *App) DeleteLaunchProfile(id string) error {
	return a.launcher.DeleteProfile(a.ctx, id)
}

// ParseLaunchArgsForGUI parses a launch arguments string into GUI fields.
func (a *App) ParseLaunchArgsForGUI(args string) LaunchArgsParsedDTO {
	p := launcher.ParseLaunchArgsForGUI(args)
	return toLaunchArgsParsedDTO(p)
}

// MergeLaunchArgsForGUI builds a launch arguments string from GUI state.
func (a *App) MergeLaunchArgsForGUI(dto LaunchArgsParsedDTO) string {
	return launcher.MergeLaunchArgsForGUI(fromLaunchArgsParsedDTO(dto))
}

// --- Settings bindings ---

// GetLogRetentionDays returns the log retention setting.
func (a *App) GetLogRetentionDays() (int, error) {
	return a.settings.GetLogRetentionDays(a.ctx)
}

// SetLogRetentionDays saves the log retention setting.
func (a *App) SetLogRetentionDays(days int) error {
	return a.settings.SetLogRetentionDays(a.ctx, days)
}

// GetOutputLogPath returns the output_log.txt path.
func (a *App) GetOutputLogPath() (string, error) {
	return a.settings.GetOutputLogPath(a.ctx)
}

// SaveOutputLogPath saves the output_log.txt path.
func (a *App) SaveOutputLogPath(path string) error {
	return a.settings.SaveOutputLogPath(a.ctx, path)
}

// GetPathSettings returns VRChat/Steam/output_log path settings.
func (a *App) GetPathSettings() (PathSettingsDTO, error) {
	ps, err := a.settings.GetPathSettings(a.ctx)
	if err != nil {
		return PathSettingsDTO{}, err
	}
	return toPathSettingsDTO(ps), nil
}

// SetPathSettings saves path settings.
func (a *App) SetPathSettings(dto PathSettingsDTO) error {
	return a.settings.SetPathSettings(a.ctx, toPathSettings(dto))
}

// ValidatePath checks if the path exists and is accessible.
func (a *App) ValidatePath(path string) bool {
	return a.settings.ValidatePath(path)
}

// OpenFileDialog opens a native file picker and returns the selected file path.
// title: dialog title, defaultDir: initial directory (empty = default), filterPattern: e.g. "*.txt" or "*.exe" (empty = all files)
func (a *App) OpenFileDialog(title, defaultDir, filterPattern string) (string, error) {
	opts := runtime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: defaultDir,
	}
	if filterPattern != "" {
		opts.Filters = []runtime.FileFilter{
			{DisplayName: "Filtered", Pattern: filterPattern},
			{DisplayName: "All Files", Pattern: "*"},
		}
	}
	return runtime.OpenFileDialog(a.ctx, opts)
}

// OpenDirectoryDialog opens a native directory picker and returns the selected directory path.
// title: dialog title, defaultDir: initial directory (empty = default)
func (a *App) OpenDirectoryDialog(title, defaultDir string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: defaultDir,
	})
}

// --- Media bindings ---

// Screenshots returns screenshots (optional worldId filter).
func (a *App) Screenshots(worldId string) ([]ScreenshotDTO, error) {
	filter := &media.ScreenshotFilter{}
	if worldId != "" {
		filter.WorldID = worldId
	}
	list, err := a.media.ListScreenshots(a.ctx, filter)
	if err != nil {
		return nil, err
	}
	return toScreenshotDTOs(list), nil
}

// SearchScreenshots returns screenshots matching the filter (worldId, worldName, dateFrom, dateTo).
func (a *App) SearchScreenshots(filter ScreenshotSearchDTO) ([]ScreenshotDTO, error) {
	f := toScreenshotFilter(filter)
	list, err := a.media.ListScreenshots(a.ctx, f)
	if err != nil {
		return nil, err
	}
	return toScreenshotDTOs(list), nil
}

// GetScreenshot returns a screenshot by ID, or nil if not found.
func (a *App) GetScreenshot(id string) (*ScreenshotDTO, error) {
	s, err := a.media.GetScreenshot(a.ctx, id)
	if err != nil || s == nil {
		return nil, err
	}
	return toScreenshotDTO(s), nil
}

// ScreenshotThumbnailDataURL returns a JPEG data URL for the screenshot thumbnail (for WebView; avoids file://).
// Uses the DB cache when valid; otherwise builds and stores a thumbnail from the source file (lazy fill for legacy rows).
func (a *App) ScreenshotThumbnailDataURL(id string) (string, error) {
	return a.media.ScreenshotThumbnailDataURL(a.ctx, id)
}

// OpenScreenshotExternally opens the screenshot file with the OS default application.
func (a *App) OpenScreenshotExternally(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("screenshot id is empty")
	}
	s, err := a.media.GetScreenshot(a.ctx, id)
	if err != nil {
		return err
	}
	if s == nil {
		return fmt.Errorf("screenshot not found")
	}
	return desktop.OpenFileWithDefaultApp(s.FilePath)
}

// RevealScreenshotInFileManager opens the file manager and shows the screenshot file where supported.
func (a *App) RevealScreenshotInFileManager(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("screenshot id is empty")
	}
	s, err := a.media.GetScreenshot(a.ctx, id)
	if err != nil {
		return err
	}
	if s == nil {
		return fmt.Errorf("screenshot not found")
	}
	return desktop.RevealInFileManager(s.FilePath)
}

const galleryScanProgressEvent = "gallery:scan-progress"

// galleryScanDoneEvent is emitted when ScanScreenshotDir finishes (success, error, or cancel).
const galleryScanDoneEvent = "gallery:scan-done"

// galleryScreenshotsChangedEvent is emitted when the picture folder watcher ingests a new screenshot row.
const galleryScreenshotsChangedEvent = "gallery:screenshots-changed"

// activityEncountersChangedEvent is emitted when a new encounter row is written from the log watcher.
const activityEncountersChangedEvent = "activity:encounters-changed"

const galleryScanProgressEmitMinInterval = 90 * time.Millisecond

// scanProgressEmitter throttles gallery:scan-progress EventsEmit; flush sends the latest pending payload.
type scanProgressEmitter struct {
	ctx        context.Context
	mu         sync.Mutex
	lastEmit   time.Time
	pending    usecase.ScanProgress
	hasPending bool
}

func (e *scanProgressEmitter) emit(p usecase.ScanProgress) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pending = p
	e.hasPending = true
	if time.Since(e.lastEmit) >= galleryScanProgressEmitMinInterval {
		runtime.EventsEmit(e.ctx, galleryScanProgressEvent, toScanProgressDTO(e.pending))
		e.lastEmit = time.Now()
		e.hasPending = false
	}
}

func (e *scanProgressEmitter) flush() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.hasPending {
		return
	}
	runtime.EventsEmit(e.ctx, galleryScanProgressEvent, toScanProgressDTO(e.pending))
	e.lastEmit = time.Now()
	e.hasPending = false
}

// ScanScreenshotDir scans a directory for screenshots.
func (a *App) ScanScreenshotDir(path string) (int, error) {
	a.galleryScanMu.Lock()
	for a.galleryScanCancel != nil {
		cancelFn := a.galleryScanCancel
		a.galleryScanMu.Unlock()
		cancelFn()
		a.galleryScanWG.Wait()
		a.galleryScanMu.Lock()
	}
	a.galleryScanWG.Add(1)
	scanCtx, cancel := context.WithCancel(a.ctx)
	a.galleryScanCancel = cancel
	a.galleryScanMu.Unlock()

	defer a.galleryScanWG.Done()
	defer func() {
		cancel()
		a.galleryScanMu.Lock()
		a.galleryScanCancel = nil
		a.galleryScanMu.Unlock()
	}()

	em := &scanProgressEmitter{ctx: a.ctx}
	var count int
	var err error
	defer func() {
		em.flush()
		dto := GalleryScanDoneDTO{Count: count}
		if err != nil {
			dto.Error = err.Error()
			if errors.Is(err, context.Canceled) {
				dto.Cancelled = true
			}
		}
		runtime.EventsEmit(a.ctx, galleryScanDoneEvent, dto)
	}()

	count, err = a.media.ScanDirectory(scanCtx, path, em.emit)
	return count, err
}

// IsGalleryScanning reports whether a gallery folder scan is in progress.
func (a *App) IsGalleryScanning() bool {
	a.galleryScanMu.Lock()
	defer a.galleryScanMu.Unlock()
	return a.galleryScanCancel != nil
}

func (a *App) stopGalleryScanAndWait() {
	a.galleryScanMu.Lock()
	cancelFn := a.galleryScanCancel
	a.galleryScanMu.Unlock()
	if cancelFn != nil {
		cancelFn()
	}
	a.galleryScanWG.Wait()
}

// ReindexScreenshotDir re-extracts metadata for existing screenshots under path.
// Returns the number of updated records.
func (a *App) ReindexScreenshotDir(path string) (int, error) {
	return a.media.ReindexScreenshots(a.ctx, path)
}

// --- Activity bindings ---

// GetActivityStats returns aggregated play stats for the date range.
func (a *App) GetActivityStats(fromISO, toISO string) (ActivityStatsDTO, error) {
	stats, err := a.activity.GetActivityStats(a.ctx, fromISO, toISO)
	if err != nil {
		return ActivityStatsDTO{}, err
	}
	return toActivityStatsDTO(stats), nil
}

// Encounters returns user encounters.
func (a *App) Encounters() ([]UserEncounterDTO, error) {
	list, err := a.activity.ListEncountersWithContext(a.ctx, nil)
	if err != nil {
		return nil, err
	}
	return toEncounterDTOsFromContext(list), nil
}

// EncountersByVRCUserID returns encounters for the given VRChat user id. Empty id yields an empty slice.
func (a *App) EncountersByVRCUserID(vrcUserID string) ([]UserEncounterDTO, error) {
	if strings.TrimSpace(vrcUserID) == "" {
		return []UserEncounterDTO{}, nil
	}
	list, err := a.activity.ListEncountersWithContext(a.ctx, &activity.EncounterFilter{VRCUserID: vrcUserID})
	if err != nil {
		return nil, err
	}
	return toEncounterDTOsFromContext(list), nil
}

// EncountersByWorldID returns encounters in the given world. Empty id yields an empty slice.
func (a *App) EncountersByWorldID(worldID string) ([]UserEncounterDTO, error) {
	if strings.TrimSpace(worldID) == "" {
		return []UserEncounterDTO{}, nil
	}
	list, err := a.activity.ListEncountersWithContext(a.ctx, &activity.EncounterFilter{WorldID: worldID})
	if err != nil {
		return nil, err
	}
	return toEncounterDTOsFromContext(list), nil
}

// OpenVRChatLogFolder opens the configured output_log directory (or default VRChat log dir) in the file manager.
func (a *App) OpenVRChatLogFolder() error {
	p, err := a.settings.GetOutputLogPath(a.ctx)
	if err != nil {
		return err
	}
	var dir string
	if strings.TrimSpace(p) != "" {
		abs, err := filepath.Abs(filepath.Clean(p))
		if err != nil {
			return err
		}
		info, err := os.Stat(abs)
		if err != nil {
			return err
		}
		if info.IsDir() {
			dir = abs
		} else {
			dir = filepath.Dir(abs)
		}
	} else {
		dir = defaultVRChatOutputLogDir()
	}
	return desktop.OpenFolderInFileManager(dir)
}

// RotateEncounters runs the retention cleanup.
func (a *App) RotateEncounters() (int64, error) {
	return a.activity.RotateEncounters(a.ctx)
}

// --- Identity bindings ---

// Login authenticates with VRChat and saves credentials to OS keyring.
func (a *App) Login(username, password, twoFactorCode string) LoginResultDTO {
	if err := a.identity.Login(a.ctx, username, password, twoFactorCode); err != nil {
		return LoginResultDTO{OK: false, Error: err.Error()}
	}
	return LoginResultDTO{OK: true}
}

// Logout clears stored credentials.
func (a *App) Logout() error {
	return a.identity.Logout(a.ctx)
}

// IsLoggedIn returns true if we have stored credentials.
func (a *App) IsLoggedIn() (bool, error) {
	return a.identity.IsLoggedIn(a.ctx)
}

// GetVRChatCurrentUser returns the logged-in user's profile from the VRChat API.
// When forceRefresh is true, bypasses the local self-profile cache and refetches from the API.
func (a *App) GetVRChatCurrentUser(forceRefresh bool) (VRChatCurrentUserDTO, error) {
	u, err := a.identity.GetCurrentUser(a.ctx, forceRefresh)
	if err != nil {
		return VRChatCurrentUserDTO{}, err
	}
	if u == nil {
		return VRChatCurrentUserDTO{}, errors.New("empty current user")
	}
	return VRChatCurrentUserDTO{
		ID:                             u.ID,
		DisplayName:                    u.DisplayName,
		Username:                       u.Username,
		Status:                         u.Status,
		StatusDescription:              u.StatusDescription,
		State:                          u.State,
		CurrentAvatarThumbnailImageURL: u.CurrentAvatarThumbnailImageURL,
		UserIcon:                       u.UserIcon,
		ProfilePicOverrideThumbnail:    u.ProfilePicOverrideThumbnail,
	}, nil
}

// RefreshFriends fetches friends from API and updates cache.
func (a *App) RefreshFriends() error {
	return a.identity.RefreshFriends(a.ctx)
}

// Friends returns cached friends.
func (a *App) Friends() ([]UserCacheDTO, error) {
	list, err := a.identity.ListFriends(a.ctx)
	if err != nil {
		return nil, err
	}
	return toUserCacheDTOs(list), nil
}

// SetFavorite updates a friend's favorite flag.
func (a *App) SetFavorite(vrcUserID string, favorite bool) error {
	return a.identity.SetFavorite(a.ctx, vrcUserID, favorite)
}

// SetStatus changes the user's VRChat status.
func (a *App) SetStatus(status string) error {
	return a.identity.SetStatus(a.ctx, status)
}

// --- Automation bindings ---

// ListAutomationRules returns all automation rules.
func (a *App) ListAutomationRules() ([]AutomationRuleDTO, error) {
	list, err := a.automation.ListRules(a.ctx)
	if err != nil {
		return nil, err
	}
	return toAutomationRuleDTOs(list), nil
}

// SaveAutomationRule persists an automation rule.
func (a *App) SaveAutomationRule(rule AutomationRuleDTO) error {
	return a.automation.SaveRule(a.ctx, toAutomationRule(rule))
}

// DeleteAutomationRule removes an automation rule by ID.
func (a *App) DeleteAutomationRule(id string) error {
	return a.automation.DeleteRule(a.ctx, id)
}

// ToggleAutomationRule enables or disables an automation rule.
func (a *App) ToggleAutomationRule(id string, enabled bool) error {
	return a.automation.ToggleRule(a.ctx, id, enabled)
}

// --- DB Maintenance bindings ---

// VacuumDb runs SQLite VACUUM to optimize the database.
func (a *App) VacuumDb() error {
	return a.dbMaintenance.VacuumDb(a.ctx)
}

// ClearEncounters deletes all user encounters. Returns affected row count.
func (a *App) ClearEncounters() (int64, error) {
	return a.dbMaintenance.ClearEncounters(a.ctx)
}

// ClearScreenshots deletes all screenshots. Returns affected row count.
func (a *App) ClearScreenshots() (int64, error) {
	a.stopGalleryScanAndWait()
	return a.dbMaintenance.ClearScreenshots(a.ctx)
}

// ClearFriendsCache deletes all cached friends. Returns affected row count.
func (a *App) ClearFriendsCache() (int64, error) {
	return a.dbMaintenance.ClearFriendsCache(a.ctx)
}

// --- VRChat Config bindings ---

// VRChatConfigExists checks if config.json exists.
func (a *App) VRChatConfigExists() (bool, error) {
	return a.vrchatConfig.Exists()
}

// GetVRChatConfig reads the current config.json.
func (a *App) GetVRChatConfig() (VRChatConfigDTO, error) {
	cfg, err := a.vrchatConfig.Get()
	if err != nil {
		return VRChatConfigDTO{}, err
	}
	return toVRChatConfigDTO(cfg), nil
}

// SaveVRChatConfig writes config.json.
func (a *App) SaveVRChatConfig(dto VRChatConfigDTO) error {
	return a.vrchatConfig.Save(fromVRChatConfigDTO(dto))
}

// DeleteVRChatConfig removes config.json.
func (a *App) DeleteVRChatConfig() error {
	return a.vrchatConfig.Delete()
}

// DefaultVRChatPictureFolder returns the folder VRChat uses when picture_output_folder
// is unset: filepath.Join(home, "Pictures", "VRChat") (e.g. ~/Pictures/VRChat on Unix,
// %USERPROFILE%\Pictures\VRChat on Windows).
//
// Limitation: this does not resolve the shell “My Pictures” location (e.g. Windows folder
// redirection to another drive). VRChat may follow the OS special folder; Go’s stdlib has
// no direct equivalent, so this path is a conventional default and may differ from the
// actual save location when Pictures is redirected.
func (a *App) DefaultVRChatPictureFolder() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Pictures", "VRChat"), nil
}

// getVRChatConfigPath returns the path to VRChat's config.json.
// On Windows: %LocalAppData%Low\VRChat\VRChat\config.json
// On other OS: falls back to ~/.local/share/VRChat/VRChat/config.json
func getVRChatConfigPath() string {
	if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
		// Windows: %LOCALAPPDATA% is typically C:\Users\<user>\AppData\Local
		// config.json lives in LocalLow, which is ../LocalLow relative to Local
		return filepath.Join(filepath.Dir(dir), "LocalLow", "VRChat", "VRChat", "config.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "config.json")
	}
	return filepath.Join(home, ".local", "share", "VRChat", "VRChat", "config.json")
}
