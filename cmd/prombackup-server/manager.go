package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/hansmi/prombackup/internal/snapshotstream"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

//go:embed template/*.tmpl
var contentTemplate embed.FS

var errSnapshotInUse = errors.New("snapshot in use")

var snapshotNameRe = regexp.MustCompile(`(?i)^[a-z0-9]+-[a-z0-9]+$`)

func validateSnapshotName(name string) error {
	if !snapshotNameRe.MatchString(name) {
		return fmt.Errorf("invalid snapshot name: %q", name)
	}

	return nil
}

func writeJsonResponse(w http.ResponseWriter, code int, header http.Header, payload any) {
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		code := http.StatusInternalServerError
		http.Error(w, err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", mime.FormatMediaType("application/json", map[string]string{
		"charset": "utf-8",
	}))

	for k, vv := range header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(code)
	w.Write(body)
	w.Write([]byte{'\n'})
}

type adminAPI interface {
	Snapshot(ctx context.Context, skipHead bool) (promv1.SnapshotResult, error)
}

type managerOptions struct {
	logger      Logger
	registry    prometheus.Registerer
	admin       adminAPI
	snapshotDir string
}

type manager struct {
	logger           Logger
	admin            adminAPI
	snapshotRoot     fs.FS
	snapshotRootPath string
	downloadLifetime time.Duration
	template         *template.Template

	mu        sync.Mutex
	downloads map[string]*snapshotstream.Stream
}

func newManager(opts managerOptions) (*manager, error) {
	var err error

	m := &manager{
		admin:            opts.admin,
		snapshotRoot:     os.DirFS(opts.snapshotDir),
		snapshotRootPath: opts.snapshotDir,
		logger:           opts.logger,
		downloadLifetime: 15 * time.Minute,

		downloads: map[string]*snapshotstream.Stream{},
	}

	if m.logger == nil {
		m.logger = log.New(io.Discard, "", 0)
	}

	if m.template, err = template.ParseFS(contentTemplate, "template/*.tmpl"); err != nil {
		return nil, err
	}

	f := promauto.With(opts.registry)
	f.NewGaugeFunc(prometheus.GaugeOpts{
		Subsystem: "download",
		Name:      "tracked_count",
	}, m.downloadsCountMetric)

	return m, nil
}

func (m *manager) checkSnapshotBeforeRemove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, d := range m.downloads {
		if d.Name() == name {
			return fmt.Errorf("%w: download %s", errSnapshotInUse, d.ID())
		}
	}

	return nil
}

func (m *manager) downloadsCountMetric() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	return float64(len(m.downloads))
}

func (m *manager) deleteDownloadAfter(id string, delay time.Duration) {
	time.AfterFunc(delay, func() {
		m.deleteDownload(id)
	})
}

func (m *manager) deleteDownload(id string) {
	m.logger.Printf("Download %s has expired", id)

	m.mu.Lock()
	delete(m.downloads, id)
	m.mu.Unlock()
}
