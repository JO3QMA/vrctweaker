package vrchatosc

import (
	"context"
	"net"
	"sync"
	"time"
)

// Listener receives OSC packets from VRChat on the configured listen address.
type Listener struct {
	mu           sync.RWMutex
	conn         *net.UDPConn
	listenAddr   string
	lastMute     *bool
	lastPacket   time.Time
	listenError  string
	onMuteChange func(muted bool)
}

// NewListener creates an OSC UDP listener.
func NewListener() *Listener {
	return &Listener{}
}

// SetOnMuteChange registers a callback when MuteSelf changes.
func (l *Listener) SetOnMuteChange(fn func(muted bool)) {
	l.mu.Lock()
	l.onMuteChange = fn
	l.mu.Unlock()
}

// Start binds the UDP socket and reads packets until ctx is cancelled.
func (l *Listener) Start(ctx context.Context, listenAddr string) error {
	l.Stop()
	l.mu.Lock()
	l.listenAddr = listenAddr
	l.lastMute = nil
	l.lastPacket = time.Time{}
	l.listenError = ""
	l.mu.Unlock()

	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		l.setListenError(err.Error())
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		l.setListenError(err.Error())
		return err
	}

	l.mu.Lock()
	l.conn = conn
	if la, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		l.listenAddr = la.String()
	}
	l.mu.Unlock()

	go l.readLoop(ctx, conn)
	return nil
}

// Stop closes the UDP listener.
func (l *Listener) Stop() {
	l.mu.Lock()
	conn := l.conn
	l.conn = nil
	l.mu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}

// Snapshot returns the latest received MuteSelf state.
func (l *Listener) Snapshot() Snapshot {
	l.mu.RLock()
	defer l.mu.RUnlock()
	s := Snapshot{
		ListenAddr:  l.listenAddr,
		ListenError: l.listenError,
		Listening:   l.conn != nil,
	}
	if l.lastMute != nil {
		v := *l.lastMute
		s.MuteKnown = true
		s.Muted = v
	}
	if !l.lastPacket.IsZero() {
		s.LastPacketAt = l.lastPacket
		s.Connected = time.Since(l.lastPacket) < 60*time.Second
	}
	return s
}

// Snapshot is a point-in-time view of OSC listener state.
type Snapshot struct {
	ListenAddr   string
	ListenError  string
	Listening    bool
	Connected    bool
	MuteKnown    bool
	Muted        bool
	LastPacketAt time.Time
}

func (l *Listener) setListenError(msg string) {
	l.mu.Lock()
	l.listenError = msg
	l.mu.Unlock()
}

func (l *Listener) readLoop(ctx context.Context, conn *net.UDPConn) {
	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			l.setListenError(err.Error())
			return
		}
		if muted, ok := ParseMuteSelf(buf[:n]); ok {
			var cb func(bool)
			l.mu.Lock()
			changed := l.lastMute == nil || *l.lastMute != muted
			v := muted
			l.lastMute = &v
			l.lastPacket = time.Now()
			cb = l.onMuteChange
			l.mu.Unlock()
			if changed && cb != nil {
				cb(muted)
			}
		}
	}
}
