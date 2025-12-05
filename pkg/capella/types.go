package capella

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

// ----------------------------------------------------------------------------
// GeoJSON Types - using orb library
// ----------------------------------------------------------------------------

// Geometry is an alias for geojson.Geometry for backwards compatibility.
type Geometry = geojson.Geometry

// Point creates a GeoJSON Point geometry.
func Point(lon, lat float64) *geojson.Geometry {
	return geojson.NewGeometry(orb.Point{lon, lat})
}

// Polygon creates a GeoJSON Polygon geometry from a linear ring.
func Polygon(coordinates [][][]float64) *geojson.Geometry {
	return common.NewGeoJSONPolygon(coordinates)
}

// BoundingBox is an alias for common.BoundingBox.
type BoundingBox = common.BoundingBox

// BBoxToPolygon converts a bounding box to a GeoJSON Polygon geometry.
func BBoxToPolygon(bbox BoundingBox) *geojson.Geometry {
	return common.NewGeoJSONFromBBox(bbox)
}

// Feature represents a GeoJSON Feature with typed properties.
type Feature[T any] = common.Feature[T]

// ----------------------------------------------------------------------------
// Collection Tier
// ----------------------------------------------------------------------------

// CollectionTier represents the priority tier for satellite tasking.
type CollectionTier string

const (
	TierUrgent   CollectionTier = "urgent"
	TierPriority CollectionTier = "priority"
	TierStandard CollectionTier = "standard"
	TierFlexible CollectionTier = "flexible"
	TierRoutine  CollectionTier = "routine"
)

// ----------------------------------------------------------------------------
// Collection Type
// ----------------------------------------------------------------------------

// CollectionType represents the imaging mode for satellite collection.
type CollectionType string

const (
	CollectionSpotlight      CollectionType = "spotlight"
	CollectionSpotlightUltra CollectionType = "spotlight_ultra"
	CollectionSpotlightWide  CollectionType = "spotlight_wide"
	CollectionStripmap100    CollectionType = "stripmap_100"
	CollectionStripmap50     CollectionType = "stripmap_50"
	CollectionStripmap20     CollectionType = "stripmap_20"
)

// ----------------------------------------------------------------------------
// Instrument Mode
// ----------------------------------------------------------------------------

// InstrumentMode represents the SAR instrument mode.
type InstrumentMode string

const (
	ModeSpotlight        InstrumentMode = "spotlight"
	ModeStripmap         InstrumentMode = "stripmap"
	ModeSlidingSpotlight InstrumentMode = "sliding_spotlight"
)

// ----------------------------------------------------------------------------
// Product Type
// ----------------------------------------------------------------------------

// ProductType represents the SAR product type.
type ProductType string

const (
	ProductSLC  ProductType = "SLC"
	ProductGEO  ProductType = "GEO"
	ProductGEC  ProductType = "GEC"
	ProductSICD ProductType = "SICD"
	ProductSIDD ProductType = "SIDD"
	ProductCPHD ProductType = "CPHD"
)

// ----------------------------------------------------------------------------
// Polarization
// ----------------------------------------------------------------------------

// Polarization represents the SAR polarization.
type Polarization string

const (
	PolarizationHH Polarization = "HH"
	PolarizationVV Polarization = "VV"
)

// ----------------------------------------------------------------------------
// Orbit State
// ----------------------------------------------------------------------------

// OrbitState represents the satellite orbit state.
type OrbitState string

const (
	OrbitAscending  OrbitState = "ascending"
	OrbitDescending OrbitState = "descending"
	OrbitEither     OrbitState = "either"
)

// ----------------------------------------------------------------------------
// Look Direction
// ----------------------------------------------------------------------------

// LookDirection represents the SAR look direction.
type LookDirection string

const (
	LookLeft   LookDirection = "left"
	LookRight  LookDirection = "right"
	LookEither LookDirection = "either"
)

// ----------------------------------------------------------------------------
// Orbital Plane
// ----------------------------------------------------------------------------

// OrbitalPlane represents the satellite orbital plane.
type OrbitalPlane string

const (
	OrbitalPlane45 OrbitalPlane = "45"
	OrbitalPlane53 OrbitalPlane = "53"
	OrbitalPlane97 OrbitalPlane = "97"
)

// ----------------------------------------------------------------------------
// Archive Holdback
// ----------------------------------------------------------------------------

// ArchiveHoldback represents the archive holdback period.
type ArchiveHoldback string

const (
	ArchiveNone      ArchiveHoldback = "none"
	Archive30Day     ArchiveHoldback = "30_day"
	Archive1Year     ArchiveHoldback = "1_year"
	ArchivePermanent ArchiveHoldback = "permanent"
)

// ----------------------------------------------------------------------------
// Local Time
// ----------------------------------------------------------------------------

// LocalTime represents local time preference for collection.
type LocalTime string

const (
	LocalTimeDay     LocalTime = "day"
	LocalTimeNight   LocalTime = "night"
	LocalTimeAnytime LocalTime = "anytime"
)

// ----------------------------------------------------------------------------
// Processing Status
// ----------------------------------------------------------------------------

// ProcessingStatus represents the processing status of a request.
type ProcessingStatus string

const (
	ProcessingQueued     ProcessingStatus = "queued"
	ProcessingProcessing ProcessingStatus = "processing"
	ProcessingCompleted  ProcessingStatus = "completed"
	ProcessingError      ProcessingStatus = "error"
)

// ----------------------------------------------------------------------------
// Accessibility Status
// ----------------------------------------------------------------------------

// AccessibilityStatus represents the accessibility status of a target.
type AccessibilityStatus string

const (
	AccessibilityUnknown      AccessibilityStatus = "unknown"
	AccessibilityAccessible   AccessibilityStatus = "accessible"
	AccessibilityInaccessible AccessibilityStatus = "inaccessible"
	AccessibilityRejected     AccessibilityStatus = "rejected"
)

// ----------------------------------------------------------------------------
// Task Status
// ----------------------------------------------------------------------------

// TaskStatus represents the status of a tasking request.
type TaskStatus string

const (
	TaskReceived  TaskStatus = "received"
	TaskReview    TaskStatus = "review"
	TaskSubmitted TaskStatus = "submitted"
	TaskActive    TaskStatus = "active"
	TaskAccepted  TaskStatus = "accepted"
	TaskRejected  TaskStatus = "rejected"
	TaskExpired   TaskStatus = "expired"
	TaskCompleted TaskStatus = "completed"
	TaskCanceled  TaskStatus = "canceled"
	TaskError     TaskStatus = "error"
	TaskFailed    TaskStatus = "failed"
	TaskApproved  TaskStatus = "approved"
)

// ----------------------------------------------------------------------------
// Order Status
// ----------------------------------------------------------------------------

// OrderStatus represents the status of an order.
type OrderStatus string

const (
	OrderPending    OrderStatus = "pending"
	OrderProcessing OrderStatus = "processing"
	OrderCompleted  OrderStatus = "completed"
	OrderFailed     OrderStatus = "failed"
	OrderCanceled   OrderStatus = "canceled"
)

// ----------------------------------------------------------------------------
// Common Types - using common package
// ----------------------------------------------------------------------------

// TimeWindow is an alias for common.TimeWindow.
type TimeWindow = common.TimeWindow

// Link is an alias for common.Link.
type Link = common.Link

// Asset is an alias for common.Asset.
type Asset = common.Asset
