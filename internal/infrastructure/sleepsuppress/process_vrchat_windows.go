//go:build windows

package sleepsuppress

import (
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type vrChatProcessChecker struct{}

// NewVRChatProcessChecker returns a checker that looks for VRChat.exe (case-insensitive).
func NewVRChatProcessChecker() ProcessChecker {
	return vrChatProcessChecker{}
}

func (vrChatProcessChecker) VRChatRunning() (bool, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return false, err
	}
	defer windows.CloseHandle(snapshot)

	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	if err = windows.Process32First(snapshot, &pe); err != nil {
		return false, err
	}
	for {
		name := windows.UTF16ToString(pe.ExeFile[:])
		if strings.EqualFold(name, "VRChat.exe") {
			return true, nil
		}
		err = windows.Process32Next(snapshot, &pe)
		if err != nil {
			if err == windows.ERROR_NO_MORE_FILES {
				return false, nil
			}
			return false, err
		}
	}
}
