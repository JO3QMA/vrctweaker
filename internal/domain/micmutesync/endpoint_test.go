package micmutesync

import "testing"

func TestParseEndpoint_default(t *testing.T) {
	ep, err := ParseEndpoint("")
	if err != nil {
		t.Fatal(err)
	}
	if ep.InPort != 9000 || ep.OutHost != "127.0.0.1" || ep.OutPort != 9001 {
		t.Fatalf("default endpoint: %+v", ep)
	}
	if ep.ListenAddr() != "127.0.0.1:9001" {
		t.Fatalf("ListenAddr: %s", ep.ListenAddr())
	}
}

func TestParseEndpoint_custom(t *testing.T) {
	ep, err := ParseEndpoint("9100:127.0.0.1:9101")
	if err != nil {
		t.Fatal(err)
	}
	if ep.InPort != 9100 || ep.OutPort != 9101 {
		t.Fatalf("custom endpoint: %+v", ep)
	}
}

func TestParseEndpoint_invalid(t *testing.T) {
	if _, err := ParseEndpoint("bad"); err == nil {
		t.Fatal("expected error")
	}
}

func TestPlatformAvailable(t *testing.T) {
	if !PlatformAvailable("windows") {
		t.Fatal("windows should be available")
	}
	if PlatformAvailable("linux") {
		t.Fatal("linux should not be available in v1")
	}
}
