package vrchatosc

import (
	"fmt"
	"net"

	"vrchat-tweaker/internal/domain/micmutesync"
)

// Sender transmits OSC commands to VRChat.
type Sender struct{}

// NewSender creates an OSC sender.
func NewSender() *Sender {
	return &Sender{}
}

// ToggleVoice sends the /input/Voice pulse (Toggle Voice ON) to VRChat.
func (s *Sender) ToggleVoice(endpoint string) error {
	ep, err := micmutesync.ParseEndpoint(endpoint)
	if err != nil {
		return err
	}
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ep.OutHost, ep.InPort))
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	for _, pkt := range []struct {
		tag string
		arg *int32
	}{
		{"T", nil},
		{"F", nil},
	} {
		if _, err := conn.Write(buildOSCPacket("/input/Voice", pkt.tag, pkt.arg)); err != nil {
			return err
		}
	}
	return nil
}

// SetMute toggles VRChat mic until it matches desired when current is known.
func (s *Sender) SetMute(endpoint string, currentKnown, current, desired bool) error {
	if !micmutesync.NeedsMuteToggle(currentKnown, current, desired) {
		return nil
	}
	return s.ToggleVoice(endpoint)
}
