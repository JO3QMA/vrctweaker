//go:build !windows

package singleinstance

import "errors"

var errPeerCredUnsupported = errors.New("peer credentials unsupported on this platform")

type syscallRawConn interface {
	Control(f func(fd uintptr)) error
}
