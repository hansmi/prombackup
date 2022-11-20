package testutils

import (
	"io"
	"log"
	"testing"
)

func LogOutput(t *testing.T, w io.Writer) func() {
	orig := log.Writer()

	log.SetOutput(w)

	return func() {
		log.SetOutput(orig)
	}
}
