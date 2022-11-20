package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hansmi/prombackup/api"
)

func (m *manager) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var skipHead bool

	if raw := r.Form.Get("skip_head"); raw != "" {
		if value, err := strconv.ParseBool(raw); err != nil {
			http.Error(w, fmt.Sprintf("Parsing skip_head: %v", err.Error()), http.StatusBadRequest)
			return
		} else {
			skipHead = value
		}
	}

	result, err := m.admin.Snapshot(r.Context(), skipHead)
	if err != nil {
		http.Error(w, fmt.Sprintf("Creating snapshot failed: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	if err := validateSnapshotName(result.Name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	downloadUrlValues := url.Values{
		"name": {result.Name},
	}

	if downloadFormat := r.Form.Get("download_format"); downloadFormat != "" {
		downloadUrlValues.Set("format", downloadFormat)
	}

	writeJsonResponse(w, http.StatusSeeOther, http.Header{
		"Location": {
			(&url.URL{
				Path:     "./download",
				RawQuery: downloadUrlValues.Encode(),
			}).String(),
		},
	}, api.SnapshotResult{
		Name: result.Name,
	})
}
