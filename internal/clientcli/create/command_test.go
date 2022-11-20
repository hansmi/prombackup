package create

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/ref"
	"github.com/hansmi/prombackup/internal/testutils"
)

var errTest = errors.New("test error")

type fakeClient struct {
	snapshotResult api.SnapshotResult
	snapshotError  error
	downloadBody   string
	downloadResult api.DownloadResult
	downloadStatus api.DownloadStatus
}

func (c *fakeClient) Snapshot(context.Context, api.SnapshotOptions) (*api.SnapshotResult, error) {
	return &c.snapshotResult, c.snapshotError
}

func (c *fakeClient) Download(ctx context.Context, opts api.DownloadOptions) (*api.DownloadResult, error) {
	if w, err := opts.BodyWriter(c.downloadResult); err != nil {
		return nil, fmt.Errorf("BodyWriter() failed: %v", err)
	} else if _, err := io.WriteString(w, c.downloadBody); err != nil {
		return nil, fmt.Errorf("WriteString() failed: %v", err)
	}

	return &c.downloadResult, nil
}

func (c *fakeClient) DownloadStatus(context.Context, api.DownloadStatusOptions) (*api.DownloadStatus, error) {
	return &c.downloadStatus, nil
}

func TestCommand(t *testing.T) {
	defer testutils.LogOutput(t, io.Discard)()

	outputFile := filepath.Join(t.TempDir(), "output.txt")

	for _, tc := range []struct {
		name         string
		args         []string
		client       *fakeClient
		wantErr      error
		readBodyFrom string
		wantBody     string
	}{
		{
			name: "success",
			client: &fakeClient{
				downloadBody: "test body",
				downloadResult: api.DownloadResult{
					Filename: "success.txt",
				},
				downloadStatus: api.DownloadStatus{
					Finished: &api.DownloadStatusFinished{
						Success:   true,
						Sha256Hex: "63efb315ed71cc7e5a1fc202434bb3aec2091e7838707e148a017faebb7464fe",
					},
				},
			},
			readBodyFrom: "success.txt",
			wantBody:     "test body",
		},
		{
			name: "snapshot failure",
			client: &fakeClient{
				snapshotError: errTest,
			},
			wantErr: errTest,
		},
		{
			name: "download failure",
			client: &fakeClient{
				downloadBody: "failing body",
				downloadResult: api.DownloadResult{
					Filename: "../../bad/path/../failure.txt",
				},
				downloadStatus: api.DownloadStatus{
					Finished: &api.DownloadStatusFinished{
						Success:   false,
						ErrorText: ref.Ref("test error"),
					},
				},
			},
			wantErr:      ErrDownloadFailed,
			readBodyFrom: "failure.txt",
			wantBody:     "failing body",
		},
		{
			name: "write to specified file",
			args: []string{
				"--output", outputFile,
			},
			client: &fakeClient{
				downloadBody: "test body for file",
				downloadStatus: api.DownloadStatus{
					Finished: &api.DownloadStatusFinished{
						Success:   true,
						Sha256Hex: "677ad9085d46959bccc7aa5f433efc383cf39a3d459099d4455e34e52505c7bf",
					},
				},
			},
			readBodyFrom: outputFile,
			wantBody:     "test body for file",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer testutils.Chdir(t, t.TempDir())()

			fs := flag.NewFlagSet("", flag.ContinueOnError)

			var c Command

			c.SetFlags(fs)

			if err := fs.Parse(tc.args); err != nil {
				t.Errorf("Flag parsing failed: %v", err)
			}

			err := c.execute(context.Background(), tc.client)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if tc.readBodyFrom != "" {
				if content, err := os.ReadFile(tc.readBodyFrom); err != nil {
					t.Errorf("ReadFile() failed: %v", err)
				} else if diff := cmp.Diff(tc.wantBody, string(content)); diff != "" {
					t.Errorf("Body diff (-want +got):\n%s", diff)
				}
			}
		})
	}
}
