package media

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EmbedEnrichmentMetadata writes instance/participant enrichment into an image file atomically.
func EmbedEnrichmentMetadata(path string, fields EnrichmentFields) error {
	path = filepath.Clean(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	ext := strings.ToLower(filepath.Ext(path))
	var updated []byte
	switch ext {
	case ".jpg", ".jpeg":
		updated, err = embedEnrichmentJPEG(data, fields)
	case ".png":
		updated, err = embedEnrichmentPNG(data, fields)
	default:
		return fmt.Errorf("unsupported image extension: %s", ext)
	}
	if err != nil {
		return err
	}
	return atomicReplaceFile(path, updated)
}

func embedEnrichmentJPEG(data []byte, fields EnrichmentFields) ([]byte, error) {
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return nil, fmt.Errorf("invalid JPEG")
	}
	xmp := extractXMPFromJPEG(data)
	merged := MergeEnrichmentIntoXMP(xmp, fields)
	if merged == xmp && xmp != "" {
		return data, nil
	}
	payload := append([]byte(jpegXMPNamespacePrefix), []byte(merged)...)
	return replaceOrInsertJPEGAPP1(data, payload), nil
}

func embedEnrichmentPNG(data []byte, fields EnrichmentFields) ([]byte, error) {
	const pngSignature = "\x89PNG\r\n\x1a\n"
	if len(data) < len(pngSignature)+12 || !bytes.Equal(data[:8], []byte(pngSignature)) {
		return nil, fmt.Errorf("invalid PNG")
	}
	xmp := extractXMPFromPNG(data)
	merged := MergeEnrichmentIntoXMP(xmp, fields)
	if merged == xmp && xmp != "" {
		return data, nil
	}
	return replaceOrInsertPNGXMP(data, merged), nil
}

func replaceOrInsertJPEGAPP1(data, payload []byte) []byte {
	segLen := 2 + len(payload)
	segment := []byte{0xFF, 0xE1, byte(segLen >> 8), byte(segLen & 0xff)}
	segment = append(segment, payload...)

	pos := 2
	insertAt := 2
	replaced := false
	var out []byte
	out = append(out, data[0:2]...)
	for pos+4 <= len(data) {
		if data[pos] != 0xFF {
			break
		}
		marker := data[pos+1]
		if marker == 0xD9 || marker == 0xDA {
			break
		}
		segLenField := int(binary.BigEndian.Uint16(data[pos+2 : pos+4]))
		if segLenField < 2 {
			break
		}
		segStart := pos
		segEnd := pos + 2 + segLenField
		if segEnd > len(data) {
			break
		}
		if marker == 0xE1 {
			chunk := data[pos+4 : segEnd]
			if bytes.HasPrefix(chunk, []byte(jpegXMPNamespacePrefix)) || bytes.Contains(chunk, []byte("<x:xmpmeta")) {
				if !replaced {
					out = append(out, segment...)
					replaced = true
				}
				pos = segEnd
				continue
			}
		}
		out = append(out, data[segStart:segEnd]...)
		pos = segEnd
	}
	if !replaced {
		newOut := make([]byte, 0, len(out)+len(segment)+len(data)-insertAt)
		newOut = append(newOut, out...)
		newOut = append(newOut, segment...)
		newOut = append(newOut, data[insertAt:]...)
		return newOut
	}
	out = append(out, data[pos:]...)
	return out
}

func replaceOrInsertPNGXMP(data []byte, xmp string) []byte {
	pos := 8
	var out []byte
	out = append(out, data[:8]...)
	replaced := false
	for pos+12 <= len(data) {
		length := int(binary.BigEndian.Uint32(data[pos : pos+4]))
		chunkType := string(data[pos+4 : pos+8])
		chunkStart := pos
		chunkEnd := pos + 12 + length
		if chunkEnd > len(data) {
			break
		}
		if chunkType == "iTXt" {
			kw, text := parseITXTKeywordAndText(data[pos+8 : pos+8+length])
			if kw == pngXMPKeyword || strings.Contains(text, "<x:xmpmeta") {
				if !replaced {
					out = append(out, buildPNGITXtChunk(pngXMPKeyword, xmp)...)
					replaced = true
				}
				pos = chunkEnd
				continue
			}
		}
		out = append(out, data[chunkStart:chunkEnd]...)
		if chunkType == "IEND" && !replaced {
			out = append(out[:len(out)-(12+length)], buildPNGITXtChunk(pngXMPKeyword, xmp)...)
			out = append(out, data[chunkStart:chunkEnd]...)
			replaced = true
		}
		pos = chunkEnd
	}
	if !replaced {
		return appendPNGDataChunk(data[:len(data)-12], "iTXt", buildPNGITXtPayload(pngXMPKeyword, xmp))
	}
	return out
}

func buildPNGITXtChunk(keyword, text string) []byte {
	return appendPNGDataChunk(nil, "iTXt", buildPNGITXtPayload(keyword, text))
}

func buildPNGITXtPayload(keyword, text string) []byte {
	payload := append([]byte(keyword), 0, 0)
	payload = append(payload, []byte("en")...)
	payload = append(payload, 0, 0)
	payload = append(payload, []byte(text)...)
	return payload
}

func appendPNGDataChunk(data []byte, typ string, payload []byte) []byte {
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(payload)))
	data = append(data, lenBuf[:]...)
	data = append(data, []byte(typ)...)
	data = append(data, payload...)
	data = append(data, 0, 0, 0, 0)
	return data
}

func atomicReplaceFile(path string, data []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".screenshot-enrich-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(info.Mode()); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// ReadEnrichmentFieldsFromFile extracts enrichment fields from an on-disk screenshot.
func ReadEnrichmentFieldsFromFile(path string) (EnrichmentFields, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return EnrichmentFields{}, err
	}
	ext := strings.ToLower(filepath.Ext(path))
	var xmp string
	switch ext {
	case ".jpg", ".jpeg":
		xmp = extractXMPFromJPEG(data)
	case ".png":
		xmp = extractXMPFromPNG(data)
	}
	return ParseEnrichmentXMP(xmp), nil
}

// HasMetadataTakenAt reports whether the file has XMP or JPEG EXIF capture time.
func HasMetadataTakenAt(path string) (bool, error) {
	meta, err := Extract(path)
	if err != nil {
		return false, err
	}
	return meta.TakenAt != nil, nil
}
