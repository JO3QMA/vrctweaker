//go:build windows

package locale

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// Detect returns the user's default Windows UI locale mapped to an app language code.
func Detect() string {
	kernel32 := windows.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetUserDefaultLocaleName")
	buf := make([]uint16, 85) // LOCALE_NAME_MAX_LENGTH
	r0, _, _ := proc.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r0 == 0 {
		return "en"
	}
	name := windows.UTF16ToString(buf)
	if name == "" {
		return "en"
	}
	return MapToAppLanguage(name)
}
