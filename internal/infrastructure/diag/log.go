package diag

import "log"

// Logger logs non-fatal diagnostics. Optional; pass Nop when unused.
type Logger func(format string, args ...any)

// Nop discards log output.
func Nop(string, ...any) {}

// Std returns a Logger that writes to log.Default().
func Std() Logger {
	return func(format string, args ...any) {
		log.Printf(format, args...)
	}
}
