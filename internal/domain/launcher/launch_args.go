package launcher

import (
	"strconv"
	"strings"
)

// ScreenMode is the display mode: fullscreen, windowed, or popupwindow (virtual fullscreen).
const (
	ScreenModeFullscreen  = "fullscreen"
	ScreenModeWindowed    = "windowed"
	ScreenModePopupWindow = "popupwindow"
)

// VrMode is the VR launch mode: desktop (-no-vr), none (無設定), or vr (-vr).
const (
	VrModeDesktop = "desktop"
	VrModeNone    = ""
	VrModeVR      = "vr"
)

// RenderBackend is the exclusive render/display mode: default, d3d11, vulkan, or nographics.
const (
	RenderBackendDefault    = ""
	RenderBackendD3D11      = "d3d11"
	RenderBackendVulkan     = "vulkan"
	RenderBackendNoGraphics = "nographics"
)

// LaunchArgsParsed holds GUI-friendly parsed launch arguments.
type LaunchArgsParsed struct {
	// 一般設定
	VrMode     string // desktop|""|vr (replaces NoVR+VR)
	ClearCache bool   // --clear-cache (app-specific, not passed to VRChat)
	ScreenMode string // fullscreen|windowed|popupwindow (replaces Fullscreen+Windowed)
	// 詳細設定
	FPFC            bool   // -fpfc (First Person Flying Camera)
	ScreenWidth     int    // -screen-width N, 0=omit
	ScreenHeight    int    // -screen-height N, 0=omit
	FPS             int    // --fps=N, 0=omit
	Safe            bool   // -safe
	NoSplash        bool   // -nosplash
	NoAudio         bool   // -noaudio
	SkipRegistry    bool   // --skip-registry-install
	RenderBackend   string // ""|d3d11|vulkan|nographics (exclusive: -force-d3d11, -force-vulkan, -nographics)
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
	popupwindow           = "-popupwindow"
	screenWidthArg        = "-screen-width"
	screenHeightArg       = "-screen-height"
	fpsPrefix             = "--fps="
	safe                  = "-safe"
	nosplash              = "-nosplash"
	noaudio               = "-noaudio"
	skipRegistry          = "--skip-registry-install"
	forceD3D11            = "-force-d3d11"
	forceVulkan           = "-force-vulkan"
	nographics            = "-nographics"
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
			p.VrMode = VrModeDesktop
		case tok == vr:
			p.VrMode = VrModeVR
		case tok == clearCache:
			p.ClearCache = true
		case tok == screenFull:
			if i+1 < len(tokens) {
				if tokens[i+1] == "1" {
					p.ScreenMode = ScreenModeFullscreen
				}
				i++
			}
		case tok == fpfc:
			p.FPFC = true
		case tok == windowed:
			p.ScreenMode = ScreenModeWindowed
		case tok == popupwindow:
			p.ScreenMode = ScreenModePopupWindow
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
		case tok == nographics:
			p.RenderBackend = RenderBackendNoGraphics
		case tok == forceD3D11:
			p.RenderBackend = RenderBackendD3D11
		case tok == forceVulkan:
			p.RenderBackend = RenderBackendVulkan
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
	switch p.VrMode {
	case VrModeDesktop:
		parts = append(parts, noVrShort)
	case VrModeVR:
		parts = append(parts, vr)
	}
	if p.ClearCache {
		parts = append(parts, clearCache)
	}
	if p.FPFC {
		parts = append(parts, fpfc)
	}
	switch p.ScreenMode {
	case ScreenModeFullscreen:
		parts = append(parts, fullscreen1)
	case ScreenModeWindowed:
		parts = append(parts, windowed)
	case ScreenModePopupWindow:
		parts = append(parts, popupwindow)
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
	switch p.RenderBackend {
	case RenderBackendD3D11:
		parts = append(parts, forceD3D11)
	case RenderBackendVulkan:
		parts = append(parts, forceVulkan)
	case RenderBackendNoGraphics:
		parts = append(parts, nographics)
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
