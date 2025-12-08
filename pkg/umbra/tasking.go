package umbra

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

// TaskStatus represents the lifecycle status of a task.
type TaskStatus string

const (
	TaskStatusReceived        TaskStatus = "RECEIVED"
	TaskStatusSubmitted       TaskStatus = "SUBMITTED"
	TaskStatusReview          TaskStatus = "REVIEW"
	TaskStatusAccepted        TaskStatus = "ACCEPTED"
	TaskStatusActive          TaskStatus = "ACTIVE"
	TaskStatusScheduled       TaskStatus = "SCHEDULED"
	TaskStatusRejected        TaskStatus = "REJECTED"
	TaskStatusExpired         TaskStatus = "EXPIRED"
	TaskStatusTasked          TaskStatus = "TASKED"
	TaskStatusTransmitted     TaskStatus = "TRANSMITTED"
	TaskStatusIncomplete      TaskStatus = "INCOMPLETE"
	TaskStatusProcessing      TaskStatus = "PROCESSING"
	TaskStatusProcessed       TaskStatus = "PROCESSED"
	TaskStatusDelivering      TaskStatus = "DELIVERING"
	TaskStatusDelivered       TaskStatus = "DELIVERED"
	TaskStatusCancelRequested TaskStatus = "CANCEL_REQUESTED"
	TaskStatusCanceled        TaskStatus = "CANCELED"
	TaskStatusError           TaskStatus = "ERROR"
	TaskStatusAnomaly         TaskStatus = "ANOMALY"
	TaskStatusCompleted       TaskStatus = "COMPLETED"
)

// IsTerminal returns true if the status is a terminal state.
func (s TaskStatus) IsTerminal() bool {
	switch s {
	case TaskStatusDelivered, TaskStatusRejected, TaskStatusCanceled,
		TaskStatusError, TaskStatusAnomaly, TaskStatusCompleted, TaskStatusExpired:
		return true
	}
	return false
}

// StatusChange represents a historical status transition.
type StatusChange struct {
	Status    TaskStatus `json:"status"`
	Timestamp time.Time  `json:"timestamp"`
}

// Task represents a tasking request to the Umbra constellation.
type Task struct {
	ID                   string                `json:"id"`
	TaskName             string                `json:"taskName,omitempty"`
	UserOrderID          string                `json:"userOrderId,omitempty"`
	Status               TaskStatus            `json:"status"`
	ImagingMode          ImagingMode           `json:"imagingMode"`
	SpotlightConstraints *SpotlightConstraints `json:"spotlightConstraints,omitempty"`
	ScanConstraints      *ScanConstraints      `json:"scanConstraints,omitempty"`
	WindowStartAt        time.Time             `json:"windowStartAt"`
	WindowEndAt          time.Time             `json:"windowEndAt"`
	DeliveryConfigID     string                `json:"deliveryConfigId,omitempty"`
	ProductTypes         []ProductType         `json:"productTypes,omitempty"`
	CollectIDs           []string              `json:"collectIds,omitempty"`
	CreatedAt            time.Time             `json:"createdAt"`
	UpdatedAt            time.Time             `json:"updatedAt"`
	StatusHistory        []StatusChange        `json:"statusHistory,omitempty"`
	OrganizationID       string                `json:"organizationId,omitempty"`
	UserID               string                `json:"userId,omitempty"`
	SatelliteIDs         []string              `json:"satelliteIds,omitempty"`
	Tags                 []string              `json:"tags,omitempty"`
}

// CreateTaskRequest contains parameters for creating a new task.
type CreateTaskRequest struct {
	TaskName             string                `json:"taskName,omitempty"`
	UserOrderID          string                `json:"userOrderId,omitempty"`
	ImagingMode          ImagingMode           `json:"imagingMode"`
	SpotlightConstraints *SpotlightConstraints `json:"spotlightConstraints,omitempty"`
	ScanConstraints      *ScanConstraints      `json:"scanConstraints,omitempty"`
	WindowStartAt        time.Time             `json:"windowStartAt"`
	WindowEndAt          time.Time             `json:"windowEndAt"`
	DeliveryConfigID     string                `json:"deliveryConfigId,omitempty"`
	ProductTypes         []ProductType         `json:"productTypes,omitempty"`
	SatelliteIDs         []string              `json:"satelliteIds,omitempty"`
	Tags                 []string              `json:"tags,omitempty"`
}

// ListTasksOptions contains optional filters for listing tasks.
type ListTasksOptions struct {
	Status        []TaskStatus `url:"status,omitempty"`
	Limit         int          `url:"limit,omitempty"`
	Offset        int          `url:"offset,omitempty"`
	CreatedAfter  *time.Time   `url:"createdAfter,omitempty"`
	CreatedBefore *time.Time   `url:"createdBefore,omitempty"`
}

// TaskListResponse contains a paginated list of tasks.
type TaskListResponse struct {
	Tasks      []Task `json:"tasks"`
	TotalCount int    `json:"totalCount"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

// CreateTask creates a new task.
// POST /tasking/tasks
func (c *Client) CreateTask(ctx context.Context, req *CreateTaskRequest) (*Task, error) {
	body, err := common.MarshalBody(req)
	if err != nil {
		return nil, err
	}
	var t Task
	err = c.DoRaw(ctx, http.MethodPost, c.BaseURL().JoinPath("tasking", "tasks"), body, http.StatusCreated, &t)
	return &t, err
}

// GetTask retrieves a task by ID.
// GET /tasking/tasks/{id}
func (c *Client) GetTask(ctx context.Context, id string) (*Task, error) {
	var t Task
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("tasking", "tasks", id), nil, http.StatusOK, &t)
	return &t, err
}

// ListTasks retrieves all tasks with optional filtering.
// GET /tasking/tasks
func (c *Client) ListTasks(ctx context.Context, opts *ListTasksOptions) (*TaskListResponse, error) {
	u := c.BaseURL().JoinPath("tasking", "tasks")
	if opts != nil {
		q := u.Query()
		if opts.Limit > 0 {
			q.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Offset > 0 {
			q.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
		for _, s := range opts.Status {
			q.Add("status", string(s))
		}
		if opts.CreatedAfter != nil {
			q.Set("createdAfter", opts.CreatedAfter.Format(time.RFC3339))
		}
		if opts.CreatedBefore != nil {
			q.Set("createdBefore", opts.CreatedBefore.Format(time.RFC3339))
		}
		u.RawQuery = q.Encode()
	}
	var resp TaskListResponse
	err := c.DoRaw(ctx, http.MethodGet, u, nil, http.StatusOK, &resp)
	return &resp, err
}

// CancelTask cancels an active task.
// PATCH /tasking/tasks/{id}/cancel
func (c *Client) CancelTask(ctx context.Context, id string) (*Task, error) {
	var t Task
	err := c.DoRaw(ctx, http.MethodPatch, c.BaseURL().JoinPath("tasking", "tasks", id, "cancel"), nil, http.StatusOK, &t)
	return &t, err
}

// TaskSearchRequest contains parameters for searching tasks.
type TaskSearchRequest struct {
	Limit  *int                   `json:"limit,omitempty"`
	Skip   *int                   `json:"skip,omitempty"`
	Query  map[string]interface{} `json:"query,omitempty"`
	SortBy string                 `json:"sortBy,omitempty"`
	Order  string                 `json:"order,omitempty"`
}

// SearchTasks returns an iterator over task search results with automatic pagination.
// POST /tasking/tasks/search
func (c *Client) SearchTasks(ctx context.Context, req TaskSearchRequest) iter.Seq2[Task, error] {
	if req.Limit == nil {
		limit := defaultSearchLimit
		req.Limit = &limit
	}
	if req.Skip == nil {
		skip := 0
		req.Skip = &skip
	}

	return func(yield func(Task, error) bool) {
		for {
			body, err := common.MarshalBody(req)
			if err != nil {
				var zero Task
				yield(zero, err)
				return
			}

			var tasks []Task
			err = c.DoRaw(ctx, http.MethodPost, c.BaseURL().JoinPath("tasking", "tasks", "search"), body, http.StatusOK, &tasks)
			if err != nil {
				var zero Task
				yield(zero, err)
				return
			}

			for _, task := range tasks {
				if !yield(task, nil) {
					return
				}
			}

			// If we got fewer results than the limit, we've reached the end
			if len(tasks) < *req.Limit {
				return
			}

			// Advance to next page
			*req.Skip += *req.Limit
		}
	}
}

// TaskStatusCallback is called for each status update during polling.
type TaskStatusCallback func(task *Task)

// WaitForTaskStatus polls until the task reaches the target status or a terminal state.
func (c *Client) WaitForTaskStatus(ctx context.Context, taskID string, targetStatus TaskStatus, opts *WaitOptions) (*Task, error) {
	if opts == nil {
		opts = &WaitOptions{
			PollInterval: 30 * time.Second,
			Timeout:      24 * time.Hour,
		}
	}

	deadline := time.Now().Add(opts.Timeout)
	ticker := time.NewTicker(opts.PollInterval)
	defer ticker.Stop()

	for {
		t, err := c.GetTask(ctx, taskID)
		if err != nil {
			return nil, err
		}

		if t.Status == targetStatus || t.Status.IsTerminal() {
			return t, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for task %s to reach status %s", taskID, targetStatus)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

// WaitForTaskDelivery polls until the task is delivered or fails.
func (c *Client) WaitForTaskDelivery(ctx context.Context, taskID string, opts *WaitOptions) (*Task, error) {
	return c.WaitForTaskStatus(ctx, taskID, TaskStatusDelivered, opts)
}

// WaitForTaskDeliveryWithCallback polls with status callbacks.
func (c *Client) WaitForTaskDeliveryWithCallback(ctx context.Context, taskID string, callback TaskStatusCallback, opts *WaitOptions) (*Task, error) {
	if opts == nil {
		opts = &WaitOptions{
			PollInterval: 30 * time.Second,
			Timeout:      24 * time.Hour,
		}
	}

	deadline := time.Now().Add(opts.Timeout)
	ticker := time.NewTicker(opts.PollInterval)
	defer ticker.Stop()

	var lastStatus TaskStatus

	for {
		t, err := c.GetTask(ctx, taskID)
		if err != nil {
			return nil, err
		}

		if t.Status != lastStatus {
			lastStatus = t.Status
			if callback != nil {
				callback(t)
			}
		}

		if t.Status == TaskStatusDelivered || t.Status.IsTerminal() {
			return t, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for task %s delivery", taskID)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
