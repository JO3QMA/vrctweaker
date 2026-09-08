//go:build !windows

package singleinstance

// ActivateWindowByTitle is a no-op on non-Windows platforms.
func ActivateWindowByTitle(_ string) error {
	return nil
}
