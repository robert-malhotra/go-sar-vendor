package airbus

import (
	"context"
	"net/http"
)

// ListBaskets returns all baskets for the current user.
// GET /sar/baskets
func (c *Client) ListBaskets(ctx context.Context) ([]Basket, error) {
	var out []Basket
	err := c.doRequest(ctx, http.MethodGet, c.BaseURL().JoinPath("sar", "baskets"), nil, http.StatusOK, &out)
	return out, err
}

// CreateBasket creates a new basket.
// POST /sar/baskets
func (c *Client) CreateBasket(ctx context.Context, req *CreateBasketRequest) (*Basket, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var out Basket
	err = c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "baskets"), body, http.StatusOK, &out)
	return &out, err
}

// GetBasket retrieves a basket by ID.
// GET /sar/baskets/{basketId}
func (c *Client) GetBasket(ctx context.Context, basketID string) (*Basket, error) {
	var out Basket
	err := c.doRequest(ctx, http.MethodGet, c.BaseURL().JoinPath("sar", "baskets", basketID), nil, http.StatusOK, &out)
	return &out, err
}

// UpdateBasket updates basket parameters.
// Only provided fields will be updated.
// PATCH /sar/baskets/{basketId}
func (c *Client) UpdateBasket(ctx context.Context, basketID string, req *UpdateBasketRequest) (*Basket, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var out Basket
	err = c.doRequest(ctx, http.MethodPatch, c.BaseURL().JoinPath("sar", "baskets", basketID), body, http.StatusOK, &out)
	return &out, err
}

// ReplaceBasket replaces all basket parameters.
// PUT /sar/baskets/{basketId}
func (c *Client) ReplaceBasket(ctx context.Context, basketID string, req *ReplaceBasketRequest) (*Basket, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var out Basket
	err = c.doRequest(ctx, http.MethodPut, c.BaseURL().JoinPath("sar", "baskets", basketID), body, http.StatusOK, &out)
	return &out, err
}

// DeleteBasket deletes a basket.
// DELETE /sar/baskets/{basketId}
func (c *Client) DeleteBasket(ctx context.Context, basketID string) error {
	return c.doRequest(ctx, http.MethodDelete, c.BaseURL().JoinPath("sar", "baskets", basketID), nil, http.StatusNoContent, nil)
}

// AddItemsToBasket adds acquisitions or items to a basket.
// You can add either by acquisition ID (from catalogue) or by item UUID (from feasibility).
// POST /sar/baskets/{basketId}/addItems
func (c *Client) AddItemsToBasket(ctx context.Context, basketID string, req *AddItemsRequest) (*Basket, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var out Basket
	err = c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "baskets", basketID, "addItems"), body, http.StatusOK, &out)
	return &out, err
}

// RemoveItemsFromBasket removes items from a basket.
// POST /sar/baskets/{basketId}/removeItems
func (c *Client) RemoveItemsFromBasket(ctx context.Context, basketID string, req *RemoveItemsRequest) (*Basket, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var out Basket
	err = c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "baskets", basketID, "removeItems"), body, http.StatusOK, &out)
	return &out, err
}

// RearrangeBasketItems rearranges items within a basket.
// The items array should contain all item IDs in the desired order.
// POST /sar/baskets/{basketId}/rearrangeItems
func (c *Client) RearrangeBasketItems(ctx context.Context, basketID string, req *RearrangeItemsRequest) (*Basket, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var out Basket
	err = c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "baskets", basketID, "rearrangeItems"), body, http.StatusOK, &out)
	return &out, err
}

// SubmitBasket submits a basket as an order.
// The basket must have a purpose set and contain at least one item.
// POST /sar/baskets/{basketId}/submit
func (c *Client) SubmitBasket(ctx context.Context, basketID string) (*Order, error) {
	var out Order
	err := c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "baskets", basketID, "submit"), nil, http.StatusOK, &out)
	return &out, err
}
