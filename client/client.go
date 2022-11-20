package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/hansmi/prombackup/api"
)

var ErrRequestFailed = errors.New("request failed")
var ErrResponseIncomplete = errors.New("response incomplete")

func errorFromResponse(resp *http.Response) error {
	msg := fmt.Sprintf("%s %s returned %q", resp.Request.Method, resp.Request.URL.Redacted(), resp.Status)

	if body, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024)); err == nil {
		if body := string(bytes.TrimSpace(body)); body != "" {
			return fmt.Errorf("%w: %s: %q", ErrRequestFailed, msg, body)
		}
	}

	return fmt.Errorf("%w: %s", ErrRequestFailed, msg)
}

func disableRedirects(client *http.Client) *http.Client {
	clone := *client
	clone.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &clone
}

type Logger interface {
	Print(...any)
	Printf(string, ...any)
}

// Options defines the options for a new client.
type Options struct {
	// URL of the server to connect to.
	Address string

	Logger Logger

	// Customize the User-Agent header sent to the server.
	UserAgent string

	// Client is used to make HTTP requests. If not provided http.DefaultClient
	// is used.
	Client *http.Client
}

// New returns a new API client.
func New(opts Options) (api.Interface, error) {
	ep, err := url.Parse(opts.Address)
	if err != nil {
		return nil, err
	}
	ep.Path = strings.TrimRight(ep.Path, "/")

	h := &httpClient{
		endpoint: ep,
		logger:   opts.Logger,
	}

	if h.logger == nil {
		h.logger = log.New(io.Discard, "", 0)
	}

	cl := opts.Client

	if cl == nil {
		cl = http.DefaultClient
	}

	cl = disableRedirects(cl)

	h.doReq = func(req *http.Request) (*http.Response, error) {
		if opts.UserAgent != "" {
			req.Header.Set("User-Agent", opts.UserAgent)
		}

		return cl.Do(req)
	}

	return h, nil
}

type httpClient struct {
	endpoint *url.URL
	logger   Logger
	doReq    func(*http.Request) (*http.Response, error)
}

func (h *httpClient) newRequest(ctx context.Context, method string, u *url.URL) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, u.String(), nil)
}

func (h *httpClient) buildURL(ep string, query url.Values) *url.URL {
	u := *h.endpoint
	u.Path = path.Join(u.Path, ep)
	u.RawQuery = query.Encode()
	return &u
}
