package discordrpc

import "os"

// ClientID is the bundled Discord Application client ID.
// Release builds set this via -ldflags; VRCTWEAKER_DISCORD_CLIENT_ID overrides at runtime.
var ClientID = "000000000000000000"

// clientSecret may be set at link time for release builds; VRCTWEAKER_DISCORD_CLIENT_SECRET overrides at runtime.
var clientSecret = ""

func init() {
	loadDiscordEnvFromFiles()
	if v := os.Getenv(envDiscordClientID); v != "" {
		ClientID = v
	}
	if v := os.Getenv(envDiscordClientSecret); v != "" {
		clientSecret = v
	}
}

// ClientSecret returns the OAuth client secret for Discord token exchange.
func ClientSecret() string {
	if clientSecret != "" {
		return clientSecret
	}
	return os.Getenv(envDiscordClientSecret)
}
