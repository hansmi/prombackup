package snapshotstream

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/hansmi/prombackup/api"
)

const MiB = 1024 * 1024

func makeFilesystem(b *testing.B, fileSize int) fs.FS {
	fileData := make([]byte, fileSize)

	if _, err := rand.Read(fileData); err != nil {
		b.Fatalf("Read() failed: %v", err)
	}

	return &fstest.MapFS{
		"file": {
			Data: fileData,
		},
	}
}

func BenchmarkStreamWriteArchive(b *testing.B) {
	var output bytes.Buffer

	output.Grow((128 + 16) * MiB)

	for _, format := range api.ArchiveFormatAll {
		b.Run(format.Name(), func(b *testing.B) {
			for _, fileSize := range []int{
				1 * MiB,
				8 * MiB,
				16 * MiB,
				128 * MiB,
			} {
				b.Run(fmt.Sprintf("%.1fMiB", float32(fileSize)/MiB), func(b *testing.B) {
					s, err := New(Options{
						Name:   "stream",
						Root:   makeFilesystem(b, fileSize),
						Format: format,
					})
					if err != nil {
						b.Fatalf("New() failed: %v", err)
					}

					output.Grow(fileSize)
					b.SetBytes(int64(fileSize))
					b.ResetTimer()

					for i := 0; i < b.N; i++ {
						if err := s.WriteArchive(&output); err != nil {
							b.Errorf("WriteArchive() failed: %v", err)
						}

						output.Reset()
					}
				})
			}
		})
	}
}
