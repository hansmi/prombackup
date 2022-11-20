package main

import (
	"net/http"
	"sort"

	"github.com/hansmi/prombackup/api"
)

func (m *manager) handleRoot(w http.ResponseWriter, r *http.Request) {
	type downloadInfo struct {
		ID   string
		Name string
	}

	var data struct {
		Downloads      []downloadInfo
		ArchiveFormats []api.ArchiveFormat
	}

	data.ArchiveFormats = api.ArchiveFormatAll

	m.mu.Lock()
	for _, d := range m.downloads {
		data.Downloads = append(data.Downloads, downloadInfo{
			ID:   d.ID(),
			Name: d.Name(),
		})
	}
	m.mu.Unlock()

	sort.Slice(data.Downloads, func(a, b int) bool {
		return data.Downloads[a].ID > data.Downloads[b].ID
	})

	if err := m.template.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
