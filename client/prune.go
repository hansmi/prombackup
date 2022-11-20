package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/apiendpoints"
)

func (h *httpClient) Prune(ctx context.Context, opts api.PruneOptions) (*api.PruneResult, error) {
	queryValues := url.Values{}

	if opts.KeepWithin != 0 {
		queryValues.Set("keep_within", opts.KeepWithin.String())
	}

	req, err := h.newRequest(ctx, http.MethodPost, h.buildURL(apiendpoints.Prune, queryValues))
	if err != nil {
		return nil, err
	}

	resp, err := h.doReq(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result api.PruneResult

	switch resp.StatusCode {
	case http.StatusOK:
		body, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}

	default:
		return nil, errorFromResponse(resp)
	}

	return &result, nil
}
