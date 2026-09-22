package wailsapp

import "runtime"

// GetLogicalProcessorCount returns the number of logical CPUs on this machine.
func (a *App) GetLogicalProcessorCount() (int, error) {
	_ = a
	n := runtime.NumCPU()
	if n < 1 {
		return 1, nil
	}
	return n, nil
}
