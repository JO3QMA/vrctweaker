//go:build windows

package appupdate

import "testing"

func TestUtf8BOM_forPowerShell51(t *testing.T) {
	if len(utf8BOM) != 3 || utf8BOM[0] != 0xEF || utf8BOM[1] != 0xBB || utf8BOM[2] != 0xBF {
		t.Fatalf("utf8BOM=%v", utf8BOM)
	}
}
