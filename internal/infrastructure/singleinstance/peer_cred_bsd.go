//go:build darwin || freebsd || openbsd || netbsd

package singleinstance

import (
	"golang.org/x/sys/unix"
)

func peerUID(raw syscallRawConn) (int, error) {
	var uid int
	var ctlErr error
	if err := raw.Control(func(fd uintptr) {
		cred, err := unix.GetsockoptUcred(int(fd), unix.SOL_LOCAL, unix.LOCAL_PEERCRED)
		if err != nil {
			ctlErr = err
			return
		}
		uid = int(cred.Uid)
	}); err != nil {
		return 0, err
	}
	if ctlErr != nil {
		return 0, ctlErr
	}
	return uid, nil
}
