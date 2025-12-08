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

// CreateOrder creates a new order for downloading imagery.
// POST /compute/ops/orders/v2
func (c *Client) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	body, err := common.MarshalBody(req)
	if err != nil {
		return nil, err
	}
	var order Order
	err = c.DoRaw(ctx, http.MethodPost, c.ordersBaseURL, body, http.StatusAccepted, &order)
	return &order, err
}

// GetOrder retrieves an order by ID.
// GET /compute/ops/orders/v2/{id}
func (c *Client) GetOrder(ctx context.Context, id string) (*Order, error) {
	var order Order
	err := c.DoRaw(ctx, http.MethodGet, c.OrdersURL(id), nil, http.StatusOK, &order)
	return &order, err
}

// CancelOrder cancels an order.
// PUT /compute/ops/orders/v2/{id}
func (c *Client) CancelOrder(ctx context.Context, id string) error {
	// Cancellation is done by sending a PUT request with an empty body
	return c.DoRaw(ctx, http.MethodPut, c.OrdersURL(id), nil, http.StatusOK, nil)
}

// ListOrders retrieves all orders with optional filtering.
// Returns an iterator that handles pagination automatically.
// GET /compute/ops/orders/v2
func (c *Client) ListOrders(ctx context.Context, opts *ListOrdersOptions) iter.Seq2[Order, error] {
	return func(yield func(Order, error) bool) {
		u := c.ordersBaseURL

		for {
			q := u.Query()
			if opts != nil {
				if opts.Limit > 0 {
					q.Set("limit", fmt.Sprintf("%d", opts.Limit))
				} else {
					q.Set("limit", fmt.Sprintf("%d", defaultSearchLimit))
				}
				if opts.Name != "" {
					q.Set("name", opts.Name)
				}
				if opts.SourceType != "" {
					q.Set("source_type", string(opts.SourceType))
				}
				if opts.DestinationRef != "" {
					q.Set("destination_ref", opts.DestinationRef)
				}
				if opts.Hosting != nil {
					q.Set("hosting", fmt.Sprintf("%t", *opts.Hosting))
				}
				for _, s := range opts.State {
					q.Add("state", string(s))
				}
			} else {
				q.Set("limit", fmt.Sprintf("%d", defaultSearchLimit))
			}
			u.RawQuery = q.Encode()

			var resp ordersListResponse
			if err := c.DoRaw(ctx, http.MethodGet, u, nil, http.StatusOK, &resp); err != nil {
				var zero Order
				yield(zero, err)
				return
			}

			for _, order := range resp.Orders {
				if !yield(order, nil) {
					return
				}
			}

			// Check if there are more pages
			if resp.Links.Next == "" {
				return
			}

			nextURL, err := url.Parse(resp.Links.Next)
			if err != nil {
				var zero Order
				yield(zero, err)
				return
			}
			u = nextURL
			opts = nil
		}
	}
}

// ordersListResponse is the response structure for listing orders.
type ordersListResponse struct {
	Orders []Order `json:"orders"`
	Links  struct {
		Self string `json:"_self,omitempty"`
		Next string `json:"_next,omitempty"`
	} `json:"_links,omitempty"`
}

// WaitForOrder polls until the order reaches a terminal state.
func (c *Client) WaitForOrder(ctx context.Context, id string, opts *WaitOptions) (*Order, error) {
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
		order, err := c.GetOrder(ctx, id)
		if err != nil {
			return nil, err
		}

		if order.State.IsTerminal() {
			return order, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for order %s to reach terminal state", id)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

// WaitForOrderSuccess polls until the order succeeds or fails.
func (c *Client) WaitForOrderSuccess(ctx context.Context, id string, opts *WaitOptions) (*Order, error) {
	order, err := c.WaitForOrder(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	if order.State == OrderStateFailed {
		hints := ""
		if len(order.ErrorHints) > 0 {
			hints = fmt.Sprintf(": %v", order.ErrorHints)
		}
		return order, fmt.Errorf("order %s failed%s", id, hints)
	}

	return order, nil
}

// GetOrderResults returns the download links for a completed order.
func (c *Client) GetOrderResults(ctx context.Context, id string) ([]Link, error) {
	order, err := c.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	if order.Links == nil || len(order.Links.Results) == 0 {
		return nil, nil
	}

	return order.Links.Results, nil
}
