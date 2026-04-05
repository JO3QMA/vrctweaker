package vrchatapi

import (
	"sync"
	"testing"
)

// TestClient_SetAuthToken_ConcurrentRace verifies that concurrent reads and writes
// of authToken via SetAuthToken do not cause data races (go test -race).
func TestClient_SetAuthToken_ConcurrentRace(t *testing.T) {
	c := NewClient("")
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(2)
		go func(tok string) {
			defer wg.Done()
			c.SetAuthToken(tok)
		}("token-" + string(rune('A'+i%26)))
		go func() {
			defer wg.Done()
			_ = c.GetAuthToken()
		}()
	}
	wg.Wait()
}
