package main

import (
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/apiendpoints"
)

func TestPrune(t *testing.T) {
	for _, tc := range []struct {
		name       string
		method     string
		target     url.URL
		wantCode   int
		wantBodyRe *regexp.Regexp
		want       *api.PruneResult
	}{
		{
			name:   "success",
			method: http.MethodPost,
			target: url.URL{
				Path: apiendpoints.Prune,
			},
			wantCode: http.StatusOK,
			want:     &api.PruneResult{},
		},
		{
			name: "wrong method",
			target: url.URL{
				Path: apiendpoints.Prune,
			},
			wantCode:   http.StatusMethodNotAllowed,
			wantBodyRe: regexp.MustCompile(`(?i)^Method\b`),
		},
		{
			name:   "bad skip_head",
			method: http.MethodPost,
			target: url.URL{
				Path:     apiendpoints.Prune,
				RawQuery: "keep_within=<foo>",
			},
			wantCode:   http.StatusBadRequest,
			wantBodyRe: regexp.MustCompile(`(?i): invalid duration\b`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, err := newManager(managerOptions{
				snapshotDir: t.TempDir(),
			})
			if err != nil {
				t.Fatalf("newManager() failed: %v", err)
			}

			handlerTest{
				handler:        newRouter(m, nil),
				method:         tc.method,
				target:         tc.target,
				wantStatusCode: tc.wantCode,
				wantBodyMatch:  tc.wantBodyRe,
				wantBodyJson:   tc.want,
			}.do(t)
		})
	}
}
