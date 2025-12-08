package umbra

import (
	"context"
	"net/http"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

// DeliveryType represents the type of delivery destination.
type DeliveryType string

const (
	DeliveryTypeS3UmbraRole DeliveryType = "S3_UMBRA_ROLE"
	DeliveryTypeGCPWIF      DeliveryType = "GCP_WIF"
)

// DeliveryConfigStatus represents the status of a delivery configuration.
type DeliveryConfigStatus string

const (
	DeliveryConfigStatusActive     DeliveryConfigStatus = "ACTIVE"
	DeliveryConfigStatusInactive   DeliveryConfigStatus = "INACTIVE"
	DeliveryConfigStatusUnverified DeliveryConfigStatus = "UNVERIFIED"
)

// PackagingOption represents how files are packaged before delivery.
type PackagingOption string

const (
	PackagingNone                    PackagingOption = ""
	PackagingZipEachAssetWithMetadata PackagingOption = "ZIP_EACH_ASSET_WITH_METADATA"
	PackagingZipAllAssetsWithMetadata PackagingOption = "ZIP_ALL_ASSETS_WITH_METADATA"
)

// DeliveryOptions contains destination-specific configuration.
type DeliveryOptions struct {
	// S3 options
	Bucket     string `json:"bucket,omitempty"`
	Path       string `json:"path,omitempty"`
	Region     string `json:"region,omitempty"`
	IsGovcloud bool   `json:"isGovcloud,omitempty"`

	// GCP options
	ProjectID  string `json:"projectId,omitempty"`
	BucketName string `json:"bucketName,omitempty"`
}

// DeliveryConfig represents a delivery destination configuration.
type DeliveryConfig struct {
	ID                   string               `json:"id"`
	Name                 string               `json:"name,omitempty"`
	Type                 DeliveryType         `json:"type"`
	Status               DeliveryConfigStatus `json:"status"`
	Bucket               string               `json:"bucket,omitempty"`
	Path                 string               `json:"path,omitempty"`
	Region               string               `json:"region,omitempty"`
	IsGovcloud           bool                 `json:"isGovcloud,omitempty"`
	ProjectID            string               `json:"projectId,omitempty"`
	BucketName           string               `json:"bucketName,omitempty"`
	FileNamingConvention string               `json:"fileNamingConvention,omitempty"`
	Packaging            PackagingOption      `json:"packaging,omitempty"`
	CreatedAt            time.Time            `json:"createdAt"`
	UpdatedAt            time.Time            `json:"updatedAt"`
}

// CreateDeliveryConfigRequest contains parameters for creating a delivery config.
type CreateDeliveryConfigRequest struct {
	Name                 string          `json:"name,omitempty"`
	Type                 DeliveryType    `json:"type"`
	Bucket               string          `json:"bucket,omitempty"`
	Path                 string          `json:"path,omitempty"`
	Region               string          `json:"region,omitempty"`
	IsGovcloud           bool            `json:"isGovcloud,omitempty"`
	ProjectID            string          `json:"projectId,omitempty"`
	BucketName           string          `json:"bucketName,omitempty"`
	FileNamingConvention string          `json:"fileNamingConvention,omitempty"`
	Packaging            PackagingOption `json:"packaging,omitempty"`
}

// CreateDeliveryConfig creates a new delivery configuration.
// POST /delivery/delivery-config/
func (c *Client) CreateDeliveryConfig(ctx context.Context, req *CreateDeliveryConfigRequest) (*DeliveryConfig, error) {
	body, err := common.MarshalBody(req)
	if err != nil {
		return nil, err
	}
	var dc DeliveryConfig
	err = c.DoRaw(ctx, http.MethodPost, c.BaseURL().JoinPath("delivery", "delivery-config"), body, http.StatusCreated, &dc)
	return &dc, err
}

// GetDeliveryConfig retrieves a delivery configuration by ID.
// GET /delivery/delivery-config/{id}
func (c *Client) GetDeliveryConfig(ctx context.Context, id string) (*DeliveryConfig, error) {
	var dc DeliveryConfig
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("delivery", "delivery-config", id), nil, http.StatusOK, &dc)
	return &dc, err
}

// ListDeliveryConfigs retrieves all delivery configurations.
// GET /delivery/delivery-configs
func (c *Client) ListDeliveryConfigs(ctx context.Context) ([]DeliveryConfig, error) {
	var configs []DeliveryConfig
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("delivery", "delivery-configs"), nil, http.StatusOK, &configs)
	return configs, err
}

// VerifyDeliveryConfig triggers verification of a delivery configuration.
// POST /delivery/delivery-config/verify
func (c *Client) VerifyDeliveryConfig(ctx context.Context, id string) (*DeliveryConfig, error) {
	body, err := common.MarshalBody(map[string]string{"deliveryConfigId": id})
	if err != nil {
		return nil, err
	}
	var dc DeliveryConfig
	err = c.DoRaw(ctx, http.MethodPost, c.BaseURL().JoinPath("delivery", "delivery-config", "verify"), body, http.StatusOK, &dc)
	return &dc, err
}

// DeleteDeliveryConfig deletes a delivery configuration.
// DELETE /delivery/delivery-config/{id}
func (c *Client) DeleteDeliveryConfig(ctx context.Context, id string) error {
	return c.DoRaw(ctx, http.MethodDelete, c.BaseURL().JoinPath("delivery", "delivery-config", id), nil, http.StatusNoContent, nil)
}

// GetCollectMetadataSchema retrieves the schema for collect metadata.
// GET /delivery/collect-metadata/schema
func (c *Client) GetCollectMetadataSchema(ctx context.Context) (map[string]interface{}, error) {
	var schema map[string]interface{}
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("delivery", "collect-metadata", "schema"), nil, http.StatusOK, &schema)
	return schema, err
}
