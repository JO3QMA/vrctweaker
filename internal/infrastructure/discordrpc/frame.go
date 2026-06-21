package discordrpc

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

const (
	opHandshake = 0
	opFrame     = 1
	opClose     = 2
)

// writeFrame writes a Discord IPC frame.
func writeFrame(w io.Writer, opcode int32, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header[0:4], uint32(opcode))
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(body)))
	if _, err := w.Write(header); err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

// readFrame reads a Discord IPC frame.
func readFrame(r io.Reader) (opcode int32, payload json.RawMessage, err error) {
	header := make([]byte, 8)
	if _, err = io.ReadFull(r, header); err != nil {
		return 0, nil, err
	}
	opcode = int32(binary.LittleEndian.Uint32(header[0:4]))
	length := binary.LittleEndian.Uint32(header[4:8])
	if length == 0 {
		return opcode, nil, nil
	}
	body := make([]byte, length)
	if _, err = io.ReadFull(r, body); err != nil {
		return 0, nil, err
	}
	return opcode, json.RawMessage(body), nil
}

type frameMessage struct {
	Cmd   string          `json:"cmd"`
	Evt   string          `json:"evt"`
	Nonce string          `json:"nonce"`
	Args  json.RawMessage `json:"args"`
	Data  json.RawMessage `json:"data"`
}

func parseVoiceSettingsMute(data json.RawMessage) (bool, bool) {
	if len(data) == 0 {
		return false, false
	}
	var v struct {
		Mute bool `json:"mute"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return false, false
	}
	return v.Mute, true
}

func commandPayload(cmd string, args any, nonce string) map[string]any {
	return map[string]any{
		"cmd":   cmd,
		"args":  args,
		"nonce": nonce,
	}
}

func fmtNonce() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
