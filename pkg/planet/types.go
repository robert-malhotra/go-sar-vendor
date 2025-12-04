package planet

import (
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/robert.malhotra/go-sar-vendor/pkg/common"
)

// Common type aliases for convenience.
type (
	ListOptions = common.ListOptions
	WaitOptions = common.WaitOptions
)

// defaultSearchLimit is the default page size for paginated requests.
const defaultSearchLimit = 100

// NewPointGeometry creates a GeoJSON Point from longitude and latitude.
func NewPointGeometry(lon, lat float64) *geojson.Geometry {
	return geojson.NewGeometry(orb.Point{lon, lat})
}

// NewPolygonGeometry creates a GeoJSON Polygon from coordinates.
// coords is a slice of rings, where each ring is a slice of [lon, lat] pairs.
func NewPolygonGeometry(coords [][][2]float64) *geojson.Geometry {
	rings := make([]orb.Ring, len(coords))
	for i, ring := range coords {
		orbRing := make(orb.Ring, len(ring))
		for j, pt := range ring {
			orbRing[j] = orb.Point{pt[0], pt[1]}
		}
		rings[i] = orbRing
	}
	return geojson.NewGeometry(orb.Polygon(rings))
}

// BBoxToPolygon creates a GeoJSON Polygon from a bounding box [minLon, minLat, maxLon, maxLat].
func BBoxToPolygon(minLon, minLat, maxLon, maxLat float64) *geojson.Geometry {
	coords := [][][2]float64{{
		{minLon, minLat},
		{maxLon, minLat},
		{maxLon, maxLat},
		{minLon, maxLat},
		{minLon, minLat}, // Close the ring
	}}
	return NewPolygonGeometry(coords)
}

// =============================================================================
// Tasking API Types
// =============================================================================

// TaskingOrderStatus represents the status of a tasking order.
type TaskingOrderStatus string

const (
	TaskingOrderStatusReceived           TaskingOrderStatus = "RECEIVED"
	TaskingOrderStatusPending            TaskingOrderStatus = "PENDING"
	TaskingOrderStatusInProgress         TaskingOrderStatus = "IN_PROGRESS"
	TaskingOrderStatusExpired            TaskingOrderStatus = "EXPIRED"
	TaskingOrderStatusFulfilled          TaskingOrderStatus = "FULFILLED"
	TaskingOrderStatusFailed             TaskingOrderStatus = "FAILED"
	TaskingOrderStatusCancelled          TaskingOrderStatus = "CANCELLED"
	TaskingOrderStatusRequested          TaskingOrderStatus = "REQUESTED"
	TaskingOrderStatusFinalizing         TaskingOrderStatus = "FINALIZING"
	TaskingOrderStatusPendingCancellation TaskingOrderStatus = "PENDING_CANCELLATION"
	TaskingOrderStatusRejected           TaskingOrderStatus = "REJECTED"
)

// IsTerminal returns true if the status is a terminal state.
func (s TaskingOrderStatus) IsTerminal() bool {
	switch s {
	case TaskingOrderStatusFulfilled, TaskingOrderStatusFailed,
		TaskingOrderStatusCancelled, TaskingOrderStatusExpired, TaskingOrderStatusRejected:
		return true
	}
	return false
}

// SchedulingType represents the tasking scheduling type.
type SchedulingType string

const (
	SchedulingTypeFlexible   SchedulingType = "FLEXIBLE"
	SchedulingTypeLockIn     SchedulingType = "LOCK_IN"
	SchedulingTypeMonitoring SchedulingType = "MONITORING"
	SchedulingTypeExpress    SchedulingType = "EXPRESS"
	SchedulingTypeAssured    SchedulingType = "ASSURED"
	SchedulingTypeArchive    SchedulingType = "ARCHIVE"
)

// TaskingOrderType represents the type of tasking order.
type TaskingOrderType string

const (
	TaskingOrderTypeImage  TaskingOrderType = "IMAGE"
	TaskingOrderTypeVideo  TaskingOrderType = "VIDEO"
	TaskingOrderTypeStereo TaskingOrderType = "STEREO"
)

// SatelliteType represents supported satellite types.
type SatelliteType string

const (
	SatelliteTypeSkySat  SatelliteType = "SKYSAT"
	SatelliteTypePelican SatelliteType = "PELICAN"
	SatelliteTypeTanager SatelliteType = "TANAGER"
)

// TaskingOrder represents a tasking order.
type TaskingOrder struct {
	ID                           string             `json:"id"`
	Name                         string             `json:"name"`
	Status                       TaskingOrderStatus `json:"status"`
	Geometry                     *geojson.Geometry  `json:"geometry,omitempty"`
	OriginalGeometry             *geojson.Geometry  `json:"original_geometry,omitempty"`
	FulfilledGeometry            *geojson.Geometry  `json:"fulfilled_geometry,omitempty"`
	ImagingWindow                *string            `json:"imaging_window,omitempty"`
	PLNumber                     string             `json:"pl_number,omitempty"`
	Product                      string             `json:"product,omitempty"`
	SchedulingType               SchedulingType     `json:"scheduling_type,omitempty"`
	OrderType                    TaskingOrderType   `json:"order_type,omitempty"`
	StartTime                    *time.Time         `json:"start_time,omitempty"`
	EndTime                      *time.Time         `json:"end_time,omitempty"`
	EarlyStart                   bool               `json:"early_start,omitempty"`
	SatElevationAngleMin         *float64           `json:"sat_elevation_angle_min,omitempty"`
	SatElevationAngleMax         *float64           `json:"sat_elevation_angle_max,omitempty"`
	CloudThreshold               *float64           `json:"cloud_threshold,omitempty"`
	SatelliteTypes               []SatelliteType    `json:"satellite_types,omitempty"`
	DataProducts                 []string           `json:"data_products,omitempty"`
	AssetTypes                   []string           `json:"asset_types,omitempty"`
	NStereoPoV                   *int               `json:"n_stereo_pov,omitempty"`
	ExclusivityDays              *int               `json:"exclusivity_days,omitempty"`
	IsCancellable                bool               `json:"is_cancellable,omitempty"`
	CancellableUntil             *time.Time         `json:"cancellable_until,omitempty"`
	RequestedSqKm                float64            `json:"requested_sqkm,omitempty"`
	FulfilledSqKm                float64            `json:"fulfilled_sqkm,omitempty"`
	CaptureCount                 int                `json:"capture_count,omitempty"`
	CaptureStatusQueuedCount     int                `json:"capture_status_queued_count,omitempty"`
	CaptureStatusProcessingCount int                `json:"capture_status_processing_count,omitempty"`
	CaptureStatusPublishedCount  int                `json:"capture_status_published_count,omitempty"`
	CaptureStatusFailedCount     int                `json:"capture_status_failed_count,omitempty"`
	EstimatedQuotaCost           float64            `json:"estimated_quota_cost,omitempty"`
	UsedQuota                    float64            `json:"used_quota,omitempty"`
	CreatedBy                    string             `json:"created_by,omitempty"`
	CreatedTime                  *time.Time         `json:"created_time,omitempty"`
	UpdatedTime                  *time.Time         `json:"updated_time,omitempty"`
	FulfilledTime                *time.Time         `json:"fulfilled_time,omitempty"`
	LastAcquiredTime             *time.Time         `json:"last_acquired_time,omitempty"`
	NextPlannedAcquisitionTime   *time.Time         `json:"next_planned_acquisition_time,omitempty"`
	RRule                        string             `json:"rrule,omitempty"`
	RequestedItemIDs             []string           `json:"requested_item_ids,omitempty"`
}

// CreateTaskingOrderRequest represents a request to create a tasking order.
type CreateTaskingOrderRequest struct {
	Name                 string            `json:"name"`
	Geometry             *geojson.Geometry `json:"geometry,omitempty"`
	ImagingWindow        *string           `json:"imaging_window,omitempty"`
	PLNumber             string            `json:"pl_number,omitempty"`
	Product              string            `json:"product,omitempty"`
	SchedulingType       SchedulingType    `json:"scheduling_type,omitempty"`
	StartTime            *time.Time        `json:"start_time,omitempty"`
	EndTime              *time.Time        `json:"end_time,omitempty"`
	SatElevationAngleMin *float64          `json:"sat_elevation_angle_min,omitempty"`
	SatElevationAngleMax *float64          `json:"sat_elevation_angle_max,omitempty"`
	CloudThreshold       *float64          `json:"cloud_threshold,omitempty"`
	SatelliteTypes       []SatelliteType   `json:"satellite_types,omitempty"`
	NStereoPoV           *int              `json:"n_stereo_pov,omitempty"`
	ExclusivityDays      *int              `json:"exclusivity_days,omitempty"`
	RRule                string            `json:"rrule,omitempty"`
	EarlyStart           bool              `json:"early_start,omitempty"`
}

// UpdateTaskingOrderRequest represents a request to update a tasking order.
type UpdateTaskingOrderRequest struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Name      string     `json:"name,omitempty"`
}

// ListTaskingOrdersOptions represents options for listing tasking orders.
type ListTaskingOrdersOptions struct {
	Status             []TaskingOrderStatus `url:"status__in,omitempty"`
	SchedulingType     SchedulingType       `url:"scheduling_type,omitempty"`
	PLNumber           string               `url:"pl_number,omitempty"`
	Product            string               `url:"product,omitempty"`
	NameContains       string               `url:"name__icontains,omitempty"`
	CreatedTimeGTE     *time.Time           `url:"created_time__gte,omitempty"`
	CreatedTimeLTE     *time.Time           `url:"created_time__lte,omitempty"`
	StartTimeGTE       *time.Time           `url:"start_time__gte,omitempty"`
	StartTimeLTE       *time.Time           `url:"start_time__lte,omitempty"`
	GeometryIntersects string               `url:"geometry__intersects,omitempty"`
	Ordering           string               `url:"ordering,omitempty"`
	Limit              int                  `url:"limit,omitempty"`
	Offset             int                  `url:"offset,omitempty"`
}

// TaskingOrderPricing represents pricing information for an order.
type TaskingOrderPricing struct {
	OrderID            string                 `json:"order_id,omitempty"`
	Units              string                 `json:"units"`
	EstimatedQuotaCost float64                `json:"estimated_quota_cost"`
	DeterminedBy       string                 `json:"determined_by"`
	PricingModel       map[string]interface{} `json:"pricing_model,omitempty"`
	ReplacedOrders     []ReplacedOrderPricing `json:"replaced_orders,omitempty"`
}

// ReplacedOrderPricing represents pricing for a replaced order.
type ReplacedOrderPricing struct {
	OrderID string  `json:"order_id"`
	Name    string  `json:"name"`
	Cost    float64 `json:"cost"`
}

// CaptureStatus represents the status of a capture.
type CaptureStatus string

const (
	CaptureStatusScheduled  CaptureStatus = "SCHEDULED"
	CaptureStatusRemoved    CaptureStatus = "REMOVED"
	CaptureStatusQueued     CaptureStatus = "QUEUED"
	CaptureStatusProcessing CaptureStatus = "PROCESSING"
	CaptureStatusFailed     CaptureStatus = "FAILED"
	CaptureStatusDeriving   CaptureStatus = "DERIVING"
	CaptureStatusPublished  CaptureStatus = "PUBLISHED"
)

// Capture represents a capture from a tasking order.
type Capture struct {
	ID                     string            `json:"id"`
	OrderID                string            `json:"order_id"`
	Status                 CaptureStatus     `json:"status"`
	StatusDescription      string            `json:"status_description,omitempty"`
	CapturedArea           *geojson.Geometry `json:"captured_area,omitempty"`
	AreaOfInterest         *geojson.Geometry `json:"area_of_interest,omitempty"`
	Fulfilling             bool              `json:"fulfilling"`
	GroundID               string            `json:"ground_id,omitempty"`
	ItemIDs                []string          `json:"item_ids,omitempty"`
	SceneIDs               []string          `json:"scene_ids,omitempty"`
	ItemTypes              []string          `json:"item_types,omitempty"`
	CloudCover             *float64          `json:"cloud_cover,omitempty"`
	DeliveredAssetTypes    []string          `json:"delivered_asset_types,omitempty"`
	SatelliteType          SatelliteType     `json:"satellite_type,omitempty"`
	StripID                string            `json:"strip_id,omitempty"`
	PLNumber               string            `json:"pl_number,omitempty"`
	Product                string            `json:"product,omitempty"`
	OrderName              string            `json:"order_name,omitempty"`
	Assessment             string            `json:"assessment,omitempty"`
	AssessmentTime         *time.Time        `json:"assessment_time,omitempty"`
	CreatedTime            *time.Time        `json:"created_time,omitempty"`
	UpdatedTime            *time.Time        `json:"updated_time,omitempty"`
	PlannedAcquisitionTime *time.Time        `json:"planned_acquisition_time,omitempty"`
	AcquiredTime           *time.Time        `json:"acquired_time,omitempty"`
	PublishedTime          *time.Time        `json:"published_time,omitempty"`
}

// ListCapturesOptions represents options for listing captures.
type ListCapturesOptions struct {
	OrderID     string          `url:"order_id,omitempty"`
	Status      []CaptureStatus `url:"status__in,omitempty"`
	Fulfilling  *bool           `url:"fulfilling,omitempty"`
	Ordering    string          `url:"ordering,omitempty"`
	Limit       int             `url:"limit,omitempty"`
	Offset      int             `url:"offset,omitempty"`
}

// =============================================================================
// Imaging Windows (Feasibility) API Types
// =============================================================================

// ImagingWindowSearchStatus represents the status of an imaging window search.
type ImagingWindowSearchStatus string

const (
	ImagingWindowSearchStatusCreated    ImagingWindowSearchStatus = "CREATED"
	ImagingWindowSearchStatusInProgress ImagingWindowSearchStatus = "IN_PROGRESS"
	ImagingWindowSearchStatusDone       ImagingWindowSearchStatus = "DONE"
	ImagingWindowSearchStatusFailed     ImagingWindowSearchStatus = "FAILED"
)

// IsTerminal returns true if the search status is a terminal state.
func (s ImagingWindowSearchStatus) IsTerminal() bool {
	return s == ImagingWindowSearchStatusDone || s == ImagingWindowSearchStatusFailed
}

// AssuredTaskingTier represents the tier for assured tasking.
type AssuredTaskingTier string

const (
	AssuredTaskingTierNotApplicable AssuredTaskingTier = "NOT_APPLICABLE"
	AssuredTaskingTierStandard      AssuredTaskingTier = "STANDARD"
	AssuredTaskingTierExpress       AssuredTaskingTier = "EXPRESS"
)

// SensitivityMode represents sensitivity mode for hyperspectral imagery.
type SensitivityMode string

const (
	SensitivityModeNotApplicable SensitivityMode = "NOT_APPLICABLE"
	SensitivityModeGlint         SensitivityMode = "GLINT"
	SensitivityModeStandard      SensitivityMode = "STANDARD"
	SensitivityModeMedium        SensitivityMode = "MEDIUM"
	SensitivityModeHigh          SensitivityMode = "HIGH"
	SensitivityModeMax           SensitivityMode = "MAX"
	SensitivityModePushbroom     SensitivityMode = "PUSHBROOM"
)

// ImagingWindowSearchRequest represents a request to search for imaging windows.
type ImagingWindowSearchRequest struct {
	Geometry             *geojson.Geometry `json:"geometry"`
	PLNumber             string            `json:"pl_number,omitempty"`
	Product              string            `json:"product,omitempty"`
	StartTime            *time.Time        `json:"start_time,omitempty"`
	EndTime              *time.Time        `json:"end_time,omitempty"`
	SatElevationAngleMin *float64          `json:"sat_elevation_angle_min,omitempty"`
	SatElevationAngleMax *float64          `json:"sat_elevation_angle_max,omitempty"`
	OffNadirAngleMin     *float64          `json:"off_nadir_angle_min,omitempty"`
	OffNadirAngleMax     *float64          `json:"off_nadir_angle_max,omitempty"`
	SatelliteTypes       []SatelliteType   `json:"satellite_types,omitempty"`
	SensitivityMode      SensitivityMode   `json:"sensitivity_mode,omitempty"`
}

// ImagingWindowSearch represents an imaging window search.
type ImagingWindowSearch struct {
	ID                   string                    `json:"id"`
	Status               ImagingWindowSearchStatus `json:"status"`
	Geometry             *geojson.Geometry         `json:"geometry,omitempty"`
	PLNumber             string                    `json:"pl_number,omitempty"`
	Product              string                    `json:"product,omitempty"`
	StartTime            *time.Time                `json:"start_time,omitempty"`
	EndTime              *time.Time                `json:"end_time,omitempty"`
	SatElevationAngleMin *float64                  `json:"sat_elevation_angle_min,omitempty"`
	SatElevationAngleMax *float64                  `json:"sat_elevation_angle_max,omitempty"`
	OffNadirAngleMin     *float64                  `json:"off_nadir_angle_min,omitempty"`
	OffNadirAngleMax     *float64                  `json:"off_nadir_angle_max,omitempty"`
	ImagingWindows       []ImagingWindow           `json:"imaging_windows,omitempty"`
	ErrorCode            string                    `json:"error_code,omitempty"`
	ErrorMessage         string                    `json:"error_message,omitempty"`
}

// ImagingWindow represents a single imaging opportunity.
type ImagingWindow struct {
	ID                      string                 `json:"id"`
	Geometry                *ImagingWindowGeometry `json:"geometry,omitempty"`
	Product                 string                 `json:"product,omitempty"`
	PLNumber                string                 `json:"pl_number,omitempty"`
	StartTime               time.Time              `json:"start_time"`
	EndTime                 time.Time              `json:"end_time"`
	StartOffNadir           float64                `json:"start_off_nadir,omitempty"`
	EndOffNadir             float64                `json:"end_off_nadir,omitempty"`
	GroundSampleDistance    float64                `json:"ground_sample_distance,omitempty"`
	GroundSampleDistanceMax float64                `json:"ground_sample_distance_max,omitempty"`
	SolarZenithAngleMin     float64                `json:"solar_zenith_angle_min,omitempty"`
	SolarZenithAngleMax     float64                `json:"solar_zenith_angle_max,omitempty"`
	SunElevationAngleMin    float64                `json:"sun_elevation_angle_min,omitempty"`
	SunElevationAngleMax    float64                `json:"sun_elevation_angle_max,omitempty"`
	SatAzimuthAngleStart    float64                `json:"sat_azimuth_angle_start,omitempty"`
	SatAzimuthAngleEnd      float64                `json:"sat_azimuth_angle_end,omitempty"`
	LowLight                bool                   `json:"low_light,omitempty"`
	AssuredTaskingTier      AssuredTaskingTier     `json:"assured_tasking_tier,omitempty"`
	SensitivityMode         SensitivityMode        `json:"sensitivity_mode,omitempty"`
	SatelliteType           SatelliteType          `json:"satellite_type,omitempty"`
	CloudForecast           []CloudForecast        `json:"cloud_forecast,omitempty"`
	ConflictingOrders       map[string]interface{} `json:"conflicting_orders,omitempty"`
	PricingDetails          *PricingDetails        `json:"pricing_details,omitempty"`
	CreatedTime             *time.Time             `json:"created_time,omitempty"`
}

// ImagingWindowGeometry represents geometry for an imaging window.
type ImagingWindowGeometry struct {
	GeoJSON *geojson.Geometry `json:"geojson,omitempty"`
}

// CloudForecast represents cloud cover forecast information.
type CloudForecast struct {
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Prediction   float64   `json:"prediction"`
	Historical   float64   `json:"historical,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	UpdatedTime  time.Time `json:"updated_time,omitempty"`
}

// PricingDetails represents pricing details for an imaging window.
type PricingDetails struct {
	Units              string                 `json:"units"`
	EstimatedQuotaCost float64                `json:"estimated_quota_cost"`
	DeterminedBy       string                 `json:"determined_by"`
	PricingModel       map[string]interface{} `json:"pricing_model,omitempty"`
	ReplacedOrders     []ReplacedOrder        `json:"replaced_orders,omitempty"`
}

// ReplacedOrder represents an order that would be replaced.
type ReplacedOrder struct {
	OrderID string  `json:"order_id"`
	Name    string  `json:"name"`
	Cost    float64 `json:"cost"`
}

// =============================================================================
// Orders API v2 Types
// =============================================================================

// OrderState represents the state of an order.
type OrderState string

const (
	OrderStateQueued    OrderState = "queued"
	OrderStateRunning   OrderState = "running"
	OrderStateSuccess   OrderState = "success"
	OrderStatePartial   OrderState = "partial"
	OrderStateFailed    OrderState = "failed"
	OrderStateCancelled OrderState = "cancelled"
)

// IsTerminal returns true if the order state is a terminal state.
func (s OrderState) IsTerminal() bool {
	switch s {
	case OrderStateSuccess, OrderStatePartial, OrderStateFailed, OrderStateCancelled:
		return true
	}
	return false
}

// SourceType represents the source type for an order.
type SourceType string

const (
	SourceTypeScenes   SourceType = "scenes"
	SourceTypeBasemaps SourceType = "basemaps"
)

// Order represents an order from the Orders API.
type Order struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	State         OrderState          `json:"state"`
	CreatedOn     time.Time           `json:"created_on"`
	LastModified  time.Time           `json:"last_modified,omitempty"`
	ErrorHints    []string            `json:"error_hints,omitempty"`
	Products      []ProductSpec       `json:"products,omitempty"`
	Tools         []Tool              `json:"tools,omitempty"`
	Delivery      *DeliveryConfig     `json:"delivery,omitempty"`
	Hosting       *HostingConfig      `json:"hosting,omitempty"`
	Notifications *NotificationConfig `json:"notifications,omitempty"`
	OrderType     string              `json:"order_type,omitempty"`
	SourceType    SourceType          `json:"source_type,omitempty"`
	Links         *OrderLinks         `json:"_links,omitempty"`
}

// OrderLinks represents links associated with an order.
type OrderLinks struct {
	Self    string `json:"_self,omitempty"`
	Results []Link `json:"results,omitempty"`
}

// Link represents a generic link.
type Link struct {
	Name     string `json:"name,omitempty"`
	Location string `json:"location,omitempty"`
	Delivery string `json:"delivery,omitempty"`
	Expires  string `json:"expires_at,omitempty"`
}

// ProductSpec represents a product specification in an order.
type ProductSpec struct {
	ItemIDs       []string `json:"item_ids"`
	ItemType      string   `json:"item_type"`
	ProductBundle string   `json:"product_bundle"`
}

// CreateOrderRequest represents a request to create an order.
type CreateOrderRequest struct {
	Name          string              `json:"name"`
	Products      []ProductSpec       `json:"products"`
	Tools         []Tool              `json:"tools,omitempty"`
	Delivery      *DeliveryConfig     `json:"delivery,omitempty"`
	Hosting       *HostingConfig      `json:"hosting,omitempty"`
	Notifications *NotificationConfig `json:"notifications,omitempty"`
	OrderType     string              `json:"order_type,omitempty"`
	SourceType    SourceType          `json:"source_type,omitempty"`
	Metadata      map[string]string   `json:"metadata,omitempty"`
}

// Tool represents a raster processing tool.
type Tool struct {
	Type       string                 `json:"type,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ClipTool creates a clip tool configuration.
func ClipTool(aoi *geojson.Geometry) Tool {
	return Tool{
		Type: "clip",
		Parameters: map[string]interface{}{
			"aoi": aoi,
		},
	}
}

// ReprojectTool creates a reproject tool configuration.
func ReprojectTool(projection string) Tool {
	return Tool{
		Type: "reproject",
		Parameters: map[string]interface{}{
			"projection": projection,
		},
	}
}

// CompositeTool creates a composite tool configuration.
func CompositeTool() Tool {
	return Tool{
		Type:       "composite",
		Parameters: map[string]interface{}{},
	}
}

// FileFormatTool creates a file format tool configuration.
func FileFormatTool(format string) Tool {
	return Tool{
		Type: "file_format",
		Parameters: map[string]interface{}{
			"format": format,
		},
	}
}

// DeliveryConfig represents delivery configuration.
type DeliveryConfig struct {
	AmazonS3           *AmazonS3Delivery           `json:"amazon_s3,omitempty"`
	AzureBlobStorage   *AzureBlobStorageDelivery   `json:"azure_blob_storage,omitempty"`
	GoogleCloudStorage *GoogleCloudStorageDelivery `json:"google_cloud_storage,omitempty"`
	GoogleEarthEngine  *GoogleEarthEngineDelivery  `json:"google_earth_engine,omitempty"`
	OracleCloudStorage *OracleCloudStorageDelivery `json:"oracle_cloud_storage,omitempty"`
	Destination        *DestinationRef             `json:"destination,omitempty"`
	ArchiveType        string                      `json:"archive_type,omitempty"`
	ArchiveFilename    string                      `json:"archive_filename,omitempty"`
	SingleArchive      bool                        `json:"single_archive,omitempty"`
}

// AmazonS3Delivery represents S3 delivery configuration.
type AmazonS3Delivery struct {
	Bucket             string `json:"bucket"`
	AWSRegion          string `json:"aws_region"`
	AWSAccessKeyID     string `json:"aws_access_key_id"`
	AWSSecretAccessKey string `json:"aws_secret_access_key"`
	PathPrefix         string `json:"path_prefix,omitempty"`
}

// AzureBlobStorageDelivery represents Azure delivery configuration.
type AzureBlobStorageDelivery struct {
	Account    string `json:"account"`
	Container  string `json:"container"`
	SASToken   string `json:"sas_token"`
	PathPrefix string `json:"path_prefix,omitempty"`
}

// GoogleCloudStorageDelivery represents GCS delivery configuration.
type GoogleCloudStorageDelivery struct {
	Bucket      string `json:"bucket"`
	Credentials string `json:"credentials"`
	PathPrefix  string `json:"path_prefix,omitempty"`
}

// GoogleEarthEngineDelivery represents GEE delivery configuration.
type GoogleEarthEngineDelivery struct {
	Project    string `json:"project"`
	Collection string `json:"collection"`
}

// OracleCloudStorageDelivery represents Oracle Cloud delivery configuration.
type OracleCloudStorageDelivery struct {
	Bucket              string `json:"bucket"`
	Namespace           string `json:"namespace"`
	Region              string `json:"region"`
	CustomerAccessKeyID string `json:"customer_access_key_id"`
	CustomerSecretKey   string `json:"customer_secret_key"`
	PathPrefix          string `json:"path_prefix,omitempty"`
}

// DestinationRef represents a reference to a saved destination.
type DestinationRef struct {
	Ref        string `json:"ref"`
	PathPrefix string `json:"path_prefix,omitempty"`
}

// HostingConfig represents hosting configuration.
type HostingConfig struct {
	SentinelHub *SentinelHubHosting `json:"sentinel_hub,omitempty"`
}

// SentinelHubHosting represents Sentinel Hub hosting configuration.
type SentinelHubHosting struct {
	CollectionID        string `json:"collection_id,omitempty"`
	CreateConfiguration bool   `json:"create_configuration,omitempty"`
	ConfigurationID     string `json:"configuration_id,omitempty"`
}

// NotificationConfig represents notification configuration.
type NotificationConfig struct {
	Webhook *WebhookNotification `json:"webhook,omitempty"`
	Email   bool                 `json:"email,omitempty"`
}

// WebhookNotification represents webhook notification configuration.
type WebhookNotification struct {
	URL      string `json:"url"`
	PerOrder bool   `json:"per_order,omitempty"`
}

// ListOrdersOptions represents options for listing orders.
type ListOrdersOptions struct {
	State          []OrderState `url:"state,omitempty"`
	Name           string       `url:"name,omitempty"`
	SourceType     SourceType   `url:"source_type,omitempty"`
	Hosting        *bool        `url:"hosting,omitempty"`
	DestinationRef string       `url:"destination_ref,omitempty"`
	Limit          int          `url:"limit,omitempty"`
}

// paginatedResponse is a generic paginated response structure.
type paginatedResponse[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
	Results  []T    `json:"results"`
}
