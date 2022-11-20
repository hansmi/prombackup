package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/pruner"
)

func (m *manager) defaultPruneOptions() pruner.Options {
	return pruner.Options{
		Root:           m.snapshotRootPath,
		Logger:         m.logger,
		PreRemoveCheck: m.checkSnapshotBeforeRemove,
	}
}

func (m *manager) handlePrune(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	opts := m.defaultPruneOptions()

	if raw := r.Form.Get("keep_within"); raw != "" {
		if value, err := time.ParseDuration(raw); err != nil {
			http.Error(w, fmt.Sprintf("Parsing keep_within: %v", err.Error()), http.StatusBadRequest)
			return
		} else {
			opts.KeepWithin = value
		}
	}

	if err := pruner.Prune(r.Context(), opts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJsonResponse(w, http.StatusOK, nil, api.PruneResult{})
}
