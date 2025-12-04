# Umbra Canopy Go Client Design Document

## 1. Overview

### 1.1 Purpose

This document describes the design of a Go client library for the Umbra Canopy API. Canopy provides self-service access to Umbra's constellation of synthetic aperture radar (SAR) satellites, enabling customers to task satellites for new data collection, check feasibility of imaging requests, track task lifecycle, manage delivery configurations, and access collected imagery via STAC-compliant endpoints.

### 1.2 Scope

The Go client will provide idiomatic Go bindings for all Canopy API functionality, including tasking and feasibility operations, collect management, STAC catalog access, archive search, and delivery configuration management.

### 1.3 API Base URLs

| Environment | Base URL |
|-------------|----------|
| Production | `https://api.canopy.umbra.space` |
| Sandbox | `https://api.canopy.prod.umbra-sandbox.space` |

---

## 2. Authentication

### 2.1 Token-Based Authentication

The Canopy API uses Bearer token authentication. Access tokens are obtained from the Canopy account page and expire after 12-24 hours.

```go
type AuthConfig struct {
    // AccessToken is the Bearer token for API authentication.
    AccessToken string
}
```

### 2.2 Client Initialization

```go
// Client represents a Canopy API client.
type Client struct {
    baseURL    string
    httpClient *http.Client
    auth       AuthConfig
}

// NewClient creates a new Canopy API client.
func NewClient(accessToken string, opts ...Option) *Client

// NewSandboxClient creates a new Canopy API client configured for the sandbox environment.
func NewSandboxClient(accessToken string, opts ...Option) *Client
```

### 2.3 Configuration Options

```go
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option

// WithBaseURL overrides the default base URL.
func WithBaseURL(url string) Option

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) Option

// WithRetry configures automatic retry with exponential backoff.
func WithRetry(maxRetries int, baseDelay time.Duration) Option
```

---

## 3. Core Data Types

### 3.1 Task Types

```go
// TaskStatus represents the lifecycle status of a task.
type TaskStatus string

const (
    TaskStatusReceived    TaskStatus = "RECEIVED"
    TaskStatusSubmitted   TaskStatus = "SUBMITTED"
    TaskStatusAccepted    TaskStatus = "ACCEPTED"
    TaskStatusActive      TaskStatus = "ACTIVE"
    TaskStatusScheduled   TaskStatus = "SCHEDULED"
    TaskStatusTasked      TaskStatus = "TASKED"
    TaskStatusTransmitted TaskStatus = "TRANSMITTED"
    TaskStatusIncomplete  TaskStatus = "INCOMPLETE"
    TaskStatusProcessing  TaskStatus = "PROCESSING"
    TaskStatusProcessed   TaskStatus = "PROCESSED"
    TaskStatusDelivering  TaskStatus = "DELIVERING"
    TaskStatusDelivered   TaskStatus = "DELIVERED"   // Terminal
    TaskStatusRejected    TaskStatus = "REJECTED"    // Terminal
    TaskStatusCanceled    TaskStatus = "CANCELED"    // Terminal
    TaskStatusError       TaskStatus = "ERROR"       // Terminal
    TaskStatusAnomaly     TaskStatus = "ANOMALY"     // Terminal
)

// IsTerminal returns true if the status is a terminal state.
func (s TaskStatus) IsTerminal() bool

// ImagingMode represents the satellite imaging mode.
type ImagingMode string

const (
    ImagingModeSpotlight ImagingMode = "SPOTLIGHT"
    ImagingModeScan      ImagingMode = "SCAN"
)

// Polarization represents the radar polarization type.
type Polarization string

const (
    PolarizationVV Polarization = "VV"
    PolarizationHH Polarization = "HH"
)

// Task represents a tasking request to the Umbra constellation.
type Task struct {
    ID                   string                `json:"id"`
    TaskName             string                `json:"taskName,omitempty"`
    UserOrderID          string                `json:"userOrderId,omitempty"`
    Status               TaskStatus            `json:"status"`
    ImagingMode          ImagingMode           `json:"imagingMode"`
    SpotlightConstraints *SpotlightConstraints `json:"spotlightConstraints,omitempty"`
    ScanConstraints      *ScanConstraints      `json:"scanConstraints,omitempty"`
    WindowStartAt        time.Time             `json:"windowStartAt"`
    WindowEndAt          time.Time             `json:"windowEndAt"`
    DeliveryConfigID     string                `json:"deliveryConfigId,omitempty"`
    ProductTypes         []ProductType         `json:"productTypes,omitempty"`
    CollectIDs           []string              `json:"collectIds,omitempty"`
    CreatedAt            time.Time             `json:"createdAt"`
    UpdatedAt            time.Time             `json:"updatedAt"`
    StatusHistory        []StatusChange        `json:"statusHistory,omitempty"`
}

// StatusChange represents a historical status transition.
type StatusChange struct {
    Status    TaskStatus `json:"status"`
    Timestamp time.Time  `json:"timestamp"`
}

// ProductType represents the type of data product.
type ProductType string

const (
    ProductTypeGEC      ProductType = "GEC"      // Geocoded Ellipsoid Corrected
    ProductTypeSICD     ProductType = "SICD"     // Sensor Independent Complex Data
    ProductTypeSIDD     ProductType = "SIDD"     // Sensor Independent Detected Data
    ProductTypeCPHD     ProductType = "CPHD"     // Compensated Phase History Data
    ProductTypeMetadata ProductType = "METADATA" // JSON metadata
)
```

### 3.2 Spotlight Constraints

```go
// SpotlightConstraints defines imaging parameters for spotlight mode.
type SpotlightConstraints struct {
    Geometry                    GeoJSONGeometry `json:"geometry"`
    Polarization                Polarization    `json:"polarization"`
    RangeResolutionMinMeters    float64         `json:"rangeResolutionMinMeters"`
    MultilookFactor             int             `json:"multilookFactor"`
    GrazingAngleMinDegrees      float64         `json:"grazingAngleMinDegrees"`
    GrazingAngleMaxDegrees      float64         `json:"grazingAngleMaxDegrees"`
    TargetAzimuthAngleStartDegrees float64      `json:"targetAzimuthAngleStartDegrees"`
    TargetAzimuthAngleEndDegrees   float64      `json:"targetAzimuthAngleEndDegrees"`
    SceneSizeOption             string          `json:"sceneSizeOption,omitempty"`
}

// GeoJSONGeometry represents a GeoJSON geometry object.
type GeoJSONGeometry struct {
    Type        string          `json:"type"`
    Coordinates json.RawMessage `json:"coordinates"`
}

// NewPointGeometry creates a GeoJSON Point from longitude and latitude.
func NewPointGeometry(lon, lat float64) GeoJSONGeometry

// NewPolygonGeometry creates a GeoJSON Polygon from coordinates.
func NewPolygonGeometry(coords [][][2]float64) GeoJSONGeometry
```

### 3.3 Scan Constraints

```go
// ScanConstraints defines imaging parameters for scan mode.
type ScanConstraints struct {
    StartPoint               GeoJSONGeometry `json:"startPoint"`
    EndPoint                 GeoJSONGeometry `json:"endPoint"`
    Polarization             Polarization    `json:"polarization"`
    RangeResolutionMinMeters float64         `json:"rangeResolutionMinMeters"`
    GrazingAngleMinDegrees   float64         `json:"grazingAngleMinDegrees"`
    GrazingAngleMaxDegrees   float64         `json:"grazingAngleMaxDegrees"`
}
```

### 3.4 Feasibility Types

```go
// FeasibilityStatus represents the status of a feasibility request.
type FeasibilityStatus string

const (
    FeasibilityStatusReceived  FeasibilityStatus = "RECEIVED"
    FeasibilityStatusCompleted FeasibilityStatus = "COMPLETED"
    FeasibilityStatusError     FeasibilityStatus = "ERROR"
)

// Feasibility represents a feasibility request and its results.
type Feasibility struct {
    ID                   string                `json:"id"`
    Status               FeasibilityStatus     `json:"status"`
    ImagingMode          ImagingMode           `json:"imagingMode"`
    SpotlightConstraints *SpotlightConstraints `json:"spotlightConstraints,omitempty"`
    ScanConstraints      *ScanConstraints      `json:"scanConstraints,omitempty"`
    WindowStartAt        time.Time             `json:"windowStartAt"`
    WindowEndAt          time.Time             `json:"windowEndAt"`
    Opportunities        []Opportunity         `json:"opportunities,omitempty"`
    CreatedAt            time.Time             `json:"createdAt"`
    UpdatedAt            time.Time             `json:"updatedAt"`
}

// Opportunity represents a feasible imaging window.
type Opportunity struct {
    WindowStartAt               time.Time `json:"windowStartAt"`
    WindowEndAt                 time.Time `json:"windowEndAt"`
    DurationSec                 float64   `json:"durationSec"`
    GrazingAngleStartDegrees    float64   `json:"grazingAngleStartDegrees"`
    GrazingAngleEndDegrees      float64   `json:"grazingAngleEndDegrees"`
    GroundRangeStartKm          float64   `json:"groundRangeStartKm"`
    GroundRangeEndKm            float64   `json:"groundRangeEndKm"`
    SlantRangeStartKm           float64   `json:"slantRangeStartKm"`
    SlantRangeEndKm             float64   `json:"slantRangeEndKm"`
    SquintAngleStartDegrees     float64   `json:"squintAngleStartDegrees"`
    SquintAngleEndDegrees       float64   `json:"squintAngleEndDegrees"`
    TargetAzimuthAngleStartDegrees float64 `json:"targetAzimuthAngleStartDegrees"`
    TargetAzimuthAngleEndDegrees   float64 `json:"targetAzimuthAngleEndDegrees"`
    SatelliteID                 string    `json:"satelliteId,omitempty"`
}
```

### 3.5 Collect Types

```go
// CollectStatus represents the lifecycle status of a collect.
type CollectStatus string

const (
    CollectStatusScheduled   CollectStatus = "SCHEDULED"
    CollectStatusTasked      CollectStatus = "TASKED"
    CollectStatusTransmitted CollectStatus = "TRANSMITTED"
    CollectStatusProcessing  CollectStatus = "PROCESSING"
    CollectStatusProcessed   CollectStatus = "PROCESSED"
    CollectStatusDelivering  CollectStatus = "DELIVERING"
    CollectStatusDelivered   CollectStatus = "DELIVERED"
    CollectStatusSuperseded  CollectStatus = "SUPERSEDED"
    CollectStatusFailed      CollectStatus = "FAILED"
)

// Collect represents a satellite data collection.
type Collect struct {
    ID           string        `json:"id"`
    TaskID       string        `json:"taskId"`
    Status       CollectStatus `json:"status"`
    SatelliteID  string        `json:"satelliteId"`
    CollectStart time.Time     `json:"collectStart"`
    CollectEnd   time.Time     `json:"collectEnd"`
    CreatedAt    time.Time     `json:"createdAt"`
    UpdatedAt    time.Time     `json:"updatedAt"`
}
```

### 3.6 Delivery Configuration Types

```go
// DeliveryType represents the type of delivery destination.
type DeliveryType string

const (
    DeliveryTypeS3UmbraRole DeliveryType = "S3_UMBRA_ROLE"
    DeliveryTypeGCPWIF      DeliveryType = "GCP_WIF"
    DeliveryTypeAzureBlob   DeliveryType = "AZURE_BLOB"
)

// DeliveryConfigStatus represents the status of a delivery configuration.
type DeliveryConfigStatus string

const (
    DeliveryConfigStatusPending  DeliveryConfigStatus = "PENDING"
    DeliveryConfigStatusActive   DeliveryConfigStatus = "ACTIVE"
    DeliveryConfigStatusInactive DeliveryConfigStatus = "INACTIVE"
)

// DeliveryConfig represents a delivery destination configuration.
type DeliveryConfig struct {
    ID                   string               `json:"id"`
    Name                 string               `json:"name"`
    Type                 DeliveryType         `json:"type"`
    Status               DeliveryConfigStatus `json:"status"`
    Options              DeliveryOptions      `json:"options"`
    FileNamingConvention string               `json:"fileNamingConvention,omitempty"`
    CreatedAt            time.Time            `json:"createdAt"`
    UpdatedAt            time.Time            `json:"updatedAt"`
}

// DeliveryOptions contains destination-specific configuration.
type DeliveryOptions struct {
    // S3 options
    Bucket      string `json:"bucket,omitempty"`
    Path        string `json:"path,omitempty"`
    Region      string `json:"region,omitempty"`
    IsGovcloud  bool   `json:"isGovcloud,omitempty"`
    
    // GCP options
    ProjectID   string `json:"projectId,omitempty"`
    BucketName  string `json:"bucketName,omitempty"`
    
    // Azure options
    ContainerName string `json:"containerName,omitempty"`
    StorageAccount string `json:"storageAccount,omitempty"`
}
```

---

## 4. API Services

### 4.1 Tasks Service

```go
// TasksService handles task-related API operations.
type TasksService struct {
    client *Client
}

// CreateTaskRequest contains parameters for creating a new task.
type CreateTaskRequest struct {
    TaskName             string                `json:"taskName,omitempty"`
    UserOrderID          string                `json:"userOrderId,omitempty"`
    ImagingMode          ImagingMode           `json:"imagingMode"`
    SpotlightConstraints *SpotlightConstraints `json:"spotlightConstraints,omitempty"`
    ScanConstraints      *ScanConstraints      `json:"scanConstraints,omitempty"`
    WindowStartAt        time.Time             `json:"windowStartAt"`
    WindowEndAt          time.Time             `json:"windowEndAt"`
    DeliveryConfigID     string                `json:"deliveryConfigId,omitempty"`
    ProductTypes         []ProductType         `json:"productTypes,omitempty"`
}

// Create creates a new task.
// POST /tasking/tasks
func (s *TasksService) Create(ctx context.Context, req *CreateTaskRequest) (*Task, error)

// Get retrieves a task by ID.
// GET /tasking/tasks/{id}
func (s *TasksService) Get(ctx context.Context, id string) (*Task, error)

// List retrieves all tasks with optional filtering.
// GET /tasking/tasks
func (s *TasksService) List(ctx context.Context, opts *ListTasksOptions) (*TaskListResponse, error)

// Cancel cancels an active task.
// POST /tasking/tasks/{id}/cancel
func (s *TasksService) Cancel(ctx context.Context, id string) (*Task, error)

// ListTasksOptions contains optional filters for listing tasks.
type ListTasksOptions struct {
    Status       []TaskStatus `url:"status,omitempty"`
    Limit        int          `url:"limit,omitempty"`
    Offset       int          `url:"offset,omitempty"`
    CreatedAfter *time.Time   `url:"createdAfter,omitempty"`
    CreatedBefore *time.Time  `url:"createdBefore,omitempty"`
}

// TaskListResponse contains a paginated list of tasks.
type TaskListResponse struct {
    Tasks      []Task `json:"tasks"`
    TotalCount int    `json:"totalCount"`
    Limit      int    `json:"limit"`
    Offset     int    `json:"offset"`
}
```

### 4.2 Feasibilities Service

```go
// FeasibilitiesService handles feasibility-related API operations.
type FeasibilitiesService struct {
    client *Client
}

// CreateFeasibilityRequest contains parameters for a feasibility check.
type CreateFeasibilityRequest struct {
    ImagingMode          ImagingMode           `json:"imagingMode"`
    SpotlightConstraints *SpotlightConstraints `json:"spotlightConstraints,omitempty"`
    ScanConstraints      *ScanConstraints      `json:"scanConstraints,omitempty"`
    WindowStartAt        time.Time             `json:"windowStartAt"`
    WindowEndAt          time.Time             `json:"windowEndAt"`
}

// Create submits a new feasibility request.
// POST /tasking/feasibilities
func (s *FeasibilitiesService) Create(ctx context.Context, req *CreateFeasibilityRequest) (*Feasibility, error)

// Get retrieves a feasibility request by ID.
// GET /tasking/feasibilities/{id}
func (s *FeasibilitiesService) Get(ctx context.Context, id string) (*Feasibility, error)

// List retrieves all feasibility requests.
// GET /tasking/feasibilities
func (s *FeasibilitiesService) List(ctx context.Context, opts *ListFeasibilitiesOptions) (*FeasibilityListResponse, error)

// WaitForCompletion polls until the feasibility request is complete or times out.
func (s *FeasibilitiesService) WaitForCompletion(ctx context.Context, id string, opts *WaitOptions) (*Feasibility, error)

// WaitOptions configures polling behavior.
type WaitOptions struct {
    PollInterval time.Duration
    Timeout      time.Duration
}
```

### 4.3 Collects Service

```go
// CollectsService handles collect-related API operations.
type CollectsService struct {
    client *Client
}

// Get retrieves a collect by ID.
// GET /tasking/collects/{id}
func (s *CollectsService) Get(ctx context.Context, id string) (*Collect, error)

// List retrieves all collects with optional filtering.
// GET /tasking/collects
func (s *CollectsService) List(ctx context.Context, opts *ListCollectsOptions) (*CollectListResponse, error)

// ListCollectsOptions contains optional filters for listing collects.
type ListCollectsOptions struct {
    TaskID        string          `url:"taskId,omitempty"`
    Status        []CollectStatus `url:"status,omitempty"`
    Limit         int             `url:"limit,omitempty"`
    Offset        int             `url:"offset,omitempty"`
}
```

### 4.4 STAC Service

```go
// STACService handles STAC catalog operations.
type STACService struct {
    client *Client
}

// STACItem represents a STAC catalog item.
type STACItem struct {
    ID         string                 `json:"id"`
    Type       string                 `json:"type"`
    Geometry   GeoJSONGeometry        `json:"geometry"`
    Properties map[string]interface{} `json:"properties"`
    Assets     map[string]STACAsset   `json:"assets"`
    Links      []STACLink             `json:"links"`
    Collection string                 `json:"collection"`
}

// STACAsset represents an asset in a STAC item.
type STACAsset struct {
    Href        string            `json:"href"`
    Type        string            `json:"type"`
    Description string            `json:"description,omitempty"`
    Created     time.Time         `json:"created,omitempty"`
    Alternate   *AlternateAssets  `json:"alternate,omitempty"`
}

// AlternateAssets contains alternate access methods for an asset.
type AlternateAssets struct {
    S3Signed *SignedURL `json:"s3_signed,omitempty"`
}

// SignedURL contains a pre-signed URL for direct download.
type SignedURL struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    Href        string `json:"href"`
}

// STACLink represents a link in a STAC item.
type STACLink struct {
    Rel  string `json:"rel"`
    Type string `json:"type,omitempty"`
    Href string `json:"href"`
}

// GetItem retrieves a STAC item by collection and item ID.
// GET /stac/collections/{collectionId}/items/{itemId}
func (s *STACService) GetItem(ctx context.Context, collectionID, itemID string) (*STACItem, error)

// GetItemV2 retrieves a STAC item using the v2 API.
// GET /v2/stac/collections/{collectionId}/items/{itemId}
func (s *STACService) GetItemV2(ctx context.Context, collectionID, itemID string) (*STACItem, error)

// Search searches STAC items with CQL2 filters.
// POST /stac/search
func (s *STACService) Search(ctx context.Context, req *STACSearchRequest) (*STACSearchResponse, error)

// SearchV2 searches STAC items using the v2 API.
// POST /v2/stac/search
func (s *STACService) SearchV2(ctx context.Context, req *STACSearchRequest) (*STACSearchResponse, error)

// STACSearchRequest contains search parameters.
type STACSearchRequest struct {
    Limit      int              `json:"limit,omitempty"`
    BBox       []float64        `json:"bbox,omitempty"`
    Datetime   string           `json:"datetime,omitempty"`
    Intersects *GeoJSONGeometry `json:"intersects,omitempty"`
    FilterLang string           `json:"filter-lang,omitempty"`
    Filter     interface{}      `json:"filter,omitempty"`
}

// STACSearchResponse contains search results.
type STACSearchResponse struct {
    Type     string                 `json:"type"`
    Features []STACItem             `json:"features"`
    Context  map[string]interface{} `json:"context,omitempty"`
    Links    []STACLink             `json:"links,omitempty"`
}
```

### 4.5 Archive Service

```go
// ArchiveService handles archive catalog operations.
type ArchiveService struct {
    client *Client
}

// Search searches the archive catalog.
// POST /archive/search
func (s *ArchiveService) Search(ctx context.Context, req *ArchiveSearchRequest) (*STACSearchResponse, error)

// GetThumbnail retrieves a thumbnail image for an archive item.
// GET /archive/thumbnail/{archiveId}
func (s *ArchiveService) GetThumbnail(ctx context.Context, archiveID string) ([]byte, error)

// GetCollectionItems lists items in a collection.
// GET /archive/collections/{collectionId}/items
func (s *ArchiveService) GetCollectionItems(ctx context.Context, collectionID string, opts *ListOptions) (*STACSearchResponse, error)

// ArchiveSearchRequest contains archive search parameters.
type ArchiveSearchRequest struct {
    Limit        int              `json:"limit,omitempty"`
    BBox         []float64        `json:"bbox,omitempty"`
    Datetime     string           `json:"datetime,omitempty"`
    Intersects   *GeoJSONGeometry `json:"intersects,omitempty"`
    Collections  []string         `json:"collections,omitempty"`
    FilterLang   string           `json:"filter-lang,omitempty"`
    Filter       interface{}      `json:"filter,omitempty"`
}
```

### 4.6 Delivery Configs Service

```go
// DeliveryConfigsService handles delivery configuration operations.
type DeliveryConfigsService struct {
    client *Client
}

// Create creates a new delivery configuration.
// POST /delivery/configs
func (s *DeliveryConfigsService) Create(ctx context.Context, req *CreateDeliveryConfigRequest) (*DeliveryConfig, error)

// Get retrieves a delivery configuration by ID.
// GET /delivery/configs/{id}
func (s *DeliveryConfigsService) Get(ctx context.Context, id string) (*DeliveryConfig, error)

// List retrieves all delivery configurations.
// GET /delivery/configs
func (s *DeliveryConfigsService) List(ctx context.Context) ([]DeliveryConfig, error)

// Validate validates a delivery configuration.
// POST /delivery/configs/{id}/validate
func (s *DeliveryConfigsService) Validate(ctx context.Context, id string) (*DeliveryConfig, error)

// Delete deletes a delivery configuration.
// DELETE /delivery/configs/{id}
func (s *DeliveryConfigsService) Delete(ctx context.Context, id string) error

// CreateDeliveryConfigRequest contains parameters for creating a delivery config.
type CreateDeliveryConfigRequest struct {
    Name                 string          `json:"name"`
    Type                 DeliveryType    `json:"type"`
    Options              DeliveryOptions `json:"options"`
    FileNamingConvention string          `json:"fileNamingConvention,omitempty"`
}
```

---

## 5. Error Handling

### 5.1 Error Types

```go
// APIError represents an error returned by the Canopy API.
type APIError struct {
    StatusCode int    `json:"-"`
    Code       string `json:"code,omitempty"`
    Message    string `json:"message"`
    Details    string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
    return fmt.Sprintf("canopy: %s (status %d)", e.Message, e.StatusCode)
}

// IsNotFound returns true if the error indicates a resource was not found.
func IsNotFound(err error) bool

// IsRateLimited returns true if the error indicates rate limiting.
func IsRateLimited(err error) bool

// IsUnauthorized returns true if the error indicates authentication failure.
func IsUnauthorized(err error) bool

// RateLimitError includes rate limit information.
type RateLimitError struct {
    APIError
    RetryAfter time.Duration
}
```

### 5.2 Rate Limiting

The client implements automatic retry with exponential backoff for rate-limited requests (HTTP 429).

```go
// RetryConfig configures retry behavior.
type RetryConfig struct {
    MaxRetries    int
    BaseDelay     time.Duration
    MaxDelay      time.Duration
    Jitter        float64  // Random jitter factor (0-1)
}

// DefaultRetryConfig returns sensible defaults.
func DefaultRetryConfig() *RetryConfig {
    return &RetryConfig{
        MaxRetries: 5,
        BaseDelay:  time.Second,
        MaxDelay:   30 * time.Second,
        Jitter:     0.2,
    }
}
```

---

## 6. Helper Functions

### 6.1 Task Creation Helpers

```go
// NewSpotlightTask creates a task request for spotlight imaging.
func NewSpotlightTask(lon, lat float64, windowStart, windowEnd time.Time, opts ...TaskOption) *CreateTaskRequest

// NewSpotlightTaskFromOpportunity creates a task request from a feasibility opportunity.
func NewSpotlightTaskFromOpportunity(feasibility *Feasibility, opportunityIndex int, taskName string) *CreateTaskRequest

// TaskOption configures optional task parameters.
type TaskOption func(*CreateTaskRequest)

// WithTaskName sets the task name.
func WithTaskName(name string) TaskOption

// WithUserOrderID sets the user order ID.
func WithUserOrderID(id string) TaskOption

// WithDeliveryConfig sets the delivery configuration.
func WithDeliveryConfig(configID string) TaskOption

// WithProductTypes sets the product types to deliver.
func WithProductTypes(types ...ProductType) TaskOption

// WithResolution sets the range resolution.
func WithResolution(meters float64) TaskOption

// WithGrazingAngle sets the grazing angle constraints.
func WithGrazingAngle(min, max float64) TaskOption

// WithAzimuthAngle sets the target azimuth angle range.
func WithAzimuthAngle(start, end float64) TaskOption
```

### 6.2 Polling Helpers

```go
// Poller provides utilities for polling task status.
type Poller struct {
    client   *Client
    interval time.Duration
}

// NewPoller creates a new poller with the specified interval.
func NewPoller(client *Client, interval time.Duration) *Poller

// WaitForTaskStatus polls until the task reaches the target status or a terminal state.
func (p *Poller) WaitForTaskStatus(ctx context.Context, taskID string, targetStatus TaskStatus) (*Task, error)

// WaitForTaskDelivery polls until the task is delivered or fails.
func (p *Poller) WaitForTaskDelivery(ctx context.Context, taskID string) (*Task, error)

// TaskStatusCallback is called for each status update during polling.
type TaskStatusCallback func(task *Task)

// WaitForTaskDeliveryWithCallback polls with status callbacks.
func (p *Poller) WaitForTaskDeliveryWithCallback(ctx context.Context, taskID string, callback TaskStatusCallback) (*Task, error)
```

### 6.3 CQL2 Filter Builder

```go
// CQL2 provides a builder for CQL2 filter expressions.
type CQL2 struct {}

// Equal creates an equality filter.
func (CQL2) Equal(property string, value interface{}) map[string]interface{}

// GreaterThan creates a greater-than filter.
func (CQL2) GreaterThan(property string, value interface{}) map[string]interface{}

// LessThan creates a less-than filter.
func (CQL2) LessThan(property string, value interface{}) map[string]interface{}

// And combines multiple filters with AND logic.
func (CQL2) And(filters ...map[string]interface{}) map[string]interface{}

// Or combines multiple filters with OR logic.
func (CQL2) Or(filters ...map[string]interface{}) map[string]interface{}

// In creates an "in" filter for matching multiple values.
func (CQL2) In(values []interface{}, property string) map[string]interface{}
```

---

## 7. Package Structure

```
github.com/yourorg/canopy-go/
├── canopy.go           # Client initialization and core types
├── auth.go             # Authentication handling
├── errors.go           # Error types and handling
├── retry.go            # Retry logic with exponential backoff
├── options.go          # Client configuration options
│
├── tasks.go            # TasksService implementation
├── feasibilities.go    # FeasibilitiesService implementation
├── collects.go         # CollectsService implementation
├── stac.go             # STACService implementation
├── archive.go          # ArchiveService implementation
├── delivery.go         # DeliveryConfigsService implementation
│
├── types/              # Shared type definitions
│   ├── task.go
│   ├── feasibility.go
│   ├── collect.go
│   ├── stac.go
│   ├── delivery.go
│   └── geometry.go
│
├── helpers/            # Helper utilities
│   ├── poller.go
│   ├── cql2.go
│   └── builders.go
│
└── examples/           # Usage examples
    ├── create_task/
    ├── feasibility_check/
    ├── download_data/
    └── archive_search/
```

---

## 8. Usage Examples

### 8.1 Basic Task Creation

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/yourorg/canopy-go"
)

func main() {
    client := canopy.NewClient("your-access-token")

    // Create a spotlight task
    task, err := client.Tasks.Create(context.Background(), &canopy.CreateTaskRequest{
        TaskName:    "My First Task",
        ImagingMode: canopy.ImagingModeSpotlight,
        SpotlightConstraints: &canopy.SpotlightConstraints{
            Geometry:                 canopy.NewPointGeometry(-122.4194, 37.7749),
            Polarization:             canopy.PolarizationVV,
            RangeResolutionMinMeters: 1.0,
            MultilookFactor:          1,
            GrazingAngleMinDegrees:   30,
            GrazingAngleMaxDegrees:   70,
            TargetAzimuthAngleStartDegrees: 0,
            TargetAzimuthAngleEndDegrees:   360,
        },
        WindowStartAt: time.Now().Add(24 * time.Hour),
        WindowEndAt:   time.Now().Add(48 * time.Hour),
        ProductTypes:  []canopy.ProductType{canopy.ProductTypeGEC, canopy.ProductTypeSICD},
    })
    if err != nil {
        log.Fatalf("Failed to create task: %v", err)
    }

    log.Printf("Created task: %s (status: %s)", task.ID, task.Status)
}
```

### 8.2 Feasibility Check and Task from Opportunity

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/yourorg/canopy-go"
)

func main() {
    client := canopy.NewClient("your-access-token")
    ctx := context.Background()

    // Create feasibility request
    feas, err := client.Feasibilities.Create(ctx, &canopy.CreateFeasibilityRequest{
        ImagingMode: canopy.ImagingModeSpotlight,
        SpotlightConstraints: &canopy.SpotlightConstraints{
            Geometry:                 canopy.NewPointGeometry(-122.4194, 37.7749),
            Polarization:             canopy.PolarizationVV,
            RangeResolutionMinMeters: 1.0,
            MultilookFactor:          1,
            GrazingAngleMinDegrees:   30,
            GrazingAngleMaxDegrees:   70,
            TargetAzimuthAngleStartDegrees: 0,
            TargetAzimuthAngleEndDegrees:   360,
        },
        WindowStartAt: time.Now().Add(24 * time.Hour),
        WindowEndAt:   time.Now().Add(7 * 24 * time.Hour),
    })
    if err != nil {
        log.Fatalf("Failed to create feasibility: %v", err)
    }

    // Wait for opportunities
    feas, err = client.Feasibilities.WaitForCompletion(ctx, feas.ID, &canopy.WaitOptions{
        PollInterval: 2 * time.Second,
        Timeout:      5 * time.Minute,
    })
    if err != nil {
        log.Fatalf("Failed waiting for feasibility: %v", err)
    }

    if len(feas.Opportunities) == 0 {
        log.Fatal("No opportunities found")
    }

    // Create task from first opportunity
    task, err := client.Tasks.Create(ctx, canopy.NewSpotlightTaskFromOpportunity(feas, 0, "Task from Opportunity"))
    if err != nil {
        log.Fatalf("Failed to create task: %v", err)
    }

    log.Printf("Created task: %s", task.ID)
}
```

### 8.3 Poll for Delivery and Download

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/yourorg/canopy-go"
)

func main() {
    client := canopy.NewClient("your-access-token")
    ctx := context.Background()

    taskID := "existing-task-id"

    // Create poller
    poller := canopy.NewPoller(client, 30*time.Second)

    // Wait for delivery with status updates
    task, err := poller.WaitForTaskDeliveryWithCallback(ctx, taskID, func(t *canopy.Task) {
        log.Printf("Task status: %s", t.Status)
    })
    if err != nil {
        log.Fatalf("Task failed: %v", err)
    }

    // Get STAC item for download links
    for _, collectID := range task.CollectIDs {
        item, err := client.STAC.GetItemV2(ctx, "umbra-sar", collectID)
        if err != nil {
            log.Printf("Failed to get STAC item: %v", err)
            continue
        }

        for name, asset := range item.Assets {
            if asset.Alternate != nil && asset.Alternate.S3Signed != nil {
                log.Printf("Download %s: %s", name, asset.Alternate.S3Signed.Href)
            }
        }
    }
}
```

### 8.4 Archive Search

```go
package main

import (
    "context"
    "log"

    "github.com/yourorg/canopy-go"
)

func main() {
    client := canopy.NewClient("your-access-token")
    ctx := context.Background()

    // Search archive with CQL2 filter
    cql := canopy.CQL2{}
    
    results, err := client.Archive.Search(ctx, &canopy.ArchiveSearchRequest{
        Limit:       10,
        Collections: []string{"umbra-sar"},
        Datetime:    "2024-01-01T00:00:00Z/2024-12-31T23:59:59Z",
        BBox:        []float64{-120.0, 34.0, -119.0, 35.0},
        FilterLang:  "cql2-json",
        Filter: cql.And(
            cql.Equal("sar:resolution_range", 1),
            cql.GreaterThan("view:sun_elevation", 50),
        ),
    })
    if err != nil {
        log.Fatalf("Search failed: %v", err)
    }

    for _, item := range results.Features {
        log.Printf("Found item: %s", item.ID)
    }
}
```

---

## 9. Testing Strategy

### 9.1 Unit Tests

Unit tests mock the HTTP layer using `httptest` to verify request formatting and response parsing.

```go
func TestTasksService_Create(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "/tasking/tasks", r.URL.Path)
        assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
        
        // Return mock response
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(Task{ID: "task-123", Status: TaskStatusReceived})
    }))
    defer server.Close()
    
    client := NewClient("test-token", WithBaseURL(server.URL))
    task, err := client.Tasks.Create(context.Background(), &CreateTaskRequest{...})
    
    assert.NoError(t, err)
    assert.Equal(t, "task-123", task.ID)
}
```

### 9.2 Integration Tests

Integration tests run against the sandbox environment with a valid access token.

```go
//go:build integration

func TestIntegration_TaskLifecycle(t *testing.T) {
    token := os.Getenv("CANOPY_SANDBOX_TOKEN")
    if token == "" {
        t.Skip("CANOPY_SANDBOX_TOKEN not set")
    }
    
    client := NewSandboxClient(token)
    ctx := context.Background()
    
    // Create task, poll for completion, verify status transitions
    // ...
}
```

---

## 10. Dependencies

| Package | Purpose |
|---------|---------|
| `net/http` | HTTP client |
| `encoding/json` | JSON serialization |
| `context` | Request context and cancellation |
| `time` | Time handling |
| `github.com/google/go-querystring` | URL query parameter encoding |
| `golang.org/x/time/rate` | Client-side rate limiting (optional) |

---

## 11. Version Compatibility

The client targets the following API versions:

| API Component | Version |
|---------------|---------|
| Tasking API | v1 |
| STAC API | v1 and v2 |
| Archive Catalog | v1 |
| Delivery API | v1 |

The client provides methods for both STAC v1 and v2 endpoints to support migration.

---

## 12. Future Considerations

1. **Scan Mode Support**: When scan mode becomes generally available, add full support for scan constraints and tasks.

2. **Webhook Support**: If Canopy adds webhook notifications, implement webhook handlers for async status updates.

3. **Streaming Downloads**: Add helpers for streaming large asset downloads directly to disk.

4. **OpenAPI Generation**: Consider generating client code from OpenAPI spec when available.

5. **Observability**: Add hooks for metrics, tracing (OpenTelemetry), and structured logging.

---

## 13. References

- [Canopy API Documentation](https://docs.canopy.umbra.space/docs/introduction)
- [Canopy API Reference](https://docs.canopy.umbra.space/reference/create_task)
- [STAC Specification](https://stacspec.org/)
- [CQL2 Filter Extension](https://github.com/opengeospatial/ogcapi-features/tree/master/cql2)
