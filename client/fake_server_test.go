package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type fakeServer struct {
	method string
	path   string

	wantQuery url.Values

	responseCode   int
	responseHeader map[string]string
	responseBody   string
}

func (s fakeServer) start(t *testing.T) *httptest.Server {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if diff := cmp.Diff(s.wantQuery, r.URL.Query(), cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Query diff (-want +got):\n%s", diff)
		}

		if r.Method == s.method && r.URL.Path == s.path {
			for k, v := range s.responseHeader {
				w.Header().Set(k, v)
			}
			w.WriteHeader(s.responseCode)
			io.WriteString(w, s.responseBody)
			return
		}

		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}))

	t.Cleanup(ts.Close)

	return ts
}
