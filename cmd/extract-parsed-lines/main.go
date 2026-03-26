package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

func main() {
	inPath := flag.String("in", "", "input VRChat output_log.txt path")
	outPath := flag.String("out", "", "output path (matched lines only)")
	flag.Parse()

	if *inPath == "" || *outPath == "" {
		_, _ = fmt.Fprintln(os.Stderr, "usage: extract-parsed-lines -in <input> -out <output>")
		os.Exit(2)
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create output dir: %v\n", err)
		os.Exit(1)
	}

	in, err := os.Open(*inPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to open input: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(*outPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create output: %v\n", err)
		os.Exit(1)
	}

	parser := activity.NewLogParser()

	// output_log.txt lines are typically timestamped; keep a moving fallback so
	// contiguous lines without timestamps still get a reasonable OccurredAt.
	lastTime := time.Time{}

	sc := bufio.NewScanner(in)
	bw := bufio.NewWriter(out)
	defer func() {
		_ = bw.Flush()
		_ = out.Close()
	}()

	for sc.Scan() {
		raw := sc.Text()
		at := activity.ParseVRChatTimestamp(raw, lastTime)
		if !at.IsZero() {
			lastTime = at
		}

		events, _ := parser.ParseLine(raw, at)
		if len(events) == 0 {
			continue
		}

		_, _ = bw.WriteString(raw)
		_, _ = bw.WriteString("\n")
	}
	if err := sc.Err(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed while scanning input: %v\n", err)
		os.Exit(1)
	}
}
