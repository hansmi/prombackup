package main

import (
	"testing"
	"testing/fstest"

	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/snapshotstream"
)

func TestDeleteDownload(t *testing.T) {
	m, err := newManager(managerOptions{
		snapshotDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("newManager() failed: %v", err)
	}

	testStream, err := snapshotstream.New(snapshotstream.Options{
		Name:   "to delete",
		Root:   &fstest.MapFS{},
		Format: api.ArchiveTar,
	})
	if err != nil {
		t.Error(err)
	}

	m.downloads[testStream.ID()] = testStream

	m.deleteDownload(testStream.ID())

	if len(m.downloads) != 0 {
		t.Errorf("Download wasn't deleted: %v", m.downloads)
	}
}
