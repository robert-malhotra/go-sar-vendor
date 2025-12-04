package umbra

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"time"
)

// CollectStatus represents the lifecycle status of a collect.
type CollectStatus string

const (
	// Early stages
	CollectStatusScheduled  CollectStatus = "SCHEDULED"
	CollectStatusCommanded  CollectStatus = "COMMANDED"
	CollectStatusCollected  CollectStatus = "COLLECTED"
	CollectStatusTransmitted CollectStatus = "TRANSMITTED"

	// Data reception
	CollectStatusParsed     CollectStatus = "PARSED"
	CollectStatusIncomplete CollectStatus = "INCOMPLETE"
	CollectStatusAssembled  CollectStatus = "ASSEMBLED"
	CollectStatusStaged     CollectStatus = "STAGED"

	// Processing
	CollectStatusProcessing      CollectStatus = "PROCESSING"
	CollectStatusStalled         CollectStatus = "STALLED"
	CollectStatusProcessingDelay CollectStatus = "PROCESSING_DELAY"
	CollectStatusProcessed       CollectStatus = "PROCESSED"

	// Delivery
	CollectStatusDelivering      CollectStatus = "DELIVERING"
	CollectStatusDelivered       CollectStatus = "DELIVERED"
	CollectStatusDeliveryDelayed CollectStatus = "DELIVERY_DELAYED"
	CollectStatusDeliveryStalled CollectStatus = "DELIVERY_STALLED"
	CollectStatusDeliveryError   CollectStatus = "DELIVERY_ERROR"

	// Terminal states
	CollectStatusCanceled            CollectStatus = "CANCELED"
	CollectStatusFailed              CollectStatus = "FAILED"
	CollectStatusCorrupt             CollectStatus = "CORRUPT"
	CollectStatusSuperseded          CollectStatus = "SUPERSEDED"
	CollectStatusUnknownCollectStatus CollectStatus = "UNKNOWN_COLLECT_STATUS"

	// Legacy - keeping for backwards compatibility
	CollectStatusTasked CollectStatus = "TASKED"
)

// IsTerminal returns true if the collect status is a terminal state.
func (s CollectStatus) IsTerminal() bool {
	switch s {
	case CollectStatusCanceled, CollectStatusFailed, CollectStatusCorrupt,
		CollectStatusSuperseded, CollectStatusUnknownCollectStatus:
		return true
	}
	return false
}

// Collect represents a satellite data collection.
type Collect struct {
	ID           string        `json:"id"`
	TaskID       string        `json:"taskId"`
	Status       CollectStatus `json:"status"`
	SatelliteID  string        `json:"satelliteId"`
	CollectStart time.Time     `json:"collectStart,omitempty"`
	CollectEnd   time.Time     `json:"collectEnd,omitempty"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

// ListCollectsOptions contains optional filters for listing collects.
type ListCollectsOptions struct {
	TaskID string          `url:"taskId,omitempty"`
	Status []CollectStatus `url:"status,omitempty"`
	Limit  int             `url:"limit,omitempty"`
	Offset int             `url:"offset,omitempty"`
}

// CollectListResponse contains a paginated list of collects.
type CollectListResponse struct {
	Collects   []Collect `json:"collects"`
	TotalCount int       `json:"totalCount"`
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`
}

// GetCollect retrieves a collect by ID.
// GET /tasking/collects/{id}
func (c *Client) GetCollect(ctx context.Context, id string) (*Collect, error) {
	var col Collect
	err := c.doRequest(ctx, http.MethodGet, c.BaseURL().JoinPath("tasking", "collects", id), nil, http.StatusOK, &col)
	return &col, err
}

// ListCollects retrieves all collects with optional filtering.
// GET /tasking/collects
func (c *Client) ListCollects(ctx context.Context, opts *ListCollectsOptions) (*CollectListResponse, error) {
	u := c.BaseURL().JoinPath("tasking", "collects")
	if opts != nil {
		q := u.Query()
		if opts.TaskID != "" {
			q.Set("taskId", opts.TaskID)
		}
		if opts.Limit > 0 {
			q.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Offset > 0 {
			q.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
		for _, s := range opts.Status {
			q.Add("status", string(s))
		}
		u.RawQuery = q.Encode()
	}
	var resp CollectListResponse
	err := c.doRequest(ctx, http.MethodGet, u, nil, http.StatusOK, &resp)
	return &resp, err
}

// GetProductConstraints retrieves product constraints for an imaging mode.
// GET /tasking/products/{mode}/constraints
func (c *Client) GetProductConstraints(ctx context.Context, mode ImagingMode) ([]ProductConstraint, error) {
	var pc []ProductConstraint
	err := c.doRequest(ctx, http.MethodGet, c.BaseURL().JoinPath("tasking", "products", string(mode), "constraints"), nil, http.StatusOK, &pc)
	return pc, err
}

// CollectSearchRequest contains parameters for searching collects.
type CollectSearchRequest struct {
	Limit  *int                   `json:"limit,omitempty"`
	Skip   *int                   `json:"skip,omitempty"`
	Query  map[string]interface{} `json:"query,omitempty"`
	SortBy string                 `json:"sortBy,omitempty"`
	Order  string                 `json:"order,omitempty"`
}

// SearchCollects returns an iterator over collect search results with automatic pagination.
// POST /tasking/collects/search
func (c *Client) SearchCollects(ctx context.Context, req CollectSearchRequest) iter.Seq2[Collect, error] {
	if req.Limit == nil {
		limit := defaultSearchLimit
		req.Limit = &limit
	}
	if req.Skip == nil {
		skip := 0
		req.Skip = &skip
	}

	return func(yield func(Collect, error) bool) {
		for {
			body, err := marshalBody(req)
			if err != nil {
				var zero Collect
				yield(zero, err)
				return
			}

			var collects []Collect
			err = c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("tasking", "collects", "search"), body, http.StatusOK, &collects)
			if err != nil {
				var zero Collect
				yield(zero, err)
				return
			}

			for _, col := range collects {
				if !yield(col, nil) {
					return
				}
			}

			// If we got fewer results than the limit, we've reached the end
			if len(collects) < *req.Limit {
				return
			}

			// Advance to next page
			*req.Skip += *req.Limit
		}
	}
}
