package capella

import (
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

// ---------------------------------------------------------------
// Tasking – data models, listing & searching helpers
// ---------------------------------------------------------------

// ---------- ENUMS ------------------------------------------------

type CollectionTier string

const (
	TierUrgent   CollectionTier = "urgent"
	TierPriority CollectionTier = "priority"
	TierStandard CollectionTier = "standard"
	TierFlexible CollectionTier = "flexible"
	TierRoutine  CollectionTier = "routine"
)

type ArchiveHoldback string

const (
	ArchiveNone      ArchiveHoldback = "none"
	Archive30Day     ArchiveHoldback = "30_day"
	Archive1Year     ArchiveHoldback = "1_year"
	ArchivePermanent ArchiveHoldback = "permanent"
)

type TaskStatus string

const (
	StatusReceived  TaskStatus = "received"
	StatusReview    TaskStatus = "review"
	StatusSubmitted TaskStatus = "submitted"
	StatusActive    TaskStatus = "active"
	StatusAccepted  TaskStatus = "accepted"
	StatusRejected  TaskStatus = "rejected"
	StatusExpired   TaskStatus = "expired"
	StatusCompleted TaskStatus = "completed"
	StatusCanceled  TaskStatus = "canceled"
	StatusError     TaskStatus = "error"
	StatusAnomaly   TaskStatus = "anomaly"
	StatusRetrying  TaskStatus = "retrying"
	StatusApproved  TaskStatus = "approved" // cost-review transition
)

// ---------- PRODUCT TYPES ---------------------------------------

type ProductType string

const (
	ProductSLC  ProductType = "SLC"
	ProductGEO  ProductType = "GEO"
	ProductGEC  ProductType = "GEC"
	ProductSICD ProductType = "SICD"
	ProductSIDD ProductType = "SIDD"
	ProductCPHD ProductType = "CPHD"
	ProductVC   ProductType = "VC"
)

// ---------- CONSTRAINTS / PROCESSING ----------------------------

type CollectConstraints struct {
	LookDirection   *string  `json:"lookDirection,omitempty"`
	AscDesc         *string  `json:"ascDesc,omitempty"`
	OrbitalPlanes   []int    `json:"orbitalPlanes,omitempty"`
	OffNadirMin     *float64 `json:"offNadirMin,omitempty"`
	OffNadirMax     *float64 `json:"offNadirMax,omitempty"`
	AzimuthAngleMin *float64 `json:"azimuthAngleMin,omitempty"`
	AzimuthAngleMax *float64 `json:"azimuthAngleMax,omitempty"`
	GrazingAngleMin *float64 `json:"grazingAngleMin,omitempty"`
	GrazingAngleMax *float64 `json:"grazingAngleMax,omitempty"`
	Squint          *string  `json:"squint,omitempty"`
	MaxSquintAngle  *int     `json:"maxSquintAngle,omitempty"`
}

type ProcessingConfig struct {
	ProductTypes []ProductType `json:"productTypes,omitempty"`
}

// ---------- REQUEST / RESPONSE MODELS ---------------------------

type TaskingRequestProperties struct {
	TaskingRequestName        string              `json:"taskingrequestName,omitempty"`
	TaskingRequestDescription string              `json:"taskingrequestDescription,omitempty"`
	OrgID                     string              `json:"orgId"`
	UserID                    string              `json:"userId"`
	WindowOpen                time.Time           `json:"windowOpen"`
	WindowClose               time.Time           `json:"windowClose"`
	CollectionTier            CollectionTier      `json:"collectionTier"`
	ArchiveHoldback           ArchiveHoldback     `json:"archiveHoldback,omitempty"`
	CollectConstraints        *CollectConstraints `json:"collectConstraints,omitempty"`
	CollectionType            string              `json:"collectionType"`
	ProcessingConfig          *ProcessingConfig   `json:"processingConfig,omitempty"`
	PreApproval               *bool               `json:"preApproval,omitempty"`
	CustomAttribute1          string              `json:"customAttribute1,omitempty"`
	CustomAttribute2          string              `json:"customAttribute2,omitempty"`
}

type TaskingRequest struct {
	Type       string                   `json:"type"` // always "Feature"
	Geometry   GeoJSONGeometry          `json:"geometry"`
	Properties TaskingRequestProperties `json:"properties"`
}

type StatusEntry struct {
	Time    time.Time  `json:"time"`
	Code    TaskStatus `json:"code"`
	Message string     `json:"message"`
}

type ConflictingTask struct {
	TaskingRequestID   string         `json:"taskingrequestId"`
	TaskingRequestName string         `json:"taskingrequestName"`
	RepeatRequestID    string         `json:"repeatrequestId"`
	CollectionTier     CollectionTier `json:"collectionTier"`
	WindowOpen         time.Time      `json:"windowOpen"`
	WindowClose        time.Time      `json:"windowClose"`
}

type TaskingRequestPropertiesResponse struct {
	TaskingRequestProperties

	TaskingRequestID    string              `json:"taskingrequestId"`
	ProcessingStatus    ProcessingStatus    `json:"processingStatus"`
	AccessibilityStatus AccessibilityStatus `json:"accessibilityStatus,omitempty"`
	CollectionStatus    string              `json:"collectionStatus,omitempty"`
}

type TaskingRequestResponse struct {
	Type             string                           `json:"type"`
	Geometry         GeoJSONGeometry                  `json:"geometry"`
	Properties       TaskingRequestPropertiesResponse `json:"properties"`
	StatusHistory    []StatusEntry                    `json:"statusHistory,omitempty"`
	ConflictingTasks *[]ConflictingTask               `json:"conflictingTasks,omitempty"`
}

type TaskingRequestsPagedResponse struct {
	Results     []TaskingRequestResponse `json:"results"`
	CurrentPage int                      `json:"currentPage"`
	TotalPages  int                      `json:"totalPages"`
}

// ----------------------------------------------------------------
// BASIC CRUD HELPERS
// ----------------------------------------------------------------

func (c *Client) CreateTaskingRequest(apiKey string, req TaskingRequest) (*TaskingRequestResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	h, err := c.newRequest(apiKey, http.MethodPost, "/task", body)
	if err != nil {
		return nil, err
	}
	var resp TaskingRequestResponse
	return &resp, c.do(h, &resp)
}

func (c *Client) GetTaskingRequest(apiKey, id string) (*TaskingRequestResponse, error) {
	h, err := c.newRequest(apiKey, http.MethodGet, path.Join("/task", id), nil)
	if err != nil {
		return nil, err
	}
	var resp TaskingRequestResponse
	return &resp, c.do(h, &resp)
}

func (c *Client) ApproveTaskingRequest(apiKey, id string) (*TaskingRequestResponse, error) {
	payload := struct {
		Status TaskStatus `json:"status"`
	}{Status: StatusApproved}
	body, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}
	h, err := c.newRequest(apiKey, http.MethodPatch, path.Join("/task", id), body)
	if err != nil {
		return nil, err
	}
	var resp TaskingRequestResponse
	return &resp, c.do(h, &resp)
}

// ----------------------------------------------------------------
// PAGED LISTING – PAGE FETCH UTILITY
// ----------------------------------------------------------------

type PagedTasksParams struct {
	CustomerID     string
	OrganizationID string
	ResellerID     string
	Page           int
	Limit          int
	Sort           string
	Order          string
}

func (c *Client) fetchTasksPage(apiKey string, p PagedTasksParams) (*TaskingRequestsPagedResponse, error) {
	v := url.Values{}
	if p.CustomerID != "" {
		v.Set("customerId", p.CustomerID)
	}
	if p.OrganizationID != "" {
		v.Set("organizationId", p.OrganizationID)
	}
	if p.ResellerID != "" {
		v.Set("resellerId", p.ResellerID)
	}
	if p.Page > 0 {
		v.Set("page", strconv.Itoa(p.Page))
	}
	if p.Limit > 0 {
		v.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Sort != "" {
		v.Set("sort", p.Sort)
	}
	if p.Order != "" {
		v.Set("order", p.Order)
	}

	h, err := c.newRequest(apiKey, http.MethodGet, "/tasks/paged", nil)
	if err != nil {
		return nil, err
	}
	h.URL.RawQuery = v.Encode()

	var resp TaskingRequestsPagedResponse
	return &resp, c.do(h, &resp)
}

// ----------------------------------------------------------------
// STREAMING ITERATOR USING iter.Seq2
// ----------------------------------------------------------------
//
// ListTasksPaged returns an iter.Seq2 that lazily emits
// (TaskingRequestResponse, error) pairs:
//
//	seq := cli.ListTasksPaged(apiKey, capella.PagedTasksParams{CustomerID: "cust-1"})
//	for t, err := range iter.Pull2(seq) { … }
//
// *  The generator transparently walks through /tasks/paged until all
//    pages are exhausted.
// *  If an API call fails, it yields a nil-task with the error.
// *  Stop early by breaking the `range`; no clean-up required.

func (c *Client) ListTasksPaged(apiKey string, p PagedTasksParams) iter.Seq2[TaskingRequestResponse, error] {
	// Provide sensible defaults
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 {
		p.Limit = 25
	}

	return iter.Seq2[TaskingRequestResponse, error](func(yield func(TaskingRequestResponse, error) bool) {
		page := p.Page
		for {
			p.Page = page

			resp, err := c.fetchTasksPage(apiKey, p)
			if err != nil {
				yield(TaskingRequestResponse{}, err)
				return
			}

			for _, t := range resp.Results {
				if !yield(t, nil) { // caller said “stop”
					return
				}
			}

			if page >= resp.TotalPages {
				return // no more pages
			}
			page++
		}
	})
}

// ----------------------------------------------------------------
// ADVANCED SEARCH  (POST /tasks/search)
// ----------------------------------------------------------------

type TaskSearchRequest struct {
	Sort  string      `json:"sort,omitempty"`
	Order string      `json:"order,omitempty"` // asc|desc
	Page  int         `json:"page,omitempty"`
	Limit int         `json:"limit,omitempty"`
	Query interface{} `json:"query,omitempty"`
}

func (c *Client) SearchTasks(apiKey string, body TaskSearchRequest) (*TaskingRequestsPagedResponse, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal search body: %w", err)
	}
	h, err := c.newRequest(apiKey, http.MethodPost, "/tasks/search", b)
	if err != nil {
		return nil, err
	}
	var resp TaskingRequestsPagedResponse
	return &resp, c.do(h, &resp)
}
