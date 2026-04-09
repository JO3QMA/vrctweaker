package locale

import "strings"

// MapToAppLanguage maps an OS locale string (e.g. ja-JP, en_US.UTF-8) to an app UI code:
// ja, en, ko, zh-TW, zh-CN. Unknown values become "en".
func MapToAppLanguage(raw string) string {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" || s == "c" || s == "posix" {
		return "en"
	}
	if i := strings.IndexByte(s, '.'); i >= 0 {
		s = s[:i]
	}
	if i := strings.IndexByte(s, '@'); i >= 0 {
		s = s[:i]
	}
	s = strings.ReplaceAll(s, "_", "-")

	switch {
	case strings.HasPrefix(s, "ja"):
		return "ja"
	case strings.HasPrefix(s, "ko"):
		return "ko"
	case strings.HasPrefix(s, "zh-tw") || strings.HasPrefix(s, "zh-hant"):
		return "zh-TW"
	case strings.HasPrefix(s, "zh-cn") || strings.HasPrefix(s, "zh-hans") || strings.HasPrefix(s, "zh-sg"):
		return "zh-CN"
	case strings.HasPrefix(s, "zh-hk") || strings.HasPrefix(s, "zh-mo"):
		return "zh-TW"
	case strings.HasPrefix(s, "zh"):
		return "zh-CN"
	default:
		return "en"
	}
}
