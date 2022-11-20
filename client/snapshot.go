package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/apiendpoints"
)

func (h *httpClient) Snapshot(ctx context.Context, opts api.SnapshotOptions) (*api.SnapshotResult, error) {
	u := h.buildURL(apiendpoints.Snapshot, url.Values{
		"skip_head": {strconv.FormatBool(opts.SkipHead)},
	})

	req, err := h.newRequest(ctx, http.MethodPost, u)
	if err != nil {
		return nil, err
	}

	resp, err := h.doReq(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result api.SnapshotResult

	switch resp.StatusCode {
	case http.StatusOK, http.StatusSeeOther:
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
