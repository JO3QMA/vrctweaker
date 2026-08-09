package launcher

import (
	"math"
	"strconv"
	"strings"
)

// ScreenMode is the display mode: fullscreen, windowed, or popupwindow (virtual fullscreen).
const (
	ScreenModeFullscreen  = "fullscreen"
	ScreenModeWindowed    = "windowed"
	ScreenModePopupWindow = "popupwindow"
)

// VideoDecoding is the video decode mode: default, software, or hardware.
const (
	VideoDecodingDefault  = ""
	VideoDecodingSoftware = "software"
	VideoDecodingHardware = "hardware"
)

// PriorityOmit is used for ProcessPriority and MainThreadPriority when not set.
const PriorityOmit = -999

// LaunchArgsParsed holds GUI-friendly parsed launch arguments.
// Per https://docs.vrchat.com/docs/launch-options
// JSON tags are for Wails IPC (camelCase); not used for CLI parsing.
type LaunchArgsParsed struct {
	// 一般設定
	NoVR       bool   `json:"noVr"`       // -no-vr or --no-vr (デスクトップモード)
	ScreenMode string `json:"screenMode"` // fullscreen|windowed|popupwindow (replaces Fullscreen+Windowed)
	// 詳細設定
	ScreenWidth                 int    `json:"screenWidth"`                 // -screen-width N, 0=omit
	ScreenHeight                int    `json:"screenHeight"`                // -screen-height N, 0=omit
	FPS                         int    `json:"fps"`                         // --fps=N, 0=omit
	SkipRegistry                bool   `json:"skipRegistry"`                // --skip-registry-install
	ProcessPriority             int    `json:"processPriority"`             // --process-priority=N, -2..2, PriorityOmit=omit
	MainThreadPriority          int    `json:"mainThreadPriority"`          // --main-thread-priority=N, -2..2, PriorityOmit=omit
	Monitor                     int    `json:"monitor"`                     // -monitor N (1-based), 0=omit
	Profile                     int    `json:"profile"`                     // --profile=X, -1=omit
	EnableDebugGui              bool   `json:"enableDebugGui"`              // --enable-debug-gui
	EnableSDKLogLevels          bool   `json:"enableSDKLogLevels"`          // --enable-sdk-log-levels
	EnableUdonDebugLogging      bool   `json:"enableUdonDebugLogging"`      // --enable-udon-debug-logging
	Midi                        string `json:"midi"`                        // --midi=deviceName, empty=omit
	WatchWorlds                 bool   `json:"watchWorlds"`                 // --watch-worlds
	WatchAvatars                bool   `json:"watchAvatars"`                // --watch-avatars
	IgnoreTrackers              string `json:"ignoreTrackers"`              // --ignore-trackers=serial1,serial2
	VideoDecoding               string `json:"videoDecoding"`               // ""|software|hardware
	DisableAMDStutterWorkaround bool   `json:"disableAMDStutterWorkaround"` // --disable-amd-stutter-workaround
	OSC                         string `json:"osc"`                         // --osc=inPort:outIP:outPort
	Affinity                    string `json:"affinity"`                    // --affinity=<hex>
	EnforceWorldServerChecks    bool   `json:"enforceWorldServerChecks"`    // --enforce-world-server-checks
	// IK 2.0 (https://docs.vrchat.com/docs/ik-20-features-and-options)
	CustomArmRatio             float64 `json:"customArmRatio"`             // --custom-arm-ratio=N, 0=omit
	DisableShoulderTracking    bool    `json:"disableShoulderTracking"`    // --disable-shoulder-tracking
	EnableIKDebugLogging       bool    `json:"enableIkDebugLogging"`       // --enable-ik-debug-logging
	CalibrationRange           float64 `json:"calibrationRange"`           // --calibration-range=N, 0=omit
	FreezeTrackingOnDisconnect bool    `json:"freezeTrackingOnDisconnect"` // --freeze-tracking-on-disconnect
	Custom                     string  `json:"custom"`                     // remaining args as-is
}

var (
	noVrShort                   = "-no-vr"
	noVrLong                    = "--no-vr"
	screenFull                  = "-screen-fullscreen"
	fullscreen1                 = "-screen-fullscreen 1"
	windowed                    = "-windowed"
	popupwindow                 = "-popupwindow"
	screenWidthArg              = "-screen-width"
	screenHeightArg             = "-screen-height"
	fpsPrefix                   = "--fps="
	skipRegistry                = "--skip-registry-install"
	processPriorityPrefix       = "--process-priority="
	mainThreadPriorityPrefix    = "--main-thread-priority="
	monitorArg                  = "-monitor"
	profilePrefix               = "--profile="
	enableDebugGui              = "--enable-debug-gui"
	enableSDKLogLevels          = "--enable-sdk-log-levels"
	enableUdonDebugLogging      = "--enable-udon-debug-logging"
	midiPrefix                  = "--midi="
	watchWorlds                 = "--watch-worlds"
	watchAvatars                = "--watch-avatars"
	ignoreTrackersPrefix        = "--ignore-trackers="
	disableHwVideoDecoding      = "--disable-hw-video-decoding"
	enableHwVideoDecoding       = "--enable-hw-video-decoding"
	disableAmdStutterWorkaround = "--disable-amd-stutter-workaround"
	oscPrefix                   = "--osc="
	affinityPrefix              = "--affinity="
	enforceWorldServerChecks    = "--enforce-world-server-checks"
	customArmRatioPrefix        = "--custom-arm-ratio="
	disableShoulderTracking     = "--disable-shoulder-tracking"
	enableIKDebugLogging        = "--enable-ik-debug-logging"
	calibrationRangePrefix      = "--calibration-range="
	freezeTrackingOnDisconnect  = "--freeze-tracking-on-disconnect"
)

// ParseLaunchArgsForGUI parses a launch arguments string into GUI fields.
// Detects known args; everything else goes to Custom. Order of custom args is preserved.
func ParseLaunchArgsForGUI(args string) *LaunchArgsParsed {
	p := &LaunchArgsParsed{
		Profile:            -1,
		ProcessPriority:    PriorityOmit,
		MainThreadPriority: PriorityOmit,
	}
	if args == "" {
		return p
	}
	tokens := parseLaunchArgsTokens(normalizeIKQuotedArgs(args))
	var customParts []string
	i := 0
	for i < len(tokens) {
		tok := tokens[i]
		switch {
		case tok == noVrShort || tok == noVrLong:
			p.NoVR = true
		case tok == screenFull:
			if i+1 < len(tokens) {
				if tokens[i+1] == "1" {
					p.ScreenMode = ScreenModeFullscreen
				}
				i++
			}
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
		case tok == monitorArg:
			if i+1 < len(tokens) {
				if n, err := strconv.Atoi(tokens[i+1]); err == nil && n >= 1 {
					p.Monitor = n
				}
				i++
			}
		case strings.HasPrefix(tok, fpsPrefix):
			if v := strings.TrimPrefix(tok, fpsPrefix); v != "" {
				if n, err := strconv.Atoi(v); err == nil && n > 0 {
					p.FPS = n
				}
			}
		case strings.HasPrefix(tok, profilePrefix):
			if v := strings.TrimPrefix(tok, profilePrefix); v != "" {
				if n, err := strconv.Atoi(v); err == nil && n >= 0 {
					p.Profile = n
				}
			}
		case tok == skipRegistry:
			p.SkipRegistry = true
		case tok == enableDebugGui:
			p.EnableDebugGui = true
		case tok == enableSDKLogLevels:
			p.EnableSDKLogLevels = true
		case tok == enableUdonDebugLogging:
			p.EnableUdonDebugLogging = true
		case strings.HasPrefix(tok, midiPrefix):
			if v := strings.TrimPrefix(tok, midiPrefix); v != "" {
				p.Midi = v
			}
		case tok == watchWorlds:
			p.WatchWorlds = true
		case tok == watchAvatars:
			p.WatchAvatars = true
		case strings.HasPrefix(tok, ignoreTrackersPrefix):
			if v := strings.TrimPrefix(tok, ignoreTrackersPrefix); v != "" {
				p.IgnoreTrackers = v
			}
		case tok == disableHwVideoDecoding:
			p.VideoDecoding = VideoDecodingSoftware
		case tok == enableHwVideoDecoding:
			p.VideoDecoding = VideoDecodingHardware
		case tok == disableAmdStutterWorkaround:
			p.DisableAMDStutterWorkaround = true
		case strings.HasPrefix(tok, oscPrefix):
			if v := strings.TrimPrefix(tok, oscPrefix); v != "" {
				p.OSC = v
			}
		case strings.HasPrefix(tok, affinityPrefix):
			if v := strings.TrimPrefix(tok, affinityPrefix); v != "" {
				p.Affinity = v
			}
		case strings.HasPrefix(tok, processPriorityPrefix):
			if v := strings.TrimPrefix(tok, processPriorityPrefix); v != "" {
				if n, err := strconv.Atoi(v); err == nil && n >= -2 && n <= 2 {
					p.ProcessPriority = n
				}
			}
		case strings.HasPrefix(tok, mainThreadPriorityPrefix):
			if v := strings.TrimPrefix(tok, mainThreadPriorityPrefix); v != "" {
				if n, err := strconv.Atoi(v); err == nil && n >= -2 && n <= 2 {
					p.MainThreadPriority = n
				}
			}
		case tok == enforceWorldServerChecks:
			p.EnforceWorldServerChecks = true
		case strings.HasPrefix(tok, customArmRatioPrefix):
			if n, ok := parsePositiveFiniteFloat(strings.TrimPrefix(tok, customArmRatioPrefix)); ok {
				p.CustomArmRatio = n
			}
		case tok == disableShoulderTracking:
			p.DisableShoulderTracking = true
		case tok == enableIKDebugLogging:
			p.EnableIKDebugLogging = true
		case strings.HasPrefix(tok, calibrationRangePrefix):
			if n, ok := parsePositiveFiniteFloat(strings.TrimPrefix(tok, calibrationRangePrefix)); ok {
				p.CalibrationRange = n
			}
		case tok == freezeTrackingOnDisconnect:
			p.FreezeTrackingOnDisconnect = true
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

// normalizeIKQuotedArgs rewrites only known IK value options from --key="v" / --key='v'
// to --key=v so the quote tokenizer still yields one token. Other Custom args keep quotes.
func normalizeIKQuotedArgs(s string) string {
	if s == "" {
		return s
	}
	for _, prefix := range []string{customArmRatioPrefix, calibrationRangePrefix} {
		s = stripQuotedValueAfterPrefix(s, prefix)
	}
	return s
}

func stripQuotedValueAfterPrefix(s, prefix string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		idx := strings.Index(s[i:], prefix)
		if idx < 0 {
			b.WriteString(s[i:])
			break
		}
		idx += i
		b.WriteString(s[i:idx])
		b.WriteString(prefix)
		restStart := idx + len(prefix)
		if restStart >= len(s) {
			i = restStart
			break
		}
		quote := s[restStart]
		if quote != '"' && quote != '\'' {
			i = restStart
			continue
		}
		endRel := strings.IndexByte(s[restStart+1:], quote)
		if endRel < 0 {
			i = restStart
			continue
		}
		b.WriteString(s[restStart+1 : restStart+1+endRel])
		i = restStart + 1 + endRel + 1
	}
	return b.String()
}

func stripLaunchArgQuotes(v string) string {
	if len(v) >= 2 {
		if (v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'') {
			return v[1 : len(v)-1]
		}
	}
	return v
}

func parsePositiveFiniteFloat(v string) (float64, bool) {
	n, err := strconv.ParseFloat(stripLaunchArgQuotes(v), 64)
	if err != nil || !(n > 0) || math.IsInf(n, 0) {
		return 0, false
	}
	return n, true
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
	if p.Monitor >= 1 {
		parts = append(parts, monitorArg, strconv.Itoa(p.Monitor))
	}
	if p.FPS > 0 {
		parts = append(parts, fpsPrefix+strconv.Itoa(p.FPS))
	}
	if p.Profile >= 0 {
		parts = append(parts, profilePrefix+strconv.Itoa(p.Profile))
	}
	if p.SkipRegistry {
		parts = append(parts, skipRegistry)
	}
	if p.EnableDebugGui {
		parts = append(parts, enableDebugGui)
	}
	if p.EnableSDKLogLevels {
		parts = append(parts, enableSDKLogLevels)
	}
	if p.EnableUdonDebugLogging {
		parts = append(parts, enableUdonDebugLogging)
	}
	if p.Midi != "" {
		parts = append(parts, midiPrefix+p.Midi)
	}
	if p.WatchWorlds {
		parts = append(parts, watchWorlds)
	}
	if p.WatchAvatars {
		parts = append(parts, watchAvatars)
	}
	if p.IgnoreTrackers != "" {
		parts = append(parts, ignoreTrackersPrefix+p.IgnoreTrackers)
	}
	switch p.VideoDecoding {
	case VideoDecodingSoftware:
		parts = append(parts, disableHwVideoDecoding)
	case VideoDecodingHardware:
		parts = append(parts, enableHwVideoDecoding)
	}
	if p.DisableAMDStutterWorkaround {
		parts = append(parts, disableAmdStutterWorkaround)
	}
	if p.OSC != "" {
		parts = append(parts, oscPrefix+p.OSC)
	}
	if p.Affinity != "" {
		parts = append(parts, affinityPrefix+p.Affinity)
	}
	if p.ProcessPriority >= -2 && p.ProcessPriority <= 2 {
		parts = append(parts, processPriorityPrefix+strconv.Itoa(p.ProcessPriority))
	}
	if p.MainThreadPriority >= -2 && p.MainThreadPriority <= 2 {
		parts = append(parts, mainThreadPriorityPrefix+strconv.Itoa(p.MainThreadPriority))
	}
	if p.EnforceWorldServerChecks {
		parts = append(parts, enforceWorldServerChecks)
	}
	if p.CustomArmRatio > 0 {
		parts = append(parts, customArmRatioPrefix+strconv.FormatFloat(p.CustomArmRatio, 'f', -1, 64))
	}
	if p.DisableShoulderTracking {
		parts = append(parts, disableShoulderTracking)
	}
	if p.EnableIKDebugLogging {
		parts = append(parts, enableIKDebugLogging)
	}
	if p.CalibrationRange > 0 {
		parts = append(parts, calibrationRangePrefix+strconv.FormatFloat(p.CalibrationRange, 'f', -1, 64))
	}
	if p.FreezeTrackingOnDisconnect {
		parts = append(parts, freezeTrackingOnDisconnect)
	}
	if p.Custom != "" {
		parts = append(parts, strings.TrimSpace(p.Custom))
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}
