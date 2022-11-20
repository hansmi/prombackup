package main

import (
	"net/http"
	"net/url"
	"regexp"
	"testing"
)

func TestRoot(t *testing.T) {
	for _, tc := range []struct {
		name       string
		method     string
		target     url.URL
		wantCode   int
		wantBodyRe *regexp.Regexp
	}{
		{
			name:   "success",
			method: http.MethodGet,
			target: url.URL{
				Path: "/",
			},
			wantCode:   http.StatusOK,
			wantBodyRe: regexp.MustCompile(`(?m)^<html>`),
		},
		{
			name:   "wrong method",
			method: http.MethodPost,
			target: url.URL{
				Path: "/",
			},
			wantCode:   http.StatusMethodNotAllowed,
			wantBodyRe: regexp.MustCompile(`(?i)^Method\b`),
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
			}.do(t)
		})
	}
}
