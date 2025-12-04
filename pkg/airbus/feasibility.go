package airbus

import (
	"context"
	"net/http"
)

// SearchFeasibility searches for possible future acquisitions (tasking).
// This is used to check what imaging opportunities are available for a given
// area of interest and time window before placing a tasking order.
// POST /sar/feasibility
func (c *Client) SearchFeasibility(ctx context.Context, req *FeasibilityRequest) (*FeatureCollection, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var out FeatureCollection
	err = c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "feasibility"), body, http.StatusOK, &out)
	return &out, err
}

// GetConflicts checks for conflicts between items.
// Conflicts can occur when items have overlapping acquisition windows.
// POST /sar/conflicts
func (c *Client) GetConflicts(ctx context.Context, req *ConflictsRequest) (*ConflictsResponse, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var out ConflictsResponse
	err = c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "conflicts"), body, http.StatusOK, &out)
	return &out, err
}

// GetSwathEditInfo retrieves swath editing information for items.
// This shows the editable geometry bounds for acquisitions.
// POST /sar/swathedit
func (c *Client) GetSwathEditInfo(ctx context.Context, req *SwathEditRequest) (*SwathEditResponse, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var out SwathEditResponse
	err = c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "swathedit"), body, http.StatusOK, &out)
	return &out, err
}

// UpdateSwathEdit updates swath geometry for items.
// This allows customizing the acquisition footprint within allowed bounds.
// PATCH /sar/swathedit
func (c *Client) UpdateSwathEdit(ctx context.Context, req *UpdateSwathRequest) (*SwathEditResponse, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var out SwathEditResponse
	err = c.doRequest(ctx, http.MethodPatch, c.BaseURL().JoinPath("sar", "swathedit"), body, http.StatusOK, &out)
	return &out, err
}

// CreateStacks creates datastacks from template items.
// A datastack is a series of acquisitions over the same area at regular intervals.
// POST /sar/stacks
func (c *Client) CreateStacks(ctx context.Context, req *CreateStacksRequest) (*StacksResponse, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var out StacksResponse
	err = c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "stacks"), body, http.StatusOK, &out)
	return &out, err
}
