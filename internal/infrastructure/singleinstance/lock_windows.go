//go:build windows

package singleinstance

import (
	"fmt"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

type winGuard struct {
	mutexName   string
	eventName   string
	windowTitle string
	mutex       windows.Handle
	event       windows.Handle
	listener    sync.WaitGroup
}

func newPlatformGuard(name, windowTitle string) platformGuard {
	return &winGuard{
		mutexName:   `Local\` + name + `_SingleInstance`,
		eventName:   `Local\` + name + `_Activate`,
		windowTitle: windowTitle,
	}
}

func (w *winGuard) acquire() (bool, error) {
	mutexName, err := windows.UTF16PtrFromString(w.mutexName)
	if err != nil {
		return false, fmt.Errorf("singleinstance mutex name: %w", err)
	}
	mutex, err := windows.CreateMutex(nil, true, mutexName)
	if err != nil {
		return false, fmt.Errorf("singleinstance CreateMutex: %w", err)
	}
	if windows.GetLastError() == windows.ERROR_ALREADY_EXISTS {
		_ = windows.CloseHandle(mutex)
		return false, nil
	}

	eventName, err := windows.UTF16PtrFromString(w.eventName)
	if err != nil {
		_ = windows.ReleaseMutex(mutex)
		_ = windows.CloseHandle(mutex)
		return false, fmt.Errorf("singleinstance event name: %w", err)
	}
	event, err := windows.CreateEvent(nil, false, false, eventName)
	if err != nil {
		_ = windows.ReleaseMutex(mutex)
		_ = windows.CloseHandle(mutex)
		return false, fmt.Errorf("singleinstance CreateEvent: %w", err)
	}

	w.mutex = mutex
	w.event = event
	return true, nil
}

func (w *winGuard) notifyExisting() error {
	signalErr := w.signalEvent()
	if signalErr == nil {
		return nil
	}
	fallbackErr := w.activateWindowByTitle(w.windowTitle)
	if fallbackErr == nil {
		return nil
	}
	return fmt.Errorf("singleinstance: existing instance could not be activated (signal: %w; fallback: %w)", signalErr, fallbackErr)
}

func (w *winGuard) signalEvent() error {
	eventName, err := windows.UTF16PtrFromString(w.eventName)
	if err != nil {
		return err
	}
	event, err := windows.OpenEvent(windows.EVENT_MODIFY_STATE, false, eventName)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(event)
	if err := windows.SetEvent(event); err != nil {
		return err
	}
	return nil
}

func (w *winGuard) start(stop <-chan struct{}, dispatch func()) error {
	eventHandle := w.event
	if eventHandle == 0 {
		return nil
	}
	w.listener.Add(1)
	go func() {
		defer w.listener.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			status, err := windows.WaitForSingleObject(eventHandle, 200)
			if err != nil {
				return
			}
			switch status {
			case windows.WAIT_OBJECT_0:
				dispatch()
			case uint32(windows.WAIT_FAILED):
				return
			}
		}
	}()
	return nil
}

func (w *winGuard) release() {
	w.listener.Wait()
	if w.event != 0 {
		_ = windows.CloseHandle(w.event)
		w.event = 0
	}
	if w.mutex != 0 {
		_ = windows.ReleaseMutex(w.mutex)
		_ = windows.CloseHandle(w.mutex)
		w.mutex = 0
	}
}

var (
	modUser32               = windows.NewLazySystemDLL("user32.dll")
	procFindWindowW         = modUser32.NewProc("FindWindowW")
	procShowWindow          = modUser32.NewProc("ShowWindow")
	procSetForegroundWindow = modUser32.NewProc("SetForegroundWindow")
)

const swRestore = 9

func (w *winGuard) activateWindowByTitle(title string) error {
	if title == "" {
		return fmt.Errorf("window title not configured")
	}
	titlePtr, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return err
	}
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd == 0 {
		if errno := windows.GetLastError(); errno != windows.ERROR_SUCCESS {
			return errno
		}
		return fmt.Errorf("window not found")
	}
	procShowWindow.Call(hwnd, swRestore)
	if ret, _, _ := procSetForegroundWindow.Call(hwnd); ret == 0 {
		return fmt.Errorf("SetForegroundWindow failed")
	}
	return nil
}
