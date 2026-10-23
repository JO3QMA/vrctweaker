//go:build windows

package tray

import (
	"fmt"
	"testing"
	"time"
)

func TestFinishStop_unblocksStopWaiterOnStartFailure(t *testing.T) {
	m := &windowsManager{}
	stopCh := m.stopWaiters.add()
	m.stopping = true

	m.finishStop(fmt.Errorf("start failed"))

	select {
	case err := <-stopCh:
		if err == nil {
			t.Fatal("expected error")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stop waiter not completed")
	}
}
