//go:build windows

package discordrpc

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Microsoft/go-winio"
)

type platformVoiceClient struct {
	store TokenStore

	mu           sync.RWMutex
	conn         net.Conn
	waiter       *frameWaiter
	connected    bool
	authorized   bool
	muteKnown    bool
	muted        bool
	err          string
	onMuteChange func(bool)
	readDone     chan struct{}
}

func newPlatformVoiceClient(store TokenStore) VoiceClient {
	return &platformVoiceClient{store: store, readDone: make(chan struct{})}
}

func (c *platformVoiceClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeLocked()
	if !clientIDConfigured() {
		c.err = "discord_client_id_missing"
		return fmt.Errorf(c.err)
	}
	conn, err := dialDiscordPipe()
	if err != nil {
		c.err = "discord_not_running"
		return err
	}
	if err := writeFrame(conn, opHandshake, map[string]any{"v": 1, "client_id": ClientID}); err != nil {
		_ = conn.Close()
		c.err = err.Error()
		return err
	}
	c.conn = conn
	c.connected = true
	c.waiter = newFrameWaiter()
	c.err = ""
	go c.readLoop(conn)
	if err := c.authenticateLocked(); err != nil {
		c.err = err.Error()
		return err
	}
	if err := c.subscribeVoiceSettingsLocked(); err != nil {
		c.err = err.Error()
		return err
	}
	if err := c.refreshVoiceSettingsLocked(); err != nil {
		c.err = err.Error()
		return err
	}
	return nil
}

func (c *platformVoiceClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeLocked()
}

func (c *platformVoiceClient) closeLocked() {
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	c.connected = false
	c.authorized = false
}

func (c *platformVoiceClient) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return Snapshot{
		Available:   true,
		Connected:   c.connected,
		Authorized:  c.authorized,
		MuteKnown:   c.muteKnown,
		Muted:       c.muted,
		Error:       c.err,
		ClientIDSet: clientIDConfigured(),
	}
}

func (c *platformVoiceClient) SetMute(muted bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.authorized || c.conn == nil {
		return fmt.Errorf("discord_not_authorized")
	}
	nonce := fmtNonce()
	c.waiter.register(nonce)
	defer c.waiter.unregister(nonce)
	if err := writeFrame(c.conn, opFrame, commandPayload("SET_VOICE_SETTINGS", map[string]any{"mute": muted}, nonce)); err != nil {
		return err
	}
	msg, err := c.waiter.wait(nonce, 5*time.Second)
	if err != nil {
		return err
	}
	if msg.Data != nil {
		if m, ok := parseVoiceSettingsMute(msg.Data); ok {
			c.muteKnown = true
			c.muted = m
		}
	} else {
		c.muteKnown = true
		c.muted = muted
	}
	return nil
}

func (c *platformVoiceClient) SetOnMuteChange(fn func(bool)) {
	c.mu.Lock()
	c.onMuteChange = fn
	c.mu.Unlock()
}

func (c *platformVoiceClient) authenticateLocked() error {
	token := ""
	if c.store != nil {
		token = c.store.GetAccessToken()
	}
	if token != "" {
		if err := c.sendAuthenticateLocked(token); err == nil {
			return nil
		}
	}
	return c.authorizeLocked()
}

func (c *platformVoiceClient) sendAuthenticateLocked(token string) error {
	nonce := fmtNonce()
	c.waiter.register(nonce)
	defer c.waiter.unregister(nonce)
	if err := writeFrame(c.conn, opFrame, commandPayload("AUTHENTICATE", map[string]any{"access_token": token}, nonce)); err != nil {
		return err
	}
	msg, err := c.waiter.wait(nonce, 8*time.Second)
	if err != nil {
		return err
	}
	if len(msg.Data) == 0 {
		return fmt.Errorf("discord_authenticate_failed")
	}
	c.authorized = true
	if c.store != nil {
		_ = c.store.SetAccessToken(token)
	}
	return nil
}

func (c *platformVoiceClient) authorizeLocked() error {
	nonce := fmtNonce()
	args := map[string]any{
		"client_id":    ClientID,
		"scopes":       []string{"rpc"},
		"redirect_uri": "http://127.0.0.1",
	}
	c.waiter.register(nonce)
	defer c.waiter.unregister(nonce)
	if err := writeFrame(c.conn, opFrame, commandPayload("AUTHORIZE", args, nonce)); err != nil {
		return err
	}
	msg, err := c.waiter.wait(nonce, 60*time.Second)
	if err != nil {
		return err
	}
	var auth struct {
		Code string `json:"code"`
	}
	if len(msg.Data) > 0 {
		_ = json.Unmarshal(msg.Data, &auth)
	}
	if auth.Code == "" {
		return fmt.Errorf("discord_authorize_denied")
	}
	token, err := exchangeCode(auth.Code)
	if err != nil {
		return err
	}
	return c.sendAuthenticateLocked(token)
}

func (c *platformVoiceClient) subscribeVoiceSettingsLocked() error {
	nonce := fmtNonce()
	args := map[string]any{"evt": "VOICE_SETTINGS_UPDATE"}
	c.waiter.register(nonce)
	defer c.waiter.unregister(nonce)
	if err := writeFrame(c.conn, opFrame, commandPayload("SUBSCRIBE", args, nonce)); err != nil {
		return err
	}
	_, err := c.waiter.wait(nonce, 5*time.Second)
	return err
}

func (c *platformVoiceClient) refreshVoiceSettingsLocked() error {
	nonce := fmtNonce()
	c.waiter.register(nonce)
	defer c.waiter.unregister(nonce)
	if err := writeFrame(c.conn, opFrame, commandPayload("GET_VOICE_SETTINGS", map[string]any{}, nonce)); err != nil {
		return err
	}
	msg, err := c.waiter.wait(nonce, 5*time.Second)
	if err != nil {
		return err
	}
	if m, ok := parseVoiceSettingsMute(msg.Data); ok {
		c.muteKnown = true
		c.muted = m
	}
	return nil
}

func (c *platformVoiceClient) readLoop(conn net.Conn) {
	for {
		_, payload, err := readFrame(conn)
		if err != nil {
			c.mu.Lock()
			c.connected = false
			c.authorized = false
			c.err = "discord_disconnected"
			c.mu.Unlock()
			return
		}
		var msg frameMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			continue
		}
		if c.waiter != nil && c.waiter.deliver(msg) {
			continue
		}
		if msg.Evt != "VOICE_SETTINGS_UPDATE" {
			continue
		}
		muted, ok := parseVoiceSettingsMute(msg.Data)
		if !ok {
			continue
		}
		var cb func(bool)
		c.mu.Lock()
		changed := !c.muteKnown || c.muted != muted
		c.muteKnown = true
		c.muted = muted
		cb = c.onMuteChange
		c.mu.Unlock()
		if changed && cb != nil {
			cb(muted)
		}
	}
}

func dialDiscordPipe() (net.Conn, error) {
	var lastErr error
	for i := 0; i < 10; i++ {
		path := fmt.Sprintf(`\\.\pipe\discord-ipc-%d`, i)
		conn, err := winio.DialPipe(path, nil)
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("discord pipe not found")
	}
	return nil, lastErr
}
