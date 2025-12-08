package airbus

import (
	"context"
	"net/http"
)

// Ping checks basic availability of the API.
// Returns nil if the API is available.
// GET /sar/ping
func (c *Client) Ping(ctx context.Context) error {
	return c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("sar", "ping"), nil, http.StatusOK, nil)
}

// Health returns the health status of the API.
// GET /sar/health
func (c *Client) Health(ctx context.Context) (*HealthStatus, error) {
	var out HealthStatus
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("sar", "health"), nil, http.StatusOK, &out)
	return &out, err
}
