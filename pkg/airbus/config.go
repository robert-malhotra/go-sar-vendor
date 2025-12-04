package airbus

import (
	"context"
	"net/http"
)

// GetConfig retrieves the entire user configuration.
// GET /sar/config
func (c *Client) GetConfig(ctx context.Context) (*Config, error) {
	var out Config
	err := c.doRequest(ctx, http.MethodGet, c.BaseURL().JoinPath("sar", "config"), nil, http.StatusOK, &out)
	return &out, err
}

// GetPermissions retrieves user permissions.
// GET /sar/config/permissions
func (c *Client) GetPermissions(ctx context.Context) (*Permissions, error) {
	var out Permissions
	err := c.doRequest(ctx, http.MethodGet, c.BaseURL().JoinPath("sar", "config", "permissions"), nil, http.StatusOK, &out)
	return &out, err
}

// GetSettings retrieves user settings.
// GET /sar/config/settings
func (c *Client) GetSettings(ctx context.Context) (*Settings, error) {
	var out Settings
	err := c.doRequest(ctx, http.MethodGet, c.BaseURL().JoinPath("sar", "config", "settings"), nil, http.StatusOK, &out)
	return &out, err
}

// GetCustomers retrieves available customers (for resellers).
// GET /sar/config/customers
func (c *Client) GetCustomers(ctx context.Context) ([]Customer, error) {
	var out []Customer
	err := c.doRequest(ctx, http.MethodGet, c.BaseURL().JoinPath("sar", "config", "customers"), nil, http.StatusOK, &out)
	return out, err
}

// GetOrderTemplates retrieves available order templates.
// GET /sar/config/orderTemplates
func (c *Client) GetOrderTemplates(ctx context.Context) ([]OrderTemplate, error) {
	var out []OrderTemplate
	err := c.doRequest(ctx, http.MethodGet, c.BaseURL().JoinPath("sar", "config", "orderTemplates"), nil, http.StatusOK, &out)
	return out, err
}

// GetAssociatedUsers retrieves associated users.
// GET /sar/config/associations
func (c *Client) GetAssociatedUsers(ctx context.Context) ([]Association, error) {
	var out []Association
	err := c.doRequest(ctx, http.MethodGet, c.BaseURL().JoinPath("sar", "config", "associations"), nil, http.StatusOK, &out)
	return out, err
}

// GetReceivingStations retrieves allowed receiving stations.
// This is only relevant for direct-access customers.
// GET /sar/config/receivingStations
func (c *Client) GetReceivingStations(ctx context.Context) ([]ReceivingStation, error) {
	var out []ReceivingStation
	err := c.doRequest(ctx, http.MethodGet, c.BaseURL().JoinPath("sar", "config", "receivingStations"), nil, http.StatusOK, &out)
	return out, err
}
