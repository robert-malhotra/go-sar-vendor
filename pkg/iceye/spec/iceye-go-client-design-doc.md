# ICEYE API Go Client — Design Document

**Version:** 2.0
**Date:** December 2024
**Status:** Final
**API Version:** v1 (Company, Tasking, Catalog, Delivery)

---

## 1. Executive Summary

This document outlines the design for a Go client library that provides programmatic access to the ICEYE API Platform. The ICEYE platform enables direct access to the world's largest SAR (Synthetic Aperture Radar) satellite constellation, providing capabilities for satellite tasking, catalog browsing, and purchasing archived imagery.

The Go client provides an idiomatic, type-safe interface to the four main API domains:
- **Company API v1** — Contract and budget management
- **Tasking API v1** — Satellite task scheduling and management
- **Catalog API v1** — Archive browsing and purchasing
- **Delivery API v1** — Product delivery management

---

## 2. API Overview

### 2.1 Base Configuration

| Property | Value |
|----------|-------|
| Base URL | `https://platform.iceye.com/api` |
| Authentication | OAuth2 Client Credentials / Resource Owner Password |
| Token Endpoint | Provided per-customer |
| Token Validity | ~3600 seconds (1 hour) |
| API Versions | Company v1, Tasking v1, Catalog v1, Delivery v1 |

### 2.2 Authentication Flow

The ICEYE API supports two OAuth2 authentication flows:

**Client Credentials Flow (Recommended)**
```
POST {TOKEN_URL}
Authorization: Basic {base64(client_id:client_secret)}
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
```

**Resource Owner Password Flow (Legacy)**
```
POST {TOKEN_URL}
Authorization: Basic {API_KEY}
Content-Type: application/x-www-form-urlencoded

grant_type=password
username={API_USERNAME}
password={API_PASSWORD}
```

**Response:**
```json
{
  "token_type": "Bearer",
  "expires_in": 3600,
  "access_token": "XXXXXXXX",
  "scope": "catalog.read deliveries.read orders.read..."
}
```

---

## 3. Package Structure

```
github.com/yourorg/iceye-go/
├── iceye/
│   ├── client.go      # Main client, configuration, HTTP handling
│   ├── types.go       # Common types (Point, TimeWindow, etc.)
│   ├── errors.go      # Error types and handling
│   ├── company.go     # Company API (contracts, summary)
│   ├── tasking.go     # Tasking API (tasks, products, pricing)
│   ├── catalog.go     # Catalog API (items, search, purchases)
│   └── delivery.go    # Delivery API (deliveries, location configs)
│
├── examples/
│   ├── authentication/
│   ├── tasking/
│   ├── catalog/
│   └── delivery/
│
└── go.mod
```

---

## 4. Core Types

### 4.1 Client Configuration

```go
package iceye

type Config struct {
    // BaseURL is the API base URL (default: https://platform.iceye.com/api)
    BaseURL string

    // TokenURL is the OAuth2 token endpoint URL
    TokenURL string

    // ClientID for OAuth2 Client Credentials flow
    ClientID string

    // ClientSecret for OAuth2 Client Credentials flow
    ClientSecret string

    // ResourceOwner auth (legacy) - if set, uses password grant instead
    ResourceOwner *ResourceOwnerAuth

    // HTTPClient allows custom HTTP client (for testing, retries, etc.)
    HTTPClient *http.Client

    // UserAgent for API requests
    UserAgent string

    // Timeout for API requests
    Timeout time.Duration
}

type ResourceOwnerAuth struct {
    APIKey   string // Base64-encoded client credentials
    Username string
    Password string
}

type Client struct {
    cfg   Config
    mu    sync.Mutex // guards token+exp
    token string
    exp   time.Time
}
```

### 4.2 Common Types

```go
// Point represents a WGS84 coordinate point
type Point struct {
    Lat float64 `json:"lat"`
    Lon float64 `json:"lon"`
}

// TimeWindow represents a time range
type TimeWindow struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
}

// IncidenceAngle represents an angle range in degrees
type IncidenceAngle struct {
    Min float64 `json:"min"`
    Max float64 `json:"max"`
}

// BoundingBox represents a geographic bounding box [minLon, minLat, maxLon, maxLat]
type BoundingBox [4]float64

// Price represents a currency amount in minor units
type Price struct {
    Amount   int64  `json:"amount"`   // In minor currency unit (e.g., cents)
    Currency string `json:"currency"` // ISO 4217 code (e.g., "EUR", "USD")
}

// DeliveryLocation specifies where products are delivered
type DeliveryLocation struct {
    ConfigID string `json:"configID,omitempty"`
    Method   string `json:"method"` // "s3"
    Path     string `json:"path"`
}

// NotificationConfig specifies webhook notifications
type NotificationConfig struct {
    Webhook *WebhookConfig `json:"webhook,omitempty"`
}

type WebhookConfig struct {
    ID string `json:"id"`
}
```

### 4.3 Error Types

```go
// Error represents an ICEYE API error (RFC 7807 Problem Details)
type Error struct {
    Status int    `json:"status"`
    Code   string `json:"code"`
    Detail string `json:"detail"`
    Title  string `json:"title,omitempty"`
    Type   string `json:"type,omitempty"`
}

func (e *Error) Error() string {
    if e.Detail != "" {
        return fmt.Sprintf("iceye: %s (%d) – %s", e.Code, e.Status, e.Detail)
    }
    return fmt.Sprintf("iceye: %s (%d)", e.Code, e.Status)
}

// Common error codes
const (
    ErrCodeInsufficientFunds = "ERR_INSUFFICIENT_FUNDS"
    ErrCodeInvalidContract   = "ERR_INVALID_CONTRACT"
    ErrCodeTaskNotFound      = "ERR_TASK_NOT_FOUND"
    ErrCodeSceneUnavailable  = "ERR_SCENE_UNAVAILABLE"
)
```

---

## 5. Company API v1

Base path: `/company/v1`

### 5.1 Types

```go
// Contract represents an ICEYE contract
type Contract struct {
    ID                 string             `json:"id"`
    Name               string             `json:"name"`
    Start              time.Time          `json:"start"`
    End                time.Time          `json:"end"`
    DeliveryLocations  []DeliveryLocation `json:"deliveryLocations,omitempty"`
    ImagingModes       *OptionConfig      `json:"imagingModes,omitempty"`
    Priority           *OptionConfig      `json:"priority,omitempty"`
    Exclusivity        *OptionConfig      `json:"exclusivity,omitempty"`
    SLA                *OptionConfig      `json:"sla,omitempty"`
    EULA               *OptionConfig      `json:"eula,omitempty"`
    CatalogCollections *OptionConfig      `json:"catalogCollections,omitempty"`
}

// OptionConfig specifies allowed values and default for an option
type OptionConfig struct {
    Allowed []string `json:"allowed"`
    Default string   `json:"default"`
}

// Summary represents contract budget information (AccountSummary)
type Summary struct {
    ContractID        string `json:"contractID"`
    Currency          string `json:"currency"`          // "EUR" or "USD"
    SpendLimit        int64  `json:"spendLimit"`        // Minor currency unit
    OnHold            int64  `json:"onHold"`            // Minor currency unit
    ConsolidatedSpent int64  `json:"consolidatedSpent"` // Minor currency unit
}

// ContractsResponse is the paginated response for listing contracts
type ContractsResponse struct {
    Data   []Contract `json:"data"`
    Cursor string     `json:"cursor,omitempty"`
}
```

### 5.2 Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/company/v1/contracts` | List all contracts |
| GET | `/company/v1/contracts/{contractID}` | Get contract details |
| GET | `/company/v1/contracts/{contractID}/summary` | Get budget summary |

### 5.3 Client Methods

```go
// ListContracts retrieves all contracts for the authenticated company
// GET /company/v1/contracts
func (c *Client) ListContracts(ctx context.Context, pageSize int) iter.Seq2[ContractsResponse, error]

// GetContract retrieves a specific contract by ID
// GET /company/v1/contracts/{contractID}
func (c *Client) GetContract(ctx context.Context, contractID string) (*Contract, error)

// GetSummary retrieves budget summary for a contract
// GET /company/v1/contracts/{contractID}/summary
func (c *Client) GetSummary(ctx context.Context, contractID string) (*Summary, error)
```

---

## 6. Tasking API v1

Base path: `/tasking/v1`

### 6.1 Enums

```go
// TaskStatus represents the lifecycle state of a task
type TaskStatus string

const (
    TaskStatusReceived  TaskStatus = "RECEIVED"
    TaskStatusActive    TaskStatus = "ACTIVE"
    TaskStatusFulfilled TaskStatus = "FULFILLED"
    TaskStatusRejected  TaskStatus = "REJECTED"
    TaskStatusCanceled  TaskStatus = "CANCELED"
    TaskStatusDone      TaskStatus = "DONE"
    TaskStatusFailed    TaskStatus = "FAILED"
)

// Exclusivity represents data exclusivity options
type Exclusivity string

const (
    ExclusivityPrivate Exclusivity = "PRIVATE"
    ExclusivityPublic  Exclusivity = "PUBLIC"
)

// Priority represents task scheduling priority
type Priority string

const (
    PriorityBackground Priority = "BACKGROUND"
    PriorityCommercial Priority = "COMMERCIAL"
)

// LookSide represents satellite look direction
type LookSide string

const (
    LookSideLeft  LookSide = "LEFT"
    LookSideRight LookSide = "RIGHT"
    LookSideAny   LookSide = "ANY"
)

// PassDirection represents orbital pass direction
type PassDirection string

const (
    PassDirectionAscending  PassDirection = "ASCENDING"
    PassDirectionDescending PassDirection = "DESCENDING"
    PassDirectionAny        PassDirection = "ANY"
)

// AdditionalProductType represents additional SAR product types beyond standard
type AdditionalProductType string

const (
    AdditionalProductTypeSIDD AdditionalProductType = "SIDD"
    AdditionalProductTypeSICD AdditionalProductType = "SICD"
)

// EULA represents end-user license agreement types
type EULA string

const (
    EULAStandard   EULA = "STANDARD"
    EULAGovernment EULA = "GOVERNMENT"
    EULAMulti      EULA = "MULTI"
)
```

### 6.2 Task Types

```go
// Task represents a satellite imaging task
type Task struct {
    ID                     string                  `json:"id"`
    ContractID             string                  `json:"contractID"`
    PointOfInterest        Point                   `json:"pointOfInterest"`
    AcquisitionWindow      TimeWindow              `json:"acquisitionWindow"`
    ImagingMode            string                  `json:"imagingMode"`
    Status                 TaskStatus              `json:"status"`
    Exclusivity            Exclusivity             `json:"exclusivity,omitempty"`
    Priority               Priority                `json:"priority,omitempty"`
    SLA                    string                  `json:"sla,omitempty"`
    EULA                   EULA                    `json:"eula,omitempty"`
    AdditionalProductTypes []AdditionalProductType `json:"additionalProductTypes,omitempty"`
    IncidenceAngle         *IncidenceAngle         `json:"incidenceAngle,omitempty"`
    LookSide               LookSide                `json:"lookSide,omitempty"`
    PassDirection          PassDirection           `json:"passDirection,omitempty"`
    DeliveryLocations      []DeliveryLocation      `json:"deliveryLocations,omitempty"`
    CreatedAt              time.Time               `json:"createdAt"`
    UpdatedAt              time.Time               `json:"updatedAt"`
}

// CreateTaskRequest represents parameters for creating a new task
type CreateTaskRequest struct {
    // Required fields
    ContractID        string     `json:"contractID"`
    PointOfInterest   Point      `json:"pointOfInterest"`
    AcquisitionWindow TimeWindow `json:"acquisitionWindow"`
    ImagingMode       string     `json:"imagingMode"`

    // Optional fields
    Exclusivity            Exclusivity             `json:"exclusivity,omitempty"`
    Priority               Priority                `json:"priority,omitempty"`
    SLA                    string                  `json:"sla,omitempty"`
    EULA                   EULA                    `json:"eula,omitempty"`
    AdditionalProductTypes []AdditionalProductType `json:"additionalProductTypes,omitempty"`
    IncidenceAngle         *IncidenceAngle         `json:"incidenceAngle,omitempty"`
    LookSide               LookSide                `json:"lookSide,omitempty"`
    PassDirection          PassDirection           `json:"passDirection,omitempty"`
    DeliveryLocations      []DeliveryLocation      `json:"deliveryLocations,omitempty"`
}

// TaskScene represents imaging parameters for a scheduled task
type TaskScene struct {
    ImagingTime   TimeWindow    `json:"imagingTime"`
    Duration      int           `json:"duration"` // seconds
    LookSide      LookSide      `json:"lookSide"`
    PassDirection PassDirection `json:"passDirection"`
}

// TaskProduct represents a SAR data product from a completed task
type TaskProduct struct {
    Type   string           `json:"type"`
    Assets map[string]Asset `json:"assets"`
}

// TaskPrice represents a price quotation for a task
type TaskPrice struct {
    Amount   int64  `json:"amount"`   // Minor currency unit (e.g., cents)
    Currency string `json:"currency"` // ISO 4217 code
}

// TaskPriceRequest contains all parameters for getting a task price quote
type TaskPriceRequest struct {
    ContractID      string      `json:"contractID"`
    PointOfInterest Point       `json:"pointOfInterest"`
    ImagingMode     string      `json:"imagingMode"`
    Exclusivity     Exclusivity `json:"exclusivity"`
    Priority        Priority    `json:"priority"`
    SLA             string      `json:"sla"`
    EULA            EULA        `json:"eula"`
}

// ListTasksOptions for filtering task lists
type ListTasksOptions struct {
    ContractID    string
    CreatedAfter  *time.Time
    CreatedBefore *time.Time
}

// TasksResponse is the paginated response for listing tasks
type TasksResponse struct {
    Data   []Task `json:"data"`
    Cursor string `json:"cursor,omitempty"`
}
```

### 6.3 Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/tasking/v1/price` | Get task price quote |
| POST | `/tasking/v1/tasks` | Create task |
| GET | `/tasking/v1/tasks` | List tasks |
| GET | `/tasking/v1/tasks/{taskID}` | Get task |
| PATCH | `/tasking/v1/tasks/{taskID}` | Cancel task (set status to CANCELED) |
| GET | `/tasking/v1/tasks/{taskID}/products` | List task products |
| GET | `/tasking/v1/tasks/{taskID}/products/{productType}` | Get specific product |
| GET | `/tasking/v1/tasks/{taskID}/scene` | Get task scene |

### 6.4 Client Methods

```go
// GetTaskPrice gets a price quotation for task parameters
// GET /tasking/v1/price
func (c *Client) GetTaskPrice(ctx context.Context, req *TaskPriceRequest) (*TaskPrice, error)

// CreateTask creates a new satellite imaging task
// POST /tasking/v1/tasks
func (c *Client) CreateTask(ctx context.Context, req *CreateTaskRequest) (*Task, error)

// ListTasks lists tasks with optional filters
// GET /tasking/v1/tasks
func (c *Client) ListTasks(ctx context.Context, pageSize int, opts *ListTasksOptions) iter.Seq2[[]Task, error]

// GetTask retrieves a task by ID
// GET /tasking/v1/tasks/{taskID}
func (c *Client) GetTask(ctx context.Context, taskID string) (*Task, error)

// CancelTask cancels an active task by setting its status to CANCELED
// PATCH /tasking/v1/tasks/{taskID}
func (c *Client) CancelTask(ctx context.Context, taskID string) (*Task, error)

// ListTaskProducts retrieves all available products for a fulfilled task
// GET /tasking/v1/tasks/{taskID}/products
func (c *Client) ListTaskProducts(ctx context.Context, taskID string) ([]TaskProduct, error)

// GetTaskProduct retrieves a specific product type for a task
// GET /tasking/v1/tasks/{taskID}/products/{productType}
func (c *Client) GetTaskProduct(ctx context.Context, taskID string, productType string) (*TaskProduct, error)

// GetTaskScene retrieves planned imaging parameters for a task
// GET /tasking/v1/tasks/{taskID}/scene
func (c *Client) GetTaskScene(ctx context.Context, taskID string) (*TaskScene, error)
```

---

## 7. Catalog API v1

Base path: `/catalog/v1`

### 7.1 Types

```go
// STACItem represents a STAC-compliant catalog item (GeoJSON Feature)
type STACItem struct {
    ID          string               `json:"id"`
    Type        string               `json:"type"` // "Feature"
    StacVersion string               `json:"stac_version"`
    Geometry    Geometry             `json:"geometry"`
    BBox        BoundingBox          `json:"bbox"`
    Properties  ItemProperties       `json:"properties"`
    Assets      map[string]ItemAsset `json:"assets"`
    Links       []Link               `json:"links,omitempty"`
}

// ItemProperties contains STAC item properties
type ItemProperties struct {
    // Temporal
    StartTime time.Time `json:"start_time,omitempty"`
    EndTime   time.Time `json:"end_time,omitempty"`

    // Imaging parameters
    InstrumentMode       string   `json:"instrument_mode,omitempty"`
    ProductType          string   `json:"product_type,omitempty"`
    ObservationDirection string   `json:"observation_direction,omitempty"`
    SatelliteLookAngle   float64  `json:"satellite_look_angle,omitempty"`
    IncidenceAngle       float64  `json:"incidence_angle,omitempty"`
    Polarizations        []string `json:"polarizations,omitempty"`
    ImageMode            string   `json:"image_mode,omitempty"`
    OrbitState           string   `json:"orbit_state,omitempty"`
}

// ItemAsset represents an asset in a STAC item
type ItemAsset struct {
    Href        string           `json:"href"`
    Title       string           `json:"title,omitempty"`
    Type        string           `json:"type,omitempty"` // MIME type
    Roles       []string         `json:"roles,omitempty"`
    Coordinates *ThumbnailCoords `json:"coordinates,omitempty"`
}

// ThumbnailCoords contains corner coordinates for thumbnails
type ThumbnailCoords struct {
    TopLeft     [2]float64 `json:"top_left"`
    TopRight    [2]float64 `json:"top_right"`
    BottomLeft  [2]float64 `json:"bottom_left"`
    BottomRight [2]float64 `json:"bottom_right"`
}

// Link represents a STAC link
type Link struct {
    Href  string `json:"href"`
    Rel   string `json:"rel"`
    Type  string `json:"type,omitempty"`
    Title string `json:"title,omitempty"`
}

// PurchaseStatus represents the status of a catalog purchase
type PurchaseStatus string

const (
    PurchaseStatusReceived PurchaseStatus = "received"
    PurchaseStatusActive   PurchaseStatus = "active"
    PurchaseStatusClosed   PurchaseStatus = "closed"
    PurchaseStatusCanceled PurchaseStatus = "canceled"
    PurchaseStatusFailed   PurchaseStatus = "failed"
)

// Purchase represents a catalog purchase (PurchasedItem)
type Purchase struct {
    ID           string         `json:"id"`
    CustomerName string         `json:"customerName"`
    ContractName string         `json:"contractName"`
    CreatedAt    time.Time      `json:"createdAt"`
    Status       PurchaseStatus `json:"status"`
    Reference    string         `json:"reference,omitempty"`
}

// PurchaseRequest for creating a catalog purchase
type PurchaseRequest struct {
    ItemIDs       []string `json:"itemIds"`
    ContractID    string   `json:"contractId"`
    CompanyName   string   `json:"companyName,omitempty"`
    ContactPerson string   `json:"contactPerson,omitempty"`
    Reference     string   `json:"reference,omitempty"`
}

// PurchaseResponse is returned when creating a purchase (202 Accepted)
type PurchaseResponse struct {
    PurchaseID string `json:"purchaseId"`
}

// ListItemsOptions for GET /catalog/v1/items
type ListItemsOptions struct {
    IDs      []string     // Specific item IDs to retrieve
    BBox     *BoundingBox // Bounding box filter [minLon, minLat, maxLon, maxLat]
    Datetime string       // RFC 3339 datetime or interval
    SortBy   []string     // Properties with +/- prefix for asc/desc
}

// SearchRequest for POST /catalog/v1/search
type SearchRequest struct {
    IDs      []string               `json:"ids,omitempty"`
    BBox     *BoundingBox           `json:"bbox,omitempty"`
    Datetime string                 `json:"datetime,omitempty"`
    Limit    int                    `json:"limit,omitempty"`
    Query    map[string]QueryFilter `json:"query,omitempty"`
    SortBy   []SortCondition        `json:"sortby,omitempty"`
}

// QueryFilter for advanced STAC queries
// Supported operators: eq, neq, gt, lt, gte, lte, startsWith, endsWith, contains, in
type QueryFilter struct {
    Eq         any    `json:"eq,omitempty"`
    Neq        any    `json:"neq,omitempty"`
    Lt         any    `json:"lt,omitempty"`
    Lte        any    `json:"lte,omitempty"`
    Gt         any    `json:"gt,omitempty"`
    Gte        any    `json:"gte,omitempty"`
    StartsWith string `json:"startsWith,omitempty"`
    EndsWith   string `json:"endsWith,omitempty"`
    Contains   string `json:"contains,omitempty"`
    In         []any  `json:"in,omitempty"`
}

// SortCondition for sorting search results
type SortCondition struct {
    Field     string `json:"field"`
    Direction string `json:"direction"` // "asc" or "desc"
}

// CatalogResponse is the paginated response for catalog items
type CatalogResponse struct {
    Data   []STACItem `json:"data"`
    Cursor string     `json:"cursor,omitempty"`
}

// PurchasesResponse is the paginated response for listing purchases
type PurchasesResponse struct {
    Data   []Purchase `json:"data"`
    Cursor string     `json:"cursor,omitempty"`
}
```

### 7.2 Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/catalog/v1/items` | List catalog items (max 200, default 10) |
| POST | `/catalog/v1/search` | Search catalog items |
| POST | `/catalog/v1/purchases` | Purchase catalog items (returns 202) |
| GET | `/catalog/v1/purchases` | List purchases |
| GET | `/catalog/v1/purchases/{purchaseID}` | Get purchase status |
| GET | `/catalog/v1/purchases/{purchaseID}/products` | List products in purchase |

### 7.3 Client Methods

```go
// ListCatalogItems lists catalog items with optional filters
// GET /catalog/v1/items
func (c *Client) ListCatalogItems(ctx context.Context, pageSize int, opts *ListItemsOptions) iter.Seq2[CatalogResponse, error]

// SearchCatalogItems performs an advanced catalog search
// POST /catalog/v1/search
func (c *Client) SearchCatalogItems(ctx context.Context, req *SearchRequest) iter.Seq2[CatalogResponse, error]

// PurchaseCatalogItems purchases catalog items
// POST /catalog/v1/purchases
func (c *Client) PurchaseCatalogItems(ctx context.Context, req *PurchaseRequest) (*PurchaseResponse, error)

// ListPurchases lists all purchases for the authenticated user
// GET /catalog/v1/purchases
func (c *Client) ListPurchases(ctx context.Context, pageSize int) iter.Seq2[PurchasesResponse, error]

// GetPurchase retrieves purchase status by ID
// GET /catalog/v1/purchases/{purchaseID}
func (c *Client) GetPurchase(ctx context.Context, purchaseID string) (*Purchase, error)

// ListPurchasedProducts retrieves all products from a completed purchase
// GET /catalog/v1/purchases/{purchaseID}/products
func (c *Client) ListPurchasedProducts(ctx context.Context, purchaseID string) ([]STACItem, error)
```

---

## 8. Delivery API v1

Base path: `/delivery/v1`

### 8.1 Types

```go
// DeliveryStatus represents the status of a delivery
type DeliveryStatus string

const (
    DeliveryStatusSuccess DeliveryStatus = "SUCCESS"
    DeliveryStatusPending DeliveryStatus = "PENDING"
    DeliveryStatusFailed  DeliveryStatus = "FAILED"
)

// DeliveryLocationConfigStatus represents location config status
type DeliveryLocationConfigStatus string

const (
    DeliveryLocationConfigStatusActive   DeliveryLocationConfigStatus = "active"
    DeliveryLocationConfigStatusInactive DeliveryLocationConfigStatus = "inactive"
)

// DeliveryLocationConfig represents a delivery location configuration
type DeliveryLocationConfig struct {
    ID     string                       `json:"id"`
    Method string                       `json:"method"` // "s3"
    Config S3Config                     `json:"config"`
    Status DeliveryLocationConfigStatus `json:"status"`
}

// S3Config contains S3-specific delivery configuration
type S3Config struct {
    Endpoint string `json:"endpoint"`
    Bucket   string `json:"bucket"`
    Region   string `json:"region"`
    KeyID    string `json:"keyID"`
}

// Delivery represents a delivery response
type Delivery struct {
    ID                string                    `json:"id"`
    Status            DeliveryStatus            `json:"status"`
    URL               string                    `json:"url,omitempty"`
    ItemIDs           []string                  `json:"itemIDs"`
    DeliveryLocations []DeliveryLocation        `json:"deliveryLocations"`
    Notifications     *DeliveryNotificationRef  `json:"notifications,omitempty"`
}

// DeliveryNotificationRef references a notification subscription
type DeliveryNotificationRef struct {
    Subscription *SubscriptionRef `json:"subscription,omitempty"`
}

// SubscriptionRef references a notification subscription by ID
type SubscriptionRef struct {
    ID string `json:"id"`
}

// CreateDeliveryRequest for creating a new delivery
type CreateDeliveryRequest struct {
    ItemIDs           []string           `json:"itemIDs"`
    ContractID        string             `json:"contractID"`
    DeliveryLocations []DeliveryLocation `json:"deliveryLocations"`
}

// DeliveriesResponse is the paginated response for listing deliveries
type DeliveriesResponse struct {
    Data   []Delivery `json:"data"`
    Cursor string     `json:"cursor,omitempty"`
}

// ListDeliveriesOptions for filtering delivery lists
type ListDeliveriesOptions struct {
    Type string // Optional type filter
}
```

### 8.2 Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/delivery/v1/deliveries/location-configs` | List delivery location configs |
| GET | `/delivery/v1/deliveries` | List deliveries (limit 1-100) |
| GET | `/delivery/v1/deliveries/{ID}` | Get a delivery |
| POST | `/delivery/v1/deliveries` | Create a delivery |

### 8.3 Client Methods

```go
// ListDeliveryLocationConfigs retrieves all delivery location configurations
// GET /delivery/v1/deliveries/location-configs
func (c *Client) ListDeliveryLocationConfigs(ctx context.Context) ([]DeliveryLocationConfig, error)

// ListDeliveries lists deliveries with optional filters
// GET /delivery/v1/deliveries
func (c *Client) ListDeliveries(ctx context.Context, pageSize int, opts *ListDeliveriesOptions) iter.Seq2[DeliveriesResponse, error]

// GetDelivery retrieves a delivery by ID
// GET /delivery/v1/deliveries/{ID}
func (c *Client) GetDelivery(ctx context.Context, deliveryID string) (*Delivery, error)

// CreateDelivery creates a new delivery
// POST /delivery/v1/deliveries
func (c *Client) CreateDelivery(ctx context.Context, req *CreateDeliveryRequest) (*Delivery, error)
```

---

## 9. GeoJSON Support

```go
// Geometry represents a GeoJSON geometry
type Geometry struct {
    Type        string `json:"type"`
    Coordinates any    `json:"coordinates"`
}

// GeoJSONPoint creates a GeoJSON Point geometry
func GeoJSONPoint(lon, lat float64) Geometry {
    return Geometry{
        Type:        "Point",
        Coordinates: []float64{lon, lat},
    }
}

// GeoJSONPolygon creates a GeoJSON Polygon from a linear ring
func GeoJSONPolygon(coordinates [][][]float64) Geometry {
    return Geometry{
        Type:        "Polygon",
        Coordinates: coordinates,
    }
}

// BBoxToPolygon converts a bounding box to a GeoJSON polygon
func BBoxToPolygon(bbox BoundingBox) Geometry {
    minLon, minLat, maxLon, maxLat := bbox[0], bbox[1], bbox[2], bbox[3]
    return GeoJSONPolygon([][][]float64{{
        {minLon, minLat},
        {maxLon, minLat},
        {maxLon, maxLat},
        {minLon, maxLat},
        {minLon, minLat}, // Close the ring
    }})
}
```

---

## 10. Usage Examples

### 10.1 Client Initialization

```go
package main

import (
    "context"
    "log"

    "github.com/yourorg/go-sar-vendor/pkg/iceye"
)

func main() {
    client := iceye.NewClient(iceye.Config{
        BaseURL:      "https://platform.iceye.com/api",
        TokenURL:     "https://auth.iceye.com/oauth2/token",
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })

    ctx := context.Background()

    // List contracts
    for resp, err := range client.ListContracts(ctx, 100) {
        if err != nil {
            log.Fatal(err)
        }
        for _, c := range resp.Data {
            log.Printf("Contract: %s (%s)", c.Name, c.ID)
        }
    }
}
```

### 10.2 Creating a Tasking Request

```go
func createTask(ctx context.Context, client *iceye.Client, contractID string) error {
    // Define acquisition window
    now := time.Now()
    window := iceye.TimeWindow{
        Start: now.Add(48 * time.Hour),
        End:   now.Add(72 * time.Hour),
    }

    task, err := client.CreateTask(ctx, &iceye.CreateTaskRequest{
        ContractID: contractID,
        PointOfInterest: iceye.Point{
            Lat: 60.1699,
            Lon: 24.9384, // Helsinki
        },
        AcquisitionWindow: window,
        ImagingMode:       "SPOTLIGHT",
        Priority:          iceye.PriorityCommercial,
        Exclusivity:       iceye.ExclusivityPrivate,
        EULA:              iceye.EULAStandard,
    })
    if err != nil {
        return fmt.Errorf("create task: %w", err)
    }

    log.Printf("Task created: %s (status: %s)", task.ID, task.Status)
    return nil
}
```

### 10.3 Searching the Catalog

```go
func searchCatalog(ctx context.Context, client *iceye.Client) error {
    for resp, err := range client.SearchCatalogItems(ctx, &iceye.SearchRequest{
        BBox:     &iceye.BoundingBox{24.0, 59.5, 25.5, 60.5}, // Helsinki area
        Datetime: "2024-01-01T00:00:00Z/2024-06-01T00:00:00Z",
        Query: map[string]iceye.QueryFilter{
            "product_type": {In: []any{"GRD", "SLC"}},
        },
        Limit: 20,
    }) {
        if err != nil {
            return err
        }
        for _, item := range resp.Data {
            log.Printf("Found: %s (type: %s)", item.ID, item.Properties.ProductType)
        }
    }
    return nil
}
```

### 10.4 Creating a Delivery

```go
func createDelivery(ctx context.Context, client *iceye.Client, contractID string, itemIDs []string) error {
    // Get available delivery locations
    configs, err := client.ListDeliveryLocationConfigs(ctx)
    if err != nil {
        return err
    }

    if len(configs) == 0 {
        return fmt.Errorf("no delivery locations configured")
    }

    // Create delivery to first available location
    delivery, err := client.CreateDelivery(ctx, &iceye.CreateDeliveryRequest{
        ContractID: contractID,
        ItemIDs:    itemIDs,
        DeliveryLocations: []iceye.DeliveryLocation{{
            ConfigID: configs[0].ID,
            Method:   "s3",
            Path:     "/deliveries",
        }},
    })
    if err != nil {
        return err
    }

    log.Printf("Delivery created: %s (status: %s)", delivery.ID, delivery.Status)
    return nil
}
```

---

## 11. HTTP Status Codes

All APIs use standard HTTP status codes:

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created (POST /tasks) |
| 202 | Accepted (POST /purchases - async) |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 429 | Rate Limited |
| 500 | Server Error |

---

## 12. Implementation Considerations

### 12.1 Token Management

The client automatically handles token refresh with a 30-second buffer before expiration.

### 12.2 Pagination

All list endpoints use cursor-based pagination with `iter.Seq2` for Go 1.23+ range-over-func support.

### 12.3 Thread Safety

The client is thread-safe and can be used concurrently from multiple goroutines.

### 12.4 Zero Dependencies

The client uses only the Go standard library (no external dependencies).

---

## 13. References

- ICEYE API Platform Documentation: https://docs.iceye.com/constellation/api/
- Company API v1: https://docs.iceye.com/constellation/api/specification/company/v1/
- Tasking API v1: https://docs.iceye.com/constellation/api/specification/tasking/v1/
- Catalog API v1: https://docs.iceye.com/constellation/api/specification/catalog/v1/
- Delivery API v1: https://docs.iceye.com/constellation/api/specification/delivery/v1/
- STAC Specification: https://stacspec.org/
- GeoJSON RFC 7946: https://datatracker.ietf.org/doc/html/rfc7946
