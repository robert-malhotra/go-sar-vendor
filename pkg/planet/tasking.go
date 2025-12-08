package planet

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

// CreateTaskingOrder creates a new tasking order.
// POST /tasking/v2/orders/
func (c *Client) CreateTaskingOrder(ctx context.Context, req *CreateTaskingOrderRequest) (*TaskingOrder, error) {
	body, err := common.MarshalBody(req)
	if err != nil {
		return nil, err
	}
	var order TaskingOrder
	err = c.DoRaw(ctx, http.MethodPost, c.TaskingURL("orders", ""), body, http.StatusCreated, &order)
	return &order, err
}

// GetTaskingOrder retrieves a tasking order by ID.
// GET /tasking/v2/orders/{id}
func (c *Client) GetTaskingOrder(ctx context.Context, id string) (*TaskingOrder, error) {
	var order TaskingOrder
	err := c.DoRaw(ctx, http.MethodGet, c.TaskingURL("orders", id), nil, http.StatusOK, &order)
	return &order, err
}

// UpdateTaskingOrder updates an existing tasking order.
// PUT /tasking/v2/orders/{id}
func (c *Client) UpdateTaskingOrder(ctx context.Context, id string, req *UpdateTaskingOrderRequest) (*TaskingOrder, error) {
	body, err := common.MarshalBody(req)
	if err != nil {
		return nil, err
	}
	var order TaskingOrder
	err = c.DoRaw(ctx, http.MethodPut, c.TaskingURL("orders", id), body, http.StatusOK, &order)
	return &order, err
}

// CancelTaskingOrder cancels a tasking order.
// DELETE /tasking/v2/orders/{id}
func (c *Client) CancelTaskingOrder(ctx context.Context, id string) error {
	return c.DoRaw(ctx, http.MethodDelete, c.TaskingURL("orders", id), nil, http.StatusNoContent, nil)
}

// ListTaskingOrders retrieves all tasking orders with optional filtering.
// Returns an iterator that handles pagination automatically.
// GET /tasking/v2/orders/
func (c *Client) ListTaskingOrders(ctx context.Context, opts *ListTaskingOrdersOptions) iter.Seq2[TaskingOrder, error] {
	return func(yield func(TaskingOrder, error) bool) {
		u := c.TaskingURL("orders", "")

		for {
			// Build query parameters
			q := u.Query()
			if opts != nil {
				if opts.Limit > 0 {
					q.Set("limit", fmt.Sprintf("%d", opts.Limit))
				} else {
					q.Set("limit", fmt.Sprintf("%d", defaultSearchLimit))
				}
				if opts.Offset > 0 {
					q.Set("offset", fmt.Sprintf("%d", opts.Offset))
				}
				if opts.SchedulingType != "" {
					q.Set("scheduling_type", string(opts.SchedulingType))
				}
				if opts.PLNumber != "" {
					q.Set("pl_number", opts.PLNumber)
				}
				if opts.Product != "" {
					q.Set("product", opts.Product)
				}
				if opts.NameContains != "" {
					q.Set("name__icontains", opts.NameContains)
				}
				if opts.Ordering != "" {
					q.Set("ordering", opts.Ordering)
				}
				if opts.GeometryIntersects != "" {
					q.Set("geometry__intersects", opts.GeometryIntersects)
				}
				for _, s := range opts.Status {
					q.Add("status__in", string(s))
				}
				if opts.CreatedTimeGTE != nil {
					q.Set("created_time__gte", opts.CreatedTimeGTE.Format(time.RFC3339))
				}
				if opts.CreatedTimeLTE != nil {
					q.Set("created_time__lte", opts.CreatedTimeLTE.Format(time.RFC3339))
				}
				if opts.StartTimeGTE != nil {
					q.Set("start_time__gte", opts.StartTimeGTE.Format(time.RFC3339))
				}
				if opts.StartTimeLTE != nil {
					q.Set("start_time__lte", opts.StartTimeLTE.Format(time.RFC3339))
				}
			} else {
				q.Set("limit", fmt.Sprintf("%d", defaultSearchLimit))
			}
			u.RawQuery = q.Encode()

			var resp paginatedResponse[TaskingOrder]
			if err := c.DoRaw(ctx, http.MethodGet, u, nil, http.StatusOK, &resp); err != nil {
				var zero TaskingOrder
				yield(zero, err)
				return
			}

			for _, order := range resp.Results {
				if !yield(order, nil) {
					return
				}
			}

			// Check if there are more pages
			if resp.Next == "" {
				return
			}

			// Parse the next URL for the next iteration
			nextURL, err := url.Parse(resp.Next)
			if err != nil {
				var zero TaskingOrder
				yield(zero, err)
				return
			}
			u = nextURL
			opts = nil // Use the next URL directly, don't rebuild params
		}
	}
}

// GetTaskingOrderPricing retrieves pricing for a tasking order.
// GET /tasking/v2/orders/{id}/pricing
func (c *Client) GetTaskingOrderPricing(ctx context.Context, id string) (*TaskingOrderPricing, error) {
	var pricing TaskingOrderPricing
	err := c.DoRaw(ctx, http.MethodGet, c.TaskingURL("orders", id, "pricing"), nil, http.StatusOK, &pricing)
	return &pricing, err
}

// PreviewPricing previews pricing for a potential order.
// POST /tasking/v2/pricing/
func (c *Client) PreviewPricing(ctx context.Context, req *CreateTaskingOrderRequest) (*TaskingOrderPricing, error) {
	body, err := common.MarshalBody(req)
	if err != nil {
		return nil, err
	}
	var pricing TaskingOrderPricing
	err = c.DoRaw(ctx, http.MethodPost, c.TaskingURL("pricing", ""), body, http.StatusOK, &pricing)
	return &pricing, err
}

// GetCapture retrieves a capture by ID.
// GET /tasking/v2/captures/{id}
func (c *Client) GetCapture(ctx context.Context, id string) (*Capture, error) {
	var capture Capture
	err := c.DoRaw(ctx, http.MethodGet, c.TaskingURL("captures", id), nil, http.StatusOK, &capture)
	return &capture, err
}

// ListCaptures retrieves captures with optional filtering.
// Returns an iterator that handles pagination automatically.
// GET /tasking/v2/captures/
func (c *Client) ListCaptures(ctx context.Context, opts *ListCapturesOptions) iter.Seq2[Capture, error] {
	return func(yield func(Capture, error) bool) {
		u := c.TaskingURL("captures", "")

		for {
			q := u.Query()
			if opts != nil {
				if opts.Limit > 0 {
					q.Set("limit", fmt.Sprintf("%d", opts.Limit))
				} else {
					q.Set("limit", fmt.Sprintf("%d", defaultSearchLimit))
				}
				if opts.Offset > 0 {
					q.Set("offset", fmt.Sprintf("%d", opts.Offset))
				}
				if opts.OrderID != "" {
					q.Set("order_id", opts.OrderID)
				}
				if opts.Ordering != "" {
					q.Set("ordering", opts.Ordering)
				}
				if opts.Fulfilling != nil {
					q.Set("fulfilling", fmt.Sprintf("%t", *opts.Fulfilling))
				}
				for _, s := range opts.Status {
					q.Add("status__in", string(s))
				}
			} else {
				q.Set("limit", fmt.Sprintf("%d", defaultSearchLimit))
			}
			u.RawQuery = q.Encode()

			var resp paginatedResponse[Capture]
			if err := c.DoRaw(ctx, http.MethodGet, u, nil, http.StatusOK, &resp); err != nil {
				var zero Capture
				yield(zero, err)
				return
			}

			for _, capture := range resp.Results {
				if !yield(capture, nil) {
					return
				}
			}

			if resp.Next == "" {
				return
			}

			nextURL, err := url.Parse(resp.Next)
			if err != nil {
				var zero Capture
				yield(zero, err)
				return
			}
			u = nextURL
			opts = nil
		}
	}
}

// WaitForTaskingOrder polls until the tasking order reaches a terminal state.
func (c *Client) WaitForTaskingOrder(ctx context.Context, id string, opts *WaitOptions) (*TaskingOrder, error) {
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
		order, err := c.GetTaskingOrder(ctx, id)
		if err != nil {
			return nil, err
		}

		if order.Status.IsTerminal() {
			return order, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for tasking order %s to reach terminal state", id)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

// WaitForTaskingOrderFulfilled polls until the tasking order is fulfilled or reaches a terminal state.
func (c *Client) WaitForTaskingOrderFulfilled(ctx context.Context, id string, opts *WaitOptions) (*TaskingOrder, error) {
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
		order, err := c.GetTaskingOrder(ctx, id)
		if err != nil {
			return nil, err
		}

		if order.Status == TaskingOrderStatusFulfilled || order.Status.IsTerminal() {
			return order, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for tasking order %s to be fulfilled", id)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
