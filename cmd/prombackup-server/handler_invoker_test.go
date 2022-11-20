package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type handlerTest struct {
	handler http.Handler

	method string
	target url.URL

	wantStatusCode  int
	wantHeaderMatch map[string]*regexp.Regexp
	wantBodyMatch   *regexp.Regexp
	wantBodyJson    any
}

func (ht handlerTest) do(t *testing.T) (*http.Response, []byte) {
	t.Helper()

	req := httptest.NewRequest(ht.method, ht.target.String(), nil)
	rec := httptest.NewRecorder()

	ht.handler.ServeHTTP(rec, req)

	resp := rec.Result()

	if diff := cmp.Diff(ht.wantStatusCode, resp.StatusCode); diff != "" {
		t.Errorf("Response status code diff (-want +got):\n%s", diff)
	}

	for name, want := range ht.wantHeaderMatch {
		if got := resp.Header.Get(name); !want.MatchString(got) {
			t.Errorf("Response header %s doesn't match %q: %q", name, want, got)
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Reading body failed: %v", err)
	}

	if !(ht.wantBodyMatch == nil || ht.wantBodyMatch.Match(body)) {
		t.Errorf("Response body doesn't match %q: %q", ht.wantBodyMatch.String(), body)
	}

	if ht.wantBodyJson != nil {
		var wantType reflect.Type

		want := reflect.ValueOf(ht.wantBodyJson)

		switch {
		case want.Kind() == reflect.Pointer && !want.IsNil():
			want = want.Elem()
			wantType = want.Type()
		case want.Kind() == reflect.Struct:
			wantType = want.Type()
		}

		if wantType != nil {
			got := reflect.New(wantType)

			if err := json.Unmarshal(body, got.Interface()); err != nil {
				t.Errorf("Unmarshalling body %q failed: %v", body, err)
			}

			if diff := cmp.Diff(want.Interface(), got.Elem().Interface()); diff != "" {
				t.Errorf("Response body diff (-want +got):\n%s", diff)
			}
		}
	}

	return resp, body
}
