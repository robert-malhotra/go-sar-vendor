package airbus

import (
	"context"
	"net/http"
)

// GetPrices queries prices for acquisitions or items.
// You can provide either acquisition IDs (for catalogue items) or
// item UUIDs (for feasibility items).
// POST /sar/prices
func (c *Client) GetPrices(ctx context.Context, req *PricesRequest) ([]PriceResponse, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var out []PriceResponse
	err = c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "prices"), body, http.StatusOK, &out)
	return out, err
}
