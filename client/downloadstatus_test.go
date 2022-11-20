package client

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/apiendpoints"
	"github.com/hansmi/prombackup/internal/ref"
)

func TestDownloadStatus(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         api.DownloadStatusOptions
		responseCode int
		response     string
		wantQuery    url.Values
		wantErr      error
		want         *api.DownloadStatus
	}{
		{
			name:         "default",
			responseCode: http.StatusOK,
			response:     `{}`,
			wantQuery: url.Values{
				"id": {""},
			},
			want: &api.DownloadStatus{},
		},
		{
			name: "finished",
			opts: api.DownloadStatusOptions{
				ID: "8d4fa53f-17be-477e-8f09-34aa7f18f55f",
			},
			responseCode: http.StatusOK,
			response:     `{ "finished": { "success": false, "error_text": "hello" } }`,
			wantQuery: url.Values{
				"id": {"8d4fa53f-17be-477e-8f09-34aa7f18f55f"},
			},
			want: &api.DownloadStatus{
				Finished: &api.DownloadStatusFinished{
					Success:   false,
					ErrorText: ref.Ref("hello"),
				},
			},
		},
		{
			name: "error",
			opts: api.DownloadStatusOptions{
				ID: "a7372f6a-1b10-4834-9221-c7b077078f8f",
			},
			responseCode: http.StatusNotFound,
			wantQuery: url.Values{
				"id": {"a7372f6a-1b10-4834-9221-c7b077078f8f"},
			},
			wantErr: ErrRequestFailed,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts := fakeServer{
				method:       http.MethodGet,
				path:         apiendpoints.DownloadStatus,
				wantQuery:    tc.wantQuery,
				responseCode: tc.responseCode,
				responseBody: tc.response,
			}.start(t)

			c := newTestClient(t, ts)

			status, err := c.DownloadStatus(context.Background(), tc.opts)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, status); diff != "" {
				t.Errorf("Response diff (-want +got):\n%s", diff)
			}
		})
	}
}
