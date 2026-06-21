package vrchatosc

import "testing"

func TestParseMuteSelf_trueFalse(t *testing.T) {
	pkt := buildOSCPacket("/avatar/parameters/MuteSelf", "T", nil)
	muted, ok := ParseMuteSelf(pkt)
	if !ok || !muted {
		t.Fatalf("true: ok=%v muted=%v", ok, muted)
	}
	pkt = buildOSCPacket("/avatar/parameters/MuteSelf", "F", nil)
	muted, ok = ParseMuteSelf(pkt)
	if !ok || muted {
		t.Fatalf("false: ok=%v muted=%v", ok, muted)
	}
}

func TestParseMuteSelf_int(t *testing.T) {
	one := int32(1)
	zero := int32(0)
	muted, ok := ParseMuteSelf(buildOSCPacket("/avatar/parameters/MuteSelf", "i", &one))
	if !ok || !muted {
		t.Fatalf("int 1: ok=%v muted=%v", ok, muted)
	}
	muted, ok = ParseMuteSelf(buildOSCPacket("/avatar/parameters/MuteSelf", "i", &zero))
	if !ok || muted {
		t.Fatalf("int 0: ok=%v muted=%v", ok, muted)
	}
}

func TestParseMuteSelf_otherAddress(t *testing.T) {
	pkt := buildOSCPacket("/avatar/parameters/Other", "T", nil)
	if _, ok := ParseMuteSelf(pkt); ok {
		t.Fatal("expected false for other address")
	}
}
