package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

func newRouter(m *manager, registry *prometheus.Registry) http.Handler {
	r := mux.NewRouter()
	r.Use(mux.CORSMethodMiddleware(r))
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowed)

	r.HandleFunc("/", m.handleRoot).Methods(http.MethodGet)
	r.HandleFunc("/api/snapshot", m.handleSnapshot).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/api/download", m.handleDownload).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/download_status", m.handleDownloadStatus).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/prune", m.handlePrune).Methods(http.MethodPost, http.MethodOptions)

	if registry != nil {
		r.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			Registry:            registry,
			MaxRequestsInFlight: 3,
		}))
	}

	return r
}
