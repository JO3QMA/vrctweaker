package vrchatosc

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestListener_receivesMuteSelf(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l := NewListener()
	if err := l.Start(ctx, "127.0.0.1:0"); err != nil {
		t.Fatal(err)
	}
	defer l.Stop()

	snap := l.Snapshot()
	addr, err := net.ResolveUDPAddr("udp", snap.ListenAddr)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	pkt := buildOSCPacket("/avatar/parameters/MuteSelf", "T", nil)
	if _, err := conn.Write(pkt); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		s := l.Snapshot()
		if s.MuteKnown && s.Muted {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for mute; snap=%+v", l.Snapshot())
}
