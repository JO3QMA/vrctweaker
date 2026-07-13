//go:build windows

package usecase

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

const maxYTDLPVersionInfoBytes = 4 << 20

// localYTDLPFileVersionString reads FileVersion / ProductVersion from the PE VERSIONINFO
// resource (same metadata as Explorer “詳細” / file properties), without executing the exe.
func localYTDLPFileVersionString(exePath string) string {
	if exePath == "" {
		return ""
	}
	abs, err := filepath.Abs(exePath)
	if err != nil {
		log.Printf("ytdlp version: Abs(%s): %v", exePath, err)
		abs = exePath
	}
	size, err := windows.GetFileVersionInfoSize(abs, nil)
	if err != nil || size == 0 {
		if err != nil {
			log.Printf("ytdlp version: GetFileVersionInfoSize(%s): %v", abs, err)
		}
		return ""
	}
	if size > maxYTDLPVersionInfoBytes {
		log.Printf("ytdlp version: VERSIONINFO too large (%d bytes) for %s", size, abs)
		return ""
	}
	buf := make([]byte, size)
	err = windows.GetFileVersionInfo(abs, 0, size, unsafe.Pointer(&buf[0]))
	if err != nil {
		log.Printf("ytdlp version: GetFileVersionInfo(%s): %v", abs, err)
		return ""
	}
	return stringFileInfoVersionKeys(unsafe.Pointer(&buf[0]), "FileVersion", "ProductVersion")
}

// stringFileInfoVersionKeys walks StringFileInfo translation blocks and returns the first
// non-empty value among keys (e.g. FileVersion, ProductVersion) from the PE VERSIONINFO resource.
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
			if valLen%2 != 0 {
				continue
			}
			// VerQueryValue sets valLen to the value size in bytes (UTF-16 code units × 2).
			charCount := valLen / uint32(unsafe.Sizeof(uint16(0)))
			if charCount == 0 {
				continue
			}
			u16 := unsafe.Slice((*uint16)(val), charCount)
			s := windows.UTF16ToString(u16)
			s = strings.TrimSpace(s)
			if s != "" {
				return s
			}
		}
	}
	return ""
}
