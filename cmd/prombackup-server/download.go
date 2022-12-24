package main

import (
	"errors"
	"io/fs"
	"mime"
	"net/http"

	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/snapshotstream"
)

func (m *manager) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	q := r.URL.Query()

	name := q.Get("name")

	if name == "" {
		http.Error(w, "Snapshot name is required", http.StatusNotFound)
		return
	} else if err := validateSnapshotName(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	format := api.ArchiveTar

	if rawFormat := q.Get("format"); rawFormat != "" {
		format = api.ArchiveFormat(rawFormat)
	}

	dir, err := fs.Sub(m.snapshotRoot, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	s, err := snapshotstream.New(snapshotstream.Options{
		Name:   name,
		Root:   dir,
		Format: format,
	})

	if err != nil {
		code := http.StatusInternalServerError

		if errors.Is(err, snapshotstream.ErrArchiveFormat) {
			code = http.StatusBadRequest
		} else if errors.Is(err, snapshotstream.ErrNotFound) || errors.Is(err, snapshotstream.ErrInvalid) {
			code = http.StatusNotFound
		}

		http.Error(w, err.Error(), code)
		return
	}

	id := s.ID()

	m.mu.Lock()
	m.downloads[id] = s
	m.mu.Unlock()

	defer m.deleteDownloadAfter(id, m.downloadLifetime)

	header := w.Header()
	header.Set("Content-Type", s.ContentType)
	header.Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": s.Filename,
	}))
	header.Set(api.HttpHeaderDownloadID, id)

	if err := s.WriteArchive(w); err != nil {
		m.logger.Printf("Download %s failed: %v", id, err)
	} else {
		m.logger.Printf("Download %s finished: %+v", id, s.Status().Finished)
	}
}
