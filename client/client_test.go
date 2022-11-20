package client

import (
	"net/http/httptest"
	"testing"

	"github.com/hansmi/prombackup/api"
)

func newTestClient(t *testing.T, ts *httptest.Server) api.Interface {
	t.Helper()

	c, err := New(Options{
		Address: ts.URL,
		Client:  ts.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}

	return c
}
