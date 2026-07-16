package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	goruntime "runtime" // stdlib; wails v2/pkg/runtime is imported as runtime below
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/automation"
	"vrchat-tweaker/internal/domain/launcher"
	"vrchat-tweaker/internal/domain/media"
	"vrchat-tweaker/internal/domain/vrchatconfig"
	"vrchat-tweaker/internal/infrastructure/desktop"
	"vrchat-tweaker/internal/infrastructure/diag"
	"vrchat-tweaker/internal/infrastructure/filesystem"
	"vrchat-tweaker/internal/infrastructure/logwatcher"
	"vrchat-tweaker/internal/infrastructure/picturewatcher"
	"vrchat-tweaker/internal/infrastructure/sleepsuppress"
	"vrchat-tweaker/internal/infrastructure/sqlite"
	"vrchat-tweaker/internal/infrastructure/statuspage"
	"vrchat-tweaker/internal/infrastructure/vrchatapi"
	"vrchat-tweaker/internal/infrastructure/vrchatpipeline"
	"vrchat-tweaker/internal/infrastructure/ytdlpmaintain"
	"vrchat-tweaker/internal/locale"
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
	dbMaintenance    *usecase.DBMaintenanceUseCase
	ytdlp            *usecase.YTDLPMaintainUseCase
	serverStatus     *statuspage.Client
	vrchatConfigRepo vrchatconfig.ConfigRepository

	galleryScanMu     sync.Mutex
	galleryScanCancel context.CancelFunc
	galleryScanWG     sync.WaitGroup

	pipelineMu     sync.Mutex
	pipelineCancel context.CancelFunc
	pipelineWG     sync.WaitGroup

	sleepSuppressMu     sync.Mutex
	sleepSuppressCancel context.CancelFunc
	sleepSuppressWG     sync.WaitGroup

	ytdlpMaintainMu     sync.Mutex
	ytdlpMaintainCancel context.CancelFunc
	ytdlpMaintainWG     sync.WaitGroup

	activityWatchMu     sync.Mutex
	activityWatchCancel context.CancelFunc
	activityWatchWG     sync.WaitGroup

	activityIngestMu       sync.Mutex
	activityIngestAdapters map[string]*logwatcher.ActivityIngestAdapter
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

	launcherRepo := sqlite.NewLauncherProfileRepository(db)
	mediaRepo := sqlite.NewScreenshotRepository(db)
	playRepo := sqlite.NewPlaySessionRepository(db)
	encounterRepo := sqlite.NewUserEncounterRepository(db)
	userCacheRepo := sqlite.NewUserCacheRepository(db)
	worldRepo := sqlite.NewWorldInfoRepository(db)
	automationRepo := sqlite.NewAutomationRuleRepository(db)
	settingsRepo := sqlite.NewAppSettingsRepository(db)

	credStore := vrchatapi.NewAutoCredentialStore(dataDir, func(msg string) {
		runtime.LogWarning(ctx, msg)
	})
	// Do NOT set the auth token here. The frontend must call GetCredentialBlob,
	// decrypt the wrapped blob with Web Crypto, and then call UnlockVRChatSession.
	apiClient := vrchatapi.NewClient("")

	extractor := media.NewDefaultMetadataExtractor()
	const defaultNotificationTitle = "VRChat Tweaker"
	notify := func(title, message string) error {
		if title == "" {
			title = defaultNotificationTitle
		}
		return beeep.Notify(title, message, "")
	}
	a.launcher = usecase.NewLauncherUseCase(launcherRepo)
	a.media = usecase.NewMediaUseCase(mediaRepo, extractor, worldRepo, userCacheRepo)
	a.activity = usecase.NewActivityUseCase(playRepo, encounterRepo, settingsRepo, userCacheRepo, worldRepo)
	a.identity = usecase.NewIdentityUseCase(userCacheRepo, apiClient, credStore, settingsRepo, notify)
	a.automation = usecase.NewAutomationUseCase(automationRepo, a.identity)
	a.settings = usecase.NewSettingsUseCase(settingsRepo)
	a.dbMaintenance = usecase.NewDBMaintenanceUseCase(db, encounterRepo, mediaRepo, userCacheRepo, settingsRepo)
	a.ytdlp = usecase.NewYTDLPMaintainUseCase(a.settings, usecase.NewYTDLPUpdater())
	a.serverStatus = statuspage.NewClient()

	configPath := getVRChatConfigPath()
	configRepo := filesystem.NewVRChatConfigFileRepository(configPath)
	a.vrchatConfigRepo = configRepo

	// Start output_log watcher if path is configured
	a.startOutputLogWatcher(ctx)

	a.startPictureFolderWatcher(ctx)
	go a.startupGalleryIncremental()
	a.startSleepSuppressLoop()
	a.startYTDLPMaintainLoop()
}

// onShutdown persists state before the process exits (Wails lifecycle).
func (a *App) onShutdown(ctx context.Context) {
	a.stopVRChatActivityMonitor()
	a.stopSleepSuppressLoop()
	a.stopYTDLPMaintainLoop()
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
	dir, cleared, err := a.settings.EnsureOutputLogWatchDir(ctx)
	if err != nil {
		return "", err
	}
	if cleared {
		runtime.LogWarning(ctx, "cleared stale output_log_path setting; using default log folder")
	}
	if dir != "" {
		info, statErr := os.Stat(dir)
		if statErr != nil {
			return "", statErr
		}
		if !info.IsDir() {
			return "", fmt.Errorf("output_log watch path is not a directory: %s", dir)
		}
		return dir, nil
	}
	defaultDir := defaultVRChatOutputLogDir()
	if defaultDir == "" {
		return "", os.ErrNotExist
	}
	absDir, err := filepath.Abs(defaultDir)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(absDir); err != nil {
		return "", err
	}
	return absDir, nil
}

func (a *App) ingestActivityLogsBootstrap(ctx context.Context, absWatch string, parser *activity.LogParser, logger logwatcher.Logger, emitEncounters func()) {
	info, err := os.Stat(absWatch)
	if err != nil {
		runtime.LogWarning(ctx, "activity log bootstrap skipped: "+err.Error())
		return
	}
	if !info.IsDir() {
		runtime.LogWarning(ctx, "activity log bootstrap skipped: not a directory: "+absWatch)
		return
	}
	cp, _ := a.activity.GetActivityLogCheckpoint(ctx)

	files, listErr := logwatcher.ListOutputLogFiles(absWatch)
	if listErr != nil {
		return
	}
	live := bootstrapLiveLogFiles(files)
	for _, fp := range files {
		finalize := live == nil || !live[fp]
		a.ingestOneActivityLogBootstrap(ctx, absWatch, fp, parser, logger, emitEncounters, cp, finalize, nil)
	}
}

func bootstrapLiveLogFiles(paths []string) map[string]bool {
	const liveWindow = 5 * time.Second
	type fileMod struct {
		path string
		mod  time.Time
	}
	var files []fileMod
	var maxMod time.Time
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		mod := info.ModTime()
		files = append(files, fileMod{path: p, mod: mod})
		if mod.After(maxMod) {
			maxMod = mod
		}
	}
	if maxMod.IsZero() {
		return nil
	}
	live := make(map[string]bool)
	for _, f := range files {
		if maxMod.Sub(f.mod) <= liveWindow {
			live[f.path] = true
		}
	}
	return live
}

func (a *App) ingestOneActivityLogBootstrap(
	ctx context.Context,
	absWatch, filePath string,
	parser *activity.LogParser,
	logger logwatcher.Logger,
	emitEncounters func(),
	cp *usecase.ActivityLogCheckpoint,
	finalizeAtEnd bool,
	ingestAdapter *logwatcher.ActivityIngestAdapter,
) {
	absFile := absLogPath(filePath)
	if ingestAdapter == nil {
		ingestAdapter = a.activityIngestAdapterForPath(ctx, logger, emitEncounters, filePath)
	}
	off := int64(0)
	if cp != nil && matchAbsPaths(cp.WatchPath, absWatch) {
		if fc, ok := cp.FileCheckpoint(absFile); ok {
			off = fc.ByteOffset
			if st, statErr := os.Stat(filePath); statErr == nil && st != nil && st.Size() > 0 && off >= st.Size() {
				if finalizeAtEnd {
					lastVRLineTime := time.Time{}
					if fc.VRChatLineTime != "" {
						if t, parseErr := time.Parse(time.RFC3339, fc.VRChatLineTime); parseErr == nil {
							lastVRLineTime = t
						}
					}
					if lastVRLineTime.IsZero() {
						if t, tErr := logwatcher.LastVRChatLineTimeInFile(filePath); tErr == nil {
							lastVRLineTime = t
						}
					}
					if !lastVRLineTime.IsZero() {
						_ = a.activity.FinalizeOpenActivityForLogSource(ctx, ingestAdapter.LogSourcePath(), lastVRLineTime)
					}
				}
				return
			}
		}
	}
	pathCopy := filePath
	ingestAdapter.SetSuppressEncounterNotify(true)
	defer ingestAdapter.SetSuppressEncounterNotify(false)
	checkpointLines := 0
	var lastVRLineTime time.Time
	if off == 0 {
		ingestAdapter.ResetSessionContextForNewLogFile()
	} else if warmErr := logwatcher.WarmSessionCorrelatorFromLogFile(ctx, pathCopy, off, parser, ingestAdapter, logger); warmErr != nil {
		runtime.LogWarning(ctx, "activity log correlator warm: "+warmErr.Error())
	}
	_, procErr := logwatcher.ProcessOutputLogFileFromOffset(ctx, pathCopy, off, parser, ingestAdapter, logger, func(pos int64, line string) {
		ts := activity.ParseVRChatTimestamp(line, time.Time{})
		if !ts.IsZero() {
			lastVRLineTime = ts
		}
		checkpointLines++
		if checkpointLines != 1 && checkpointLines%32 != 0 {
			return
		}
		_ = a.activity.SetActivityLogFileCheckpoint(ctx, absWatch, absLogPath(pathCopy), pos, checkpointVRTime(ts))
	})
	if procErr != nil {
		if errors.Is(procErr, context.Canceled) {
			return
		}
		runtime.LogWarning(ctx, "activity log ingest: "+procErr.Error())
		return
	}
	if finalizeAtEnd && !lastVRLineTime.IsZero() {
		_ = a.activity.FinalizeOpenActivityForLogSource(ctx, ingestAdapter.LogSourcePath(), lastVRLineTime)
	}
	st, statErr := os.Stat(pathCopy)
	endOff := int64(0)
	if statErr == nil && st != nil {
		endOff = st.Size()
	}
	_ = a.activity.SetActivityLogFileCheckpoint(ctx, absWatch, absLogPath(pathCopy), endOff, checkpointVRTime(lastVRLineTime))
}

func (a *App) startOutputLogWatcher(ctx context.Context) {
	watchPath, err := a.resolveEffectiveOutputLogWatchPath(ctx)
	if err != nil || watchPath == "" {
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			runtime.LogWarning(ctx, "output_log watch path: "+err.Error())
		}
		return
	}
	info, err := os.Stat(watchPath)
	if err != nil || info == nil || !info.IsDir() {
		runtime.LogWarning(ctx, "output_log path not accessible as directory, skipping log watcher")
		return
	}

	parser := activity.NewLogParser()
	logger := appDiagLogger()
	emitEncounters := func() {
		runtime.EventsEmit(a.ctx, activityEncountersChangedEvent, struct{}{})
	}

	a.resetActivityIngestAdapterCache()
	a.ingestActivityLogsBootstrap(ctx, watchPath, parser, logger, emitEncounters)
	if _, dedupeErr := a.activity.DeduplicateEncounters(ctx); dedupeErr != nil {
		runtime.LogWarning(ctx, "activity encounter dedupe: "+dedupeErr.Error())
	}
	if _, backfillErr := a.activity.BackfillEncounterWorldContext(ctx); backfillErr != nil {
		runtime.LogWarning(ctx, "activity encounter world backfill: "+backfillErr.Error())
	}

	watchDeps := activityLogWatchDeps{
		watchPath:      watchPath,
		parser:         parser,
		logger:         logger,
		emitEncounters: emitEncounters,
	}

	watcher := logwatcher.NewMultiOutputLogWatcher(watchPath, parser, func(logPath string) logwatcher.EventHandler {
		adapter := a.activityIngestAdapterForPath(ctx, logger, emitEncounters, logPath)
		triggerHandler := logwatcher.NewAutomationTriggerHandler(a.automation, ctx, logger)
		return logwatcher.NewMultiHandler(adapter, triggerHandler)
	}, logwatcher.MultiOutputLogWatcherCallbacks{
		OnLogRotationHandoff: func(c context.Context, oldPath string) error {
			return a.handleActivityLogRotationHandoff(c, watchDeps, oldPath)
		},
		OnTailCheckpoint: func(c context.Context, path string, offset int64, lineTime time.Time) {
			if a.activity == nil {
				return
			}
			_ = a.activity.SetActivityLogFileCheckpoint(c, watchDeps.watchPath, absLogPath(path), offset, checkpointVRTime(lineTime))
		},
	}, logger)
	if startErr := watcher.Start(ctx); startErr != nil {
		runtime.LogError(ctx, "failed to start multi output_log watcher: "+startErr.Error())
		return
	}

	a.startVRChatActivityMonitor(ctx, watchPath)
	runtime.LogInfo(ctx, "output_log watcher started for "+watchPath)
}

// ValidateOutputLogPath checks if path is an existing directory (empty dirs allowed).
func (a *App) ValidateOutputLogPath(path string) bool {
	return logwatcher.OutputLogPathValid(path)
}

func (a *App) resolveVRChatPictureWatchRoot() string {
	cfg, err := a.vrchatConfigRepo.Read()
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
	if err := picturewatcher.Start(ctx, root, ingest, diag.Logger(log.Printf)); err != nil {
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

func appDiagLogger() diag.Logger {
	ll := logLogger{}
	return diag.Logger(ll.Printf)
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
// lastLaunchProfileID updates Last launch profile on success when non-empty (Profile launch).
func (a *App) LaunchVRChatWithArgs(args string, lastLaunchProfileID string) error {
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
	launchErr := a.launcher.LaunchWithArgs(a.ctx, args, vrchatPath, steamPath, outputLogPath)
	return a.setLastLaunchProfileOnSuccess(lastLaunchProfileID, launchErr)
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
	launchErr := a.launcher.LaunchVRChat(a.ctx, profileID, vrchatPath, steamPath, outputLogPath)
	return a.setLastLaunchProfileOnSuccess(profileID, launchErr)
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
func (a *App) ParseLaunchArgsForGUI(args string) launcher.LaunchArgsParsed {
	p := launcher.ParseLaunchArgsForGUI(args)
	if p == nil {
		return launcher.LaunchArgsParsed{}
	}
	return *p
}

// MergeLaunchArgsForGUI builds a launch arguments string from GUI state.
func (a *App) MergeLaunchArgsForGUI(parsed launcher.LaunchArgsParsed) string {
	return launcher.MergeLaunchArgsForGUI(&parsed)
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
func (a *App) GetPathSettings() (usecase.PathSettings, error) {
	ps, err := a.settings.GetPathSettings(a.ctx)
	if err != nil {
		return usecase.PathSettings{}, err
	}
	if ps == nil {
		return usecase.PathSettings{}, nil
	}
	return *ps, nil
}

// SetPathSettings saves path settings.
func (a *App) SetPathSettings(ps usecase.PathSettings) error {
	return a.settings.SetPathSettings(a.ctx, &ps)
}

// GetSuppressSleepWhileVRChat returns whether sleep is suppressed while VRChat.exe runs (Windows).
func (a *App) GetSuppressSleepWhileVRChat() (bool, error) {
	return a.settings.GetSuppressSleepWhileVRChat(a.ctx)
}

// SetSuppressSleepWhileVRChat enables or disables sleep suppression while VRChat.exe runs.
func (a *App) SetSuppressSleepWhileVRChat(on bool) error {
	return a.settings.SetSuppressSleepWhileVRChat(a.ctx, on)
}

// GetLanguage returns the saved UI language code (ja, en, ko, zh-TW, zh-CN), or empty if unset.
func (a *App) GetLanguage() (string, error) {
	return a.settings.GetLanguage(a.ctx)
}

// SetLanguage persists the UI language code.
func (a *App) SetLanguage(lang string) error {
	return a.settings.SetLanguage(a.ctx, lang)
}

// GetSystemLocale returns the OS-preferred UI language mapped to a supported app code.
func (a *App) GetSystemLocale() string {
	return locale.Detect()
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

// Screenshots returns screenshots (optional worldId filter) within the current picture folder.
func (a *App) Screenshots(worldId string) ([]ScreenshotDTO, error) {
	filter := &media.ScreenshotFilter{}
	if worldId != "" {
		filter.WorldID = worldId
	}
	return a.listGalleryScreenshotDTOs(filter)
}

// SearchScreenshots returns screenshots matching the filter (worldId, worldName, dateFrom, dateTo).
func (a *App) SearchScreenshots(filter ScreenshotSearchDTO) ([]ScreenshotDTO, error) {
	f := toScreenshotFilter(filter)
	return a.listGalleryScreenshotDTOs(f)
}

func (a *App) listGalleryScreenshotDTOs(filter *media.ScreenshotFilter) ([]ScreenshotDTO, error) {
	root := a.resolveVRChatPictureWatchRoot()
	list, err := a.media.ListScreenshotsInGalleryScope(a.ctx, root, filter)
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

// friendsChangedEvent is emitted when the friends cache may have changed (Pipeline or REST reconcile).
const friendsChangedEvent = "vrchat:friends-changed"

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
		runtime.EventsEmit(e.ctx, galleryScanProgressEvent, e.pending)
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
	runtime.EventsEmit(e.ctx, galleryScanProgressEvent, e.pending)
	e.lastEmit = time.Now()
	e.hasPending = false
}

// ScanScreenshotDir synchronizes a picture folder (ingest new files + selective reindex).
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
		dto := usecase.GalleryScanDone{Count: count}
		if err != nil {
			dto.Error = err.Error()
			if errors.Is(err, context.Canceled) {
				dto.Cancelled = true
			}
		}
		runtime.EventsEmit(a.ctx, galleryScanDoneEvent, dto)
	}()

	count, err = a.media.SyncPictureFolder(scanCtx, path, em.emit)
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
func (a *App) GetActivityStats(fromISO, toISO string) (activity.ActivityStats, error) {
	stats, err := a.activity.GetActivityStats(a.ctx, fromISO, toISO)
	if err != nil {
		return activity.ActivityStats{}, err
	}
	if stats == nil {
		return activity.ActivityStats{
			DailyPlaySeconds: []activity.DailyPlaySeconds{},
			TopWorlds:        []activity.TopWorldSummary{},
		}, nil
	}
	if stats.DailyPlaySeconds == nil {
		stats.DailyPlaySeconds = []activity.DailyPlaySeconds{}
	}
	if stats.TopWorlds == nil {
		stats.TopWorlds = []activity.TopWorldSummary{}
	}
	return *stats, nil
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

func pipelineEventUpdatesFriends(typ string) bool {
	switch typ {
	case "friend-delete", "friend-offline", "friend-active", "friend-online",
		"friend-location", "friend-update", "friend-add":
		return true
	default:
		return false
	}
}

// stopVRChatPipelineCancel stops the pipeline without waiting. Do not call stopVRChatPipeline
// (which waits) from callbacks invoked synchronously inside vrchatpipeline.Run — that deadlocks.
func (a *App) stopVRChatPipelineCancel() {
	var cancelFn context.CancelFunc
	a.pipelineMu.Lock()
	if a.pipelineCancel != nil {
		cancelFn = a.pipelineCancel
		a.pipelineCancel = nil
	}
	a.pipelineMu.Unlock()
	if cancelFn != nil {
		cancelFn()
	}
}

func (a *App) stopVRChatPipeline() {
	a.stopVRChatPipelineCancel()
	a.pipelineWG.Wait()
}

func (a *App) startVRChatPipeline() {
	a.pipelineMu.Lock()
	for a.pipelineCancel != nil {
		cancelFn := a.pipelineCancel
		a.pipelineCancel = nil
		a.pipelineMu.Unlock()
		cancelFn()
		a.pipelineWG.Wait()
		a.pipelineMu.Lock()
	}
	if a.identity == nil {
		a.pipelineMu.Unlock()
		return
	}
	token := a.identity.CurrentAuthToken()
	if token == "" {
		a.pipelineMu.Unlock()
		return
	}
	runCtx, cancel := context.WithCancel(a.ctx)
	a.pipelineCancel = cancel
	cfg := vrchatpipeline.Config{
		AuthToken: token,
		UserAgent: vrchatapi.UserAgent,
		OnReconnect: func(ctx context.Context) error {
			err := a.identity.PipelineReconnectRestSync(ctx)
			if err != nil {
				if errors.Is(err, vrchatapi.ErrSessionExpired) || errors.Is(err, vrchatapi.ErrNotAuthenticated) {
					a.stopVRChatPipelineCancel()
				}
				return err
			}
			runtime.EventsEmit(a.ctx, friendsChangedEvent, struct{}{})
			return nil
		},
		OnEvent: func(ctx context.Context, typ string, payload []byte) error {
			err := a.identity.HandleVRChatPipelineEvent(ctx, typ, payload)
			if err == nil && pipelineEventUpdatesFriends(typ) {
				runtime.EventsEmit(a.ctx, friendsChangedEvent, struct{}{})
			}
			return err
		},
	}
	a.pipelineWG.Add(1)
	a.pipelineMu.Unlock()
	go func() {
		defer a.pipelineWG.Done()
		err := vrchatpipeline.Run(runCtx, cfg)
		if err != nil && !errors.Is(err, context.Canceled) {
			runtime.LogWarning(a.ctx, "vrchat pipeline: "+err.Error())
		}
	}()
}

// --- Identity bindings ---

// Login authenticates with VRChat. On success the returned PlaintextToken must be
// immediately wrapped by the frontend via Web Crypto and persisted via
// PersistWrappedCredential. The token is held in Go memory for the current session.
func (a *App) Login(username, password, twoFactorCode string) LoginResultDTO {
	token, err := a.identity.Login(a.ctx, username, password, twoFactorCode)
	if err != nil {
		return LoginResultDTO{OK: false, Error: err.Error()}
	}
	a.startVRChatPipeline()
	return LoginResultDTO{OK: true, PlaintextToken: token}
}

// Logout clears stored credentials.
func (a *App) Logout() error {
	a.stopVRChatPipeline()
	return a.identity.Logout(a.ctx)
}

// IsLoggedIn returns true when the session is active (token in memory).
func (a *App) IsLoggedIn() (bool, error) {
	return a.identity.IsLoggedIn(a.ctx)
}

// HasStoredCredential returns true when the credential store has a value (blob or legacy).
// Use this on startup to decide whether to attempt unlock before showing the login form.
func (a *App) HasStoredCredential() (bool, error) {
	return a.identity.HasStoredCredential(a.ctx)
}

// GetCredentialBlob returns the raw credential value from the store.
// The frontend inspects the value to decide whether to decrypt (wrapped blob, starts with
// "VRCTWKV1:") or forward as-is (legacy plaintext) before calling UnlockVRChatSession.
func (a *App) GetCredentialBlob() (string, error) {
	return a.identity.GetCredentialBlob(a.ctx)
}

// UnlockVRChatSession sets the decrypted auth token from the frontend and loads the user profile.
// Must be called after the frontend successfully decrypts the credential blob.
func (a *App) UnlockVRChatSession(token string) error {
	if err := a.identity.UnlockSession(a.ctx, token); err != nil {
		return err
	}
	a.startVRChatPipeline()
	return nil
}

// PersistWrappedCredential saves a Web-Crypto wrapped blob to the credential store.
// The blob must start with "VRCTWKV1:" (the WrappedBlobMagic prefix).
// Call this after every successful login or key migration.
func (a *App) PersistWrappedCredential(blob string) error {
	return a.identity.PersistWrappedCredential(a.ctx, blob)
}

// ClearStoredCredential removes the credential blob from the store.
// Call when IDB key loss makes the blob unrecoverable, or for a full logout.
func (a *App) ClearStoredCredential() error {
	return a.identity.ClearStoredCredential(a.ctx)
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

// ReconcileVRChatSocialCache refreshes self and friends from the VRChat REST API (e.g. after sleep resume).
func (a *App) ReconcileVRChatSocialCache() error {
	if err := a.identity.ReconcileSocialCacheFromAPIHandled(a.ctx); err != nil {
		return err
	}
	runtime.EventsEmit(a.ctx, friendsChangedEvent, struct{}{})
	return nil
}

// Friends returns cached friends.
func (a *App) Friends() ([]UserCacheDTO, error) {
	list, err := a.identity.ListFriends(a.ctx)
	if err != nil {
		return nil, err
	}
	return toUserCacheDTOs(list), nil
}

// ResolveUserProfileForNavigation refreshes users_cache when logged in (GET /users/{id}) and returns routing hints.
func (a *App) ResolveUserProfileNavigation(vrcUserID string) (UserProfileNavigationDTO, error) {
	u, openFriends, openSelf, err := a.identity.ResolveUserProfileForNavigation(a.ctx, vrcUserID)
	if err != nil {
		return UserProfileNavigationDTO{}, err
	}
	return UserProfileNavigationDTO{
		User:              toUserCacheDTO(u),
		OpenInFriendsView: openFriends,
		OpenInSelfProfile: openSelf,
	}, nil
}

// GetSelfProfile returns the logged-in user's cached profile row (users_cache user_kind=self).
func (a *App) GetSelfProfile(forceRefresh bool) (UserCacheDTO, error) {
	u, err := a.identity.GetSelfProfile(a.ctx, forceRefresh)
	if err != nil {
		return UserCacheDTO{}, err
	}
	return toUserCacheDTO(u), nil
}

// SetFavorite updates a friend's favorite flag.
func (a *App) SetFavorite(vrcUserID string, favorite bool) error {
	return a.identity.SetFavorite(a.ctx, vrcUserID, favorite)
}

// SetStatus changes the user's VRChat status.
func (a *App) SetStatus(status string) error {
	return a.identity.SetStatus(a.ctx, status)
}

// SetStatusDescription updates the current user's VRChat status description text.
func (a *App) SetStatusDescription(description string) error {
	return a.identity.SetStatusDescription(a.ctx, description)
}

// SetStatusAndDescription updates VRChat status and description in one request.
func (a *App) SetStatusAndDescription(status, description string) error {
	return a.identity.SetStatusAndDescription(a.ctx, status, description)
}

// --- Automation bindings ---

// ListAutomationRules returns all automation rules.
func (a *App) ListAutomationRules() ([]*automation.AutomationRule, error) {
	return a.automation.ListRules(a.ctx)
}

// SaveAutomationRule persists an automation rule.
func (a *App) SaveAutomationRule(rule automation.AutomationRule) error {
	return a.automation.SaveRule(a.ctx, &rule)
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
	return a.vrchatConfigRepo.Exists()
}

// GetVRChatConfig reads the current config.json.
func (a *App) GetVRChatConfig() (VRChatConfigDTO, error) {
	cfg, err := a.vrchatConfigRepo.Read()
	if err != nil {
		return VRChatConfigDTO{}, err
	}
	return toVRChatConfigDTO(cfg), nil
}

// SaveVRChatConfig writes config.json.
func (a *App) SaveVRChatConfig(dto VRChatConfigDTO) error {
	return a.vrchatConfigRepo.Write(fromVRChatConfigDTO(dto))
}

// DeleteVRChatConfig removes config.json.
func (a *App) DeleteVRChatConfig() error {
	return a.vrchatConfigRepo.Delete()
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

type sleepSuppressSettingGetter struct {
	a *App
}

func (g sleepSuppressSettingGetter) SuppressSleepWhileVRChat(ctx context.Context) (bool, error) {
	if g.a == nil || g.a.settings == nil {
		return false, nil
	}
	return g.a.settings.GetSuppressSleepWhileVRChat(ctx)
}

func (a *App) startSleepSuppressLoop() {
	a.sleepSuppressMu.Lock()
	defer a.sleepSuppressMu.Unlock()
	if a.settings == nil {
		return
	}
	if a.sleepSuppressCancel != nil {
		a.sleepSuppressCancel()
		a.sleepSuppressWG.Wait()
		a.sleepSuppressCancel = nil
	}
	runCtx, cancel := context.WithCancel(context.Background())
	a.sleepSuppressCancel = cancel
	a.sleepSuppressWG.Add(1)
	go func() {
		defer a.sleepSuppressWG.Done()
		checker := sleepsuppress.NewVRChatProcessChecker()
		exec := sleepsuppress.NewExecutionState()
		sg := sleepSuppressSettingGetter{a: a}
		_ = sleepsuppress.Run(runCtx, 8*time.Second, sg, checker, exec)
	}()
}

func (a *App) stopSleepSuppressLoop() {
	var cancel context.CancelFunc
	a.sleepSuppressMu.Lock()
	cancel = a.sleepSuppressCancel
	a.sleepSuppressCancel = nil
	a.sleepSuppressMu.Unlock()
	if cancel != nil {
		cancel()
		a.sleepSuppressWG.Wait()
	}
}

type ytdlpMaintainSettingGetter struct {
	a *App
}

func (g ytdlpMaintainSettingGetter) YTDLPToolsReplaceMaintain(ctx context.Context) (bool, error) {
	return g.a.settings.GetYTDLPToolsReplaceMaintain(ctx)
}

type ytdlpMaintainReapplier struct {
	a *App
}

func (r ytdlpMaintainReapplier) ReapplyIfNeeded(ctx context.Context) error {
	return r.a.ytdlp.ReapplyIfNeeded(ctx)
}

type ytdlpToolsDirProvider struct {
	a *App
}

func (p ytdlpToolsDirProvider) ToolsDir() (string, error) {
	if p.a.ytdlp != nil {
		return p.a.ytdlp.ToolsDir()
	}
	tools, err := usecase.VRChatYTDLPToolsPath()
	if err != nil {
		return "", err
	}
	return filepath.Dir(tools), nil
}

func (a *App) startYTDLPMaintainLoop() {
	a.ytdlpMaintainMu.Lock()
	if a.ytdlp == nil || a.settings == nil {
		a.ytdlpMaintainMu.Unlock()
		return
	}
	if a.ytdlpMaintainCancel != nil {
		cancel := a.ytdlpMaintainCancel
		a.ytdlpMaintainCancel = nil
		a.ytdlpMaintainMu.Unlock()
		cancel()
		a.ytdlpMaintainWG.Wait()
		a.ytdlpMaintainMu.Lock()
	}
	runCtx, cancel := context.WithCancel(context.Background())
	a.ytdlpMaintainCancel = cancel
	a.ytdlpMaintainWG.Add(1)
	a.ytdlpMaintainMu.Unlock()

	go func() {
		defer a.ytdlpMaintainWG.Done()
		checker := sleepsuppress.NewVRChatProcessChecker()
		if err := ytdlpmaintain.Run(
			runCtx,
			2*time.Second,
			ytdlpMaintainSettingGetter{a: a},
			checker,
			ytdlpMaintainReapplier{a: a},
			ytdlpToolsDirProvider{a: a},
		); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("ytdlp maintain loop: %v", err)
		}
	}()
}

func (a *App) stopYTDLPMaintainLoop() {
	var cancel context.CancelFunc
	a.ytdlpMaintainMu.Lock()
	cancel = a.ytdlpMaintainCancel
	a.ytdlpMaintainCancel = nil
	a.ytdlpMaintainMu.Unlock()
	if cancel != nil {
		cancel()
		a.ytdlpMaintainWG.Wait()
	}
}

// RuntimeIsWindows reports whether the backend process is running on Windows.
func (a *App) RuntimeIsWindows() bool {
	return goruntime.GOOS == "windows"
}

// GetYTDLPMaintainStatus returns desired/effective Tools replace maintain state (no GitHub call).
func (a *App) GetYTDLPMaintainStatus() (usecase.YTDLPMaintainStatus, error) {
	if a.ytdlp == nil {
		return usecase.YTDLPMaintainStatus{
			Supported:         false,
			UnsupportedReason: "notInitialized",
		}, nil
	}
	return a.ytdlp.GetStatus(a.ctx)
}

// AcknowledgeYTDLPToolsReplaceRisk records first-enable risk acknowledgment.
func (a *App) AcknowledgeYTDLPToolsReplaceRisk() error {
	if a.ytdlp == nil {
		return errors.New("notInitialized")
	}
	return a.ytdlp.AcknowledgeRisk(a.ctx)
}

// SetYTDLPToolsReplaceMaintain enables or disables maintain (enable requires risk ack).
func (a *App) SetYTDLPToolsReplaceMaintain(on bool) error {
	if a.ytdlp == nil {
		return errors.New("notInitialized")
	}
	return usecase.WrapMaintainAPIError(a.ytdlp.SetMaintainDesired(a.ctx, on))
}

// CheckYTDLPLatestRelease queries GitHub for the latest official yt-dlp.exe.
func (a *App) CheckYTDLPLatestRelease() (usecase.YTDLPMaintainStatus, error) {
	if a.ytdlp == nil {
		return usecase.YTDLPMaintainStatus{
			Supported:         false,
			UnsupportedReason: "notInitialized",
		}, nil
	}
	return a.ytdlp.CheckLatest(a.ctx)
}

// UpdateOfficialYTDLPCache downloads latest (or given URL) into Official cache; re-links when maintain is on.
func (a *App) UpdateOfficialYTDLPCache(downloadURL, latestTag string) (usecase.YTDLPMaintainStatus, error) {
	if a.ytdlp == nil {
		return usecase.YTDLPMaintainStatus{
			Supported:         false,
			UnsupportedReason: "notInitialized",
		}, nil
	}
	return a.ytdlp.UpdateOfficialCache(a.ctx, downloadURL, latestTag)
}

// OpenYTDLPCacheFolder opens the Official yt-dlp cache directory in the file manager.
func (a *App) OpenYTDLPCacheFolder() error {
	return openYTDLPFolderInFileManager(usecase.OfficialYTDLPCachePath)
}

// OpenYTDLPToolsFolder opens VRChat's Tools directory (parent of yt-dlp.exe) in the file manager.
func (a *App) OpenYTDLPToolsFolder() error {
	return openYTDLPFolderInFileManager(usecase.VRChatYTDLPToolsPath)
}

func openYTDLPFolderInFileManager(pathResolver func() (string, error)) error {
	path, err := pathResolver()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
		return mkErr
	}
	return desktop.OpenFolderInFileManager(dir)
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
