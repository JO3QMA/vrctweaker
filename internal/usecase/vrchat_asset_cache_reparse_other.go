//go:build !windows

package usecase

// isReparsePoint is a no-op outside Windows; Unix symlinks are handled via os.ModeSymlink.
func isReparsePoint(string) bool { return false }
