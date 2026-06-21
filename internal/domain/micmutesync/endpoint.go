package micmutesync

import (
	"fmt"
	"strconv"
	"strings"
)

const DefaultOSCEndpoint = "9000:127.0.0.1:9001"

// Endpoint is VRChat's --osc=inPort:outIP:outPort form.
type Endpoint struct {
	InPort  int
	OutHost string
	OutPort int
}

// ListenAddr returns the UDP address Tweaker binds to receive OSC from VRChat.
func (e Endpoint) ListenAddr() string {
	host := e.OutHost
	if host == "" {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("%s:%d", host, e.OutPort)
}

// ParseEndpoint parses inPort:outIP:outPort. Empty input returns the default endpoint.
func ParseEndpoint(raw string) (Endpoint, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ParseEndpoint(DefaultOSCEndpoint)
	}
	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		return Endpoint{}, fmt.Errorf("invalid OSC endpoint %q: want inPort:outIP:outPort", raw)
	}
	inPort, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || inPort <= 0 || inPort > 65535 {
		return Endpoint{}, fmt.Errorf("invalid OSC inPort in %q", raw)
	}
	outHost := strings.TrimSpace(parts[1])
	if outHost == "" {
		return Endpoint{}, fmt.Errorf("invalid OSC outIP in %q", raw)
	}
	outPort, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil || outPort <= 0 || outPort > 65535 {
		return Endpoint{}, fmt.Errorf("invalid OSC outPort in %q", raw)
	}
	return Endpoint{InPort: inPort, OutHost: outHost, OutPort: outPort}, nil
}

// PlatformAvailable reports whether Mic Mute Sync is offered on this OS build.
func PlatformAvailable(goos string) bool {
	return goos == "windows"
}
