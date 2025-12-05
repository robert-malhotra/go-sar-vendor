package capella

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/paulmach/orb/geojson"
)

// TaskingService provides tasking request operations.
type TaskingService struct {
	client *Client
}

// NewTaskingService creates a new tasking service.
func NewTaskingService(client *Client) *TaskingService {
	return &TaskingService{client: client}
}

// ----------------------------------------------------------------------------
// Collect Constraints
// ----------------------------------------------------------------------------

// CollectConstraints defines the constraints for a collection request.
type CollectConstraints struct {
	// Look direction: "left", "right", or "either"
	LookDirection *LookDirection `json:"lookDirection,omitempty"`

	// Orbit direction: "ascending", "descending", or "either"
	AscDesc *OrbitState `json:"ascDesc,omitempty"`

	// Constrained set of orbital planes
	OrbitalPlanes []OrbitalPlane `json:"orbitalPlanes,omitempty"`

	// Local time windows (seconds since midnight)
	LocalTime [][]int `json:"localTime,omitempty"`

	// Off-nadir angle constraints (degrees)
	OffNadirMin *float64 `json:"offNadirMin,omitempty"`
	OffNadirMax *float64 `json:"offNadirMax,omitempty"`

	// Azimuth angle constraints (degrees)
	AzimuthAngleMin *float64 `json:"azimuthAngleMin,omitempty"`
	AzimuthAngleMax *float64 `json:"azimuthAngleMax,omitempty"`

	// Grazing angle constraints (degrees)
	GrazingAngleMin *float64 `json:"grazingAngleMin,omitempty"`
	GrazingAngleMax *float64 `json:"grazingAngleMax,omitempty"`

	// Squint configuration
	Squint         *string `json:"squint,omitempty"`
	MaxSquintAngle *int    `json:"maxSquintAngle,omitempty"`

	// Image dimensions (meters)
	ImageLength *int `json:"imageLength,omitempty"`
	ImageWidth  *int `json:"imageWidth,omitempty"`
}

// ProcessingConfig specifies the desired product types.
type ProcessingConfig struct {
	ProductTypes []ProductType `json:"productTypes,omitempty"`
}

// ----------------------------------------------------------------------------
// Tasking Request Models
// ----------------------------------------------------------------------------

// TaskingRequestProperties defines the properties of a tasking request.
type TaskingRequestProperties struct {
	TaskingRequestName        string              `json:"taskingrequestName,omitempty"`
	TaskingRequestDescription string              `json:"taskingrequestDescription,omitempty"`
	OrgID                     string              `json:"orgId,omitempty"`
	UserID                    string              `json:"userId,omitempty"`
	WindowOpen                time.Time           `json:"windowOpen"`
	WindowClose               time.Time           `json:"windowClose"`
	CollectionTier            CollectionTier      `json:"collectionTier"`
	CollectionType            CollectionType      `json:"collectionType"`
	ArchiveHoldback           ArchiveHoldback     `json:"archiveHoldback,omitempty"`
	CollectConstraints        *CollectConstraints `json:"collectConstraints,omitempty"`
	ProcessingConfig          *ProcessingConfig   `json:"processingConfig,omitempty"`
	PreApproval               *bool               `json:"preApproval,omitempty"`
	CustomAttribute1          string              `json:"customAttribute1,omitempty"`
	CustomAttribute2          string              `json:"customAttribute2,omitempty"`
}

// TaskingRequest represents a tasking request submission.
type TaskingRequest struct {
	Type       string                   `json:"type"` // always "Feature"
	Geometry   *geojson.Geometry                 `json:"geometry"`
	Properties TaskingRequestProperties `json:"properties"`
}

// TaskingRequestPropertiesResponse extends properties with API-generated fields.
type TaskingRequestPropertiesResponse struct {
	TaskingRequestProperties
	TaskingRequestID    string              `json:"taskingrequestId"`
	Status              TaskStatus          `json:"status,omitempty"`
	ProcessingStatus    ProcessingStatus    `json:"processingStatus,omitempty"`
	AccessibilityStatus AccessibilityStatus `json:"accessibilityStatus,omitempty"`
	CollectionStatus    string              `json:"collectionStatus,omitempty"`
	CreatedAt           time.Time           `json:"createdAt,omitempty"`
	UpdatedAt           time.Time           `json:"updatedAt,omitempty"`
}

// StatusEntry represents a status history entry.
type StatusEntry struct {
	Time    time.Time  `json:"time"`
	Code    TaskStatus `json:"code"`
	Message string     `json:"message,omitempty"`
}

// ConflictingTask represents a task that conflicts with the current request.
type ConflictingTask struct {
	TaskingRequestID   string         `json:"taskingrequestId"`
	TaskingRequestName string         `json:"taskingrequestName"`
	RepeatRequestID    string         `json:"repeatrequestId,omitempty"`
	CollectionTier     CollectionTier `json:"collectionTier"`
	WindowOpen         time.Time      `json:"windowOpen"`
	WindowClose        time.Time      `json:"windowClose"`
}

// TaskingRequestResponse represents the response for a tasking request.
type TaskingRequestResponse struct {
	Type             string                           `json:"type"`
	Geometry         *geojson.Geometry                `json:"geometry,omitempty"`
	Properties       TaskingRequestPropertiesResponse `json:"properties"`
	StatusHistory    []StatusEntry                    `json:"statusHistory,omitempty"`
	ConflictingTasks []ConflictingTask                `json:"conflictingTasks,omitempty"`
}

// TaskingRequestsPagedResponse represents a paginated list of tasking requests.
type TaskingRequestsPagedResponse struct {
	Results     []TaskingRequestResponse `json:"results"`
	CurrentPage int                      `json:"currentPage"`
	TotalPages  int                      `json:"totalPages"`
	TotalItems  int                      `json:"totalItems,omitempty"`
}

// ----------------------------------------------------------------------------
// CRUD Operations
// ----------------------------------------------------------------------------

// CreateTask submits a new tasking request.
func (s *TaskingService) CreateTask(ctx context.Context, req TaskingRequest) (*TaskingRequestResponse, error) {
	// Set default type if not specified
	if req.Type == "" {
		req.Type = "Feature"
	}

	var resp TaskingRequestResponse
	if err := s.client.Do(ctx, http.MethodPost, "/task", 0, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTask retrieves a tasking request by ID.
func (s *TaskingService) GetTask(ctx context.Context, taskID string) (*TaskingRequestResponse, error) {
	var resp TaskingRequestResponse
	if err := s.client.Do(ctx, http.MethodGet, "/task/"+taskID, 0, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ApproveTask approves a tasking request (cost review).
func (s *TaskingService) ApproveTask(ctx context.Context, taskID string) (*TaskingRequestResponse, error) {
	payload := struct {
		Status TaskStatus `json:"status"`
	}{Status: TaskApproved}

	var resp TaskingRequestResponse
	if err := s.client.Do(ctx, http.MethodPatch, "/task/"+taskID, 0, payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CancelTask cancels a tasking request.
func (s *TaskingService) CancelTask(ctx context.Context, taskID string) (*TaskingRequestResponse, error) {
	payload := struct {
		Status TaskStatus `json:"status"`
	}{Status: TaskCanceled}

	var resp TaskingRequestResponse
	if err := s.client.Do(ctx, http.MethodPatch, "/task/"+taskID, 0, payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ----------------------------------------------------------------------------
// Retasking
// ----------------------------------------------------------------------------

// Retask submits a retask request for an existing task.
func (s *TaskingService) Retask(ctx context.Context, taskID string, req TaskingRequest) (*TaskingRequestResponse, error) {
	if req.Type == "" {
		req.Type = "Feature"
	}

	var resp TaskingRequestResponse
	if err := s.client.Do(ctx, http.MethodPost, "/task/"+taskID+"/retask", 0, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RetaskFromCollect submits a retask request from an archive collect.
func (s *TaskingService) RetaskFromCollect(ctx context.Context, collectID string, req TaskingRequest) (*TaskingRequestResponse, error) {
	if req.Type == "" {
		req.Type = "Feature"
	}

	var resp TaskingRequestResponse
	if err := s.client.Do(ctx, http.MethodPost, "/collects/"+collectID+"/retask", 0, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ----------------------------------------------------------------------------
// Pagination and Listing
// ----------------------------------------------------------------------------

// ListTasksParams defines parameters for listing tasking requests.
type ListTasksParams struct {
	CustomerID     string
	OrganizationID string
	ResellerID     string
	Page           int
	Limit          int
	Sort           string
	Order          string // "asc" or "desc"
}

// fetchTasksPage fetches a single page of tasking requests.
func (s *TaskingService) fetchTasksPage(ctx context.Context, params ListTasksParams) (*TaskingRequestsPagedResponse, error) {
	v := url.Values{}

	if params.CustomerID != "" {
		v.Set("customerId", params.CustomerID)
	}
	if params.OrganizationID != "" {
		v.Set("organizationId", params.OrganizationID)
	}
	if params.ResellerID != "" {
		v.Set("resellerId", params.ResellerID)
	}
	if params.Page > 0 {
		v.Set("page", strconv.Itoa(params.Page))
	}
	if params.Limit > 0 {
		v.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Sort != "" {
		v.Set("sort", params.Sort)
	}
	if params.Order != "" {
		v.Set("order", params.Order)
	}

	u := s.client.BuildURL("/tasks/paged")
	u.RawQuery = v.Encode()

	var resp TaskingRequestsPagedResponse
	if err := s.client.doRequest(ctx, http.MethodGet, u, nil, 0, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ListTasks returns an iterator over all tasking requests with automatic pagination.
func (s *TaskingService) ListTasks(ctx context.Context, params ListTasksParams) iter.Seq2[TaskingRequestResponse, error] {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 25
	}

	return func(yield func(TaskingRequestResponse, error) bool) {
		page := params.Page

		for {
			params.Page = page

			resp, err := s.fetchTasksPage(ctx, params)
			if err != nil {
				yield(TaskingRequestResponse{}, err)
				return
			}

			for _, task := range resp.Results {
				if !yield(task, nil) {
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

// ----------------------------------------------------------------------------
// Search
// ----------------------------------------------------------------------------

// TaskSearchRequest defines the search request body.
type TaskSearchRequest struct {
	Sort  string `json:"sort,omitempty"`
	Order string `json:"order,omitempty"` // "asc" or "desc"
	Page  int    `json:"page,omitempty"`
	Limit int    `json:"limit,omitempty"`
	Query any    `json:"query,omitempty"`
}

// SearchTasks performs an advanced search on tasking requests.
func (s *TaskingService) SearchTasks(ctx context.Context, req TaskSearchRequest) (*TaskingRequestsPagedResponse, error) {
	var resp TaskingRequestsPagedResponse
	if err := s.client.Do(ctx, http.MethodPost, "/tasks/search", 0, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SearchTasksIterator returns an iterator over search results with automatic pagination.
func (s *TaskingService) SearchTasksIterator(ctx context.Context, req TaskSearchRequest) iter.Seq2[TaskingRequestResponse, error] {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 25
	}

	return func(yield func(TaskingRequestResponse, error) bool) {
		page := req.Page

		for {
			req.Page = page

			resp, err := s.SearchTasks(ctx, req)
			if err != nil {
				yield(TaskingRequestResponse{}, err)
				return
			}

			for _, task := range resp.Results {
				if !yield(task, nil) {
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

// ----------------------------------------------------------------------------
// Collection Types
// ----------------------------------------------------------------------------

// CollectionTypeInfo represents information about a collection type.
type CollectionTypeInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Resolution  float64 `json:"resolution,omitempty"`
	MinArea     float64 `json:"minArea,omitempty"`
	MaxArea     float64 `json:"maxArea,omitempty"`
}

// GetCollectionTypes retrieves available collection types.
func (s *TaskingService) GetCollectionTypes(ctx context.Context) ([]CollectionTypeInfo, error) {
	var resp []CollectionTypeInfo
	if err := s.client.Do(ctx, http.MethodGet, "/collectiontypes", 0, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ----------------------------------------------------------------------------
// Repeat Requests
// ----------------------------------------------------------------------------

// RepeatRequestProperties defines the properties of a repeat request.
type RepeatRequestProperties struct {
	RepeatRequestName        string              `json:"repeatrequestName,omitempty"`
	RepeatRequestDescription string              `json:"repeatrequestDescription,omitempty"`
	OrgID                    string              `json:"orgId,omitempty"`
	UserID                   string              `json:"userId,omitempty"`
	WindowOpen               time.Time           `json:"windowOpen"`
	WindowClose              time.Time           `json:"windowClose"`
	RepeatInterval           int                 `json:"repeatInterval"` // Days between collections
	CollectionTier           CollectionTier      `json:"collectionTier"`
	CollectionType           CollectionType      `json:"collectionType"`
	CollectConstraints       *CollectConstraints `json:"collectConstraints,omitempty"`
	ProcessingConfig         *ProcessingConfig   `json:"processingConfig,omitempty"`
}

// RepeatRequest represents a repeat tasking request.
type RepeatRequest struct {
	Type       string                  `json:"type"` // always "Feature"
	Geometry   *geojson.Geometry                `json:"geometry"`
	Properties RepeatRequestProperties `json:"properties"`
}

// RepeatRequestPropertiesResponse extends properties with API-generated fields.
type RepeatRequestPropertiesResponse struct {
	RepeatRequestProperties
	RepeatRequestID string     `json:"repeatrequestId"`
	Status          TaskStatus `json:"status,omitempty"`
	CreatedAt       time.Time  `json:"createdAt,omitempty"`
	UpdatedAt       time.Time  `json:"updatedAt,omitempty"`
}

// RepeatRequestResponse represents the response for a repeat request.
type RepeatRequestResponse struct {
	Type       string                          `json:"type"`
	Geometry   *geojson.Geometry                        `json:"geometry"`
	Properties RepeatRequestPropertiesResponse `json:"properties"`
}

// CreateRepeatRequest submits a new repeat tasking request.
func (s *TaskingService) CreateRepeatRequest(ctx context.Context, req RepeatRequest) (*RepeatRequestResponse, error) {
	if req.Type == "" {
		req.Type = "Feature"
	}

	var resp RepeatRequestResponse
	if err := s.client.Do(ctx, http.MethodPost, "/repeat-requests", 0, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CancelRepeatRequest cancels a repeat tasking request.
func (s *TaskingService) CancelRepeatRequest(ctx context.Context, repeatRequestID string) (*RepeatRequestResponse, error) {
	payload := struct {
		Status TaskStatus `json:"status"`
	}{Status: TaskCanceled}

	var resp RepeatRequestResponse
	if err := s.client.Do(ctx, http.MethodPatch, "/repeat-requests/"+repeatRequestID, 0, payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ----------------------------------------------------------------------------
// Builder Pattern
// ----------------------------------------------------------------------------

// TaskingRequestBuilder provides a fluent API for building tasking requests.
type TaskingRequestBuilder struct {
	req TaskingRequest
}

// NewTaskingRequestBuilder creates a new tasking request builder.
func NewTaskingRequestBuilder() *TaskingRequestBuilder {
	return &TaskingRequestBuilder{
		req: TaskingRequest{
			Type: "Feature",
		},
	}
}

// Point sets a point geometry for the request.
func (b *TaskingRequestBuilder) Point(lon, lat float64) *TaskingRequestBuilder {
	b.req.Geometry = Point(lon, lat)
	return b
}

// Polygon sets a polygon geometry for the request.
func (b *TaskingRequestBuilder) Polygon(coordinates [][][]float64) *TaskingRequestBuilder {
	b.req.Geometry = Polygon(coordinates)
	return b
}

// BBox sets a bounding box geometry for the request.
func (b *TaskingRequestBuilder) BBox(bbox BoundingBox) *TaskingRequestBuilder {
	b.req.Geometry = BBoxToPolygon(bbox)
	return b
}

// Name sets the tasking request name.
func (b *TaskingRequestBuilder) Name(name string) *TaskingRequestBuilder {
	b.req.Properties.TaskingRequestName = name
	return b
}

// Description sets the tasking request description.
func (b *TaskingRequestBuilder) Description(desc string) *TaskingRequestBuilder {
	b.req.Properties.TaskingRequestDescription = desc
	return b
}

// Window sets the collection time window.
func (b *TaskingRequestBuilder) Window(open, close time.Time) *TaskingRequestBuilder {
	b.req.Properties.WindowOpen = open
	b.req.Properties.WindowClose = close
	return b
}

// Tier sets the collection tier.
func (b *TaskingRequestBuilder) Tier(tier CollectionTier) *TaskingRequestBuilder {
	b.req.Properties.CollectionTier = tier
	return b
}

// Type sets the collection type.
func (b *TaskingRequestBuilder) Type(ct CollectionType) *TaskingRequestBuilder {
	b.req.Properties.CollectionType = ct
	return b
}

// PreApproval sets the pre-approval flag.
func (b *TaskingRequestBuilder) PreApproval(preApproval bool) *TaskingRequestBuilder {
	b.req.Properties.PreApproval = &preApproval
	return b
}

// Constraints sets the collect constraints.
func (b *TaskingRequestBuilder) Constraints(constraints CollectConstraints) *TaskingRequestBuilder {
	b.req.Properties.CollectConstraints = &constraints
	return b
}

// Products sets the desired product types.
func (b *TaskingRequestBuilder) Products(products ...ProductType) *TaskingRequestBuilder {
	b.req.Properties.ProcessingConfig = &ProcessingConfig{
		ProductTypes: products,
	}
	return b
}

// Build returns the constructed TaskingRequest.
func (b *TaskingRequestBuilder) Build() TaskingRequest {
	return b.req
}
