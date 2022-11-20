package snapshotstream

import (
	"regexp"
	"testing"
)

func TestNewID(t *testing.T) {
	idRe := regexp.MustCompile(`^\d{14}_[0-9a-f]{4,}$`)

	for i := 0; i < 100; i++ {
		if got := newID(); !idRe.MatchString(got) {
			t.Errorf("newID() returned %q, want match for %q", got, idRe.String())
		}
	}
}
