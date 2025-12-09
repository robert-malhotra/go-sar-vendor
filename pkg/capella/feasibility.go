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
// Access Constraints
// ----------------------------------------------------------------------------

// AccessConstraints defines the constraints for an access request.
type AccessConstraints struct {
	// Look direction: "left", "right", or "either"
	LookDirection *LookDirection `json:"lookDirection,omitempty"`

	// Pass direction: "asc", "dsc", or "either"
	AscDsc *OrbitState `json:"ascDsc,omitempty"`

	// Constrained set of orbital planes (e.g., ["45", "53"])
	OrbitalPlanes []OrbitalPlane `json:"orbitalPlanes,omitempty"`

	// Local time windows (seconds since midnight)
	LocalTime [][]int `json:"localTime,omitempty"`

	// Off-nadir angle constraints (degrees)
	OffNadirMin *float64 `json:"offNadirMin,omitempty"`
	OffNadirMax *float64 `json:"offNadirMax,omitempty"`

	// Grazing angle constraints (degrees)
	GrazingAngleMin *float64 `json:"grazingAngleMin,omitempty"`
	GrazingAngleMax *float64 `json:"grazingAngleMax,omitempty"`

	// Azimuth angle constraints (degrees)
	AzimuthAngleMin *float64 `json:"azimuthAngleMin,omitempty"`
	AzimuthAngleMax *float64 `json:"azimuthAngleMax,omitempty"`

	// Desired scene dimensions (meters)
	ImageLength *int `json:"imageLength,omitempty"`
	ImageWidth  *int `json:"imageWidth,omitempty"`
}

// ----------------------------------------------------------------------------
// Access Request Models
// ----------------------------------------------------------------------------

// AccessRequestProperties defines the properties of an access request.
type AccessRequestProperties struct {
	AccessRequestName        string             `json:"accessrequestName,omitempty"`
	AccessRequestDescription string             `json:"accessrequestDescription,omitempty"`
	AccessRequestType        string             `json:"accessrequestType,omitempty"`
	OrgID                    string             `json:"orgId,omitempty"`
	UserID                   string             `json:"userId,omitempty"`
	WindowOpen               time.Time          `json:"windowOpen"`
	WindowClose              time.Time          `json:"windowClose"`
	AccessConstraints        *AccessConstraints `json:"accessConstraints,omitempty"`
}

// AccessRequest represents an access/feasibility request.
type AccessRequest struct {
	Type       string                  `json:"type"` // always "Feature"
	Geometry   *geojson.Geometry                `json:"geometry"`
	Properties AccessRequestProperties `json:"properties"`
}

// AccessRequestPropertiesResponse extends properties with API-generated fields.
type AccessRequestPropertiesResponse struct {
	AccessRequestProperties
	AccessRequestID      string              `json:"accessrequestId"`
	ProcessingStatus     ProcessingStatus    `json:"processingStatus"`
	AccessibilityStatus  AccessibilityStatus `json:"accessibilityStatus"`
	AccessibilityMessage string              `json:"accessibilityMessage,omitempty"`
	CreatedAt            time.Time           `json:"createdAt,omitempty"`
	UpdatedAt            time.Time           `json:"updatedAt,omitempty"`
}

// AccessRequestResponse represents the response for an access request.
type AccessRequestResponse struct {
	Type       string                          `json:"type"`
	Geometry   *geojson.Geometry                        `json:"geometry"`
	Properties AccessRequestPropertiesResponse `json:"properties"`
}

// AccessWindow represents an available collection window.
type AccessWindow struct {
	WindowOpen    time.Time `json:"windowOpen"`
	WindowClose   time.Time `json:"windowClose"`
	OrbitalPlane  string    `json:"orbitalPlane,omitempty"`
	LookDirection string    `json:"lookDirection,omitempty"`
	AscDesc       string    `json:"ascDesc,omitempty"`
	OffNadir      float64   `json:"offNadir,omitempty"`
	GrazingAngle  float64   `json:"grazingAngle,omitempty"`
	AzimuthAngle  float64   `json:"azimuthAngle,omitempty"`
}

// AccessRequestDetailResponse includes the access windows.
type AccessRequestDetailResponse struct {
	AccessRequestResponse
	AccessWindows []AccessWindow `json:"accessWindows,omitempty"`
}

// ----------------------------------------------------------------------------
// CRUD Operations
// ----------------------------------------------------------------------------

// CreateAccessRequest submits a new access/feasibility request.
func (c *Client) CreateAccessRequest(ctx context.Context, req AccessRequest) (*AccessRequestResponse, error) {
	if req.Type == "" {
		req.Type = "Feature"
	}

	var resp AccessRequestResponse
	if err := c.Do(ctx, http.MethodPost, "/ma/accessrequests", 0, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAccessRequest retrieves an access request by ID.
func (c *Client) GetAccessRequest(ctx context.Context, accessRequestID string) (*AccessRequestResponse, error) {
	var resp AccessRequestResponse
	if err := c.Do(ctx, http.MethodGet, "/ma/accessrequests/"+accessRequestID, 0, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAccessRequestDetail retrieves an access request with access windows.
func (c *Client) GetAccessRequestDetail(ctx context.Context, accessRequestID string) (*AccessRequestDetailResponse, error) {
	var resp AccessRequestDetailResponse
	if err := c.Do(ctx, http.MethodGet, "/ma/accessrequests/"+accessRequestID+"/detail", 0, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteAccessRequest deletes an access request.
func (c *Client) DeleteAccessRequest(ctx context.Context, accessRequestID string) error {
	return c.Do(ctx, http.MethodDelete, "/ma/accessrequests/"+accessRequestID, 0, nil, nil)
}

// ----------------------------------------------------------------------------
// Polling and Waiting
// ----------------------------------------------------------------------------

// WaitForAccessRequest polls the access request status until processing completes.
func (c *Client) WaitForAccessRequest(ctx context.Context, accessRequestID string, pollInterval time.Duration) (*AccessRequestResponse, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			resp, err := c.GetAccessRequest(ctx, accessRequestID)
			if err != nil {
				return nil, err
			}

			switch resp.Properties.ProcessingStatus {
			case ProcessingCompleted, ProcessingError:
				return resp, nil
			}
			// Continue polling for queued/processing status
		}
	}
}

// ----------------------------------------------------------------------------
// Convenience Functions
// ----------------------------------------------------------------------------

// CheckFeasibility is a convenience function that creates an access request,
// waits for processing, and returns the result.
func (c *Client) CheckFeasibility(ctx context.Context, req AccessRequest, pollInterval time.Duration) (*AccessRequestDetailResponse, error) {
	// Create the access request
	created, err := c.CreateAccessRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create access request: %w", err)
	}

	// Wait for processing to complete
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			resp, err := c.GetAccessRequest(ctx, created.Properties.AccessRequestID)
			if err != nil {
				return nil, err
			}

			switch resp.Properties.ProcessingStatus {
			case ProcessingCompleted:
				// Get detailed response with access windows
				return c.GetAccessRequestDetail(ctx, created.Properties.AccessRequestID)
			case ProcessingError:
				return nil, fmt.Errorf("access request processing failed: %s", resp.Properties.AccessibilityMessage)
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Builder Pattern
// ----------------------------------------------------------------------------

// AccessRequestBuilder provides a fluent API for building access requests.
type AccessRequestBuilder struct {
	req AccessRequest
}

// NewAccessRequestBuilder creates a new access request builder.
func NewAccessRequestBuilder() *AccessRequestBuilder {
	return &AccessRequestBuilder{
		req: AccessRequest{
			Type: "Feature",
		},
	}
}

// Point sets a point geometry for the request.
func (b *AccessRequestBuilder) Point(lon, lat float64) *AccessRequestBuilder {
	b.req.Geometry = Point(lon, lat)
	return b
}

// Polygon sets a polygon geometry for the request.
func (b *AccessRequestBuilder) Polygon(coordinates [][][]float64) *AccessRequestBuilder {
	b.req.Geometry = Polygon(coordinates)
	return b
}

// BBox sets a bounding box geometry for the request.
func (b *AccessRequestBuilder) BBox(bbox BoundingBox) *AccessRequestBuilder {
	b.req.Geometry = BBoxToPolygon(bbox)
	return b
}

// Name sets the access request name.
func (b *AccessRequestBuilder) Name(name string) *AccessRequestBuilder {
	b.req.Properties.AccessRequestName = name
	return b
}

// Description sets the access request description.
func (b *AccessRequestBuilder) Description(desc string) *AccessRequestBuilder {
	b.req.Properties.AccessRequestDescription = desc
	return b
}

// Window sets the access time window.
func (b *AccessRequestBuilder) Window(open, close time.Time) *AccessRequestBuilder {
	b.req.Properties.WindowOpen = open
	b.req.Properties.WindowClose = close
	return b
}

// Constraints sets the access constraints.
func (b *AccessRequestBuilder) Constraints(constraints AccessConstraints) *AccessRequestBuilder {
	b.req.Properties.AccessConstraints = &constraints
	return b
}

// LookDirection sets the look direction constraint.
func (b *AccessRequestBuilder) LookDirection(dir LookDirection) *AccessRequestBuilder {
	if b.req.Properties.AccessConstraints == nil {
		b.req.Properties.AccessConstraints = &AccessConstraints{}
	}
	b.req.Properties.AccessConstraints.LookDirection = &dir
	return b
}

// OrbitDirection sets the orbit direction constraint.
func (b *AccessRequestBuilder) OrbitDirection(dir OrbitState) *AccessRequestBuilder {
	if b.req.Properties.AccessConstraints == nil {
		b.req.Properties.AccessConstraints = &AccessConstraints{}
	}
	b.req.Properties.AccessConstraints.AscDsc = &dir
	return b
}

// OffNadirRange sets the off-nadir angle range constraint.
func (b *AccessRequestBuilder) OffNadirRange(min, max float64) *AccessRequestBuilder {
	if b.req.Properties.AccessConstraints == nil {
		b.req.Properties.AccessConstraints = &AccessConstraints{}
	}
	b.req.Properties.AccessConstraints.OffNadirMin = &min
	b.req.Properties.AccessConstraints.OffNadirMax = &max
	return b
}

// Build returns the constructed AccessRequest.
func (b *AccessRequestBuilder) Build() AccessRequest {
	return b.req
}

// ----------------------------------------------------------------------------
// Listing
// ----------------------------------------------------------------------------

// ListAccessRequestsParams defines parameters for listing access requests.
type ListAccessRequestsParams struct {
	Page  int
	Limit int
}

// AccessRequestsPagedResponse represents a paginated list of access requests.
type AccessRequestsPagedResponse struct {
	Results     []AccessRequestResponse `json:"results"`
	CurrentPage int                     `json:"currentPage"`
	TotalPages  int                     `json:"totalPages"`
	TotalItems  int                     `json:"totalItems,omitempty"`
}

// ListAccessRequests returns an iterator over all access requests.
func (c *Client) ListAccessRequests(ctx context.Context, params ListAccessRequestsParams) iter.Seq2[AccessRequestResponse, error] {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 25
	}

	return func(yield func(AccessRequestResponse, error) bool) {
		page := params.Page

		for {
			path := fmt.Sprintf("/ma/accessrequests?page=%d&limit=%d", page, params.Limit)
			var resp AccessRequestsPagedResponse
			if err := c.Do(ctx, http.MethodGet, path, 0, nil, &resp); err != nil {
				yield(AccessRequestResponse{}, err)
				return
			}

			for _, item := range resp.Results {
				if !yield(item, nil) {
					return
				}
			}

			if page >= resp.TotalPages {
				return
			}
			page++
		}
	}
}
