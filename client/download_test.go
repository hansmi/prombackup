package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/apiendpoints"
)

func TestDownload(t *testing.T) {
	for _, tc := range []struct {
		name           string
		opts           api.DownloadOptions
		responseCode   int
		responseHeader map[string]string
		wantQuery      url.Values
		wantErr        error
		want           *api.DownloadResult
	}{
		{
			name:         "default",
			responseCode: http.StatusOK,
			responseHeader: map[string]string{
				api.HttpHeaderDownloadID: "896cd64c-758a-49c4-8551-5e70a30b4ec8",
				"Content-Type":           "application/octet-stream",
				"Content-Disposition":    "attachment; filename=foo.bin",
			},
			wantQuery: url.Values{
				"name": {""},
			},
			want: &api.DownloadResult{
				ID:          "896cd64c-758a-49c4-8551-5e70a30b4ec8",
				ContentType: "application/octet-stream",
				Filename:    "foo.bin",
			},
		},
		{
			name: "with name",
			opts: api.DownloadOptions{
				SnapshotName: "the-snapshot",
			},
			responseCode: http.StatusOK,
			responseHeader: map[string]string{
				api.HttpHeaderDownloadID: "ee411ed8-417e-4948-886b-74f53c62f0aa",
				"Content-Type":           "application/x-tar",
				"Content-Disposition":    "attachment; filename*=UTF-8''b%c3%a4r.tar",
			},
			wantQuery: url.Values{
				"name": {"the-snapshot"},
			},
			want: &api.DownloadResult{
				ID:          "ee411ed8-417e-4948-886b-74f53c62f0aa",
				ContentType: "application/x-tar",
				Filename:    "b\u00e4r.tar",
			},
		},
		{
			name: "with format",
			opts: api.DownloadOptions{
				SnapshotName: "tar and gzip",
				Format:       api.ArchiveTarGzip,
			},
			responseCode: http.StatusOK,
			responseHeader: map[string]string{
				api.HttpHeaderDownloadID: "5a3fcdc8-08f5-4ae1-82d5-6dd89097abb4",
				"Content-Type":           "application/gzip",
				"Content-Disposition":    "attachment; filename=bar.tgz",
			},
			wantQuery: url.Values{
				"name":   {"tar and gzip"},
				"format": {"tgz"},
			},
			want: &api.DownloadResult{
				ID:          "5a3fcdc8-08f5-4ae1-82d5-6dd89097abb4",
				ContentType: "application/gzip",
				Filename:    "bar.tgz",
			},
		},
		{
			name: "missing content-disposition",
			opts: api.DownloadOptions{
				SnapshotName: "missing c-d",
			},
			responseCode: http.StatusOK,
			responseHeader: map[string]string{
				api.HttpHeaderDownloadID: "0841c9ad-f607-46be-ba76-eccb8b29f025",
			},
			wantQuery: url.Values{
				"name": {"missing c-d"},
			},
			wantErr: ErrResponseIncomplete,
		},
		{
			name:         "missing header",
			responseCode: http.StatusOK,
			wantQuery: url.Values{
				"name": {""},
			},
			wantErr: ErrResponseIncomplete,
		},
		{
			name:         "bad ID",
			responseCode: http.StatusOK,
			responseHeader: map[string]string{
				api.HttpHeaderDownloadID: "not a valid ID",
			},
			wantQuery: url.Values{
				"name": {""},
			},
			wantErr: cmpopts.AnyError,
		},
		{
			name:         "error",
			responseCode: http.StatusNotFound,
			wantQuery: url.Values{
				"name": {""},
			},
			wantErr: ErrRequestFailed,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			responseBody := strings.Repeat("Hello world\n", 1024)

			ts := fakeServer{
				method:         http.MethodGet,
				path:           apiendpoints.Download,
				wantQuery:      tc.wantQuery,
				responseCode:   tc.responseCode,
				responseHeader: tc.responseHeader,
				responseBody:   responseBody,
			}.start(t)

			c := newTestClient(t, ts)

			var receivedResponseBody bytes.Buffer

			tc.opts.BodyWriter = func(response api.DownloadResult) (io.Writer, error) {
				if diff := cmp.Diff(tc.want, &response, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Response diff in BodyWriter (-want +got):\n%s", diff)
				}

				return &receivedResponseBody, nil
			}

			response, err := c.Download(context.Background(), tc.opts)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, response, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Response diff (-want +got):\n%s", diff)
			}

			if err == nil {
				if diff := cmp.Diff(responseBody, receivedResponseBody.String()); diff != "" {
					t.Errorf("Response body diff (-want +got):\n%s", diff)
				}
			}
		})
	}
}
