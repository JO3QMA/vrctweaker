package tray

import (
	"log"
	"sync"
)

// runAsync invokes tray menu handlers without blocking the Win32 message loop.
// wg, when non-nil, tracks in-flight handlers so the tray session can drain them before exit.
func runAsync(wg *sync.WaitGroup, fn func()) {
	if fn == nil {
		return
	}
	if wg != nil {
		wg.Add(1)
	}
	go func() {
		if wg != nil {
			defer wg.Done()
		}
		defer func() {
			if r := recover(); r != nil {
				log.Printf("tray: handler panic: %v", r)
			}
		}()
		fn()
	}()
}
