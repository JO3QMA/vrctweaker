//go:build windows

package singleinstance

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modUser32                    = windows.NewLazySystemDLL("user32.dll")
	procFindWindowW              = modUser32.NewProc("FindWindowW")
	procShowWindow               = modUser32.NewProc("ShowWindow")
	procSetForegroundWindow      = modUser32.NewProc("SetForegroundWindow")
	procBringWindowToTop         = modUser32.NewProc("BringWindowToTop")
	procGetForegroundWindow      = modUser32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = modUser32.NewProc("GetWindowThreadProcessId")
	procAttachThreadInput        = modUser32.NewProc("AttachThreadInput")
	procAllowSetForegroundWindow = modUser32.NewProc("AllowSetForegroundWindow")
	procIsIconic                 = modUser32.NewProc("IsIconic")
	procIsWindowVisible          = modUser32.NewProc("IsWindowVisible")
)

const (
	swShow    = 5
	swRestore = 9
)

// ActivateWindowByTitle finds the main window by title and brings it to the foreground.
func ActivateWindowByTitle(title string) error {
	if title == "" {
		return fmt.Errorf("window title not configured")
	}
	hwnd, err := findMainWindow(title)
	if err != nil {
		return err
	}
	allowErr := allowForegroundForWindow(hwnd)
	if err := forceForeground(hwnd); err != nil {
		if allowErr != nil {
			return fmt.Errorf("%w; %w", allowErr, err)
		}
		return err
	}
	return allowErr
}

func findMainWindow(title string) (windows.HWND, error) {
	titlePtr, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return 0, err
	}
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd == 0 {
		if errno := windows.GetLastError(); errno != windows.ERROR_SUCCESS {
			return 0, errno
		}
		return 0, fmt.Errorf("window not found")
	}
	return windows.HWND(hwnd), nil
}

func allowForegroundForWindow(hwnd windows.HWND) error {
	var pid uint32
	_, _, _ = procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return fmt.Errorf("could not determine window PID")
	}
	ret, _, _ := procAllowSetForegroundWindow.Call(uintptr(pid))
	if ret == 0 {
		return fmt.Errorf("AllowSetForegroundWindow failed")
	}
	return nil
}

func showMainWindow(hwnd windows.HWND) {
	if isIconic(hwnd) {
		_, _, _ = procShowWindow.Call(uintptr(hwnd), swRestore)
		return
	}
	if !isWindowVisible(hwnd) {
		_, _, _ = procShowWindow.Call(uintptr(hwnd), swShow)
	}
}

func isIconic(hwnd windows.HWND) bool {
	ret, _, _ := procIsIconic.Call(uintptr(hwnd))
	return ret != 0
}

func isWindowVisible(hwnd windows.HWND) bool {
	ret, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
	return ret != 0
}

func forceForeground(hwnd windows.HWND) error {
	showMainWindow(hwnd)

	foreground, _, _ := procGetForegroundWindow.Call()
	if foreground == 0 {
		if ret, _, _ := procSetForegroundWindow.Call(uintptr(hwnd)); ret == 0 {
			return fmt.Errorf("SetForegroundWindow failed")
		}
		_, _, _ = procBringWindowToTop.Call(uintptr(hwnd))
		return nil
	}
	if foreground == uintptr(hwnd) {
		_, _, _ = procBringWindowToTop.Call(uintptr(hwnd))
		return nil
	}

	fgThread, _, _ := procGetWindowThreadProcessId.Call(foreground, 0)
	targetThread, _, _ := procGetWindowThreadProcessId.Call(uintptr(hwnd), 0)
	attached := false
	if fgThread != 0 && targetThread != 0 && fgThread != targetThread {
		ret, _, _ := procAttachThreadInput.Call(fgThread, targetThread, 1)
		attached = ret != 0
	}
	if attached {
		defer procAttachThreadInput.Call(fgThread, targetThread, 0)
	}

	if ret, _, _ := procSetForegroundWindow.Call(uintptr(hwnd)); ret == 0 {
		return fmt.Errorf("SetForegroundWindow failed")
	}
	_, _, _ = procBringWindowToTop.Call(uintptr(hwnd))
	return nil
}
