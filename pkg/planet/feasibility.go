package planet

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// CreateImagingWindowSearch creates an async imaging window search.
// POST /tasking/v2/imaging-windows/search/
func (c *Client) CreateImagingWindowSearch(ctx context.Context, req *ImagingWindowSearchRequest) (*ImagingWindowSearch, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var search ImagingWindowSearch
	err = c.doRequest(ctx, http.MethodPost, c.TaskingURL("imaging-windows", "search", ""), body, http.StatusCreated, &search)
	return &search, err
}

// GetImagingWindowSearch retrieves imaging window search results.
// GET /tasking/v2/imaging-windows/search/{id}
func (c *Client) GetImagingWindowSearch(ctx context.Context, id string) (*ImagingWindowSearch, error) {
	var search ImagingWindowSearch
	err := c.doRequest(ctx, http.MethodGet, c.TaskingURL("imaging-windows", "search", id), nil, http.StatusOK, &search)
	return &search, err
}

// WaitForImagingWindowSearch polls until the search is complete.
func (c *Client) WaitForImagingWindowSearch(ctx context.Context, id string, opts *WaitOptions) (*ImagingWindowSearch, error) {
	if opts == nil {
		opts = &WaitOptions{
			PollInterval: 5 * time.Second,
			Timeout:      5 * time.Minute,
		}
	}

	deadline := time.Now().Add(opts.Timeout)
	ticker := time.NewTicker(opts.PollInterval)
	defer ticker.Stop()

	for {
		search, err := c.GetImagingWindowSearch(ctx, id)
		if err != nil {
			return nil, err
		}

		if search.Status.IsTerminal() {
			if search.Status == ImagingWindowSearchStatusFailed {
				return search, fmt.Errorf("imaging window search failed: %s", search.ErrorMessage)
			}
			return search, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for imaging window search %s to complete", id)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

// SearchImagingWindows is a convenience method that creates an imaging window search
// and waits for it to complete.
func (c *Client) SearchImagingWindows(ctx context.Context, req *ImagingWindowSearchRequest, opts *WaitOptions) (*ImagingWindowSearch, error) {
	search, err := c.CreateImagingWindowSearch(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.WaitForImagingWindowSearch(ctx, search.ID, opts)
}
