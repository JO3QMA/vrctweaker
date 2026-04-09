//go:build !windows && !darwin

package locale

import (
	"os"
	"strings"
)

func userPreferredLocale() string {
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}
