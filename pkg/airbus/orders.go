package airbus

import (
	"context"
	"net/http"

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

// ListOrders returns all orders for the current user.
// GET /sar/orders
func (c *Client) ListOrders(ctx context.Context) ([]OrderSummary, error) {
	var out []OrderSummary
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("sar", "orders"), nil, http.StatusOK, &out)
	return out, err
}

// GetOrder retrieves an order by ID or basket ID.
// GET /sar/orders/{orderIdOrBasketId}
func (c *Client) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	var out Order
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("sar", "orders", orderID), nil, http.StatusOK, &out)
	return &out, err
}

// UpdateOrder updates order parameters (e.g., notification endpoint).
// PATCH /sar/orders/{orderIdOrBasketId}
func (c *Client) UpdateOrder(ctx context.Context, orderID string, req *UpdateOrderRequest) (*Order, error) {
	body, err := common.MarshalBody(req)
	if err != nil {
		return nil, err
	}
	var out Order
	err = c.DoRaw(ctx, http.MethodPatch, c.BaseURL().JoinPath("sar", "orders", orderID), body, http.StatusOK, &out)
	return &out, err
}

// GetOrderItems retrieves items by order item IDs.
// POST /sar/orders/*/items
func (c *Client) GetOrderItems(ctx context.Context, req *GetOrderItemsRequest) (*FeatureCollection, error) {
	body, err := common.MarshalBody(req)
	if err != nil {
		return nil, err
	}
	var out FeatureCollection
	err = c.DoRaw(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "orders", "*", "items"), body, http.StatusOK, &out)
	return &out, err
}

// GetOrderItemsStatus retrieves item status with filters.
// POST /sar/orders/*/items/status
func (c *Client) GetOrderItemsStatus(ctx context.Context, req *GetOrderItemsStatusRequest) (*OrderItemsStatusResponse, error) {
	body, err := common.MarshalBody(req)
	if err != nil {
		return nil, err
	}
	var out OrderItemsStatusResponse
	err = c.DoRaw(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "orders", "*", "items", "status"), body, http.StatusOK, &out)
	return &out, err
}

// CancelOrderItems cancels ordered items.
// Items can only be cancelled if they haven't been acquired yet.
// POST /sar/orders/cancel
func (c *Client) CancelOrderItems(ctx context.Context, req *CancelItemsRequest) (*CancelItemsResponse, error) {
	body, err := common.MarshalBody(req)
	if err != nil {
		return nil, err
	}
	var out CancelItemsResponse
	err = c.DoRaw(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "orders", "cancel"), body, http.StatusOK, &out)
	return &out, err
}

// ReorderItems reorders items with different order options.
// This creates a new basket with the specified items.
// POST /sar/orders/reorder
func (c *Client) ReorderItems(ctx context.Context, req *ReorderRequest) (*Basket, error) {
	body, err := common.MarshalBody(req)
	if err != nil {
		return nil, err
	}
	var out Basket
	err = c.DoRaw(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "orders", "reorder"), body, http.StatusOK, &out)
	return &out, err
}

// SubmitOrder submits an order directly.
// This is an alternative to SubmitBasket that allows direct order submission.
// POST /sar/orders/submit
func (c *Client) SubmitOrder(ctx context.Context, req *SubmitOrderRequest) (*Order, error) {
	body, err := common.MarshalBody(req)
	if err != nil {
		return nil, err
	}
	var out Order
	err = c.DoRaw(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "orders", "submit"), body, http.StatusOK, &out)
	return &out, err
}

// UpdateOrderOptions updates processing options for items.
// PATCH /sar/orderOptions
func (c *Client) UpdateOrderOptions(ctx context.Context, req *UpdateOrderOptionsRequest) error {
	body, err := common.MarshalBody(req)
	if err != nil {
		return err
	}
	return c.DoRaw(ctx, http.MethodPatch, c.BaseURL().JoinPath("sar", "orderOptions"), body, http.StatusOK, nil)
}
