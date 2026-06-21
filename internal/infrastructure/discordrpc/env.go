package discordrpc

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const (
	envDiscordClientID     = "VRCTWEAKER_DISCORD_CLIENT_ID"
	envDiscordClientSecret = "VRCTWEAKER_DISCORD_CLIENT_SECRET"
)

func loadDiscordEnvFromFiles() {
	for _, path := range discordEnvFilePaths() {
		loadEnvFile(path)
	}
}

func discordEnvFilePaths() []string {
	var paths []string
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths,
			filepath.Join(cwd, ".env"),
			filepath.Join(cwd, ".envrc"),
		)
	}
	if exe, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), ".env"))
	}
	if dir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(dir, "vrchat-tweaker", "discord.env"))
	}
	return paths
}

func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		key, value, ok := parseEnvLine(scanner.Text())
		if !ok || !isDiscordEnvKey(key) {
			continue
		}
		if os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, value)
	}
}

func isDiscordEnvKey(key string) bool {
	return key == envDiscordClientID || key == envDiscordClientSecret
}

func parseEnvLine(line string) (key, value string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	if strings.HasPrefix(line, "export ") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
	}
	eq := strings.IndexByte(line, '=')
	if eq <= 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:eq])
	value = strings.TrimSpace(line[eq+1:])
	value = strings.Trim(value, `"'`)
	if key == "" {
		return "", "", false
	}
	return key, value, true
}

func clientIDConfigured() bool {
	return ClientID != "" && ClientID != "000000000000000000"
}
