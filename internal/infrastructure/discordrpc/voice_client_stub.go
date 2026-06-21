//go:build !windows

package discordrpc

import "fmt"

type platformVoiceClient struct{}

func newPlatformVoiceClient(_ TokenStore) VoiceClient {
	return &platformVoiceClient{}
}

func (c *platformVoiceClient) Connect() error {
	return fmt.Errorf("discord_unavailable")
}

func (c *platformVoiceClient) Close() {}

func (c *platformVoiceClient) Snapshot() Snapshot {
	return Snapshot{Available: false, Error: "platform_unavailable"}
}

func (c *platformVoiceClient) SetMute(_ bool) error {
	return fmt.Errorf("discord_unavailable")
}

func (c *platformVoiceClient) SetOnMuteChange(_ func(bool)) {}
