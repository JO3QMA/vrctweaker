//go:build !windows

package singleinstance

import (
	"log"
	"net"
	"os"
)

// authorizedPeer accepts activation requests only from processes running as the same uid.
// On Linux this uses SO_PEERCRED; on BSD/macOS LOCAL_PEERCRED via getpeereid-style lookup.
// When peer credentials are unavailable on the platform, activation is rejected (fail closed).
func authorizedPeer(conn net.Conn) bool {
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
	if err != nil {
		log.Printf("singleinstance: peer credential lookup failed: %v", err)
		return false
	}
	return uid == os.Getuid()
}
