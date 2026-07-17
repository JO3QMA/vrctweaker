package logwatcher

import "testing"

func TestNop(t *testing.T) {
	Nop("ignored %d", 1)
}

func TestStd(t *testing.T) {
	Std()("ok %s", "test")
}
