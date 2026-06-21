package micmutesync

import (
	"strings"

	"vrchat-tweaker/internal/domain/launcher"
)

// EnsureOSCInLaunchArgs adds --osc= when missing and endpoint is non-empty.
func EnsureOSCInLaunchArgs(argsStr, endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = DefaultOSCEndpoint
	}
	parsed := launcher.ParseLaunchArgsForGUI(argsStr)
	if strings.TrimSpace(parsed.OSC) != "" {
		return argsStr
	}
	parsed.OSC = endpoint
	return launcher.MergeLaunchArgsForGUI(parsed)
}
