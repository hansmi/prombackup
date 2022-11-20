package client

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/apiendpoints"
)

func TestPrune(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         api.PruneOptions
		responseCode int
		response     string
		wantQuery    url.Values
		wantErr      error
		want         *api.PruneResult
	}{
		{
			name:         "default",
			responseCode: http.StatusOK,
			response:     `{}`,
			want:         &api.PruneResult{},
		},
		{
			name:         "keep within one hour and two minutes",
			responseCode: http.StatusOK,
			response:     `{}`,
			opts: api.PruneOptions{
				KeepWithin: 62 * time.Minute,
			},
			wantQuery: url.Values{
				"keep_within": {"1h2m0s"},
			},
			want: &api.PruneResult{},
		},
		{
			name:         "error",
			responseCode: http.StatusNotFound,
			wantErr:      ErrRequestFailed,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts := fakeServer{
				method:       http.MethodPost,
				path:         apiendpoints.Prune,
				wantQuery:    tc.wantQuery,
				responseCode: tc.responseCode,
				responseBody: tc.response,
			}.start(t)

			c := newTestClient(t, ts)

			response, err := c.Prune(context.Background(), tc.opts)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, response); diff != "" {
				t.Errorf("Response diff (-want +got):\n%s", diff)
			}
		})
	}
}
