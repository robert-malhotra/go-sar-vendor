package iceye

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

// ----------------------------------------------------------------------------
// Catalog API Types
// ----------------------------------------------------------------------------

// STACItem represents a STAC-compliant catalog item (GeoJSON Feature).
type STACItem struct {
	ID          string               `json:"id"`
	Type        string               `json:"type"` // "Feature"
	StacVersion string               `json:"stac_version"`
	Geometry    *Geometry            `json:"geometry,omitempty"`
	BBox        BoundingBox          `json:"bbox"`
	Collection  string               `json:"collection"`
	Properties  ItemProperties       `json:"properties"`
	Assets      map[string]ItemAsset `json:"assets"`
	Links       []Link               `json:"links,omitempty"`
}

// ItemProperties contains STAC item properties.
type ItemProperties struct {
	// Temporal
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`

	// Imaging parameters
	InstrumentMode       string   `json:"instrument_mode,omitempty"`
	ProductType          string   `json:"product_type,omitempty"`
	ObservationDirection string   `json:"observation_direction,omitempty"`
	SatelliteLookAngle   float64  `json:"satellite_look_angle,omitempty"`
	IncidenceAngle       float64  `json:"incidence_angle,omitempty"`
	Polarizations        []string `json:"polarizations,omitempty"`
	ImageMode            string   `json:"image_mode,omitempty"`
	OrbitState           string   `json:"orbit_state,omitempty"`
}

// ItemAsset represents an asset in a STAC item.
type ItemAsset struct {
	Href        string           `json:"href"`
	Title       string           `json:"title,omitempty"`
	Type        string           `json:"type,omitempty"` // MIME type
	Roles       []string         `json:"roles,omitempty"`
	Coordinates *ThumbnailCoords `json:"coordinates,omitempty"`
}

// ThumbnailCoords contains corner coordinates for thumbnails.
type ThumbnailCoords struct {
	TopLeft     [2]float64 `json:"top_left"`
	TopRight    [2]float64 `json:"top_right"`
	BottomLeft  [2]float64 `json:"bottom_left"`
	BottomRight [2]float64 `json:"bottom_right"`
}

// Link represents a STAC link.
type Link struct {
	Href  string `json:"href"`
	Rel   string `json:"rel"`
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
}

// ListItemsOptions for GET /catalog/v1/items.
type ListItemsOptions struct {
	IDs      []string     // Specific item IDs to retrieve
	BBox     *BoundingBox // Bounding box filter [minLon, minLat, maxLon, maxLat]
	Datetime string       // RFC 3339 datetime or interval (e.g., "2021-02-12T00:00:00Z/2021-03-18T12:31:12Z")
	SortBy   []string     // Properties with +/- prefix for asc/desc (e.g., "+start_time", "-end_time")
}

// SearchRequest for POST /catalog/v1/search.
// All fields are optional.
type SearchRequest struct {
	IDs      []string               `json:"ids,omitempty"`
	BBox     *BoundingBox           `json:"bbox,omitempty"`
	Datetime string                 `json:"datetime,omitempty"`
	Limit    int                    `json:"limit,omitempty"`
	Query    map[string]QueryFilter `json:"query,omitempty"`
	SortBy   []SortCondition        `json:"sortby,omitempty"`
}

// QueryFilter for advanced STAC queries using the Query Extension.
// Supported properties: start_time, end_time, image_mode, instrument_mode,
// product_type, observation_direction, orbit_state, incidence_angle, satellite_look_angle
type QueryFilter struct {
	Eq         any    `json:"eq,omitempty"`
	Neq        any    `json:"neq,omitempty"`
	Lt         any    `json:"lt,omitempty"`
	Lte        any    `json:"lte,omitempty"`
	Gt         any    `json:"gt,omitempty"`
	Gte        any    `json:"gte,omitempty"`
	StartsWith string `json:"startsWith,omitempty"`
	EndsWith   string `json:"endsWith,omitempty"`
	Contains   string `json:"contains,omitempty"`
	In         []any  `json:"in,omitempty"`
}

// SortCondition for sorting search results.
type SortCondition struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // "asc" or "desc"
}

// CatalogResponse is the paginated response for catalog items.
type CatalogResponse struct {
	Data   []STACItem `json:"data"`
	Cursor string     `json:"cursor,omitempty"`
}

// PurchaseStatus represents the status of a catalog purchase.
type PurchaseStatus string

const (
	PurchaseStatusReceived PurchaseStatus = "received"
	PurchaseStatusCanceled PurchaseStatus = "canceled"
	PurchaseStatusActive   PurchaseStatus = "active"
	PurchaseStatusClosed   PurchaseStatus = "closed"
	PurchaseStatusFailed   PurchaseStatus = "failed"
)

// Purchase represents a catalog purchase.
type Purchase struct {
	ID           string         `json:"id"`
	CustomerName string         `json:"customerName,omitempty"`
	ContractName string         `json:"contractName,omitempty"`
	CreatedAt    time.Time      `json:"createdAt,omitempty"`
	Status       PurchaseStatus `json:"status"`
	Reference    string         `json:"reference,omitempty"`
}

// PurchaseRequest for creating a catalog purchase.
type PurchaseRequest struct {
	ItemIDs       []string `json:"itemIds"`
	ContractID    string   `json:"contractId"`
	Reference     string   `json:"reference,omitempty"`     // Max 256 characters UTF-8
	CompanyName   string   `json:"companyName,omitempty"`   // Recipient company name
	ContactPerson string   `json:"contactPerson,omitempty"` // Contact person name
}

// PurchaseResponse is returned when creating a purchase.
type PurchaseResponse struct {
	PurchaseID string `json:"purchaseId"`
}

// PurchasesResponse is the paginated response for listing purchases.
type PurchasesResponse struct {
	Data   []Purchase `json:"data"`
	Cursor string     `json:"cursor,omitempty"`
}

// ----------------------------------------------------------------------------
// Catalog API Methods
// Endpoints: https://docs.iceye.com/constellation/api/1.0/
// ----------------------------------------------------------------------------

const catalogBasePath = "/catalog/v1"

// ListCatalogItems lists catalog items with optional filters.
// Returns an iterator that yields pages of STAC items.
// Use the cursor from previous response to get subsequent pages.
//
// GET /catalog/v1/items
func (c *Client) ListCatalogItems(ctx context.Context, pageSize int, opts *ListItemsOptions) iter.Seq2[CatalogResponse, error] {
	return func(yield func(CatalogResponse, error) bool) {
		seq := common.Paginate(func(cur *string) ([]STACItem, *string, error) {
			u := &url.URL{Path: path.Join(catalogBasePath, "items")}
			q := u.Query()

			if pageSize > 0 {
				q.Set("limit", strconv.Itoa(pageSize))
			}
			if opts != nil {
				if len(opts.IDs) > 0 {
					q.Set("ids", strings.Join(opts.IDs, ","))
				}
				if opts.BBox != nil {
					q.Set("bbox", formatBBox(*opts.BBox))
				}
				if opts.Datetime != "" {
					q.Set("datetime", opts.Datetime)
				}
				if len(opts.SortBy) > 0 {
					q.Set("sortby", strings.Join(opts.SortBy, ","))
				}
			}
			if cur != nil && *cur != "" {
				q.Set("cursor", *cur)
			}
			u.RawQuery = q.Encode()

			var resp CatalogResponse
			err := c.do(ctx, http.MethodGet, u.String(), nil, &resp)
			return resp.Data, &resp.Cursor, err
		})
		for data, err := range seq {
			if !yield(CatalogResponse{Data: data}, err) {
				return
			}
		}
	}
}

// SearchCatalogItems performs an advanced catalog search.
// Returns an iterator that yields pages of STAC items.
// The cursor from a search response should be used with ListCatalogItems for pagination.
//
// POST /catalog/v1/search
func (c *Client) SearchCatalogItems(ctx context.Context, req *SearchRequest) iter.Seq2[CatalogResponse, error] {
	return func(yield func(CatalogResponse, error) bool) {
		// First page via POST /search
		var resp CatalogResponse
		u := &url.URL{Path: path.Join(catalogBasePath, "search")}
		err := c.do(ctx, http.MethodPost, u.String(), req, &resp)
		if !yield(resp, err) {
			return
		}
		if err != nil || resp.Cursor == "" {
			return
		}

		// Subsequent pages via GET /items with cursor
		cursor := resp.Cursor
		for cursor != "" {
			u := &url.URL{Path: path.Join(catalogBasePath, "items")}
			q := u.Query()
			q.Set("cursor", cursor)
			if req.Limit > 0 {
				q.Set("limit", strconv.Itoa(req.Limit))
			}
			u.RawQuery = q.Encode()

			var pageResp CatalogResponse
			err := c.do(ctx, http.MethodGet, u.String(), nil, &pageResp)
			if !yield(pageResp, err) {
				return
			}
			if err != nil {
				return
			}
			cursor = pageResp.Cursor
		}
	}
}

// PurchaseCatalogItems purchases catalog items.
// Returns the purchase ID for tracking.
//
// POST /catalog/v1/purchases
func (c *Client) PurchaseCatalogItems(ctx context.Context, req *PurchaseRequest) (*PurchaseResponse, error) {
	var resp PurchaseResponse
	u := &url.URL{Path: path.Join(catalogBasePath, "purchases")}
	if err := c.do(ctx, http.MethodPost, u.String(), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPurchase retrieves purchase status by ID.
//
// GET /catalog/v1/purchases/{purchaseID}
func (c *Client) GetPurchase(ctx context.Context, purchaseID string) (*Purchase, error) {
	var resp Purchase
	u := &url.URL{Path: path.Join(catalogBasePath, "purchases", purchaseID)}
	if err := c.do(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListPurchases lists all purchases for the authenticated user.
// Returns an iterator that yields pages of purchases.
//
// GET /catalog/v1/purchases
func (c *Client) ListPurchases(ctx context.Context, pageSize int) iter.Seq2[PurchasesResponse, error] {
	return func(yield func(PurchasesResponse, error) bool) {
		seq := common.Paginate(func(cur *string) ([]Purchase, *string, error) {
			u := &url.URL{Path: path.Join(catalogBasePath, "purchases")}
			q := u.Query()

			if pageSize > 0 {
				q.Set("limit", strconv.Itoa(pageSize))
			}
			if cur != nil && *cur != "" {
				q.Set("cursor", *cur)
			}
			u.RawQuery = q.Encode()

			var resp PurchasesResponse
			err := c.do(ctx, http.MethodGet, u.String(), nil, &resp)
			return resp.Data, &resp.Cursor, err
		})
		for data, err := range seq {
			if !yield(PurchasesResponse{Data: data}, err) {
				return
			}
		}
	}
}

// ListPurchasedProducts retrieves all products from a completed purchase.
// This endpoint does not support pagination.
//
// GET /catalog/v1/purchases/{purchaseID}/products
func (c *Client) ListPurchasedProducts(ctx context.Context, purchaseID string) ([]STACItem, error) {
	var resp struct {
		Data []STACItem `json:"data"`
	}
	u := &url.URL{Path: path.Join(catalogBasePath, "purchases", purchaseID, "products")}
	if err := c.do(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// formatBBox formats a bounding box as a comma-separated string.
func formatBBox(bbox BoundingBox) string {
	return strconv.FormatFloat(bbox[0], 'f', -1, 64) + "," +
		strconv.FormatFloat(bbox[1], 'f', -1, 64) + "," +
		strconv.FormatFloat(bbox[2], 'f', -1, 64) + "," +
		strconv.FormatFloat(bbox[3], 'f', -1, 64)
}
