//go:build !windows

package locale

import "os"

// Detect reads the OS locale from environment variables and maps it to an app language code.
func Detect() string {
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}
	if lang == "" {
		lang = os.Getenv("LC_MESSAGES")
	}
	return MapToAppLanguage(lang)
}
