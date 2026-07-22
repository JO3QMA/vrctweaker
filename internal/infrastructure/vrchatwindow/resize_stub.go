//go:build !windows

package vrchatwindow

func resize(width, height int) error {
	_, _ = width, height
	return ErrUnsupported
}
