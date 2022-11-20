package snapshotstream

import (
	"crypto/rand"
	"fmt"
	"sync/atomic"
	"time"
)

var idCounter atomic.Uint64

func newID() string {
	// Enable string-based sorting
	ts := time.Now().UTC().Format("20060102150405")

	suffix := make([]byte, 4)

	if _, err := rand.Read(suffix); err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s_%d%x", ts, idCounter.Add(1), suffix)
}
