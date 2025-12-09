package capella

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"time"

	"github.com/paulmach/orb/geojson"
)


// ----------------------------------------------------------------------------
// STAC Item Models
// ----------------------------------------------------------------------------

// STACItem represents a STAC catalog item.
type STACItem struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"` // always "Feature"
	Geometry   *geojson.Geometry          `json:"geometry"`
	BBox       []float64         `json:"bbox,omitempty"`
	Properties STACProperties    `json:"properties"`
	Links      []Link            `json:"links,omitempty"`
	Assets     map[string]Asset  `json:"assets,omitempty"`
	Collection string            `json:"collection,omitempty"`
}

// STACProperties contains the metadata properties of a STAC item.
type STACProperties struct {
	// Core properties
	DateTime      time.Time `json:"datetime"`
	StartDateTime time.Time `json:"start_datetime,omitempty"`
	EndDateTime   time.Time `json:"end_datetime,omitempty"`
	Created       time.Time `json:"created,omitempty"`
	Updated       time.Time `json:"updated,omitempty"`
	Title         string    `json:"title,omitempty"`
	Description   string    `json:"description,omitempty"`

	// Platform properties
	Platform    string   `json:"platform,omitempty"`
	Instruments []string `json:"instruments,omitempty"`
	Constellation string `json:"constellation,omitempty"`

	// SAR properties (sar:*)
	InstrumentMode     InstrumentMode `json:"sar:instrument_mode,omitempty"`
	ProductType        ProductType    `json:"sar:product_type,omitempty"`
	Polarization       []Polarization `json:"sar:polarizations,omitempty"`
	FrequencyBand      string         `json:"sar:frequency_band,omitempty"`
	CenterFrequency    float64        `json:"sar:center_frequency,omitempty"`
	PixelSpacingRange  float64        `json:"sar:pixel_spacing_range,omitempty"`
	PixelSpacingAzimuth float64       `json:"sar:pixel_spacing_azimuth,omitempty"`
	ResolutionRange    float64        `json:"sar:resolution_range,omitempty"`
	ResolutionAzimuth  float64        `json:"sar:resolution_azimuth,omitempty"`
	LooksRange         int            `json:"sar:looks_range,omitempty"`
	LooksAzimuth       int            `json:"sar:looks_azimuth,omitempty"`
	ObservationDirection LookDirection `json:"sar:observation_direction,omitempty"`

	// Satellite properties (sat:*)
	OrbitState       OrbitState `json:"sat:orbit_state,omitempty"`
	RelativeOrbit    int        `json:"sat:relative_orbit,omitempty"`
	AbsoluteOrbit    int        `json:"sat:absolute_orbit,omitempty"`
	AnxDatetime      time.Time  `json:"sat:anx_datetime,omitempty"`

	// View properties (view:*)
	IncidenceAngle float64 `json:"view:incidence_angle,omitempty"`
	Azimuth        float64 `json:"view:azimuth,omitempty"`
	OffNadir       float64 `json:"view:off_nadir,omitempty"`

	// Capella-specific properties (capella:*)
	CollectID   string  `json:"capella:collect_id,omitempty"`
	SquintAngle float64 `json:"capella:squint_angle,omitempty"`
}

// STACCollection represents a STAC collection.
type STACCollection struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "Collection"
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Keywords    []string               `json:"keywords,omitempty"`
	License     string                 `json:"license,omitempty"`
	Providers   []STACProvider         `json:"providers,omitempty"`
	Extent      STACExtent             `json:"extent,omitempty"`
	Links       []Link                 `json:"links,omitempty"`
	Summaries   map[string]any         `json:"summaries,omitempty"`
}

// STACProvider represents a STAC data provider.
type STACProvider struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	URL         string   `json:"url,omitempty"`
}

// STACExtent represents the spatial and temporal extent of a collection.
type STACExtent struct {
	Spatial  STACSpatialExtent  `json:"spatial"`
	Temporal STACTemporalExtent `json:"temporal"`
}

// STACSpatialExtent represents the spatial extent.
type STACSpatialExtent struct {
	BBox [][]float64 `json:"bbox"`
}

// STACTemporalExtent represents the temporal extent.
type STACTemporalExtent struct {
	Interval [][]string `json:"interval"`
}

// ----------------------------------------------------------------------------
// Search Parameters
// ----------------------------------------------------------------------------

// SearchParams represents the parameters for a STAC catalog search.
type SearchParams struct {
	// Bounding box [minLon, minLat, maxLon, maxLat]
	BBox []float64 `json:"bbox,omitempty"`

	// GeoJSON geometry intersection filter
	Intersects *Geometry `json:"intersects,omitempty"`

	// Collections to search
	Collections []string `json:"collections,omitempty"`

	// Item IDs to filter
	IDs []string `json:"ids,omitempty"`

	// DateTime range (RFC 3339 format: "2024-01-01T00:00:00Z/2024-06-30T23:59:59Z")
	DateTime string `json:"datetime,omitempty"`

	// CQL2 filter expression
	Filter any `json:"filter,omitempty"`

	// Filter language ("cql2-json" or "cql2-text")
	FilterLang string `json:"filter-lang,omitempty"`

	// Query filters (property comparisons)
	Query map[string]any `json:"query,omitempty"`

	// Sort by field (prefix with "-" for descending)
	SortBy string `json:"sortby,omitempty"`

	// Maximum number of results per page
	Limit int `json:"limit,omitempty"`
}

// SearchResponse represents the response from a STAC search.
type SearchResponse struct {
	Type           string     `json:"type"` // "FeatureCollection"
	Features       []STACItem `json:"features"`
	Links          []Link     `json:"links,omitempty"`
	NumberMatched  int        `json:"numberMatched,omitempty"`
	NumberReturned int        `json:"numberReturned,omitempty"`
	Context        *SearchContext `json:"context,omitempty"`
}

// SearchContext provides pagination context in search results.
type SearchContext struct {
	Returned int    `json:"returned"`
	Matched  int    `json:"matched,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Next     string `json:"next,omitempty"`
}

// ----------------------------------------------------------------------------
// Search Methods
// ----------------------------------------------------------------------------

// CatalogSearch performs a STAC catalog search.
func (c *Client) CatalogSearch(ctx context.Context, params SearchParams) (*SearchResponse, error) {
	var resp SearchResponse
	if err := c.Do(ctx, http.MethodPost, "/catalog/search", 0, params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CatalogSearchItems returns an iterator over search results with automatic pagination.
func (c *Client) CatalogSearchItems(ctx context.Context, params SearchParams) iter.Seq2[STACItem, error] {
	if params.Limit == 0 {
		params.Limit = 100
	}

	return func(yield func(STACItem, error) bool) {
		nextURL := ""

		for {
			var resp *SearchResponse
			var err error

			if nextURL != "" {
				// Fetch next page using the link URL
				resp, err = c.fetchSearchURL(ctx, nextURL)
			} else {
				resp, err = c.CatalogSearch(ctx, params)
			}

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

			// Check for next page link
			nextURL = ""
			for _, link := range resp.Links {
				if link.Rel == "next" {
					nextURL = link.Href
					break
				}
			}

			if nextURL == "" {
				return
			}
		}
	}
}

// fetchSearchURL fetches a search result page by URL.
func (c *Client) fetchSearchURL(ctx context.Context, searchURL string) (*SearchResponse, error) {
	u, err := c.BaseURL().Parse(searchURL)
	if err != nil {
		return nil, fmt.Errorf("parse search URL: %w", err)
	}

	var resp SearchResponse
	if err := c.DoRaw(ctx, http.MethodGet, u, nil, 0, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ----------------------------------------------------------------------------
// Collection Methods
// ----------------------------------------------------------------------------

// ListCollections lists available STAC collections.
func (c *Client) ListCollections(ctx context.Context) ([]STACCollection, error) {
	var resp struct {
		Collections []STACCollection `json:"collections"`
	}
	if err := c.Do(ctx, http.MethodGet, "/catalog/collections", 0, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Collections, nil
}

// GetCollection retrieves a specific STAC collection by ID.
func (c *Client) GetCollection(ctx context.Context, collectionID string) (*STACCollection, error) {
	var resp STACCollection
	if err := c.Do(ctx, http.MethodGet, "/catalog/collections/"+collectionID, 0, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ----------------------------------------------------------------------------
// Archive Export
// ----------------------------------------------------------------------------

// ArchiveExport represents an available archive export.
type ArchiveExport struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	BBox        []float64 `json:"bbox,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt,omitempty"`
	Size        int64     `json:"size,omitempty"`
}

// PresignedURL represents a presigned download URL.
type PresignedURL struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// ListArchiveExports lists available archive footprint exports.
func (c *Client) ListArchiveExports(ctx context.Context) ([]ArchiveExport, error) {
	var resp []ArchiveExport
	if err := c.Do(ctx, http.MethodGet, "/catalog/archive-export/available", 0, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetArchiveExportURL gets a presigned download URL for an archive export.
func (c *Client) GetArchiveExportURL(ctx context.Context, exportID string) (*PresignedURL, error) {
	reqBody := map[string]string{"exportId": exportID}
	var resp PresignedURL
	if err := c.Do(ctx, http.MethodPost, "/catalog/archive-export/presigned", 0, reqBody, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ----------------------------------------------------------------------------
// Query Helpers
// ----------------------------------------------------------------------------

// QueryOp represents a query operator.
type QueryOp string

const (
	QueryEq  QueryOp = "eq"
	QueryNeq QueryOp = "neq"
	QueryLt  QueryOp = "lt"
	QueryLte QueryOp = "lte"
	QueryGt  QueryOp = "gt"
	QueryGte QueryOp = "gte"
	QueryIn  QueryOp = "in"
)

// QueryFilter creates a query filter for a property.
func QueryFilter(op QueryOp, value any) map[string]any {
	return map[string]any{string(op): value}
}

// BuildQuery builds a query map from property filters.
func BuildQuery(filters map[string]map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range filters {
		result[k] = v
	}
	return result
}

// SearchParamsBuilder provides a fluent API for building search parameters.
type SearchParamsBuilder struct {
	params SearchParams
}

// NewSearchBuilder creates a new search params builder.
func NewSearchBuilder() *SearchParamsBuilder {
	return &SearchParamsBuilder{
		params: SearchParams{
			Query: make(map[string]any),
		},
	}
}

// BBox sets the bounding box filter.
func (b *SearchParamsBuilder) BBox(minLon, minLat, maxLon, maxLat float64) *SearchParamsBuilder {
	b.params.BBox = []float64{minLon, minLat, maxLon, maxLat}
	return b
}

// Collections sets the collections to search.
func (b *SearchParamsBuilder) Collections(collections ...string) *SearchParamsBuilder {
	b.params.Collections = collections
	return b
}

// DateTime sets the datetime filter.
func (b *SearchParamsBuilder) DateTime(start, end time.Time) *SearchParamsBuilder {
	b.params.DateTime = start.Format(time.RFC3339) + "/" + end.Format(time.RFC3339)
	return b
}

// InstrumentMode filters by SAR instrument mode.
func (b *SearchParamsBuilder) InstrumentMode(mode InstrumentMode) *SearchParamsBuilder {
	b.params.Query["sar:instrument_mode"] = QueryFilter(QueryEq, string(mode))
	return b
}

// ProductType filters by SAR product type.
func (b *SearchParamsBuilder) ProductType(pt ProductType) *SearchParamsBuilder {
	b.params.Query["sar:product_type"] = QueryFilter(QueryEq, string(pt))
	return b
}

// OrbitState filters by satellite orbit state.
func (b *SearchParamsBuilder) OrbitState(state OrbitState) *SearchParamsBuilder {
	b.params.Query["sat:orbit_state"] = QueryFilter(QueryEq, string(state))
	return b
}

// LookDirection filters by observation direction.
func (b *SearchParamsBuilder) LookDirection(dir LookDirection) *SearchParamsBuilder {
	b.params.Query["sar:observation_direction"] = QueryFilter(QueryEq, string(dir))
	return b
}

// IncidenceAngle filters by incidence angle range.
func (b *SearchParamsBuilder) IncidenceAngle(min, max float64) *SearchParamsBuilder {
	b.params.Query["view:incidence_angle"] = map[string]any{
		"gte": min,
		"lte": max,
	}
	return b
}

// SortBy sets the sort field.
func (b *SearchParamsBuilder) SortBy(field string, descending bool) *SearchParamsBuilder {
	if descending {
		b.params.SortBy = "-" + field
	} else {
		b.params.SortBy = field
	}
	return b
}

// Limit sets the maximum results per page.
func (b *SearchParamsBuilder) Limit(limit int) *SearchParamsBuilder {
	b.params.Limit = limit
	return b
}

// Build returns the constructed SearchParams.
func (b *SearchParamsBuilder) Build() SearchParams {
	return b.params
}
