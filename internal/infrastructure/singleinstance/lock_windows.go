//go:build windows

package singleinstance

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type winGuard struct {
	mutexName string
	eventName string
	mutex     windows.Handle
	event     windows.Handle
	listener  sync.WaitGroup
}

func newPlatformGuard(name string) platformGuard {
	return &winGuard{
		mutexName: `Local\` + name + `_SingleInstance`,
		eventName: `Local\` + name + `_Activate`,
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
	fallbackErr := w.activateWindowByTitle(DefaultWindowTitle)
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
	if w.event == 0 {
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
			status, err := windows.WaitForSingleObject(w.event, 200)
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
	if w.event != 0 {
		_ = windows.CloseHandle(w.event)
		w.event = 0
	}
	if w.mutex != 0 {
		_ = windows.ReleaseMutex(w.mutex)
		_ = windows.CloseHandle(w.mutex)
		w.mutex = 0
	}
	w.listener.Wait()
}

var (
	modUser32               = windows.NewLazySystemDLL("user32.dll")
	procFindWindowW         = modUser32.NewProc("FindWindowW")
	procShowWindow          = modUser32.NewProc("ShowWindow")
	procSetForegroundWindow = modUser32.NewProc("SetForegroundWindow")
)

const swRestore = 9

func (w *winGuard) activateWindowByTitle(title string) error {
	titlePtr, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return err
	}
	hwnd, _, err := procFindWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd == 0 {
		if err != nil && err != syscall.Errno(0) {
			return err
		}
		return fmt.Errorf("window not found")
	}
	procShowWindow.Call(hwnd, swRestore)
	procSetForegroundWindow.Call(hwnd)
	return nil
}
