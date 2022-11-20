package main

import (
	"net/http"
	"net/url"
	"regexp"
	"testing"
	"testing/fstest"

	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/apiendpoints"
	"github.com/hansmi/prombackup/internal/snapshotstream"
)

func TestDownloadStatus(t *testing.T) {
	m, err := newManager(managerOptions{
		snapshotDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("newManager() failed: %v", err)
	}

	testStream, err := snapshotstream.New(snapshotstream.Options{
		Name:   "test",
		Root:   &fstest.MapFS{},
		Format: api.ArchiveTar,
	})
	if err != nil {
		t.Error(err)
	}

	m.downloads[testStream.ID()] = testStream

	for _, tc := range []struct {
		name       string
		method     string
		target     url.URL
		wantCode   int
		wantBodyRe *regexp.Regexp
		want       *api.DownloadStatus
	}{
		{
			name: "success",
			target: url.URL{
				Path:     apiendpoints.DownloadStatus,
				RawQuery: "id=" + url.QueryEscape(testStream.ID()),
			},
			wantCode: http.StatusOK,
			want: &api.DownloadStatus{
				ID:           testStream.ID(),
				SnapshotName: "test",
			},
		},
		{
			name:   "wrong method",
			method: http.MethodPost,
			target: url.URL{
				Path: apiendpoints.DownloadStatus,
			},
			wantCode:   http.StatusMethodNotAllowed,
			wantBodyRe: regexp.MustCompile(`(?i)^Method\b`),
		},
		{
			name: "not found",
			target: url.URL{
				Path:     apiendpoints.DownloadStatus,
				RawQuery: "id=0eb97f68-3e6c-4953-ba84-981582f0a75a",
			},
			wantCode:   http.StatusNotFound,
			wantBodyRe: regexp.MustCompile(`(?i)^Download 0eb97f68.*981582f0a75a not found\b`),
		},
		{
			name: "missing everything",
			target: url.URL{
				Path: apiendpoints.DownloadStatus,
			},
			wantCode:   http.StatusNotFound,
			wantBodyRe: regexp.MustCompile(`(?i)^Download ID\b.*\brequired\b`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
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
