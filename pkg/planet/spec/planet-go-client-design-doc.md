# Planet Go Client Design Document

## Executive Summary

This document describes the design for a Go client library that interfaces with Planet's satellite imagery APIs, specifically targeting the Tasking API v2, Imaging Windows (Feasibility) API, and Orders API v2. The client follows the established patterns in this codebase, using a single-package approach with the common client infrastructure.

---

## API Coverage

### Tasking API v2

Base URL: `https://api.planet.com/tasking/v2`

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/orders/` | GET | List all tasking orders |
| `/orders/` | POST | Create a new tasking order |
| `/orders/{id}` | GET | Retrieve a specific order |
| `/orders/{id}` | PUT | Update an existing order |
| `/orders/{id}` | DELETE | Cancel an order |
| `/orders/{id}/pricing` | GET | Get order pricing details |
| `/pricing/` | POST | Preview pricing for potential order |
| `/captures/` | GET | List captures for orders |
| `/captures/{id}` | GET | Retrieve a specific capture |

### Imaging Windows API (Feasibility)

Base URL: `https://api.planet.com/tasking/v2`

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/imaging-windows/search/` | POST | Create async imaging window search |
| `/imaging-windows/search/{id}` | GET | Get imaging window search results |

### Orders API v2

Base URL: `https://api.planet.com/compute/ops`

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/orders/v2` | GET | List all orders |
| `/orders/v2` | POST | Create a new order |
| `/orders/v2/{id}` | GET | Get order details |
| `/orders/v2/{id}` | PUT | Cancel an order |

---

## Package Structure

```
pkg/planet/
├── client.go           # Main client implementation
├── client_test.go      # Client tests
├── errors.go           # Error types (delegates to common)
├── types.go            # All domain types and enums
├── tasking.go          # Tasking API operations
├── tasking_test.go     # Tasking tests
├── feasibility.go      # Imaging Windows/Feasibility API
├── feasibility_test.go # Feasibility tests
├── orders.go           # Orders API operations
├── orders_test.go      # Orders tests
└── spec/               # API specification files
```

---

## Client Design

The client embeds `common.Client` and uses API key authentication via the `api-key` header format.

```go
// Client represents a Planet API client.
type Client struct {
    *common.Client
    taskingBaseURL *url.URL  // https://api.planet.com/tasking/v2
    ordersBaseURL  *url.URL  // https://api.planet.com/compute/ops/orders/v2
}

// NewClient creates a new Planet API client.
func NewClient(apiKey string, opts ...Option) (*Client, error)
```

### Options

```go
WithHTTPClient(client *http.Client) Option
WithBaseURL(rawURL string) Option
WithTimeout(timeout time.Duration) Option
WithUserAgent(userAgent string) Option
```

---

## Authentication

Planet uses API key authentication with the header format:
```
Authorization: api-key YOUR_API_KEY
```

This is implemented as a custom authenticator that implements `common.Authenticator`.

---

## Types Overview

### Tasking Types

```go
// TaskingOrderStatus represents the status of a tasking order.
type TaskingOrderStatus string

const (
    TaskingOrderStatusReceived   TaskingOrderStatus = "RECEIVED"
    TaskingOrderStatusPending    TaskingOrderStatus = "PENDING"
    TaskingOrderStatusInProgress TaskingOrderStatus = "IN_PROGRESS"
    TaskingOrderStatusFulfilled  TaskingOrderStatus = "FULFILLED"
    TaskingOrderStatusCancelled  TaskingOrderStatus = "CANCELLED"
    // ... etc
)

// SchedulingType represents the tasking scheduling type.
type SchedulingType string

const (
    SchedulingTypeFlexible SchedulingType = "FLEXIBLE"
    SchedulingTypeAssured  SchedulingType = "ASSURED"
    // ... etc
)

// SatelliteType represents supported satellite types.
type SatelliteType string

const (
    SatelliteTypeSkySat  SatelliteType = "SKYSAT"
    SatelliteTypePelican SatelliteType = "PELICAN"
    SatelliteTypeTanager SatelliteType = "TANAGER"
)
```

### Feasibility Types

```go
// ImagingWindowSearchStatus represents the status of an imaging window search.
type ImagingWindowSearchStatus string

const (
    ImagingWindowSearchStatusCreated    ImagingWindowSearchStatus = "CREATED"
    ImagingWindowSearchStatusInProgress ImagingWindowSearchStatus = "IN_PROGRESS"
    ImagingWindowSearchStatusDone       ImagingWindowSearchStatus = "DONE"
    ImagingWindowSearchStatusFailed     ImagingWindowSearchStatus = "FAILED"
)
```

### Orders Types

```go
// OrderState represents the state of an order.
type OrderState string

const (
    OrderStateQueued    OrderState = "queued"
    OrderStateRunning   OrderState = "running"
    OrderStateSuccess   OrderState = "success"
    OrderStateFailed    OrderState = "failed"
    OrderStateCancelled OrderState = "cancelled"
)
```

---

## API Methods

### Tasking API

```go
// CreateTaskingOrder creates a new tasking order.
func (c *Client) CreateTaskingOrder(ctx context.Context, req *CreateTaskingOrderRequest) (*TaskingOrder, error)

// GetTaskingOrder retrieves a tasking order by ID.
func (c *Client) GetTaskingOrder(ctx context.Context, id string) (*TaskingOrder, error)

// ListTaskingOrders retrieves all tasking orders with optional filtering.
func (c *Client) ListTaskingOrders(ctx context.Context, opts *ListTaskingOrdersOptions) iter.Seq2[TaskingOrder, error]

// UpdateTaskingOrder updates an existing tasking order.
func (c *Client) UpdateTaskingOrder(ctx context.Context, id string, req *UpdateTaskingOrderRequest) (*TaskingOrder, error)

// CancelTaskingOrder cancels a tasking order.
func (c *Client) CancelTaskingOrder(ctx context.Context, id string) error

// GetTaskingOrderPricing retrieves pricing for a tasking order.
func (c *Client) GetTaskingOrderPricing(ctx context.Context, id string) (*TaskingOrderPricing, error)

// PreviewPricing previews pricing for a potential order.
func (c *Client) PreviewPricing(ctx context.Context, req *CreateTaskingOrderRequest) (*TaskingOrderPricing, error)

// GetCapture retrieves a capture by ID.
func (c *Client) GetCapture(ctx context.Context, id string) (*Capture, error)

// ListCaptures retrieves captures with optional filtering.
func (c *Client) ListCaptures(ctx context.Context, opts *ListCapturesOptions) iter.Seq2[Capture, error]
```

### Feasibility API

```go
// CreateImagingWindowSearch creates an async imaging window search.
func (c *Client) CreateImagingWindowSearch(ctx context.Context, req *ImagingWindowSearchRequest) (*ImagingWindowSearch, error)

// GetImagingWindowSearch retrieves imaging window search results.
func (c *Client) GetImagingWindowSearch(ctx context.Context, id string) (*ImagingWindowSearch, error)

// WaitForImagingWindowSearch polls until the search is complete.
func (c *Client) WaitForImagingWindowSearch(ctx context.Context, id string, opts *WaitOptions) (*ImagingWindowSearch, error)
```

### Orders API

```go
// CreateOrder creates a new order for downloading imagery.
func (c *Client) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error)

// GetOrder retrieves an order by ID.
func (c *Client) GetOrder(ctx context.Context, id string) (*Order, error)

// ListOrders retrieves all orders with optional filtering.
func (c *Client) ListOrders(ctx context.Context, opts *ListOrdersOptions) iter.Seq2[Order, error]

// CancelOrder cancels an order.
func (c *Client) CancelOrder(ctx context.Context, id string) error

// WaitForOrder polls until the order reaches a terminal state.
func (c *Client) WaitForOrder(ctx context.Context, id string, opts *WaitOptions) (*Order, error)
```

---

## Example Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/paulmach/orb/geojson"
    "github.com/robert.malhotra/go-sar-vendor/pkg/planet"
)

func main() {
    ctx := context.Background()

    // Create client
    client, err := planet.NewClient("your-api-key")
    if err != nil {
        log.Fatal(err)
    }

    // Create a tasking order
    start := time.Now().Add(24 * time.Hour)
    end := start.Add(7 * 24 * time.Hour)

    order, err := client.CreateTaskingOrder(ctx, &planet.CreateTaskingOrderRequest{
        Name:           "My SkySat Order",
        Geometry:       planet.NewPointGeometry(-122.4, 37.8),
        SchedulingType: planet.SchedulingTypeFlexible,
        SatelliteTypes: []planet.SatelliteType{planet.SatelliteTypeSkySat},
        StartTime:      &start,
        EndTime:        &end,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Created order: %s (status: %s)\n", order.ID, order.Status)

    // Search for imaging windows (feasibility)
    search, err := client.CreateImagingWindowSearch(ctx, &planet.ImagingWindowSearchRequest{
        Geometry:  planet.NewPointGeometry(-122.4, 37.8),
        StartTime: &start,
        EndTime:   &end,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Wait for search to complete
    search, err = client.WaitForImagingWindowSearch(ctx, search.ID, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d imaging windows\n", len(search.ImagingWindows))
}
```

---

## Error Handling

Errors delegate to the common package for consistency:

```go
type APIError = common.APIError

var (
    IsNotFound     = common.IsNotFound
    IsRateLimited  = common.IsRateLimited
    IsUnauthorized = common.IsUnauthorized
    IsBadRequest   = common.IsBadRequest
    IsForbidden    = common.IsForbidden
    IsServerError  = common.IsServerError
)
```

---

## Dependencies

- `github.com/paulmach/orb` - GeoJSON support
- `github.com/robert.malhotra/go-sar-vendor/pkg/common` - Shared client infrastructure
