package usecase

import (
	"context"
	"runtime"
	"strconv"
	"sync"
	"time"

	"vrchat-tweaker/internal/domain/micmutesync"
	"vrchat-tweaker/internal/domain/settings"
	"vrchat-tweaker/internal/infrastructure/discordrpc"
	"vrchat-tweaker/internal/infrastructure/vrchatosc"
)

const (
	keyMicMuteSyncEnabled      = "mic_mute_sync_enabled"
	keyMicMuteSyncOSCEndpoint  = "mic_mute_sync_osc_endpoint"
	keyMicMuteSyncDiscordToken = "mic_mute_sync_discord_access_token"
	echoSuppressionDuration    = 1500 * time.Millisecond
	discordReconnectInterval   = 5 * time.Second
)

// MicMuteSyncSettings is persisted Mic Mute Sync configuration.
type MicMuteSyncSettings struct {
	Enabled     bool   `json:"enabled"`
	OSCEndpoint string `json:"oscEndpoint"`
}

// MicMuteSyncStatus is the Sync Status checklist for Settings UI.
type MicMuteSyncStatus struct {
	Available           bool   `json:"available"`
	Enabled             bool   `json:"enabled"`
	OSCEndpoint         string `json:"oscEndpoint"`
	VRChatOSCListening  bool   `json:"vrchatOscListening"`
	VRChatOSCConnected  bool   `json:"vrchatOscConnected"`
	VRChatMuteKnown     bool   `json:"vrchatMuteKnown"`
	VRChatMuted         bool   `json:"vrchatMuted"`
	VRChatOSCError      string `json:"vrchatOscError,omitempty"`
	SyncEngineState     string `json:"syncEngineState"`
	SyncPauseReason     string `json:"syncPauseReason,omitempty"`
	DiscordRPCConnected bool   `json:"discordRpcConnected"`
	DiscordMuteKnown    bool   `json:"discordMuteKnown"`
	DiscordMuted        bool   `json:"discordMuted"`
	DiscordRPCError     string `json:"discordRpcError,omitempty"`
	ToggleVoiceKnown    bool   `json:"toggleVoiceKnown"`
	ToggleVoiceOK       bool   `json:"toggleVoiceOk"`
}

type settingsTokenStore struct {
	uc *MicMuteSyncUseCase
}

func (s settingsTokenStore) GetAccessToken() string {
	s.uc.mu.Lock()
	ctx := s.uc.runningCtx
	s.uc.mu.Unlock()
	if ctx == nil {
		return ""
	}
	v, _ := s.uc.repo.Get(ctx, keyMicMuteSyncDiscordToken)
	return v
}

func (s settingsTokenStore) SetAccessToken(token string) error {
	s.uc.mu.Lock()
	ctx := s.uc.runningCtx
	s.uc.mu.Unlock()
	if ctx == nil {
		return nil
	}
	return s.uc.repo.Set(ctx, keyMicMuteSyncDiscordToken, token)
}

// MicMuteSyncUseCase manages Mic Mute Sync settings and synchronization.
type MicMuteSyncUseCase struct {
	repo     settings.AppSettingsRepository
	listener *vrchatosc.Listener
	sender   *vrchatosc.Sender
	discord  discordrpc.VoiceClient
	goos     string

	mu             sync.Mutex
	runningCtx     context.Context
	echo           micmutesync.EchoGuard
	sessionReady   bool
	lastDiscordTry time.Time
}

// NewMicMuteSyncUseCase creates a MicMuteSyncUseCase.
func NewMicMuteSyncUseCase(repo settings.AppSettingsRepository, listener *vrchatosc.Listener, sender *vrchatosc.Sender) *MicMuteSyncUseCase {
	uc := &MicMuteSyncUseCase{
		repo:     repo,
		listener: listener,
		sender:   sender,
		goos:     runtime.GOOS,
	}
	uc.discord = discordrpc.NewVoiceClient(settingsTokenStore{uc: uc})
	return uc
}

// GetSettings returns saved Mic Mute Sync settings.
func (uc *MicMuteSyncUseCase) GetSettings(ctx context.Context) (MicMuteSyncSettings, error) {
	enabled, err := uc.getBool(ctx, keyMicMuteSyncEnabled, false)
	if err != nil {
		return MicMuteSyncSettings{}, err
	}
	endpoint, err := uc.repo.Get(ctx, keyMicMuteSyncOSCEndpoint)
	if err != nil {
		return MicMuteSyncSettings{}, err
	}
	return MicMuteSyncSettings{Enabled: enabled, OSCEndpoint: endpoint}, nil
}

// SaveSettings persists Mic Mute Sync settings and restarts monitoring.
func (uc *MicMuteSyncUseCase) SaveSettings(ctx context.Context, s MicMuteSyncSettings) error {
	if err := uc.repo.Set(ctx, keyMicMuteSyncEnabled, strconv.FormatBool(s.Enabled)); err != nil {
		return err
	}
	if err := uc.repo.Set(ctx, keyMicMuteSyncOSCEndpoint, s.OSCEndpoint); err != nil {
		return err
	}
	uc.mu.Lock()
	uc.sessionReady = false
	uc.mu.Unlock()
	uc.restartListener(ctx)
	if s.Enabled {
		uc.ensureDiscord(ctx)
	} else {
		uc.discord.Close()
	}
	return nil
}

// ConnectDiscord triggers Discord RPC authorization.
func (uc *MicMuteSyncUseCase) ConnectDiscord(ctx context.Context) error {
	if !micmutesync.PlatformAvailable(uc.goos) {
		return nil
	}
	uc.discord.Close()
	uc.discord = discordrpc.NewVoiceClient(settingsTokenStore{uc: uc})
	uc.wireDiscordCallbacks(ctx)
	return uc.discord.Connect()
}

// Start begins monitoring and the sync loop.
func (uc *MicMuteSyncUseCase) Start(ctx context.Context) {
	uc.mu.Lock()
	uc.runningCtx = ctx
	uc.mu.Unlock()
	uc.listener.SetOnMuteChange(func(muted bool) {
		uc.onVRChatMuteChanged(ctx, muted)
	})
	uc.wireDiscordCallbacks(ctx)
	uc.restartListener(ctx)
	go uc.runLoop(ctx)
}

func (uc *MicMuteSyncUseCase) wireDiscordCallbacks(ctx context.Context) {
	uc.discord.SetOnMuteChange(func(muted bool) {
		uc.onDiscordMuteChanged(ctx, muted)
	})
}

// Stop stops monitoring.
func (uc *MicMuteSyncUseCase) Stop() {
	uc.listener.Stop()
	uc.discord.Close()
}

// GetStatus returns the current Sync Status snapshot.
func (uc *MicMuteSyncUseCase) GetStatus(ctx context.Context) (MicMuteSyncStatus, error) {
	cfg, err := uc.GetSettings(ctx)
	if err != nil {
		return MicMuteSyncStatus{}, err
	}
	if cfg.OSCEndpoint == "" {
		cfg.OSCEndpoint = micmutesync.DefaultOSCEndpoint
	}
	_, epErr := micmutesync.ParseEndpoint(cfg.OSCEndpoint)
	snap := uc.listener.Snapshot()
	discord := uc.discord.Snapshot()
	st := MicMuteSyncStatus{
		Available:           micmutesync.PlatformAvailable(uc.goos),
		Enabled:             cfg.Enabled,
		OSCEndpoint:         cfg.OSCEndpoint,
		VRChatOSCListening:  snap.Listening,
		VRChatOSCConnected:  snap.Connected,
		VRChatMuteKnown:     snap.MuteKnown,
		VRChatMuted:         snap.Muted,
		VRChatOSCError:      snap.ListenError,
		DiscordRPCConnected: discord.Connected && discord.Authorized,
		DiscordMuteKnown:    discord.MuteKnown,
		DiscordMuted:        discord.Muted,
		DiscordRPCError:     discord.Error,
		ToggleVoiceKnown:    false,
		ToggleVoiceOK:       false,
	}
	if epErr != nil {
		st.SyncEngineState = "paused"
		st.SyncPauseReason = epErr.Error()
		return st, nil
	}
	if !st.Available {
		st.SyncEngineState = "unavailable"
		st.SyncPauseReason = "platform_unavailable"
		return st, nil
	}
	if !cfg.Enabled {
		st.SyncEngineState = "off"
		return st, nil
	}
	if snap.ListenError != "" {
		st.SyncEngineState = "paused"
		st.SyncPauseReason = snap.ListenError
		return st, nil
	}
	if !snap.Connected {
		st.SyncEngineState = "paused"
		st.SyncPauseReason = "vrchat_osc_waiting"
		return st, nil
	}
	if !discord.ClientIDSet {
		st.SyncEngineState = "paused"
		st.SyncPauseReason = "discord_client_id_missing"
		return st, nil
	}
	if !discord.Authorized {
		st.SyncEngineState = "paused"
		if discord.Error != "" {
			st.SyncPauseReason = discord.Error
		} else {
			st.SyncPauseReason = "discord_not_authorized"
		}
		return st, nil
	}
	uc.mu.Lock()
	ready := uc.sessionReady
	uc.mu.Unlock()
	if ready {
		st.SyncEngineState = "syncing"
	} else {
		st.SyncEngineState = "monitoring"
	}
	return st, nil
}

// EnsureOSCLaunchArgs injects --osc= when Mic Mute Sync is enabled.
func (uc *MicMuteSyncUseCase) EnsureOSCLaunchArgs(ctx context.Context, argsStr string) (string, error) {
	cfg, err := uc.GetSettings(ctx)
	if err != nil {
		return argsStr, err
	}
	if !cfg.Enabled || !micmutesync.PlatformAvailable(uc.goos) {
		return argsStr, nil
	}
	return micmutesync.EnsureOSCInLaunchArgs(argsStr, cfg.OSCEndpoint), nil
}

func (uc *MicMuteSyncUseCase) runLoop(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			uc.tick(ctx)
		}
	}
}

func (uc *MicMuteSyncUseCase) tick(ctx context.Context) {
	cfg, err := uc.GetSettings(ctx)
	if err != nil || !cfg.Enabled || !micmutesync.PlatformAvailable(uc.goos) {
		return
	}
	uc.ensureDiscord(ctx)
	uc.trySessionBaseline(ctx, cfg.OSCEndpoint)
}

func (uc *MicMuteSyncUseCase) ensureDiscord(ctx context.Context) {
	discord := uc.discord.Snapshot()
	if discord.Authorized {
		return
	}
	uc.mu.Lock()
	if time.Since(uc.lastDiscordTry) < discordReconnectInterval {
		uc.mu.Unlock()
		return
	}
	uc.lastDiscordTry = time.Now()
	uc.mu.Unlock()
	_ = uc.ConnectDiscord(ctx)
}

func (uc *MicMuteSyncUseCase) trySessionBaseline(ctx context.Context, endpoint string) {
	vr := uc.listener.Snapshot()
	dc := uc.discord.Snapshot()
	ready := vr.Connected && vr.MuteKnown && dc.Authorized && dc.MuteKnown
	uc.mu.Lock()
	wasReady := uc.sessionReady
	if !ready {
		uc.sessionReady = false
		uc.mu.Unlock()
		return
	}
	if wasReady {
		uc.mu.Unlock()
		return
	}
	uc.sessionReady = true
	uc.mu.Unlock()
	if !micmutesync.NeedsMuteToggle(dc.MuteKnown, dc.Muted, vr.Muted) {
		return
	}
	uc.echo.Suppress(micmutesync.SourceVRChat, echoSuppressionDuration)
	if err := uc.discord.SetMute(vr.Muted); err != nil {
		uc.mu.Lock()
		uc.sessionReady = false
		uc.mu.Unlock()
	}
}

func (uc *MicMuteSyncUseCase) onVRChatMuteChanged(ctx context.Context, muted bool) {
	cfg, err := uc.GetSettings(ctx)
	if err != nil || !cfg.Enabled {
		return
	}
	if uc.echo.ShouldIgnore(micmutesync.SourceVRChat) {
		return
	}
	dc := uc.discord.Snapshot()
	if !dc.Authorized {
		return
	}
	if !micmutesync.NeedsMuteToggle(dc.MuteKnown, dc.Muted, muted) {
		return
	}
	uc.echo.Suppress(micmutesync.SourceVRChat, echoSuppressionDuration)
	_ = uc.discord.SetMute(muted)
}

func (uc *MicMuteSyncUseCase) onDiscordMuteChanged(ctx context.Context, muted bool) {
	cfg, err := uc.GetSettings(ctx)
	if err != nil || !cfg.Enabled {
		return
	}
	if uc.echo.ShouldIgnore(micmutesync.SourceDiscord) {
		return
	}
	vr := uc.listener.Snapshot()
	if !vr.Connected {
		return
	}
	endpoint := cfg.OSCEndpoint
	if endpoint == "" {
		endpoint = micmutesync.DefaultOSCEndpoint
	}
	if !micmutesync.NeedsMuteToggle(vr.MuteKnown, vr.Muted, muted) {
		return
	}
	uc.echo.Suppress(micmutesync.SourceDiscord, echoSuppressionDuration)
	_ = uc.sender.SetMute(endpoint, vr.MuteKnown, vr.Muted, muted)
}

func (uc *MicMuteSyncUseCase) restartListener(ctx context.Context) {
	if !micmutesync.PlatformAvailable(uc.goos) {
		uc.listener.Stop()
		return
	}
	cfg, err := uc.GetSettings(ctx)
	if err != nil {
		uc.listener.Stop()
		return
	}
	ep, err := micmutesync.ParseEndpoint(cfg.OSCEndpoint)
	if err != nil {
		uc.listener.Stop()
		return
	}
	uc.mu.Lock()
	runCtx := uc.runningCtx
	uc.mu.Unlock()
	if runCtx == nil {
		runCtx = ctx
	}
	_ = uc.listener.Start(runCtx, ep.ListenAddr())
}

func (uc *MicMuteSyncUseCase) getBool(ctx context.Context, key string, def bool) (bool, error) {
	v, err := uc.repo.Get(ctx, key)
	if err != nil {
		return def, err
	}
	if v == "" {
		return def, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def, nil
	}
	return b, nil
}
