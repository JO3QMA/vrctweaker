//go:build !windows

package sleepsuppress

type stubVRChatChecker struct{}

// NewVRChatProcessChecker returns a no-op checker (always false on non-Windows).
func NewVRChatProcessChecker() ProcessChecker {
	return stubVRChatChecker{}
}

func (stubVRChatChecker) VRChatRunning() (bool, error) {
	return false, nil
}
