//go:build windows

package usecase

import "golang.org/x/sys/windows"

// isReparsePoint reports Windows junctions / mount points / symlinks via FILE_ATTRIBUTE_REPARSE_POINT.
// Fail closed: if attributes cannot be read, treat as reparse so we use os.Remove (not RemoveAll).
func isReparsePoint(path string) bool {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return true
	}
	attrs, err := windows.GetFileAttributes(p)
	if err != nil {
		return true
	}
	return attrs&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0
}
