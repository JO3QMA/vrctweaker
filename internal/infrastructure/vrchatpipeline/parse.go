package vrchatpipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// UnwrapContent returns the inner JSON payload for a Pipeline message "content" field.
// Most events double-encode objects as JSON strings; some use a raw object or a plain string ID.
func UnwrapContent(content json.RawMessage) (json.RawMessage, error) {
	if len(bytes.TrimSpace(content)) == 0 {
		return nil, nil
	}
	c := bytes.TrimSpace(content)
	if c[0] == '"' {
		var s string
		if err := json.Unmarshal(content, &s); err != nil {
			return nil, fmt.Errorf("pipeline content string: %w", err)
		}
		if len(s) > 0 && (s[0] == '{' || s[0] == '[') {
			return []byte(s), nil
		}
		out, err := json.Marshal(s)
		if err != nil {
			return nil, err
		}
		return out, nil
	}
	return content, nil
}

// ParseEnvelope extracts event type and unwrapped content from a raw WebSocket frame.
func ParseEnvelope(frame []byte) (eventType string, payload json.RawMessage, err error) {
	var env struct {
		Type    string          `json:"type"`
		Content json.RawMessage `json:"content"`
	}
	if uerr := json.Unmarshal(frame, &env); uerr != nil {
		return "", nil, fmt.Errorf("pipeline envelope: %w", uerr)
	}
	inner, werr := UnwrapContent(env.Content)
	if werr != nil {
		return "", nil, werr
	}
	return env.Type, inner, nil
}
