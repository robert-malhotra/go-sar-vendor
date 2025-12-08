package iceye

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

// ----------------------------------------------------------------------------
// Delivery API Types
// ----------------------------------------------------------------------------

// DeliveryStatus represents the status of a delivery.
type DeliveryStatus string

const (
	DeliveryStatusSuccess DeliveryStatus = "SUCCESS"
	DeliveryStatusPending DeliveryStatus = "PENDING"
	DeliveryStatusFailed  DeliveryStatus = "FAILED"
)

// DeliveryLocationConfigStatus represents location config status.
type DeliveryLocationConfigStatus string

const (
	DeliveryLocationConfigStatusActive   DeliveryLocationConfigStatus = "active"
	DeliveryLocationConfigStatusInactive DeliveryLocationConfigStatus = "inactive"
)

// DeliveryLocationConfig represents a delivery location configuration.
type DeliveryLocationConfig struct {
	ID     string                       `json:"id"`
	Method string                       `json:"method"` // "s3"
	Config S3Config                     `json:"config"`
	Status DeliveryLocationConfigStatus `json:"status"`
}

// S3Config contains S3-specific delivery configuration.
type S3Config struct {
	Endpoint string `json:"endpoint"`
	Bucket   string `json:"bucket"`
	Region   string `json:"region"`
	KeyID    string `json:"keyID"`
}

// Delivery represents a delivery response.
type Delivery struct {
	ID                string                   `json:"id"`
	Status            DeliveryStatus           `json:"status"`
	URL               string                   `json:"url,omitempty"`
	ItemIDs           []string                 `json:"itemIDs"`
	DeliveryLocations []DeliveryLocation       `json:"deliveryLocations"`
	Notifications     *DeliveryNotificationRef `json:"notifications,omitempty"`
}

// DeliveryNotificationRef references a notification subscription.
type DeliveryNotificationRef struct {
	Subscription *SubscriptionRef `json:"subscription,omitempty"`
}

// SubscriptionRef references a notification subscription by ID.
type SubscriptionRef struct {
	ID string `json:"id"`
}

// CreateDeliveryRequest for creating a new delivery.
type CreateDeliveryRequest struct {
	ItemIDs           []string           `json:"itemIDs"`
	ContractID        string             `json:"contractID"`
	DeliveryLocations []DeliveryLocation `json:"deliveryLocations"`
}

// DeliveriesResponse is the paginated response for listing deliveries.
type DeliveriesResponse struct {
	Data   []Delivery `json:"data"`
	Cursor string     `json:"cursor,omitempty"`
}

// ListDeliveriesOptions for filtering delivery lists.
type ListDeliveriesOptions struct {
	Type string // Optional type filter
}

// ----------------------------------------------------------------------------
// Delivery API Methods
// Endpoints: https://docs.iceye.com/constellation/api/specification/delivery/v1/
// ----------------------------------------------------------------------------

const deliveryBasePath = "/delivery/v1"

// ListDeliveryLocationConfigs retrieves all delivery location configurations.
//
// GET /delivery/v1/deliveries/location-configs
func (c *Client) ListDeliveryLocationConfigs(ctx context.Context) ([]DeliveryLocationConfig, error) {
	var resp []DeliveryLocationConfig
	u := &url.URL{Path: path.Join(deliveryBasePath, "deliveries", "location-configs")}
	if err := c.do(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ListDeliveries lists deliveries with optional filters.
// Returns an iterator that yields pages of deliveries.
//
// GET /delivery/v1/deliveries
func (c *Client) ListDeliveries(ctx context.Context, pageSize int, opts *ListDeliveriesOptions) iter.Seq2[DeliveriesResponse, error] {
	return func(yield func(DeliveriesResponse, error) bool) {
		seq := common.Paginate(func(cur *string) ([]Delivery, *string, error) {
			u := &url.URL{Path: path.Join(deliveryBasePath, "deliveries")}
			q := u.Query()

			if pageSize > 0 {
				// API spec says limit is 1-100
				if pageSize > 100 {
					pageSize = 100
				}
				q.Set("limit", strconv.Itoa(pageSize))
			}
			if opts != nil && opts.Type != "" {
				q.Set("type", opts.Type)
			}
			if cur != nil && *cur != "" {
				q.Set("cursor", *cur)
			}
			u.RawQuery = q.Encode()

			var resp DeliveriesResponse
			err := c.do(ctx, http.MethodGet, u.String(), nil, &resp)
			return resp.Data, &resp.Cursor, err
		})
		for data, err := range seq {
			if !yield(DeliveriesResponse{Data: data}, err) {
				return
			}
		}
	}
}

// GetDelivery retrieves a delivery by ID.
//
// GET /delivery/v1/deliveries/{ID}
func (c *Client) GetDelivery(ctx context.Context, deliveryID string) (*Delivery, error) {
	var resp Delivery
	u := &url.URL{Path: path.Join(deliveryBasePath, "deliveries", deliveryID)}
	if err := c.do(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateDelivery creates a new delivery.
//
// POST /delivery/v1/deliveries
func (c *Client) CreateDelivery(ctx context.Context, req *CreateDeliveryRequest) (*Delivery, error) {
	var resp Delivery
	u := &url.URL{Path: path.Join(deliveryBasePath, "deliveries")}
	if err := c.do(ctx, http.MethodPost, u.String(), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
