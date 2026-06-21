package vrchatpipeline

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestUnwrapContent_rawObject(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{"userId":"usr_1","location":"wrld_x:1"}`)
	out, err := UnwrapContent(raw)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(raw) {
		t.Fatalf("got %s", out)
	}
}

func TestUnwrapContent_empty(t *testing.T) {
	t.Parallel()
	out, err := UnwrapContent(json.RawMessage("   "))
	if err != nil {
		t.Fatal(err)
	}
	if out != nil {
		t.Fatalf("got %q, want nil", out)
	}
}

func TestUnwrapContent_doubleEncodedObject(t *testing.T) {
	t.Parallel()
	inner := `{"userId":"usr_1"}`
	quoted, err := json.Marshal(inner)
	if err != nil {
		t.Fatal(err)
	}
	out, err := UnwrapContent(quoted)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != inner {
		t.Fatalf("got %s want %s", out, inner)
	}
}

func TestUnwrapContent_doubleEncodedArray(t *testing.T) {
	t.Parallel()
	inner := `[1,2,3]`
	quoted, err := json.Marshal(inner)
	if err != nil {
		t.Fatal(err)
	}
	out, err := UnwrapContent(quoted)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != inner {
		t.Fatalf("got %s want %s", out, inner)
	}
}

func TestUnwrapContent_plainStringID(t *testing.T) {
	t.Parallel()
	quoted, err := json.Marshal("not_00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatal(err)
	}
	out, err := UnwrapContent(quoted)
	if err != nil {
		t.Fatal(err)
	}
	var s string
	if err := json.Unmarshal(out, &s); err != nil {
		t.Fatal(err)
	}
	if s != "not_00000000-0000-0000-0000-000000000000" {
		t.Fatalf("got %q", s)
	}
}

func TestUnwrapContent_invalidStringJSON(t *testing.T) {
	t.Parallel()
	_, err := UnwrapContent(json.RawMessage(`"unclosed`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseEnvelope(t *testing.T) {
	t.Parallel()
	frame, err := json.Marshal(map[string]string{
		"type":    "friend-offline",
		"content": `{"userId":"usr_abc"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	typ, payload, err := ParseEnvelope(frame)
	if err != nil {
		t.Fatal(err)
	}
	if typ != "friend-offline" {
		t.Fatalf("type %q", typ)
	}
	var body struct {
		UserID string `json:"userId"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatal(err)
	}
	if body.UserID != "usr_abc" {
		t.Fatalf("userId %q", body.UserID)
	}
}

func TestParseEnvelope_invalidJSON(t *testing.T) {
	t.Parallel()
	_, _, err := ParseEnvelope([]byte(`{broken`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseEnvelope_unwrapError(t *testing.T) {
	t.Parallel()
	frame := []byte(`{"type":"x","content":"bad\x"}`)
	_, _, err := ParseEnvelope(frame)
	if err == nil {
		t.Fatal("expected unwrap error")
	}
}

func TestNextBackoff(t *testing.T) {
	t.Parallel()
	const max = 60 * time.Second
	if got := nextBackoff(time.Second, max); got != 2*time.Second {
		t.Fatalf("got %v", got)
	}
	if got := nextBackoff(40*time.Second, max); got != max {
		t.Fatalf("got %v, want max", got)
	}
}

func TestSleep_completes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	if err := sleep(ctx, time.Millisecond); err != nil {
		t.Fatalf("sleep: %v", err)
	}
}

func TestSleep_cancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sleep(ctx, time.Second); err != context.Canceled {
		t.Fatalf("sleep: %v", err)
	}
}

func TestRun_validationErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	handler := func(context.Context, string, []byte) error { return nil }

	err := Run(ctx, Config{UserAgent: "ua", OnEvent: handler})
	if err == nil || !strings.Contains(err.Error(), "auth token") {
		t.Fatalf("Run empty auth: %v", err)
	}

	err = Run(ctx, Config{AuthToken: "t", OnEvent: handler})
	if err == nil || !strings.Contains(err.Error(), "user agent") {
		t.Fatalf("Run empty UA: %v", err)
	}

	err = Run(ctx, Config{AuthToken: "t", UserAgent: "ua"})
	if err == nil || !strings.Contains(err.Error(), "OnEvent") {
		t.Fatalf("Run nil OnEvent: %v", err)
	}
}

func TestUnwrapContent_plainStringNonJSON(t *testing.T) {
	t.Parallel()
	quoted, err := json.Marshal("hello")
	if err != nil {
		t.Fatal(err)
	}
	out, err := UnwrapContent(quoted)
	if err != nil {
		t.Fatal(err)
	}
	var s string
	if err := json.Unmarshal(out, &s); err != nil {
		t.Fatal(err)
	}
	if s != "hello" {
		t.Fatalf("got %q", s)
	}
}

func TestParseEnvelope_emptyContent(t *testing.T) {
	t.Parallel()
	frame, err := json.Marshal(map[string]any{
		"type":    "ping",
		"content": nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	typ, payload, err := ParseEnvelope(frame)
	if err != nil {
		t.Fatal(err)
	}
	if typ != "ping" {
		t.Fatalf("type %q", typ)
	}
	if len(payload) == 0 {
		t.Fatalf("payload empty")
	}
}

func TestRun_onReconnectSuccess(t *testing.T) {
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		<-r.Context().Done()
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	oldHost, oldScheme := pipelineDialHost, pipelineDialScheme
	pipelineDialHost = host
	pipelineDialScheme = "ws"
	t.Cleanup(func() {
		pipelineDialHost = oldHost
		pipelineDialScheme = oldScheme
	})

	called := false
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	err := Run(ctx, Config{
		AuthToken: "token",
		UserAgent: "ua",
		OnReconnect: func(context.Context) error {
			called = true
			return nil
		},
		OnEvent: func(context.Context, string, []byte) error { return nil },
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run: %v", err)
	}
	if !called {
		t.Fatal("OnReconnect not called")
	}
}

func TestRun_onReconnectError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	want := errors.New("snapshot failed")
	err := Run(ctx, Config{
		AuthToken: "t",
		UserAgent: "ua",
		OnReconnect: func(context.Context) error {
			return want
		},
		OnEvent: func(context.Context, string, []byte) error { return nil },
	})
	if !errors.Is(err, want) {
		t.Fatalf("Run: %v", err)
	}
}

func TestRun_cancelledBeforeDial(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Run(ctx, Config{
		AuthToken: "t",
		UserAgent: "ua",
		OnEvent:   func(context.Context, string, []byte) error { return nil },
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run: %v", err)
	}
}

func TestRun_dialFailureBackoffUntilCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	err := Run(ctx, Config{
		AuthToken: "invalid-token-for-test",
		UserAgent: "vrctweaker-test",
		OnEvent:   func(context.Context, string, []byte) error { return nil },
	})
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("Run: %v", err)
	}
}

func TestReadLoop_dispatchesEvents(t *testing.T) {
	upgrader := websocket.Upgrader{}
	connReady := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		connReady <- conn
		frame, merr := json.Marshal(map[string]string{
			"type":    "friend-online",
			"content": `{"userId":"usr_1"}`,
		})
		if merr != nil {
			t.Errorf("marshal: %v", merr)
			return
		}
		if werr := conn.WriteMessage(websocket.TextMessage, frame); werr != nil {
			t.Errorf("write: %v", werr)
		}
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`not-json`))
		emptyType, _ := json.Marshal(map[string]string{"type": "", "content": "{}"})
		_ = conn.WriteMessage(websocket.TextMessage, emptyType)
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = clientConn.Close() }()

	select {
	case <-connReady:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for server connection")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	var gotType string
	done := make(chan struct{})
	handler := func(_ context.Context, eventType string, payload []byte) error {
		mu.Lock()
		gotType = eventType
		mu.Unlock()
		close(done)
		cancel()
		return nil
	}

	readErr := readLoop(ctx, clientConn, handler)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event")
	}
	mu.Lock()
	typ := gotType
	mu.Unlock()
	if typ != "friend-online" {
		t.Fatalf("event type %q", typ)
	}
	if readErr != nil && !errors.Is(readErr, context.Canceled) {
		t.Fatalf("readLoop: %v", readErr)
	}
}

func TestReadLoop_onEventErrorContinues(t *testing.T) {
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		frame, _ := json.Marshal(map[string]string{"type": "ping", "content": "{}"})
		_ = conn.WriteMessage(websocket.TextMessage, frame)
		time.Sleep(50 * time.Millisecond)
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	ctx := context.Background()
	calls := 0
	handler := func(context.Context, string, []byte) error {
		calls++
		return errors.New("handler failed")
	}
	_ = readLoop(ctx, clientConn, handler)
	if calls != 1 {
		t.Fatalf("handler calls = %d, want 1", calls)
	}
}

func TestRun_reconnectsAfterDisconnect(t *testing.T) {
	upgrader := websocket.Upgrader{}
	connects := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connects++
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		_ = conn.Close()
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	oldHost, oldScheme := pipelineDialHost, pipelineDialScheme
	pipelineDialHost = host
	pipelineDialScheme = "ws"
	t.Cleanup(func() {
		pipelineDialHost = oldHost
		pipelineDialScheme = oldScheme
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer cancel()
	_ = Run(ctx, Config{
		AuthToken: "token",
		UserAgent: "ua",
		OnEvent:   func(context.Context, string, []byte) error { return nil },
	})
	if connects < 2 {
		t.Fatalf("connects = %d, want at least 2", connects)
	}
}

func TestReadLoop_cancelledContext(t *testing.T) {
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		<-r.Context().Done()
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = readLoop(ctx, clientConn, func(context.Context, string, []byte) error { return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("readLoop: %v", err)
	}
}

func TestDialPipeline_localServer(t *testing.T) {
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		<-r.Context().Done()
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	oldHost, oldScheme := pipelineDialHost, pipelineDialScheme
	pipelineDialHost = host
	pipelineDialScheme = "ws"
	t.Cleanup(func() {
		pipelineDialHost = oldHost
		pipelineDialScheme = oldScheme
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, err := dialPipeline(ctx, websocket.Dialer{HandshakeTimeout: time.Second}, "token", "ua")
	if err != nil {
		t.Fatalf("dialPipeline: %v", err)
	}
	_ = conn.Close()
}

func TestRun_readsFromLocalPipeline(t *testing.T) {
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		frame, _ := json.Marshal(map[string]string{
			"type":    "notification",
			"content": `{"id":"n1"}`,
		})
		_ = conn.WriteMessage(websocket.TextMessage, frame)
		time.Sleep(100 * time.Millisecond)
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	oldHost, oldScheme := pipelineDialHost, pipelineDialScheme
	pipelineDialHost = host
	pipelineDialScheme = "ws"
	t.Cleanup(func() {
		pipelineDialHost = oldHost
		pipelineDialScheme = oldScheme
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan string, 1)
	go func() {
		err := Run(ctx, Config{
			AuthToken: "token",
			UserAgent: "ua",
			OnEvent: func(_ context.Context, eventType string, _ []byte) error {
				done <- eventType
				cancel()
				return nil
			},
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("Run: %v", err)
		}
	}()

	select {
	case typ := <-done:
		if typ != "notification" {
			t.Fatalf("event %q", typ)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for pipeline event")
	}
}

func TestDialPipeline_invalidHost(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := dialPipeline(ctx, websocket.Dialer{HandshakeTimeout: 50 * time.Millisecond}, "token", "ua")
	if err == nil {
		t.Fatal("expected dial error")
	}
}
