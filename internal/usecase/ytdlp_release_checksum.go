package usecase

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const ytdlpSHA256SumsAssetName = "SHA2-256SUMS"

func ytdlpReleaseAssetsFromReleaseJSON(data []byte) (tagName, exeDownloadURL, sumsDownloadURL string, err error) {
	var rel githubReleaseLatest
	if err := json.Unmarshal(data, &rel); err != nil {
		return "", "", "", fmt.Errorf("parse release JSON: %w", err)
	}
	tag := strings.TrimSpace(rel.TagName)
	if tag == "" {
		return "", "", "", errors.New("release has empty tag_name")
	}
	for _, a := range rel.Assets {
		switch a.Name {
		case ytdlpExeName:
			if strings.TrimSpace(a.BrowserDownloadURL) != "" {
				exeDownloadURL = a.BrowserDownloadURL
			}
		case ytdlpSHA256SumsAssetName:
			if strings.TrimSpace(a.BrowserDownloadURL) != "" {
				sumsDownloadURL = a.BrowserDownloadURL
			}
		}
	}
	if exeDownloadURL == "" {
		return "", "", "", fmt.Errorf("asset %q not found in release", ytdlpExeName)
	}
	if sumsDownloadURL == "" {
		return "", "", "", fmt.Errorf("asset %q not found in release", ytdlpSHA256SumsAssetName)
	}
	return tag, exeDownloadURL, sumsDownloadURL, nil
}

func parseYTDlpSHA256SumForFile(sumsBody []byte, fileName string) (string, error) {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return "", errors.New("file name is empty")
	}
	sc := bufio.NewScanner(bytes.NewReader(sumsBody))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		hash, name, ok := parseGNUChecksumLine(line)
		if !ok {
			continue
		}
		if name == fileName {
			return hash, nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("read checksums: %w", err)
	}
	return "", fmt.Errorf("checksum for %q not found in SHA2-256SUMS", fileName)
}

func parseGNUChecksumLine(line string) (hashHex, fileName string, ok bool) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", "", false
	}
	hashHex = strings.ToLower(strings.TrimSpace(fields[0]))
	if len(hashHex) != 64 {
		return "", "", false
	}
	if _, err := hex.DecodeString(hashHex); err != nil {
		return "", "", false
	}
	fileName = fields[len(fields)-1]
	fileName = strings.TrimPrefix(fileName, "*")
	return hashHex, fileName, true
}

func sha256HexFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
