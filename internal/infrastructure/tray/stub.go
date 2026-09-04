//go:build !windows

package tray

type stubManager struct{}

func newManager() Manager {
	return stubManager{}
}

func (stubManager) Supported() bool { return false }

func (stubManager) Running() bool { return false }

func (stubManager) Start(Config) error { return nil }

func (stubManager) Stop() error { return nil }
