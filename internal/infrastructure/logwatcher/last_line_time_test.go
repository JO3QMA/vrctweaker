package logwatcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLastVRChatLineTimeInFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	content := []byte(
		"noise without timestamp\n" +
			"2026.06.24 08:26:40 Debug      -  [Behaviour] Joining wrld_abc:1~region(jp)\n" +
			"2026.06.24 08:26:50 Debug      -  [Behaviour] OnPlayerJoined Alice (usr_x)\n",
	)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}

	got, err := LastVRChatLineTimeInFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want, err := time.ParseInLocation("2006.01.02 15:04:05", "2026.06.24 08:26:50", time.Local)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatalf("LastVRChatLineTimeInFile = %v, want %v", got, want)
	}
}

func TestLastVRChatLineTimeInFile_emptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output_log.txt")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	got, err := LastVRChatLineTimeInFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsZero() {
		t.Fatalf("got %v, want zero time", got)
	}
}
