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

func (h *httpClient) DownloadStatus(ctx context.Context, opts api.DownloadStatusOptions) (*api.DownloadStatus, error) {
	u := h.buildURL(apiendpoints.DownloadStatus, url.Values{
		"id": {opts.ID},
	})

	req, err := h.newRequest(ctx, http.MethodGet, u)
	if err != nil {
		return nil, err
	}

	resp, err := h.doReq(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result api.DownloadStatus

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
