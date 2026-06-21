package discordrpc

// Snapshot is the Discord Voice RPC connection state.
type Snapshot struct {
	Available   bool
	Connected   bool
	Authorized  bool
	MuteKnown   bool
	Muted       bool
	Error       string
	ClientIDSet bool
}

// VoiceClient reads and writes local Discord mic mute state.
type VoiceClient interface {
	Connect() error
	Close()
	Snapshot() Snapshot
	SetMute(muted bool) error
	SetOnMuteChange(fn func(muted bool))
}

// NewVoiceClient returns a platform VoiceClient implementation.
func NewVoiceClient(tokenStore TokenStore) VoiceClient {
	return newPlatformVoiceClient(tokenStore)
}

// TokenStore persists Discord RPC OAuth tokens.
type TokenStore interface {
	GetAccessToken() string
	SetAccessToken(token string) error
}
