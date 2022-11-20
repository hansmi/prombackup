package main

import (
	"regexp"
	"strings"
	"testing"
)

func TestPrintVersion(t *testing.T) {
	var buf strings.Builder

	printVersion(&buf)

	for _, expr := range []*regexp.Regexp{
		regexp.MustCompile(`^Version\b+.*,\s+built at\b`),
		regexp.MustCompile(`(?ms)^go\s*go\d[.\d]*$`),
	} {
		if !expr.MatchString(buf.String()) {
			t.Errorf("Want match for %q, got %q", expr.String(), buf.String())
		}
	}
}
