# SAR-API Go Client Design Document

## Overview

This document outlines the design for a Go client library for the **SAR-API (OneAtlas Radar)** — a REST API for searching, ordering, and accessing TerraSAR-X and PAZ radar data. The client leverages the `common` package for HTTP handling and authentication, providing an idiomatic Go interface to all API functionality.

**API Version:** 2.7.0
**Go Version Target:** 1.23+

---

## Goals

- Provide a type-safe, idiomatic Go interface to all SAR-API endpoints
- Use the shared `common.Client` for HTTP operations and authentication
- Support all four server environments (PROD/DEV via OneAtlas and Legacy Account)
- Handle API Key → Bearer token authentication transparently
- Implement comprehensive error handling with structured error types
- Enable extensibility through functional options pattern
- Support context-based cancellation and timeouts
- Use `paulmach/orb` and `paulmach/orb/geojson` for geometry types

---

## Package Structure

```
airbus/
├── client.go           # Main client struct wrapping common.Client
├── auth.go             # API Key to Bearer token authentication
├── errors.go           # Error types
├── types.go            # Shared type definitions and enums
├── user.go             # User service endpoints
├── health.go           # Health and ping endpoints
├── config.go           # Configuration endpoints
├── catalogue.go        # Catalogue search endpoints
├── feasibility.go      # Feasibility search endpoints
├── prices.go           # Pricing endpoints
├── baskets.go          # Basket management endpoints
├── orders.go           # Order management endpoints
├── client_test.go      # Tests
└── spec/               # API specification documents
```

---

## Core Types

### Client

```go
package airbus

import (
    "net/http"
    "time"

    "github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

const (
    // DefaultBaseURL is the production OneAtlas base URL.
    DefaultBaseURL = "https://sar.api.oneatlas.airbus.com/v1"
    // DefaultTokenURL is the production authentication endpoint.
    DefaultTokenURL = "https://authenticate.foundation.api.oneatlas.airbus.com/auth/realms/IDP/protocol/openid-connect/token"

    // DevBaseURL is the development OneAtlas base URL.
    DevBaseURL = "https://dev.sar.api.oneatlas.airbus.com/v1"
    // DevTokenURL is the development authentication endpoint.
    DevTokenURL = "https://authenticate.foundation.api.oneatlas.airbus.com/auth/realms/IDP/protocol/openid-connect/token"

    // LegacyBaseURL is the legacy production base URL.
    LegacyBaseURL = "https://sar.api.intelligence.airbus.com/v1"
    // LegacyDevBaseURL is the legacy development base URL.
    LegacyDevBaseURL = "https://dev.sar.api.intelligence.airbus.com/v1"

    defaultTimeout = 30 * time.Second
)

// Client is the SAR-API client. It embeds common.Client for HTTP operations.
type Client struct {
    *common.Client
}

// Option configures a Client.
type Option func(*clientConfig)

type clientConfig struct {
    baseURL    string
    tokenURL   string
    httpClient *http.Client
    timeout    time.Duration
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option

// WithBaseURL overrides the default base URL.
func WithBaseURL(baseURL string) Option

// WithTokenURL overrides the default token URL.
func WithTokenURL(tokenURL string) Option

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) Option

// NewClient creates a new SAR-API client with the given API key.
func NewClient(apiKey string, opts ...Option) (*Client, error)

// NewDevClient creates a client configured for the development environment.
func NewDevClient(apiKey string, opts ...Option) (*Client, error)
```

---

## Authentication

The SAR-API uses API key authentication that exchanges an API key for a Bearer token.

```go
package airbus

import (
    "context"
    "sync"
    "time"
)

// APIKeyAuth implements Authenticator for Airbus API key authentication.
// It exchanges an API key for a Bearer token via the token endpoint.
type APIKeyAuth struct {
    apiKey   string
    tokenURL string
    client   *http.Client

    mu    sync.Mutex
    token string
    exp   time.Time
}

// NewAPIKeyAuth creates an authenticator that exchanges API key for bearer token.
func NewAPIKeyAuth(apiKey, tokenURL string, client *http.Client) *APIKeyAuth

func (a *APIKeyAuth) Authenticate(ctx context.Context) error
func (a *APIKeyAuth) AuthHeader() string
```

---

## Type Definitions

### Enumerations

```go
// Mission represents the satellite mission.
type Mission string

const (
    MissionTSX Mission = "TSX"
    MissionPAZ Mission = "PAZ"
)

// Satellite represents the satellite identifier.
type Satellite string

const (
    SatelliteTSX1 Satellite = "TSX-1"
    SatellitePAZ1 Satellite = "PAZ-1"
    SatelliteTDX1 Satellite = "TDX-1"
)

// SensorMode represents the SAR sensor/imaging mode.
type SensorMode string

const (
    SensorModeStaringSpotlight SensorMode = "SAR_ST_S"
    SensorModeHighResSpotlight SensorMode = "SAR_HS_S"
    SensorModeHighResSpot300   SensorMode = "SAR_HS_S_300"
    SensorModeHighResSpot150   SensorMode = "SAR_HS_S_150"
    SensorModeHighResDual      SensorMode = "SAR_HS_D"
    SensorModeHighResDual300   SensorMode = "SAR_HS_D_300"
    SensorModeHighResDual150   SensorMode = "SAR_HS_D_150"
    SensorModeSpotlight        SensorMode = "SAR_SL_S"
    SensorModeSpotlightDual    SensorMode = "SAR_SL_D"
    SensorModeStripmap         SensorMode = "SAR_SM_S"
    SensorModeStripmapDual     SensorMode = "SAR_SM_D"
    SensorModeScanSAR          SensorMode = "SAR_SC_S"
    SensorModeWideScanSAR      SensorMode = "SAR_WS_S"
)

// Polarization represents the radar polarization.
type Polarization string

const (
    PolarizationHH   Polarization = "HH"
    PolarizationVV   Polarization = "VV"
    PolarizationHV   Polarization = "HV"
    PolarizationVH   Polarization = "VH"
    PolarizationHHVV Polarization = "HHVV"
    PolarizationHHHV Polarization = "HHHV"
    PolarizationVVVH Polarization = "VVVH"
)

// PathDirection represents the orbit direction.
type PathDirection string

const (
    PathDirectionAscending  PathDirection = "ascending"
    PathDirectionDescending PathDirection = "descending"
)

// LookDirection represents the SAR look direction.
type LookDirection string

const (
    LookDirectionRight LookDirection = "R"
    LookDirectionLeft  LookDirection = "L"
)

// ProductType represents the output product type.
type ProductType string

const (
    ProductTypeSSC ProductType = "SSC" // Single Look Slant Range Complex
    ProductTypeMGD ProductType = "MGD" // Multi Look Ground Range Detected
    ProductTypeGEC ProductType = "GEC" // Geocoded Ellipsoid Corrected
    ProductTypeEEC ProductType = "EEC" // Enhanced Ellipsoid Corrected
)

// ResolutionVariant represents the processing enhancement.
type ResolutionVariant string

const (
    ResolutionVariantSE ResolutionVariant = "SE" // Spatially Enhanced
    ResolutionVariantRE ResolutionVariant = "RE" // Radiometrically Enhanced
)

// OrbitType represents the orbit processing type.
type OrbitType string

const (
    OrbitTypePremiumNRT OrbitType = "Premium NRT"
    OrbitTypeNRT        OrbitType = "NRT"
    OrbitTypeRapid      OrbitType = "rapid"
    OrbitTypeScience    OrbitType = "science"
)

// MapProjection represents the output map projection.
type MapProjection string

const (
    MapProjectionAuto MapProjection = "auto"
    MapProjectionUTM  MapProjection = "UTM"
    MapProjectionUPS  MapProjection = "UPS"
)

// GainAttenuation represents processor gain attenuation values.
type GainAttenuation int

const (
    GainAttenuation0  GainAttenuation = 0
    GainAttenuation10 GainAttenuation = 10
    GainAttenuation20 GainAttenuation = 20
)

// FeasibilityLevel represents the depth of feasibility analysis.
type FeasibilityLevel string

const (
    FeasibilityLevelSimple   FeasibilityLevel = "simple"
    FeasibilityLevelComplete FeasibilityLevel = "complete"
)

// Priority represents order priority.
type Priority string

const (
    PriorityStandard  Priority = "standard"
    PriorityPriority  Priority = "priority"
    PriorityExclusive Priority = "exclusive"
)

// ItemStatus represents the status of an order item.
type ItemStatus string

const (
    ItemStatusPlanned     ItemStatus = "planned"
    ItemStatusAcquired    ItemStatus = "acquired"
    ItemStatusProcessing  ItemStatus = "processing"
    ItemStatusProcessed   ItemStatus = "processed"
    ItemStatusDelivering  ItemStatus = "delivering"
    ItemStatusDelivered   ItemStatus = "delivered"
    ItemStatusCancelled   ItemStatus = "cancelled"
    ItemStatusFailed      ItemStatus = "failed"
)

// Service represents an SAR-API service.
type Service string

const (
    ServiceRadar    Service = "radar"
    ServiceWorldDEM Service = "worlddem"
    ServiceMgmt     Service = "mgmt"
)
```

### Request/Response Types

```go
// TimeRange represents a time window.
type TimeRange struct {
    From time.Time `json:"from"`
    To   time.Time `json:"to,omitempty"`
}

// IncidenceAngleRange represents min/max incidence angles.
type IncidenceAngleRange struct {
    Minimum float64 `json:"minimum,omitempty"`
    Maximum float64 `json:"maximum,omitempty"`
}

// OrderOptions represents processing options for an order.
type OrderOptions struct {
    ProductType          ProductType       `json:"productType,omitempty"`
    ResolutionVariant    ResolutionVariant `json:"resolutionVariant,omitempty"`
    OrbitType            OrbitType         `json:"orbitType,omitempty"`
    MapProjection        MapProjection     `json:"mapProjection,omitempty"`
    GainAttenuation      GainAttenuation   `json:"gainAttenuation,omitempty"`
    GeocodedIncidenceMask bool             `json:"geocodedIncidenceMask,omitempty"`
}

// Price represents pricing information.
type Price struct {
    Final    bool    `json:"final"`
    Total    float64 `json:"total"`
    Currency string  `json:"currency"`
}
```

---

## Service Definitions

### User Service

```go
// WhoAmI returns account information for the current user.
// GET /user/whoami
func (c *Client) WhoAmI(ctx context.Context) (*UserInfo, error)

// ChangePassword changes the account password.
// POST /user/password
func (c *Client) ChangePassword(ctx context.Context, req *ChangePasswordRequest) error

// ResetPassword requests a password reset.
// POST /user/password/reset
func (c *Client) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error

// ListNotifications retrieves user notifications.
// GET /user/notifications
func (c *Client) ListNotifications(ctx context.Context) ([]Notification, error)
```

### Health Service

```go
// Ping checks basic availability of the API.
// GET /sar/ping
func (c *Client) Ping(ctx context.Context) error

// Health returns the health status of the API.
// GET /sar/health
func (c *Client) Health(ctx context.Context) (*HealthStatus, error)
```

### Catalogue Service

```go
// SearchCatalogue searches for existing acquisitions in the archive.
// POST /sar/catalogue
func (c *Client) SearchCatalogue(ctx context.Context, req *CatalogueRequest) (*FeatureCollection, error)

// ReplicateCatalogue retrieves catalogue updates for replication.
// GET /sar/catalogue/replication
func (c *Client) ReplicateCatalogue(ctx context.Context, opts *ReplicationOptions) (*FeatureCollection, error)

// GetCatalogueRevocations retrieves revoked catalogue items.
// GET /sar/catalogue/revocation
func (c *Client) GetCatalogueRevocations(ctx context.Context, opts *RevocationOptions) (*RevocationResponse, error)

// RetrieveCatalogueItems retrieves ordered items from catalogue.
// POST /sar/catalogue/retrieve
func (c *Client) RetrieveCatalogueItems(ctx context.Context, req *RetrieveRequest) (*FeatureCollection, error)
```

### Feasibility Service

```go
// SearchFeasibility searches for possible future acquisitions (tasking).
// POST /sar/feasibility
func (c *Client) SearchFeasibility(ctx context.Context, req *FeasibilityRequest) (*FeatureCollection, error)
```

### Prices Service

```go
// GetPrices queries prices for acquisitions or items.
// POST /sar/prices
func (c *Client) GetPrices(ctx context.Context, req *PricesRequest) ([]PriceResponse, error)
```

### Baskets Service

```go
// ListBaskets returns all baskets for the current user.
// GET /sar/baskets
func (c *Client) ListBaskets(ctx context.Context) ([]Basket, error)

// CreateBasket creates a new basket.
// POST /sar/baskets
func (c *Client) CreateBasket(ctx context.Context, req *CreateBasketRequest) (*Basket, error)

// GetBasket retrieves a basket by ID.
// GET /sar/baskets/{basketId}
func (c *Client) GetBasket(ctx context.Context, basketID string) (*Basket, error)

// UpdateBasket updates basket parameters.
// PATCH /sar/baskets/{basketId}
func (c *Client) UpdateBasket(ctx context.Context, basketID string, req *UpdateBasketRequest) (*Basket, error)

// ReplaceBasket replaces basket parameters.
// PUT /sar/baskets/{basketId}
func (c *Client) ReplaceBasket(ctx context.Context, basketID string, req *ReplaceBasketRequest) (*Basket, error)

// DeleteBasket deletes a basket.
// DELETE /sar/baskets/{basketId}
func (c *Client) DeleteBasket(ctx context.Context, basketID string) error

// AddItemsToBasket adds acquisitions or items to a basket.
// POST /sar/baskets/{basketId}/addItems
func (c *Client) AddItemsToBasket(ctx context.Context, basketID string, req *AddItemsRequest) (*Basket, error)

// RemoveItemsFromBasket removes items from a basket.
// POST /sar/baskets/{basketId}/removeItems
func (c *Client) RemoveItemsFromBasket(ctx context.Context, basketID string, req *RemoveItemsRequest) (*Basket, error)

// RearrangeBasketItems rearranges items within a basket.
// POST /sar/baskets/{basketId}/rearrangeItems
func (c *Client) RearrangeBasketItems(ctx context.Context, basketID string, req *RearrangeItemsRequest) (*Basket, error)

// SubmitBasket submits a basket as an order.
// POST /sar/baskets/{basketId}/submit
func (c *Client) SubmitBasket(ctx context.Context, basketID string) (*Order, error)
```

### Orders Service

```go
// ListOrders returns all orders for the current user.
// GET /sar/orders
func (c *Client) ListOrders(ctx context.Context) ([]OrderSummary, error)

// GetOrder retrieves an order by ID or basket ID.
// GET /sar/orders/{orderIdOrBasketId}
func (c *Client) GetOrder(ctx context.Context, orderID string) (*Order, error)

// UpdateOrder updates order parameters (e.g., notify endpoint).
// PATCH /sar/orders/{orderIdOrBasketId}
func (c *Client) UpdateOrder(ctx context.Context, orderID string, req *UpdateOrderRequest) (*Order, error)

// GetOrderItems retrieves items by order item IDs.
// POST /sar/orders/*/items
func (c *Client) GetOrderItems(ctx context.Context, req *GetOrderItemsRequest) (*FeatureCollection, error)

// GetOrderItemsStatus retrieves item status with filters.
// POST /sar/orders/*/items/status
func (c *Client) GetOrderItemsStatus(ctx context.Context, req *GetOrderItemsStatusRequest) (*OrderItemsStatusResponse, error)

// CancelOrderItems cancels ordered items.
// POST /sar/orders/cancel
func (c *Client) CancelOrderItems(ctx context.Context, req *CancelItemsRequest) (*CancelItemsResponse, error)

// ReorderItems reorders items with different order options.
// POST /sar/orders/reorder
func (c *Client) ReorderItems(ctx context.Context, req *ReorderRequest) (*Basket, error)

// SubmitOrder submits an order directly.
// POST /sar/orders/submit
func (c *Client) SubmitOrder(ctx context.Context, req *SubmitOrderRequest) (*Order, error)
```

### Config Service

```go
// GetConfig retrieves the entire user configuration.
// GET /sar/config
func (c *Client) GetConfig(ctx context.Context) (*Config, error)

// GetPermissions retrieves user permissions.
// GET /sar/config/permissions
func (c *Client) GetPermissions(ctx context.Context) (*Permissions, error)

// GetSettings retrieves user settings.
// GET /sar/config/settings
func (c *Client) GetSettings(ctx context.Context) (*Settings, error)

// GetCustomers retrieves available customers (for resellers).
// GET /sar/config/customers
func (c *Client) GetCustomers(ctx context.Context) ([]Customer, error)

// GetOrderTemplates retrieves available order templates.
// GET /sar/config/orderTemplates
func (c *Client) GetOrderTemplates(ctx context.Context) ([]OrderTemplate, error)

// GetAssociatedUsers retrieves associated users.
// GET /sar/config/associations
func (c *Client) GetAssociatedUsers(ctx context.Context) ([]Association, error)

// GetReceivingStations retrieves allowed receiving stations.
// GET /sar/config/receivingStations
func (c *Client) GetReceivingStations(ctx context.Context) ([]ReceivingStation, error)
```

---

## GeoJSON Types

```go
import (
    "github.com/paulmach/orb"
    "github.com/paulmach/orb/geojson"
)

// Geometry is an alias for geojson.Geometry.
type Geometry = geojson.Geometry

// NewPointGeometry creates a GeoJSON Point.
func NewPointGeometry(lon, lat float64) *geojson.Geometry

// NewPolygonGeometry creates a GeoJSON Polygon.
func NewPolygonGeometry(coords [][][2]float64) *geojson.Geometry

// Feature represents a GeoJSON Feature with acquisition properties.
type Feature struct {
    Type       string                 `json:"type"` // "Feature"
    Geometry   *geojson.Geometry      `json:"geometry"`
    Properties AcquisitionProperties  `json:"properties"`
}

// FeatureCollection represents a GeoJSON FeatureCollection.
type FeatureCollection struct {
    Type     string    `json:"type"` // "FeatureCollection"
    Features []Feature `json:"features"`
    Limit    int       `json:"limit,omitempty"`
    Total    int       `json:"total,omitempty"`
}
```

---

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/robert-malhotra/go-sar-vendor/pkg/airbus"
)

func main() {
    // Create client with API key
    client, err := airbus.NewClient("your-api-key")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Check API availability
    if err := client.Ping(ctx); err != nil {
        log.Fatal("API unavailable:", err)
    }

    // Get user info
    user, err := client.WhoAmI(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Logged in as: %s\n", user.Username)

    // Search catalogue
    results, err := client.SearchCatalogue(ctx, &airbus.CatalogueRequest{
        AOI: airbus.NewPolygonGeometry([][][2]float64{{
            {9.0, 47.0}, {10.0, 47.0}, {10.0, 48.0}, {9.0, 48.0}, {9.0, 47.0},
        }}),
        Time: &airbus.TimeRange{
            From: time.Now().AddDate(0, -1, 0),
            To:   time.Now(),
        },
        Limit: 10,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d acquisitions\n", len(results.Features))
}
```

### Creating an Order

```go
ctx := context.Background()

// Create a basket
basket, err := client.CreateBasket(ctx, &airbus.CreateBasketRequest{
    CustomerReference: "PROJECT-2024-001",
})
if err != nil {
    log.Fatal(err)
}

// Add items to basket
basket, err = client.AddItemsToBasket(ctx, basket.BasketID, &airbus.AddItemsRequest{
    Acquisitions: []string{"TSX-1_ST_S_spot_049R_49677_D31767159_432"},
    OrderOptions: &airbus.OrderOptions{
        ProductType:       airbus.ProductTypeEEC,
        ResolutionVariant: airbus.ResolutionVariantRE,
        OrbitType:         airbus.OrbitTypeScience,
        MapProjection:     airbus.MapProjectionAuto,
        GainAttenuation:   airbus.GainAttenuation0,
    },
})
if err != nil {
    log.Fatal(err)
}

// Submit the order
order, err := client.SubmitBasket(ctx, basket.BasketID)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Order submitted: %s\n", order.OrderID)
```

---

## Testing Strategy

### Unit Tests

- Mock HTTP responses using `httptest.Server`
- Test authentication flow (API key → Bearer token)
- Test all service methods with success and error cases
- Test error parsing

### Integration Tests

- Use build tag `//go:build integration`
- Test against development environment
- Require API key via environment variables
- Cover full workflows (search → basket → order → monitor)

```go
//go:build integration

func TestIntegration_FullWorkflow(t *testing.T) {
    client, err := airbus.NewDevClient(os.Getenv("AIRBUS_API_KEY"))
    if err != nil {
        t.Fatal(err)
    }
    // Test workflow...
}
```

---

## Dependencies

**Required:**
- Go 1.23+
- `github.com/paulmach/orb` — GeoJSON geometry handling
- `github.com/robert-malhotra/go-sar-vendor/pkg/common` — Shared HTTP client

---

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 0.1.0   | TBD  | Initial release |
