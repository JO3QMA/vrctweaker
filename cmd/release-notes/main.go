package main

import (
	"flag"
	"fmt"
	"os"

	"vrchat-tweaker/internal/packaging"
)

func main() {
	changelogPath := flag.String("changelog", "CHANGELOG.md", "path to CHANGELOG.md")
	version := flag.String("version", "", "release version (without v prefix)")
	flag.Parse()
	if *version == "" {
		fmt.Fprintln(os.Stderr, "release-notes: --version is required")
		os.Exit(1)
	}
	data, err := os.ReadFile(*changelogPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	notes, err := packaging.ReleaseNotesFromChangelog(data, *version)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(notes)
}
