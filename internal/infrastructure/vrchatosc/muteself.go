package vrchatosc

import (
	"bytes"
	"encoding/binary"
)

const muteSelfAddress = "/avatar/parameters/MuteSelf"

// ParseMuteSelf parses an OSC packet and returns the mute state when the address is MuteSelf.
func ParseMuteSelf(packet []byte) (muted bool, ok bool) {
	addr, args, ok := parseOSCPacket(packet)
	if !ok || addr != muteSelfAddress || len(args) == 0 {
		return false, false
	}
	switch args[0].kind {
	case oscArgTrue:
		return true, true
	case oscArgFalse:
		return false, true
	case oscArgInt:
		return args[0].i != 0, true
	default:
		return false, false
	}
}

type oscArgKind int

const (
	oscArgTrue oscArgKind = iota
	oscArgFalse
	oscArgInt
)

type oscArg struct {
	kind oscArgKind
	i    int32
}

func parseOSCPacket(packet []byte) (address string, args []oscArg, ok bool) {
	if len(packet) < 4 {
		return "", nil, false
	}
	addrEnd := bytes.IndexByte(packet, 0)
	if addrEnd < 0 {
		return "", nil, false
	}
	address = string(packet[:addrEnd])
	pos := pad4(addrEnd + 1)
	if pos >= len(packet) {
		return address, nil, true
	}
	if packet[pos] != ',' {
		return address, nil, true
	}
	tagEnd := bytes.IndexByte(packet[pos:], 0)
	if tagEnd < 0 {
		return "", nil, false
	}
	tags := string(packet[pos+1 : pos+tagEnd])
	argPos := pad4(pos + tagEnd + 1)
	for _, tag := range tags {
		switch tag {
		case 'T':
			args = append(args, oscArg{kind: oscArgTrue})
		case 'F':
			args = append(args, oscArg{kind: oscArgFalse})
		case 'i':
			if argPos+4 > len(packet) {
				return "", nil, false
			}
			args = append(args, oscArg{kind: oscArgInt, i: int32(binary.BigEndian.Uint32(packet[argPos : argPos+4]))})
			argPos += 4
		default:
			return "", nil, false
		}
	}
	return address, args, true
}

func pad4(n int) int {
	return (n + 3) &^ 3
}
