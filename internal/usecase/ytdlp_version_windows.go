//go:build windows

package usecase

import (
	"fmt"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// localYTDLPFileVersionString reads FileVersion / ProductVersion from the PE VERSIONINFO
// resource (same metadata as Explorer “詳細” / file properties), without executing the exe.
func localYTDLPFileVersionString(exePath string) string {
	if exePath == "" {
		return ""
	}
	abs, err := filepath.Abs(exePath)
	if err != nil {
		abs = exePath
	}
	size, err := windows.GetFileVersionInfoSize(abs, nil)
	if err != nil || size == 0 {
		return ""
	}
	buf := make([]byte, size)
	err = windows.GetFileVersionInfo(abs, 0, size, unsafe.Pointer(&buf[0]))
	if err != nil {
		return ""
	}
	return stringFileInfoVersionKeys(unsafe.Pointer(&buf[0]), "FileVersion", "ProductVersion")
}

func stringFileInfoVersionKeys(block unsafe.Pointer, keys ...string) string {
	var trans unsafe.Pointer
	var transLen uint32
	err := windows.VerQueryValue(block, `\VarFileInfo\Translation`, unsafe.Pointer(&trans), &transLen)
	if err != nil || trans == nil || transLen < 4 {
		return ""
	}
	nLang := int(transLen) / 4
	langCPs := unsafe.Slice((*uint32)(trans), nLang)
	for _, langCP := range langCPs {
		lang := langCP & 0xffff
		cp := (langCP >> 16) & 0xffff
		prefix := fmt.Sprintf(`\StringFileInfo\%04x%04x\`, lang, cp)
		for _, key := range keys {
			var val unsafe.Pointer
			var valLen uint32
			sub := prefix + key
			err = windows.VerQueryValue(block, sub, unsafe.Pointer(&val), &valLen)
			if err != nil || val == nil || valLen < 2 {
				continue
			}
			n := valLen / 2
			if n == 0 {
				continue
			}
			u16 := unsafe.Slice((*uint16)(val), n)
			s := windows.UTF16ToString(u16)
			s = strings.TrimSpace(s)
			if s != "" {
				return s
			}
		}
	}
	return ""
}
