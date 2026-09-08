//go:build !windows && !linux && !darwin && !freebsd && !openbsd && !netbsd

package singleinstance

func peerUID(syscallRawConn) (int, error) {
	return 0, errPeerCredUnsupported
}
