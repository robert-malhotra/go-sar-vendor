package umbra

import (
	"context"
	"fmt"
	"iter"
	"net/http"
)

// ArchiveSearchRequest contains archive search parameters.
type ArchiveSearchRequest struct {
	Limit       int              `json:"limit,omitempty"`
	BBox        []float64        `json:"bbox,omitempty"`
	Datetime    string           `json:"datetime,omitempty"`
	Intersects  *GeoJSONGeometry `json:"intersects,omitempty"`
	Collections []string         `json:"collections,omitempty"`
	IDs         []string         `json:"ids,omitempty"`
	FilterLang  string           `json:"filter-lang,omitempty"`
	Filter      interface{}      `json:"filter,omitempty"`
}

// SearchArchive returns an iterator over archive search results with automatic pagination.
// POST /archive/search
func (c *Client) SearchArchive(ctx context.Context, req ArchiveSearchRequest) iter.Seq2[STACItem, error] {
	if req.Limit == 0 {
		req.Limit = defaultSearchLimit
	}

	return func(yield func(STACItem, error) bool) {
		for {
			body, err := marshalBody(req)
			if err != nil {
				var zero STACItem
				yield(zero, err)
				return
			}

			var resp STACSearchResponse
			err = c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("archive", "search"), body, http.StatusOK, &resp)
			if err != nil {
				var zero STACItem
				yield(zero, err)
				return
			}

			for _, item := range resp.Features {
				if !yield(item, nil) {
					return
				}
			}

			// If we got fewer results than the limit, we've reached the end
			if len(resp.Features) < req.Limit {
				return
			}

			// Check for next link for pagination
			var hasNext bool
			for _, link := range resp.Links {
				if link.Rel == "next" {
					hasNext = true
					break
				}
			}
			if !hasNext {
				return
			}
			// Archive uses link-based pagination; single page returned per call
			return
		}
	}
}

// GetArchiveThumbnail retrieves a thumbnail image for an archive item.
// GET /archive/thumbnail/{archiveId}
func (c *Client) GetArchiveThumbnail(ctx context.Context, archiveID string) ([]byte, error) {
	u := c.BaseURL().JoinPath("archive", "thumbnail", archiveID)
	return c.doRequestRaw(ctx, http.MethodGet, u, nil, http.StatusOK)
}

// GetArchiveCollectionItems lists items in a collection.
// GET /archive/collections/{collectionId}/items
func (c *Client) GetArchiveCollectionItems(ctx context.Context, collectionID string, limit, offset int) (*STACSearchResponse, error) {
	u := c.BaseURL().JoinPath("archive", "collections", collectionID, "items")
	q := u.Query()
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		q.Set("offset", fmt.Sprintf("%d", offset))
	}
	u.RawQuery = q.Encode()

	var resp STACSearchResponse
	err := c.doRequest(ctx, http.MethodGet, u, nil, http.StatusOK, &resp)
	return &resp, err
}

// GetArchiveItem retrieves a specific archive item by ID.
// GET /archive/items/{itemId}
func (c *Client) GetArchiveItem(ctx context.Context, itemID string) (*STACItem, error) {
	var item STACItem
	err := c.doRequest(ctx, http.MethodGet, c.BaseURL().JoinPath("archive", "items", itemID), nil, http.StatusOK, &item)
	return &item, err
}
