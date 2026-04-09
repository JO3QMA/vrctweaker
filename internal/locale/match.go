package locale

import (
	"strings"

	"golang.org/x/text/language"
)

// Canonical UI language codes persisted in app_settings and used by the frontend.
const (
	UILangJA   = "ja"
	UILangEN   = "en"
	UILangZHCN = "zh-CN"
	UILangZHTW = "zh-TW"
	UILangKO   = "ko"
)

var supportedTags = []language.Tag{
	language.Japanese,
	language.English,
	language.MustParse("zh-CN"),
	language.MustParse("zh-TW"),
	language.Korean,
}

var uiMatcher = language.NewMatcher(supportedTags)

var (
	baseJA     = language.MustParseBase("ja")
	baseKO     = language.MustParseBase("ko")
	baseEN     = language.MustParseBase("en")
	baseZH     = language.MustParseBase("zh")
	scriptHant = mustParseScript("Hant")
	regionTW   = language.MustParseRegion("TW")
	regionHK   = language.MustParseRegion("HK")
	regionMO   = language.MustParseRegion("MO")
)

func mustParseScript(s string) language.Script {
	sc, err := language.ParseScript(s)
	if err != nil {
		panic(err)
	}
	return sc
}

// NormalizeUILocale returns a supported code or empty if s is not a known supported value.
func NormalizeUILocale(s string) string {
	s = strings.TrimSpace(s)
	switch s {
	case UILangJA, UILangEN, UILangZHCN, UILangZHTW, UILangKO:
		return s
	default:
		return ""
	}
}

// CanonicalUILanguage maps exact canonical codes or BCP47-style tags (e.g. ja-JP, en-US)
// to a supported UI code. Empty input returns empty. Unparseable input returns empty.
func CanonicalUILanguage(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if n := NormalizeUILocale(s); n != "" {
		return n
	}
	prepped := strings.ReplaceAll(s, "_", "-")
	if i := strings.IndexAny(prepped, "."); i >= 0 {
		prepped = prepped[:i]
	}
	prepped = strings.TrimSpace(prepped)
	if prepped == "" {
		return ""
	}
	tags, _, err := language.ParseAcceptLanguage(prepped)
	if err != nil || len(tags) == 0 {
		t, perr := language.Parse(prepped)
		if perr != nil {
			return ""
		}
		tags = []language.Tag{t}
	}
	matched, _, _ := uiMatcher.Match(tags...)
	return tagToUICode(matched)
}

// CanonicalUILanguageForSet maps user-chosen language codes to a stored UI code.
// It accepts exact canonical values and BCP47 variants for ja / en / zh / ko only.
// Unsupported bases (e.g. fr) return ok=false so SetUILanguage can reject them.
func CanonicalUILanguageForSet(s string) (code string, ok bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	if n := NormalizeUILocale(s); n != "" {
		return n, true
	}
	prepped := strings.ReplaceAll(s, "_", "-")
	if i := strings.IndexAny(prepped, "."); i >= 0 {
		prepped = prepped[:i]
	}
	prepped = strings.TrimSpace(prepped)
	if prepped == "" {
		return "", false
	}
	tags, _, err := language.ParseAcceptLanguage(prepped)
	if err != nil || len(tags) == 0 {
		t, perr := language.Parse(prepped)
		if perr != nil {
			return "", false
		}
		tags = []language.Tag{t}
	}
	t := tags[0]
	b, _ := t.Base()
	switch b {
	case baseJA:
		return UILangJA, true
	case baseEN:
		return UILangEN, true
	case baseKO:
		return UILangKO, true
	case baseZH:
		return tagToUICode(t), true
	default:
		return "", false
	}
}

// MatchRawLocale maps an OS or Accept-Language style string to a supported UI code.
// Unknown or empty input falls back to UILangEN.
func MatchRawLocale(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "C") || strings.EqualFold(raw, "POSIX") {
		return UILangEN
	}
	raw = strings.ReplaceAll(raw, "_", "-")
	if i := strings.IndexAny(raw, "."); i >= 0 {
		raw = raw[:i]
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return UILangEN
	}

	tags, _, err := language.ParseAcceptLanguage(raw)
	if err != nil || len(tags) == 0 {
		t, perr := language.Parse(raw)
		if perr != nil {
			return UILangEN
		}
		tags = []language.Tag{t}
	}

	matched, _, _ := uiMatcher.Match(tags...)
	return tagToUICode(matched)
}

func tagToUICode(t language.Tag) string {
	b, _ := t.Base()
	switch b {
	case baseJA:
		return UILangJA
	case baseKO:
		return UILangKO
	case baseEN:
		return UILangEN
	case baseZH:
		script, _ := t.Script()
		region, _ := t.Region()
		if script == scriptHant {
			return UILangZHTW
		}
		if region == regionTW || region == regionHK || region == regionMO {
			return UILangZHTW
		}
		return UILangZHCN
	default:
		s := t.String()
		if strings.HasPrefix(s, "zh-TW") || strings.HasPrefix(s, "zh-Hant") {
			return UILangZHTW
		}
		if strings.HasPrefix(s, "zh") {
			return UILangZHCN
		}
		return UILangEN
	}
}

// ResolveFromOS reads the platform user locale and returns a supported UI code.
func ResolveFromOS() string {
	return MatchRawLocale(userPreferredLocale())
}
