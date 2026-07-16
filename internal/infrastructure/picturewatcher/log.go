package picturewatcher

// Logger logs non-fatal diagnostics. Optional; pass Nop when unused.
type Logger func(format string, args ...any)

// Nop discards log output.
func Nop(string, ...any) {}
