package main

import (
	"fmt"
	"io"
	"runtime/debug"
	"strings"
	"unicode"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func printVersion(w io.Writer) {
	fmt.Fprintf(w, "Version %s, commit %s, built at %s.\n", version, commit, date)

	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Fprintf(w, "Build information:\n%s\n", strings.TrimRightFunc(info.String(), unicode.IsSpace))
	}
}

func clientUserAgent() string {
	return fmt.Sprintf("prombackup/%s (%s)", version, date)
}
