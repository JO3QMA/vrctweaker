package launcher

import (
	"strconv"
	"strings"
)

// LaunchArgsParsed holds GUI-friendly parsed launch arguments.
type LaunchArgsParsed struct {
	// 一般設定
	NoVR       bool // -no-vr or --no-vr
	ClearCache bool // --clear-cache (app-specific, not passed to VRChat)
	Fullscreen bool // -screen-fullscreen 1
	// 詳細設定
	VR              bool   // -vr (強制VRモード)
	FPFC            bool   // -fpfc (First Person Flying Camera)
	Windowed        bool   // -windowed
	ScreenWidth     int    // -screen-width N, 0=omit
	ScreenHeight    int    // -screen-height N, 0=omit
	FPS             int    // --fps=N, 0=omit
	Safe            bool   // -safe
	NoSplash        bool   // -nosplash
	NoAudio         bool   // -noaudio
	SkipRegistry    bool   // --skip-registry-install
	ForceD3D11      bool   // -force-d3d11
	ForceVulkan     bool   // -force-vulkan
	Log             bool   // -log
	ProcessPriority int    // --process-priority=N, 0=omit
	Custom          string // remaining args as-is
}

var (
	noVrShort             = "-no-vr"
	noVrLong              = "--no-vr"
	clearCache            = "--clear-cache"
	screenFull            = "-screen-fullscreen"
	fullscreen1           = "-screen-fullscreen 1"
	vr                    = "-vr"
	fpfc                  = "-fpfc"
	windowed              = "-windowed"
	screenWidthArg        = "-screen-width"
	screenHeightArg       = "-screen-height"
	fpsPrefix             = "--fps="
	safe                  = "-safe"
	nosplash              = "-nosplash"
	noaudio               = "-noaudio"
	skipRegistry          = "--skip-registry-install"
	forceD3D11            = "-force-d3d11"
	forceVulkan           = "-force-vulkan"
	logArg                = "-log"
	processPriorityPrefix = "--process-priority="
)

// ParseLaunchArgsForGUI parses a launch arguments string into GUI fields.
// Detects known args; everything else goes to Custom. Order of custom args is preserved.
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
		case tok == vr:
			p.VR = true
		case tok == fpfc:
			p.FPFC = true
		case tok == windowed:
			p.Windowed = true
		case tok == screenWidthArg:
			if i+1 < len(tokens) {
				if n, err := strconv.Atoi(tokens[i+1]); err == nil && n > 0 {
					p.ScreenWidth = n
				}
				i++
			}
		case tok == screenHeightArg:
			if i+1 < len(tokens) {
				if n, err := strconv.Atoi(tokens[i+1]); err == nil && n > 0 {
					p.ScreenHeight = n
				}
				i++
			}
		case strings.HasPrefix(tok, fpsPrefix):
			if v := strings.TrimPrefix(tok, fpsPrefix); v != "" {
				if n, err := strconv.Atoi(v); err == nil && n > 0 {
					p.FPS = n
				}
			}
		case tok == safe:
			p.Safe = true
		case tok == nosplash:
			p.NoSplash = true
		case tok == noaudio:
			p.NoAudio = true
		case tok == skipRegistry:
			p.SkipRegistry = true
		case tok == forceD3D11:
			p.ForceD3D11 = true
		case tok == forceVulkan:
			p.ForceVulkan = true
		case tok == logArg:
			p.Log = true
		case strings.HasPrefix(tok, processPriorityPrefix):
			if v := strings.TrimPrefix(tok, processPriorityPrefix); v != "" {
				if n, err := strconv.Atoi(v); err == nil && n >= 0 {
					p.ProcessPriority = n
				}
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
func MergeLaunchArgsForGUI(p *LaunchArgsParsed) string {
	if p == nil {
		return ""
	}
	var parts []string
	if p.NoVR {
		parts = append(parts, noVrShort)
	}
	if p.VR {
		parts = append(parts, vr)
	}
	if p.ClearCache {
		parts = append(parts, clearCache)
	}
	if p.Fullscreen {
		parts = append(parts, fullscreen1)
	}
	if p.FPFC {
		parts = append(parts, fpfc)
	}
	if p.Windowed {
		parts = append(parts, windowed)
	}
	if p.ScreenWidth > 0 {
		parts = append(parts, screenWidthArg, strconv.Itoa(p.ScreenWidth))
	}
	if p.ScreenHeight > 0 {
		parts = append(parts, screenHeightArg, strconv.Itoa(p.ScreenHeight))
	}
	if p.FPS > 0 {
		parts = append(parts, fpsPrefix+strconv.Itoa(p.FPS))
	}
	if p.Safe {
		parts = append(parts, safe)
	}
	if p.NoSplash {
		parts = append(parts, nosplash)
	}
	if p.NoAudio {
		parts = append(parts, noaudio)
	}
	if p.SkipRegistry {
		parts = append(parts, skipRegistry)
	}
	if p.ForceD3D11 {
		parts = append(parts, forceD3D11)
	}
	if p.ForceVulkan {
		parts = append(parts, forceVulkan)
	}
	if p.Log {
		parts = append(parts, logArg)
	}
	if p.ProcessPriority > 0 {
		parts = append(parts, processPriorityPrefix+strconv.Itoa(p.ProcessPriority))
	}
	if p.Custom != "" {
		parts = append(parts, strings.TrimSpace(p.Custom))
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}
