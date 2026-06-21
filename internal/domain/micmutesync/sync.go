package micmutesync

import "time"

const (
	SourceVRChat  = "vrchat"
	SourceDiscord = "discord"
)

// EchoGuard suppresses feedback from programmatic mute changes.
type EchoGuard struct {
	suppressUntil time.Time
	source        string
}

// Suppress ignores changes from source until d elapses.
func (g *EchoGuard) Suppress(source string, d time.Duration) {
	if g == nil {
		return
	}
	g.source = source
	g.suppressUntil = time.Now().Add(d)
}

// ShouldIgnore reports whether an incoming change from source should be ignored.
func (g *EchoGuard) ShouldIgnore(source string) bool {
	if g == nil {
		return false
	}
	if time.Now().After(g.suppressUntil) {
		return false
	}
	return g.source != source
}

// NeedsMuteToggle reports whether a toggle is required to reach desired state.
func NeedsMuteToggle(currentKnown, current, desired bool) bool {
	if !currentKnown {
		return true
	}
	return current != desired
}
