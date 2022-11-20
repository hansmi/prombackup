package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/hansmi/prombackup/internal/clientcli"
	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
)

func listenAndServe(addr string, handler http.Handler) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	defer listener.Close()

	log.Printf("Listening on %s", listener.Addr())

	return http.Serve(listener, handler)
}

func main() {
	showVersion := flag.Bool("version", false, "Output version information and exit.")

	listenAddress := flag.String("listen_address", clientcli.GetenvWithFallback("PROMBACKUP_SERVER_LISTEN_ADDRESS", ":8080"),
		"Address on which to expose the API. Defaults to the PROMBACKUP_SERVER_LISTEN_ADDRESS environment variable.")
	prometheusEndpoint := flag.String("prometheus_endpoint", clientcli.GetenvWithFallback("PROMBACKUP_SERVER_PROMETHEUS_ENDPOINT", "http://:9090"),
		"HTTP address for Prometheus API. Defaults to the PROMBACKUP_SERVER_PROMETHEUS_ENDPOINT environment variable.")
	snapshotDir := flag.String("snapshot_dir", clientcli.GetenvWithFallback("PROMBACKUP_SERVER_SNAPSHOT_DIR", ""),
		"Base directory for snapshots. Defaults to the PROMBACKUP_SERVER_SNAPSHOT_DIR environment variable.")

	autopruneEnabled := flag.Bool("autoprune", clientcli.MustGetenvBool("PROMBACKUP_SERVER_AUTOPRUNE_ENABLE", false),
		"Remove snapshots in regular intervals. Defaults to the PROMBACKUP_SERVER_AUTOPRUNE_ENABLE environment variable.")
	autopruneInterval := flag.Duration("autoprune_interval", clientcli.MustGetenvDuration("PROMBACKUP_SERVER_AUTOPRUNE_INTERVAL", 15*time.Minute),
		"How often to automatically remove snapshots. The interval is randomized by a small amount. Defaults to the PROMBACKUP_SERVER_AUTOPRUNE_INTERVAL environment variable.")
	autopruneKeepWithin := flag.Duration("autoprune_keep_within", clientcli.MustGetenvDuration("PROMBACKUP_SERVER_AUTOPRUNE_KEEP_WITHIN", time.Hour),
		"Keep snapshots younger than this amount of time when automatically removing them. Defaults to the PROMBACKUP_SERVER_AUTOPRUNE_KEEP_WITHIN environment variable.")

	flag.Parse()

	if *showVersion {
		fmt.Println(version.Print("prombackup-server"))
		return
	}

	if *prometheusEndpoint == "" {
		log.Fatal("--prometheus_endpoint is required")
	}

	if *snapshotDir == "" {
		log.Fatal("--snapshot_dir is required")
	}

	rand.Seed(time.Now().UnixNano())

	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(
		prometheus.NewBuildInfoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
		version.NewCollector("prombackup_server"),
	)

	client, err := api.NewClient(api.Config{
		Address: *prometheusEndpoint,
	})
	if err != nil {
		log.Fatalf("Creating Prometheus client failed: %v", err)
	}

	m, err := newManager(managerOptions{
		logger:      log.Default(),
		registry:    prometheus.WrapRegistererWithPrefix("prombackup_server_", registry),
		admin:       promv1.NewAPI(client),
		snapshotDir: *snapshotDir,
	})
	if err != nil {
		log.Fatalf("Creating manager failed: %v", err)
	}

	if *autopruneEnabled {
		p := autopruner{
			interval: *autopruneInterval,
			opts:     m.defaultPruneOptions(),
		}
		p.opts.KeepWithin = *autopruneKeepWithin

		go p.run(context.Background())
	}

	log.Fatal(listenAndServe(*listenAddress,
		handlers.CombinedLoggingHandler(log.Writer(),
			newRouter(m, registry))))
}
