package launcher

import (
	"strings"
)

// LaunchArgsParsed holds GUI-friendly parsed launch arguments.
type LaunchArgsParsed struct {
	NoVR       bool   // -no-vr or --no-vr
	ClearCache bool   // --clear-cache (app-specific, not passed to VRChat)
	Fullscreen bool   // -screen-fullscreen 1
	Custom     string // remaining args as-is
}

var (
	noVrShort   = "-no-vr"
	noVrLong    = "--no-vr"
	clearCache  = "--clear-cache"
	screenFull  = "-screen-fullscreen"
	fullscreen1 = "-screen-fullscreen 1"
)

// ParseLaunchArgsForGUI parses a launch arguments string into GUI fields.
// Detects: -no-vr/--no-vr, --clear-cache, -screen-fullscreen 0/1.
// Everything else goes to Custom. Order of custom args is preserved.
func ParseLaunchArgsForGUI(args string) *LaunchArgsParsed {
	p := &LaunchArgsParsed{}
	if args == "" {
		return p
	}
	tokens := parseLaunchArgsTokens(args)
	var customParts []string
	i := 0
	for i < len(tokens) {
		tok := tokens[i]
		switch {
		case tok == noVrShort || tok == noVrLong:
			p.NoVR = true
		case tok == clearCache:
			p.ClearCache = true
		case tok == screenFull:
			if i+1 < len(tokens) {
				if tokens[i+1] == "1" {
					p.Fullscreen = true
				}
				i++
			}
		default:
			customParts = append(customParts, tok)
		}
		i++
	}
	p.Custom = strings.TrimSpace(strings.Join(customParts, " "))
	return p
}

// parseLaunchArgsTokens splits args string into tokens (supports quoted values).
func parseLaunchArgsTokens(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	var cur []rune
	inDouble := false
	inSingle := false
	for _, r := range s {
		switch {
		case inDouble:
			if r == '"' {
				inDouble = false
				out = append(out, string(cur))
				cur = nil
			} else {
				cur = append(cur, r)
			}
		case inSingle:
			if r == '\'' {
				inSingle = false
				out = append(out, string(cur))
				cur = nil
			} else {
				cur = append(cur, r)
			}
		case r == '"':
			inDouble = true
			cur = nil
		case r == '\'':
			inSingle = true
			cur = nil
		case r == ' ' || r == '\t':
			if len(cur) > 0 {
				out = append(out, string(cur))
				cur = nil
			}
		default:
			cur = append(cur, r)
		}
	}
	if len(cur) > 0 {
		out = append(out, string(cur))
	}
	return out
}

// MergeLaunchArgsForGUI builds a single arguments string from parsed GUI state.
// Order: -no-vr, --clear-cache, -screen-fullscreen 0/1, custom.
func MergeLaunchArgsForGUI(p *LaunchArgsParsed) string {
	if p == nil {
		return ""
	}
	var parts []string
	if p.NoVR {
		parts = append(parts, noVrShort)
	}
	if p.ClearCache {
		parts = append(parts, clearCache)
	}
	if p.Fullscreen {
		parts = append(parts, fullscreen1)
	} else {
		// Only emit 0 if Custom might have had other screen args; we omit for simplicity
		// per spec: omit = off
	}
	if p.Custom != "" {
		parts = append(parts, strings.TrimSpace(p.Custom))
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}
