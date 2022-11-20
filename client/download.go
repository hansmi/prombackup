package client

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/internal/apiendpoints"
)

func (h *httpClient) Download(ctx context.Context, opts api.DownloadOptions) (*api.DownloadResult, error) {
	queryValues := url.Values{
		"name": {opts.SnapshotName},
	}

	if opts.Format != "" {
		queryValues.Set("format", opts.Format.Name())
	}

	req, err := h.newRequest(ctx, http.MethodGet, h.buildURL(apiendpoints.Download, queryValues))
	if err != nil {
		return nil, err
	}

	resp, err := h.doReq(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result api.DownloadResult

	switch resp.StatusCode {
	case http.StatusOK:
		if id := resp.Header.Get(api.HttpHeaderDownloadID); id == "" {
			return nil, fmt.Errorf("%w: missing %s header", ErrResponseIncomplete, api.HttpHeaderDownloadID)
		} else {
			result.ID = id
		}

		if result.ContentType, result.ContentTypeParams, err = mime.ParseMediaType(resp.Header.Get("Content-Type")); err != nil {
			return nil, fmt.Errorf("invalid content-type: %w", err)
		}

		if contentDisposition := resp.Header.Get("Content-Disposition"); contentDisposition == "" {
			return nil, fmt.Errorf("%w: missing Content-Disposition header", ErrResponseIncomplete)
		} else if mediaType, params, err := mime.ParseMediaType(contentDisposition); err != nil {
			return nil, fmt.Errorf("parsing content-disposition header failed: %w", err)
		} else if mediaType == "attachment" {
			result.Filename = filepath.Base(params["filename"])
		}

		w, err := opts.BodyWriter(result)
		if err != nil {
			return nil, err
		}

		if _, err := io.Copy(w, resp.Body); err != nil {
			return nil, err
		}

	default:
		return nil, errorFromResponse(resp)
	}

	return &result, nil
}
