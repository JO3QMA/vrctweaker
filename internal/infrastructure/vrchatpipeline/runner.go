package vrchatpipeline

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

const defaultPipelineHost = "pipeline.vrchat.cloud"

// EventHandler receives normalized Pipeline event types and inner JSON payloads.
type EventHandler func(ctx context.Context, eventType string, payload []byte) error

// Config drives a reconnecting Pipeline client.
type Config struct {
	// AuthToken is the VRChat auth cookie value (same as REST Cookie "auth").
	AuthToken string
	UserAgent string
	// OnReconnect runs before each WebSocket dial (e.g. REST snapshot). If it returns a
	// non-nil error, Run exits and the connection is not attempted.
	OnReconnect func(ctx context.Context) error
	// OnEvent handles decoded events; return nil to continue.
	OnEvent EventHandler
}

// Run connects to the VRChat Pipeline with backoff until ctx is cancelled.
func Run(ctx context.Context, cfg Config) error {
	if cfg.AuthToken == "" {
		return fmt.Errorf("vrchat pipeline: empty auth token")
	}
	if cfg.UserAgent == "" {
		return fmt.Errorf("vrchat pipeline: empty user agent")
	}
	if cfg.OnEvent == nil {
		return fmt.Errorf("vrchat pipeline: nil OnEvent")
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 30 * time.Second,
	}
	backoff := time.Second
	const maxBackoff = 60 * time.Second

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if cfg.OnReconnect != nil {
			if err := cfg.OnReconnect(ctx); err != nil {
				return err
			}
		}

		conn, err := dialPipeline(ctx, dialer, cfg.AuthToken, cfg.UserAgent)
		if err != nil {
			log.Printf("vrchat pipeline: dial: %v", err)
			if sleep(ctx, backoff) != nil {
				return ctx.Err()
			}
			backoff = nextBackoff(backoff, maxBackoff)
			continue
		}
		backoff = time.Second

		readErr := readLoop(ctx, conn, cfg.OnEvent)
		_ = conn.Close()
		if readErr != nil && ctx.Err() != nil {
			return ctx.Err()
		}
		if readErr != nil {
			log.Printf("vrchat pipeline: read: %v", readErr)
		}
		if sleep(ctx, backoff) != nil {
			return ctx.Err()
		}
		backoff = nextBackoff(backoff, maxBackoff)
	}
}

func dialPipeline(ctx context.Context, dialer websocket.Dialer, token, userAgent string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "wss", Host: defaultPipelineHost, Path: "/"}
	q := u.Query()
	q.Set("authToken", token)
	u.RawQuery = q.Encode()

	header := http.Header{}
	header.Set("User-Agent", userAgent)

	dialCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	conn, _, err := dialer.DialContext(dialCtx, u.String(), header)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func readLoop(ctx context.Context, conn *websocket.Conn, onEvent EventHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		typ, payload, perr := ParseEnvelope(data)
		if perr != nil {
			log.Printf("vrchat pipeline: parse: %v", perr)
			continue
		}
		if typ == "" {
			continue
		}
		if err := onEvent(ctx, typ, payload); err != nil {
			log.Printf("vrchat pipeline: event %q: %v", typ, err)
		}
	}
}

func nextBackoff(cur, max time.Duration) time.Duration {
	n := cur * 2
	if n > max {
		return max
	}
	return n
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
