package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/automation"
	"vrchat-tweaker/internal/domain/event"
	"vrchat-tweaker/internal/domain/media"
	"vrchat-tweaker/internal/infrastructure/desktop"
	"vrchat-tweaker/internal/infrastructure/logwatcher"
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
	identityRepo := sqlite.NewFriendCacheRepository(db)
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
	a.media = usecase.NewMediaUseCase(mediaRepo, extractor)
	a.activity = usecase.NewActivityUseCase(playRepo, encounterRepo, settingsRepo)
	a.identity = usecase.NewIdentityUseCaseWithNotifier(identityRepo, apiClient, credStore, notifier)
	actionRunner := usecase.NewDefaultActionRunner(a.identity)
	a.automation = usecase.NewAutomationUseCase(automationRepo, eventBus, actionRunner)
	a.settings = usecase.NewSettingsUseCase(settingsRepo)
	a.dbMaintenance = usecase.NewDBMaintenanceUseCase(encounterRepo, mediaRepo, identityRepo, maintenanceRepo)

	a.subscribeAutomationEvents(ctx, eventBus)

	// Start output_log watcher if path is configured
	a.startOutputLogWatcher(ctx, eventBus)
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

func (a *App) startOutputLogWatcher(ctx context.Context, eventBus event.EventBus) {
	path, err := a.settings.GetOutputLogPath(ctx)
	if err != nil || path == "" {
		return
	}
	info, err := os.Stat(path)
	if err != nil || info == nil || !info.Mode().IsRegular() {
		runtime.LogWarning(ctx, "output_log_path not set or file not accessible, skipping log watcher")
		return
	}
	parser := activity.NewLogParser()
	logger := &logLogger{}
	activityHandler := logwatcher.NewActivityEventHandler(a.activity, ctx, logger)
	publishHandler := logwatcher.NewEventPublishingHandler(eventBus, ctx, logger)
	handler := logwatcher.NewMultiHandler(activityHandler, publishHandler)
	watcher := logwatcher.NewOutputLogWatcher(path, parser, handler, logger)
	if startErr := watcher.Start(ctx); startErr != nil {
		runtime.LogError(ctx, "failed to start output_log watcher: "+startErr.Error())
		return
	}
	runtime.LogInfo(ctx, "output_log watcher started for "+path)
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

// LaunchVRChat starts VRChat with the given profile ID.
// Uses path settings (VRChat/Steam paths) from app_settings when configured.
func (a *App) LaunchVRChat(profileID string) error {
	ps, err := a.settings.GetPathSettings(a.ctx)
	if err != nil {
		return err
	}
	vrchatPath := ""
	steamPath := ""
	if ps != nil {
		vrchatPath = ps.VRChatPathWindows
		steamPath = ps.SteamPathLinux
	}
	return a.launcher.LaunchVRChat(a.ctx, profileID, vrchatPath, steamPath)
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
	if ps != nil {
		vrchatPath = ps.VRChatPathWindows
		steamPath = ps.SteamPathLinux
	}
	return a.launcher.LaunchToWorld(a.ctx, "", worldID, vrchatPath, steamPath)
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

// ScanScreenshotDir scans a directory for screenshots.
func (a *App) ScanScreenshotDir(path string) (int, error) {
	return a.media.ScanDirectory(a.ctx, path)
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
	list, err := a.activity.ListEncounters(a.ctx, nil)
	if err != nil {
		return nil, err
	}
	return toEncounterDTOs(list), nil
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

// RefreshFriends fetches friends from API and updates cache.
func (a *App) RefreshFriends() error {
	return a.identity.RefreshFriends(a.ctx)
}

// Friends returns cached friends.
func (a *App) Friends() ([]FriendCacheDTO, error) {
	list, err := a.identity.ListFriends(a.ctx)
	if err != nil {
		return nil, err
	}
	return toFriendCacheDTOs(list), nil
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
	return a.dbMaintenance.ClearScreenshots(a.ctx)
}

// ClearFriendsCache deletes all cached friends. Returns affected row count.
func (a *App) ClearFriendsCache() (int64, error) {
	return a.dbMaintenance.ClearFriendsCache(a.ctx)
}
