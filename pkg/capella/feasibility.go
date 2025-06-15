package capella

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"
)

//--- Enums ---

type GeoJSONGeometryType string

const (
	Point        GeoJSONGeometryType = "Point"
	Polygon      GeoJSONGeometryType = "Polygon"
	MultiPolygon GeoJSONGeometryType = "MultiPolygon"
)

type ProcessingStatus string

const (
	Queued     ProcessingStatus = "queued"
	Processing ProcessingStatus = "processing"
	Completed  ProcessingStatus = "completed"
	Error      ProcessingStatus = "error"
)

type AccessibilityStatus string

const (
	Unknown      AccessibilityStatus = "unknown"
	Accessible   AccessibilityStatus = "accessible"
	Inaccessible AccessibilityStatus = "inaccessible"
	Rejected     AccessibilityStatus = "rejected"
)

//--- Data Models ---
// These are unchanged from the previous refactoring

// AccessRequest represents the payload for creating an access request.
type AccessRequest Feature[AccessRequestProperties]

// AccessRequestResponse represents the API response for an access request.
type AccessRequestResponse struct {
	Type       string                          `json:"type"`
	Geometry   GeoJSONGeometry                 `json:"geometry"`
	Properties AccessRequestPropertiesResponse `json:"properties"`
}

// GeoJSONGeometry represents a GeoJSON geometry object.
type GeoJSONGeometry struct {
	Type        GeoJSONGeometryType `json:"type"`
	Coordinates any                 `json:"coordinates"`
}

// AccessRequestProperties defines the properties of an access request.
type AccessRequestProperties struct {
	AccessRequestName        string             `json:"accessrequestName,omitempty"`
	AccessRequestDescription string             `json:"accessrequestDescription,omitempty"`
	AccessRequestType        string             `json:"accessrequestType,omitempty"`
	OrgID                    string             `json:"orgId"`
	UserID                   string             `json:"userId"`
	WindowOpen               time.Time          `json:"windowOpen"`
	WindowClose              time.Time          `json:"windowClose"`
	AccessConstraints        *AccessConstraints `json:"accessConstraints,omitempty"`
}

// AccessRequestPropertiesResponse extends properties with API-generated fields.
type AccessRequestPropertiesResponse struct {
	AccessRequestProperties
	AccessRequestID      string              `json:"accessrequestId"`
	ProcessingStatus     ProcessingStatus    `json:"processingStatus"`
	AccessibilityStatus  AccessibilityStatus `json:"accessibilityStatus"`
	AccessibilityMessage string              `json:"accessibilityMessage,omitempty"`
}

// AccessConstraints defines the constraints for an access request.
type AccessConstraints struct {
	// Cardinal look direction of the sensor: "left", "right", or "either".
	LookDirection *string `json:"lookDirection,omitempty"`

	// Pass direction: "asc", "dsc", or "either".
	AscDsc *string `json:"ascDsc,omitempty"`

	// Constrained set of orbital planes (e.g. ["A", "C"]).  Empty slice means
	// “no preference”.
	OrbitalPlanes []string `json:"orbitalPlanes,omitempty"`

	// One-or-more local-solar-time windows expressed as seconds-since-midnight,
	// e.g. [[0, 86400]] for “any time of day”.
	LocalTime [][]int `json:"localTime,omitempty"`

	// Sensor-to-nadir angle constraints (degrees).
	OffNadirMin *float64 `json:"offNadirMin,omitempty"`
	OffNadirMax *float64 `json:"offNadirMax,omitempty"`

	// Grazing-angle constraints (degrees).
	GrazingAngleMin *float64 `json:"grazingAngleMin,omitempty"`
	GrazingAngleMax *float64 `json:"grazingAngleMax,omitempty"`

	// Azimuth-angle constraints (degrees).
	AzimuthAngleMin *float64 `json:"azimuthAngleMin,omitempty"`
	AzimuthAngleMax *float64 `json:"azimuthAngleMax,omitempty"`

	// Desired scene dimensions (meters).
	ImageLength *int `json:"imageLength,omitempty"`
	ImageWidth  *int `json:"imageWidth,omitempty"`
}

// HTTPValidationError is the structure for a 422 response.
type HTTPValidationError struct {
	Detail []ValidationError `json:"detail"`
}

// ValidationError provides details about a specific validation failure.
type ValidationError struct {
	Loc  []string `json:"loc"`
	Msg  string   `json:"msg"`
	Type string   `json:"type"`
}

//--- API Methods ---

// CreateAccessRequest submits a new access request, authenticated by the provided API key.
func (c *Client) CreateAccessRequest(apiKey string, request AccessRequest) (*AccessRequestResponse, error) {
	endpoint := "/ma/accessrequests/"

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpReq, err := c.newRequest(apiKey, http.MethodPost, endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	var accessResponse AccessRequestResponse
	if err := c.do(httpReq, &accessResponse); err != nil {
		return nil, err
	}

	return &accessResponse, nil
}

// GetAccessRequest retrieves a specific access request, authenticated by the provided API key.
func (c *Client) GetAccessRequest(apiKey, accessRequestID string) (*AccessRequestResponse, error) {
	endpoint := path.Join("/ma/accessrequests/", accessRequestID)

	httpReq, err := c.newRequest(apiKey, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var accessResponse AccessRequestResponse
	if err := c.do(httpReq, &accessResponse); err != nil {
		return nil, err
	}

	return &accessResponse, nil
}
