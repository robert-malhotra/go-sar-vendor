package umbra

import (
	"context"
	"net/http"
)

// RestrictedAccessArea represents a geographic area with restricted access.
type RestrictedAccessArea struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Geometry    *GeoJSONGeometry `json:"geometry,omitempty"`
}

// GetRestrictedAccessAreas retrieves all restricted access areas.
// GET /tasking/restricted-access-areas
func (c *Client) GetRestrictedAccessAreas(ctx context.Context) ([]RestrictedAccessArea, error) {
	var areas []RestrictedAccessArea
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("tasking", "restricted-access-areas"), nil, http.StatusOK, &areas)
	return areas, err
}

// OrganizationSettings represents organization-level settings.
type OrganizationSettings struct {
	ID                    string `json:"id"`
	OrganizationID        string `json:"organizationId"`
	DefaultDeliveryConfig string `json:"defaultDeliveryConfigId,omitempty"`
}

// GetOrganizationSettings retrieves organization settings.
// GET /admin/organization-settings
func (c *Client) GetOrganizationSettings(ctx context.Context) (*OrganizationSettings, error) {
	var settings OrganizationSettings
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("admin", "organization-settings"), nil, http.StatusOK, &settings)
	return &settings, err
}

// ListProductConstraintsAdmin retrieves all product constraints via admin API.
// GET /admin/product-constraints
func (c *Client) ListProductConstraintsAdmin(ctx context.Context) ([]ProductConstraint, error) {
	var constraints []ProductConstraint
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("admin", "product-constraints"), nil, http.StatusOK, &constraints)
	return constraints, err
}
