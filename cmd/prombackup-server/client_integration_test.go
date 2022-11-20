package main

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/client"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func TestClientIntegration(t *testing.T) {
	admin := fakePrometheusAdmin{
		result: promv1.SnapshotResult{
			Name: "20221109T202035Z-355a5b4970d5a906",
		},
	}

	m, err := newManager(managerOptions{
		admin:       &admin,
		snapshotDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("newManager() failed: %v", err)
	}

	ts := httptest.NewServer(newRouter(m, nil))
	defer ts.Close()

	c, err := client.New(client.Options{
		Address: ts.URL,
		Client:  ts.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}

	if result, err := c.Snapshot(context.Background(), api.SnapshotOptions{}); err != nil {
		t.Errorf("Snapshot() failed: %v", err)
	} else if diff := cmp.Diff(&api.SnapshotResult{
		Name: "20221109T202035Z-355a5b4970d5a906",
	}, result); diff != "" {
		t.Errorf("Snapshot() diff (-want +got):\n%s", diff)
	}

	if _, err := c.Download(context.Background(), api.DownloadOptions{
		SnapshotName: "snapname-123",
	}); !(errors.Is(err, client.ErrRequestFailed) && strings.Contains(err.Error(), "snapshot not found:")) {
		t.Errorf("Download() failed: %v", err)
	}

	if _, err := c.DownloadStatus(context.Background(), api.DownloadStatusOptions{
		ID: "5a21f075-fc46-43db-af80-83a43a9383db",
	}); !(errors.Is(err, client.ErrRequestFailed) && strings.Contains(err.Error(), "Download 5a21f075-fc46-43db-af80-83a43a9383db not found")) {
		t.Errorf("DownloadStatus() failed: %v", err)
	}

	if _, err := c.Prune(context.Background(), api.PruneOptions{}); err != nil {
		t.Errorf("Prune() failed: %v", err)
	}
}
