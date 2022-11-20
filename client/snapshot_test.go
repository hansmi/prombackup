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
)

func TestSnapshot(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         api.SnapshotOptions
		responseCode int
		response     string
		wantQuery    url.Values
		wantErr      error
		want         *api.SnapshotResult
	}{
		{
			name:         "default",
			responseCode: http.StatusSeeOther,
			response:     `{ "name": "1234" }`,
			wantQuery: url.Values{
				"skip_head": {"false"},
			},
			want: &api.SnapshotResult{
				Name: "1234",
			},
		},
		{
			name:         "skip head",
			responseCode: http.StatusOK,
			response:     `{ "name": "without head" }`,
			opts: api.SnapshotOptions{
				SkipHead: true,
			},
			wantQuery: url.Values{
				"skip_head": {"true"},
			},
			want: &api.SnapshotResult{
				Name: "without head",
			},
		},
		{
			name:         "error",
			responseCode: http.StatusNotFound,
			wantQuery: url.Values{
				"skip_head": {"false"},
			},
			wantErr: ErrRequestFailed,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			responseHeader := map[string]string{}
			ts := fakeServer{
				method:         http.MethodPost,
				path:           apiendpoints.Snapshot,
				wantQuery:      tc.wantQuery,
				responseCode:   tc.responseCode,
				responseHeader: responseHeader,
				responseBody:   tc.response,
			}.start(t)

			responseHeader["Location"] = ts.URL

			c := newTestClient(t, ts)

			response, err := c.Snapshot(context.Background(), tc.opts)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, response); diff != "" {
				t.Errorf("Response diff (-want +got):\n%s", diff)
			}
		})
	}
}
