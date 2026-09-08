//go:build !windows && !linux && !darwin && !freebsd && !openbsd && !netbsd

package singleinstance

import "errors"

func peerUID(syscallRawConn) (int, error) {
	return 0, errors.New("peer credentials unsupported on this platform")
}
