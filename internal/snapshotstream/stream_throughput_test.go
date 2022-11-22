package snapshotstream

import (
	"bytes"
	"crypto/rand"
	"testing"
	"testing/fstest"

	"github.com/hansmi/prombackup/api"
)

func BenchmarkStreamWriteTo(b *testing.B) {
	const fileSize = 128 * 1024 * 1024

	fileData := make([]byte, fileSize)

	if _, err := rand.Read(fileData); err != nil {
		b.Fatalf("Read() failed: %v", err)
	}

	root := fstest.MapFS{
		"file": {
			Data: fileData,
		},
	}

	s, err := New(Options{
		Name:   "stream",
		Root:   root,
		Format: api.ArchiveTar,
	})
	if err != nil {
		b.Fatalf("New() failed: %v", err)
	}

	var buf bytes.Buffer

	buf.Grow(fileSize)

	b.SetBytes(fileSize)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := s.WriteTo(&buf); err != nil {
			b.Errorf("WriteTo() failed: %v", err)
		}

		buf.Reset()
	}
}
