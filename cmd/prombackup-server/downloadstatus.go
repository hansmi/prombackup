package main

import (
	"fmt"
	"net/http"
)

func (m *manager) handleDownloadStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	q := r.URL.Query()

	id := q.Get("id")
	if id == "" {
		http.Error(w, "Download ID is required", http.StatusNotFound)
		return
	}

	m.mu.Lock()
	s := m.downloads[id]
	m.mu.Unlock()

	if s == nil {
		http.Error(w, fmt.Sprintf("Download %s not found", id), http.StatusNotFound)
		return
	}

	writeJsonResponse(w, http.StatusOK, nil, s.Status())
}
