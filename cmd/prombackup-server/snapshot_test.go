package main

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/apiendpoints"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

var errTest = errors.New("test error")

type fakePrometheusAdmin struct {
	result promv1.SnapshotResult
	err    error
}

func (p *fakePrometheusAdmin) Snapshot(ctx context.Context, skipHead bool) (promv1.SnapshotResult, error) {
	return p.result, p.err
}

func TestSnapshot(t *testing.T) {
	for _, tc := range []struct {
		name       string
		admin      fakePrometheusAdmin
		method     string
		target     url.URL
		wantCode   int
		wantHeader map[string]*regexp.Regexp
		wantBodyRe *regexp.Regexp
		want       *api.SnapshotResult
	}{
		{
			name: "success",
			admin: fakePrometheusAdmin{
				result: promv1.SnapshotResult{
					Name: "20221109T202035Z-355a5b4970d5a906",
				},
			},
			method: http.MethodPost,
			target: url.URL{
				Path: apiendpoints.Snapshot,
			},
			wantCode: http.StatusSeeOther,
			wantHeader: map[string]*regexp.Regexp{
				"Location": regexp.MustCompile(`/download\?name=20221109T202035Z-355a5b4970d5a906\b`),
			},
			want: &api.SnapshotResult{
				Name: "20221109T202035Z-355a5b4970d5a906",
			},
		},
		{
			name: "wrong method",
			target: url.URL{
				Path: apiendpoints.Snapshot,
			},
			wantCode:   http.StatusMethodNotAllowed,
			wantBodyRe: regexp.MustCompile(`(?i)^Method\b`),
		},
		{
			name:   "bad skip_head",
			method: http.MethodPost,
			target: url.URL{
				Path:     apiendpoints.Snapshot,
				RawQuery: "skip_head=<foo>",
			},
			wantCode:   http.StatusBadRequest,
			wantBodyRe: regexp.MustCompile(`(?i): invalid syntax\b`),
		},
		{
			name: "prometheus error",
			admin: fakePrometheusAdmin{
				err: errTest,
			},
			method: http.MethodPost,
			target: url.URL{
				Path: apiendpoints.Snapshot,
			},
			wantCode:   http.StatusInternalServerError,
			wantBodyRe: regexp.MustCompile(`(?i)^Creating snapshot failed\b.*\btest error\b`),
		},
		{
			name:   "bad name",
			method: http.MethodPost,
			target: url.URL{
				Path: apiendpoints.Snapshot,
			},
			wantCode:   http.StatusInternalServerError,
			wantBodyRe: regexp.MustCompile(`(?i)^Invalid snapshot name\b`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, err := newManager(managerOptions{
				admin:       &tc.admin,
				snapshotDir: t.TempDir(),
			})
			if err != nil {
				t.Fatalf("newManager() failed: %v", err)
			}

			handlerTest{
				handler:         newRouter(m, nil),
				method:          tc.method,
				target:          tc.target,
				wantStatusCode:  tc.wantCode,
				wantHeaderMatch: tc.wantHeader,
				wantBodyMatch:   tc.wantBodyRe,
				wantBodyJson:    tc.want,
			}.do(t)
		})
	}
}
