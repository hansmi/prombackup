package main

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/apiendpoints"
)

func TestDownload(t *testing.T) {
	tmpdir := t.TempDir()

	if err := os.Mkdir(filepath.Join(tmpdir, "20221109T202035Z-355a5b4970d5a906"), 0o700); err != nil {
		t.Error(err)
	}

	for _, tc := range []struct {
		name       string
		method     string
		target     url.URL
		wantCode   int
		wantHeader map[string]*regexp.Regexp
		wantBodyRe *regexp.Regexp
	}{
		{
			name: "success",
			target: url.URL{
				Path:     apiendpoints.Download,
				RawQuery: "name=20221109T202035Z-355a5b4970d5a906",
			},
			wantCode: http.StatusOK,
			wantHeader: map[string]*regexp.Regexp{
				"Content-Type":           regexp.MustCompile(`(?i)^application/x-tar\b`),
				"Content-Disposition":    regexp.MustCompile(`(?i)\bfilename.*=.*\b2022.*a906\.tar\b`),
				api.HttpHeaderDownloadID: regexp.MustCompile(`(?i)^\d+_[-_0-9a-z]+$`),
			},
		},
		{
			name:   "wrong method",
			method: http.MethodPost,
			target: url.URL{
				Path: apiendpoints.Download,
			},
			wantCode:   http.StatusMethodNotAllowed,
			wantBodyRe: regexp.MustCompile(`(?i)^Method\b`),
		},
		{
			name: "not found",
			target: url.URL{
				Path: apiendpoints.Download,
			},
			wantCode: http.StatusNotFound,
		},
		{
			name: "bad name",
			target: url.URL{
				Path:     apiendpoints.Download,
				RawQuery: "name=bad_name",
			},
			wantCode:   http.StatusBadRequest,
			wantBodyRe: regexp.MustCompile(`(?i)^Invalid snapshot name\b`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, err := newManager(managerOptions{
				snapshotDir: tmpdir,
			})
			if err != nil {
				t.Fatalf("newManager() failed: %v", err)
			}

			resp, body := handlerTest{
				handler:         newRouter(m, nil),
				method:          tc.method,
				target:          tc.target,
				wantStatusCode:  tc.wantCode,
				wantHeaderMatch: tc.wantHeader,
				wantBodyMatch:   tc.wantBodyRe,
			}.do(t)

			if resp.StatusCode == http.StatusOK {
				count := 0

				for tr := tar.NewReader(bytes.NewReader(body)); ; {
					if _, err := tr.Next(); errors.Is(err, io.EOF) {
						break
					} else if err != nil {
						t.Errorf("Invalid tar archive: %v", err)
						break
					}

					count++
				}

				if count < 1 {
					t.Error("Tar archive is empty")
				}
			}
		})
	}
}
