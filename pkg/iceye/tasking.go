package iceye

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

// ----------------------------------------------------------------------------
// Tasking API Types
// ----------------------------------------------------------------------------

// ImagingMode represents available satellite imaging modes.
type ImagingMode string

const (
	ImagingModeSpotlight ImagingMode = "SPOTLIGHT"
	ImagingModeStripmap  ImagingMode = "STRIPMAP"
	ImagingModeScan      ImagingMode = "SCAN"
)

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

const (
	TaskStatusReceived  TaskStatus = "RECEIVED"
	TaskStatusActive    TaskStatus = "ACTIVE"
	TaskStatusRejected  TaskStatus = "REJECTED"
	TaskStatusFulfilled TaskStatus = "FULFILLED"
	TaskStatusDone      TaskStatus = "DONE"
	TaskStatusCanceled  TaskStatus = "CANCELED"
	TaskStatusFailed    TaskStatus = "FAILED"
)

// Priority represents task scheduling priority.
type Priority string

const (
	PriorityBackground Priority = "BACKGROUND"
	PriorityCommercial Priority = "COMMERCIAL"
)

// SLA represents service level agreement options.
type SLA string

const (
	SLA8Hours SLA = "SLA_8H"
	SLA3Hours SLA = "SLA_3H"
)

// Exclusivity represents data exclusivity options.
type Exclusivity string

const (
	ExclusivityPublic  Exclusivity = "PUBLIC"
	ExclusivityPrivate Exclusivity = "PRIVATE"
)

// EULA represents end-user license agreement types.
type EULA string

const (
	EULAStandard   EULA = "STANDARD"
	EULAGovernment EULA = "GOVERNMENT"
	EULAMulti      EULA = "MULTI"
)

// LookSide represents satellite look direction.
type LookSide string

const (
	LookSideAny   LookSide = "ANY"
	LookSideLeft  LookSide = "LEFT"
	LookSideRight LookSide = "RIGHT"
)

// PassDirection represents orbital pass direction.
type PassDirection string

const (
	PassDirectionAny        PassDirection = "ANY"
	PassDirectionAscending  PassDirection = "ASCENDING"
	PassDirectionDescending PassDirection = "DESCENDING"
)

// AdditionalProductType represents additional SAR product types beyond standard.
type AdditionalProductType string

const (
	AdditionalProductTypeSIDD AdditionalProductType = "SIDD"
	AdditionalProductTypeSICD AdditionalProductType = "SICD"
)

// Task represents a satellite imaging task.
type Task struct {
	ID                     string                  `json:"id"`
	ContractID             string                  `json:"contractID"`
	PointOfInterest        Point                   `json:"pointOfInterest"`
	AcquisitionWindow      TimeWindow              `json:"acquisitionWindow"`
	ImagingMode            string                  `json:"imagingMode"`
	Status                 TaskStatus              `json:"status"`
	Exclusivity            Exclusivity             `json:"exclusivity,omitempty"`
	Priority               Priority                `json:"priority,omitempty"`
	SLA                    string                  `json:"sla,omitempty"`
	EULA                   EULA                    `json:"eula,omitempty"`
	AdditionalProductTypes []AdditionalProductType `json:"additionalProductTypes,omitempty"`
	IncidenceAngle         *IncidenceAngle         `json:"incidenceAngle,omitempty"`
	LookSide               LookSide                `json:"lookSide,omitempty"`
	PassDirection          PassDirection           `json:"passDirection,omitempty"`
	DeliveryLocations      []DeliveryLocation      `json:"deliveryLocations,omitempty"`
	CreatedAt              time.Time               `json:"createdAt"`
	UpdatedAt              time.Time               `json:"updatedAt"`
}

// CreateTaskRequest represents parameters for creating a new task.
type CreateTaskRequest struct {
	// Required fields
	ContractID        string     `json:"contractID"`
	PointOfInterest   Point      `json:"pointOfInterest"`
	AcquisitionWindow TimeWindow `json:"acquisitionWindow"`
	ImagingMode       string     `json:"imagingMode"`

	// Optional fields
	Exclusivity            Exclusivity             `json:"exclusivity,omitempty"`
	Priority               Priority                `json:"priority,omitempty"`
	SLA                    string                  `json:"sla,omitempty"`
	EULA                   EULA                    `json:"eula,omitempty"`
	AdditionalProductTypes []AdditionalProductType `json:"additionalProductTypes,omitempty"`
	IncidenceAngle         *IncidenceAngle         `json:"incidenceAngle,omitempty"`
	LookSide               LookSide                `json:"lookSide,omitempty"`
	PassDirection          PassDirection           `json:"passDirection,omitempty"`
	DeliveryLocations      []DeliveryLocation      `json:"deliveryLocations,omitempty"`
}

// TaskScene represents imaging parameters for a scheduled task.
type TaskScene struct {
	ImagingTime   TimeWindow    `json:"imagingTime"`
	Duration      int           `json:"duration"` // seconds
	LookSide      LookSide      `json:"lookSide"`
	PassDirection PassDirection `json:"passDirection"`
}

// TaskProduct represents a SAR data product from a completed task.
type TaskProduct struct {
	Type   string           `json:"type"`
	Assets map[string]Asset `json:"assets"`
}

// TaskPrice represents a price quotation for a task.
type TaskPrice struct {
	Amount   int64  `json:"amount"`   // Minor currency unit (e.g., cents)
	Currency string `json:"currency"` // ISO 4217 code
}

// ListTasksOptions for filtering task lists.
type ListTasksOptions struct {
	ContractID    string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}

// TasksResponse is the paginated response for listing tasks.
type TasksResponse struct {
	Data   []Task `json:"data"`
	Cursor string `json:"cursor,omitempty"`
}

// TaskPriceRequest contains all parameters for getting a task price quote.
type TaskPriceRequest struct {
	ContractID      string
	PointOfInterest Point
	ImagingMode     string
	Exclusivity     Exclusivity
	Priority        Priority
	SLA             string
	EULA            EULA
}

// ----------------------------------------------------------------------------
// Tasking API Methods
// Endpoints: https://docs.iceye.com/constellation/api/1.0/
// ----------------------------------------------------------------------------

const taskingBasePath = "/tasking/v1"

// CreateTask creates a new satellite imaging task.
//
// POST /tasking/v1/tasks
func (c *Client) CreateTask(ctx context.Context, req *CreateTaskRequest) (*Task, error) {
	var resp Task
	u := &url.URL{Path: path.Join(taskingBasePath, "tasks")}
	if err := c.do(ctx, http.MethodPost, u.String(), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTask retrieves a task by ID.
//
// GET /tasking/v1/tasks/{taskID}
func (c *Client) GetTask(ctx context.Context, taskID string) (*Task, error) {
	var resp Task
	u := &url.URL{Path: path.Join(taskingBasePath, "tasks", taskID)}
	if err := c.do(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListTasks lists tasks with optional filters.
// Returns an iterator that yields pages of tasks.
//
// GET /tasking/v1/tasks
func (c *Client) ListTasks(ctx context.Context, pageSize int, opts *ListTasksOptions) iter.Seq2[[]Task, error] {
	return common.Paginate(func(cur *string) ([]Task, *string, error) {
		u := &url.URL{Path: path.Join(taskingBasePath, "tasks")}
		q := u.Query()

		if pageSize > 0 {
			q.Set("limit", strconv.Itoa(pageSize))
		}
		if opts != nil {
			if opts.ContractID != "" {
				q.Set("contractID", opts.ContractID)
			}
			if opts.CreatedAfter != nil {
				q.Set("createdAfter", opts.CreatedAfter.Format(time.RFC3339))
			}
			if opts.CreatedBefore != nil {
				q.Set("createdBefore", opts.CreatedBefore.Format(time.RFC3339))
			}
		}
		if cur != nil && *cur != "" {
			q.Set("cursor", *cur)
		}
		u.RawQuery = q.Encode()

		var resp TasksResponse
		err := c.do(ctx, http.MethodGet, u.String(), nil, &resp)
		return resp.Data, &resp.Cursor, err
	})
}

// CancelTask cancels an active task by setting its status to CANCELED.
//
// PATCH /tasking/v1/tasks/{taskID}
func (c *Client) CancelTask(ctx context.Context, taskID string) (*Task, error) {
	var resp Task
	u := &url.URL{Path: path.Join(taskingBasePath, "tasks", taskID)}
	body := map[string]string{"status": string(TaskStatusCanceled)}
	if err := c.do(ctx, http.MethodPatch, u.String(), body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTaskScene retrieves planned imaging parameters for a task.
//
// GET /tasking/v1/tasks/{taskID}/scene
func (c *Client) GetTaskScene(ctx context.Context, taskID string) (*TaskScene, error) {
	var resp TaskScene
	u := &url.URL{Path: path.Join(taskingBasePath, "tasks", taskID, "scene")}
	if err := c.do(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListTaskProducts retrieves all available products for a fulfilled task.
//
// GET /tasking/v1/tasks/{taskID}/products
func (c *Client) ListTaskProducts(ctx context.Context, taskID string) ([]TaskProduct, error) {
	var resp []TaskProduct
	u := &url.URL{Path: path.Join(taskingBasePath, "tasks", taskID, "products")}
	if err := c.do(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetTaskProduct retrieves a specific product type for a task.
//
// GET /tasking/v1/tasks/{taskID}/products/{productType}
func (c *Client) GetTaskProduct(ctx context.Context, taskID string, productType string) (*TaskProduct, error) {
	var resp TaskProduct
	u := &url.URL{Path: path.Join(taskingBasePath, "tasks", taskID, "products", productType)}
	if err := c.do(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTaskPrice gets a price quotation for task parameters.
// All parameters are passed via query string.
//
// GET /tasking/v1/price
func (c *Client) GetTaskPrice(ctx context.Context, req *TaskPriceRequest) (*TaskPrice, error) {
	u := &url.URL{Path: path.Join(taskingBasePath, "price")}
	q := u.Query()

	// Required parameters per API spec
	q.Set("contractID", req.ContractID)
	q.Set("pointOfInterest[lat]", strconv.FormatFloat(req.PointOfInterest.Lat, 'f', -1, 64))
	q.Set("pointOfInterest[lon]", strconv.FormatFloat(req.PointOfInterest.Lon, 'f', -1, 64))
	q.Set("imagingMode", req.ImagingMode)
	q.Set("exclusivity", string(req.Exclusivity))
	q.Set("priority", string(req.Priority))
	q.Set("sla", req.SLA)
	q.Set("eula", string(req.EULA))

	u.RawQuery = q.Encode()

	var resp TaskPrice
	if err := c.do(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
