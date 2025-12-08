package airbus

import (
	"context"
	"net/http"

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

// GetPrices queries prices for acquisitions or items.
// You can provide either acquisition IDs (for catalogue items) or
// item UUIDs (for feasibility items).
// POST /sar/prices
func (c *Client) GetPrices(ctx context.Context, req *PricesRequest) ([]PriceResponse, error) {
	body, err := common.MarshalBody(req)
	if err != nil {
		return nil, err
	}
	var out []PriceResponse
	err = c.DoRaw(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "prices"), body, http.StatusOK, &out)
	return out, err
}
