package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"vrchat-tweaker/internal/packaging"
)

const exitNoChangelogSection = 2

func main() {
	changelogPath := flag.String("changelog", "CHANGELOG.md", "path to CHANGELOG.md")
	version := flag.String("version", "", "release version (without v prefix)")
	flag.Parse()
	if *version == "" {
		fail("--version is required")
	}
	data, err := os.ReadFile(*changelogPath)
	if err != nil {
		fail("read changelog: %v", err)
	}
	notes, err := packaging.ReleaseNotesFromChangelog(data, *version)
	if err != nil {
		if errors.Is(err, packaging.ErrNoChangelogSection) {
			failWithCode(exitNoChangelogSection, "%v", err)
		}
		fail("%v", err)
	}
	fmt.Print(notes)
}

func fail(format string, args ...interface{}) {
	failWithCode(1, format, args...)
}

func failWithCode(code int, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "release-notes: "+format+"\n", args...)
	os.Exit(code)
}
