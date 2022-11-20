package api

import (
	"net/http"
	"testing"
)

func TestCanonicalHeaderKey(t *testing.T) {
	for _, name := range []string{
		HttpHeaderDownloadID,
	} {
		t.Run(name, func(t *testing.T) {
			if got := http.CanonicalHeaderKey(name); got != name {
				t.Errorf("Header %q is not canonical: %s", name, got)
			}
		})
	}
}
