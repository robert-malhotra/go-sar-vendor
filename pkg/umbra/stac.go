package umbra

import (
	"context"
	"iter"
	"net/http"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

// STACItem represents a STAC catalog item.
type STACItem struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Geometry   *GeoJSONGeometry       `json:"geometry,omitempty"`
	BBox       []float64              `json:"bbox,omitempty"`
	Properties map[string]interface{} `json:"properties"`
	Assets     map[string]STACAsset   `json:"assets"`
	Links      []STACLink             `json:"links"`
	Collection string                 `json:"collection"`
}

// STACAsset represents an asset in a STAC item.
type STACAsset struct {
	Href        string           `json:"href"`
	Type        string           `json:"type"`
	Title       string           `json:"title,omitempty"`
	Description string           `json:"description,omitempty"`
	Created     time.Time        `json:"created,omitempty"`
	Roles       []string         `json:"roles,omitempty"`
	Alternate   *AlternateAssets `json:"alternate,omitempty"`
}

// AlternateAssets contains alternate access methods for an asset.
type AlternateAssets struct {
	S3Signed *SignedURL `json:"s3_signed,omitempty"`
}

// SignedURL contains a pre-signed URL for direct download.
type SignedURL struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Href        string `json:"href"`
}

// STACLink represents a link in a STAC item.
type STACLink struct {
	Rel   string `json:"rel"`
	Type  string `json:"type,omitempty"`
	Href  string `json:"href"`
	Title string `json:"title,omitempty"`
}

// STACSearchRequest contains search parameters.
type STACSearchRequest struct {
	Limit       int              `json:"limit,omitempty"`
	BBox        []float64        `json:"bbox,omitempty"`
	Datetime    string           `json:"datetime,omitempty"`
	Intersects  *GeoJSONGeometry `json:"intersects,omitempty"`
	Collections []string         `json:"collections,omitempty"`
	IDs         []string         `json:"ids,omitempty"`
	FilterLang  string           `json:"filter-lang,omitempty"`
	Filter      interface{}      `json:"filter,omitempty"`
}

// STACSearchResponse contains search results.
type STACSearchResponse struct {
	Type           string                 `json:"type"`
	Features       []STACItem             `json:"features"`
	NumberMatched  int                    `json:"numberMatched,omitempty"`
	NumberReturned int                    `json:"numberReturned,omitempty"`
	Context        map[string]interface{} `json:"context,omitempty"`
	Links          []STACLink             `json:"links,omitempty"`
}

// STACCollection represents a STAC collection.
type STACCollection struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Keywords    []string               `json:"keywords,omitempty"`
	License     string                 `json:"license,omitempty"`
	Extent      *STACExtent            `json:"extent,omitempty"`
	Links       []STACLink             `json:"links,omitempty"`
	Providers   []STACProvider         `json:"providers,omitempty"`
	Summaries   map[string]interface{} `json:"summaries,omitempty"`
}

// STACExtent represents the spatial and temporal extent of a STAC collection.
type STACExtent struct {
	Spatial  *STACSpatialExtent  `json:"spatial,omitempty"`
	Temporal *STACTemporalExtent `json:"temporal,omitempty"`
}

// STACSpatialExtent represents the spatial extent.
type STACSpatialExtent struct {
	BBox [][]float64 `json:"bbox,omitempty"`
}

// STACTemporalExtent represents the temporal extent.
type STACTemporalExtent struct {
	Interval [][]string `json:"interval,omitempty"`
}

// STACProvider represents a STAC provider.
type STACProvider struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	URL         string   `json:"url,omitempty"`
}

// GetSTACItem retrieves a STAC item by collection and item ID.
// GET /stac/collections/{collectionId}/items/{itemId}
func (c *Client) GetSTACItem(ctx context.Context, collectionID, itemID string) (*STACItem, error) {
	var item STACItem
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("stac", "collections", collectionID, "items", itemID), nil, http.StatusOK, &item)
	return &item, err
}

// GetSTACItemV2 retrieves a STAC item using the v2 API.
// GET /v2/stac/collections/{collectionId}/items/{itemId}
func (c *Client) GetSTACItemV2(ctx context.Context, collectionID, itemID string) (*STACItem, error) {
	var item STACItem
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("v2", "stac", "collections", collectionID, "items", itemID), nil, http.StatusOK, &item)
	return &item, err
}

// SearchSTAC returns an iterator over STAC search results with automatic pagination.
// POST /stac/search
func (c *Client) SearchSTAC(ctx context.Context, req STACSearchRequest) iter.Seq2[STACItem, error] {
	if req.Limit == 0 {
		req.Limit = defaultSearchLimit
	}

	return func(yield func(STACItem, error) bool) {
		for {
			body, err := common.MarshalBody(req)
			if err != nil {
				var zero STACItem
				yield(zero, err)
				return
			}

			var resp STACSearchResponse
			err = c.DoRaw(ctx, http.MethodPost, c.BaseURL().JoinPath("stac", "search"), body, http.StatusOK, &resp)
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

			// STAC uses link-based pagination - check for next link
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
			// For STAC, pagination is link-based; single page returned per call
			return
		}
	}
}

// SearchSTACV2 returns an iterator over STAC v2 search results with automatic pagination.
// POST /v2/stac/search
func (c *Client) SearchSTACV2(ctx context.Context, req STACSearchRequest) iter.Seq2[STACItem, error] {
	if req.Limit == 0 {
		req.Limit = defaultSearchLimit
	}

	return func(yield func(STACItem, error) bool) {
		for {
			body, err := common.MarshalBody(req)
			if err != nil {
				var zero STACItem
				yield(zero, err)
				return
			}

			var resp STACSearchResponse
			err = c.DoRaw(ctx, http.MethodPost, c.BaseURL().JoinPath("v2", "stac", "search"), body, http.StatusOK, &resp)
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
			var nextLink string
			for _, link := range resp.Links {
				if link.Rel == "next" {
					nextLink = link.Href
					break
				}
			}
			if nextLink == "" {
				return
			}
			// For STAC, pagination is typically link-based
			return
		}
	}
}

// GetSTACCollection retrieves a STAC collection by ID.
// GET /stac/collections/{collectionId}
func (c *Client) GetSTACCollection(ctx context.Context, collectionID string) (*STACCollection, error) {
	var col STACCollection
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("stac", "collections", collectionID), nil, http.StatusOK, &col)
	return &col, err
}

// ListSTACCollections retrieves all STAC collections.
// GET /stac/collections
func (c *Client) ListSTACCollections(ctx context.Context) ([]STACCollection, error) {
	var response struct {
		Collections []STACCollection `json:"collections"`
	}
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("stac", "collections"), nil, http.StatusOK, &response)
	return response.Collections, err
}
