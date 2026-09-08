//go:build !windows

package singleinstance

import (
	"errors"
	"log"
	"net"
	"os"
)

// authorizedPeer accepts activation requests only from processes running as the same uid.
// On Linux this uses SO_PEERCRED; on BSD/macOS LOCAL_PEERCRED.
// Abstract sockets require peer credentials. Filesystem sockets in the user-private
// AppDataDir skip peer auth when credentials are unavailable on the platform.
func authorizedPeer(conn net.Conn, abstractSocket bool) bool {
	uc, ok := conn.(*net.UnixConn)
	if !ok {
		return false
	}
	raw, err := uc.SyscallConn()
	if err != nil {
		log.Printf("singleinstance: peer credential lookup failed: %v", err)
		return false
	}
	uid, err := peerUID(raw)
	if errors.Is(err, errPeerCredUnsupported) {
		if abstractSocket {
			log.Printf("singleinstance: peer credentials required for abstract activation socket on this platform")
			return false
		}
		return true
	}
	if err != nil {
		log.Printf("singleinstance: peer credential lookup failed: %v", err)
		return false
	}
	return uid == os.Getuid()
}
