// Package airbus provides a Go client for the Airbus OneAtlas SAR-API.
//
// The SAR-API provides access to TerraSAR-X and PAZ satellite radar data through
// catalogue search, feasibility analysis (tasking), ordering, and delivery workflows.
//
// Key features:
//   - API key authentication with automatic token refresh
//   - Full coverage of Catalogue, Feasibility, Baskets, Orders, and Config APIs
//   - GeoJSON support via paulmach/orb
//   - Idiomatic Go types with comprehensive type safety
//   - Thread-safe; safe for concurrent goroutines
//
// Docs: https://api.oneatlas.airbus.com/api-catalog/sar/
package airbus

import (
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

// ----------------------------------------------------------------------------
// GeoJSON Types - using orb library
// ----------------------------------------------------------------------------

// Geometry is an alias for geojson.Geometry.
type Geometry = geojson.Geometry

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

// NewMultiPolygonGeometry creates a GeoJSON MultiPolygon from coordinates.
func NewMultiPolygonGeometry(coords [][][][2]float64) *geojson.Geometry {
	polygons := make([]orb.Polygon, len(coords))
	for i, poly := range coords {
		rings := make([]orb.Ring, len(poly))
		for j, ring := range poly {
			orbRing := make(orb.Ring, len(ring))
			for k, pt := range ring {
				orbRing[k] = orb.Point{pt[0], pt[1]}
			}
			rings[j] = orbRing
		}
		polygons[i] = rings
	}
	return geojson.NewGeometry(orb.MultiPolygon(polygons))
}

// BoundingBox is an alias for common.BoundingBox.
type BoundingBox = common.BoundingBox

// BBoxToPolygon converts a bounding box to a GeoJSON Polygon geometry.
func BBoxToPolygon(bbox BoundingBox) *geojson.Geometry {
	return common.NewGeoJSONFromBBox(bbox)
}

// ----------------------------------------------------------------------------
// Mission & Satellite
// ----------------------------------------------------------------------------

// Mission represents the satellite mission.
type Mission string

const (
	MissionTSX Mission = "TSX"
	MissionPAZ Mission = "PAZ"
	MissionAll Mission = "all"
)

// Satellite represents the satellite identifier.
type Satellite string

const (
	SatelliteTSX1 Satellite = "TSX-1"
	SatellitePAZ1 Satellite = "PAZ-1"
	SatelliteTDX1 Satellite = "TDX-1"
	SatelliteAll  Satellite = "all"
)

// ----------------------------------------------------------------------------
// Sensor Mode
// ----------------------------------------------------------------------------

// SensorMode represents the SAR sensor/imaging mode.
type SensorMode string

const (
	SensorModeStaringSpotlight SensorMode = "SAR_ST_S"     // Staring Spotlight
	SensorModeHighResSpotlight SensorMode = "SAR_HS_S"     // High Resolution Spotlight
	SensorModeHighResSpot300   SensorMode = "SAR_HS_S_300" // High Resolution Spotlight 300MHz
	SensorModeHighResSpot150   SensorMode = "SAR_HS_S_150" // High Resolution Spotlight 150MHz
	SensorModeHighResDual      SensorMode = "SAR_HS_D"     // High Resolution Spotlight Dual-Pol
	SensorModeHighResDual300   SensorMode = "SAR_HS_D_300" // High Resolution Spotlight Dual-Pol 300MHz
	SensorModeHighResDual150   SensorMode = "SAR_HS_D_150" // High Resolution Spotlight Dual-Pol 150MHz
	SensorModeSpotlight        SensorMode = "SAR_SL_S"     // Spotlight
	SensorModeSpotlightDual    SensorMode = "SAR_SL_D"     // Spotlight Dual-Pol
	SensorModeStripmap         SensorMode = "SAR_SM_S"     // Stripmap
	SensorModeStripmapDual     SensorMode = "SAR_SM_D"     // Stripmap Dual-Pol
	SensorModeScanSAR          SensorMode = "SAR_SC_S"     // ScanSAR
	SensorModeWideScanSAR      SensorMode = "SAR_WS_S"     // Wide ScanSAR
	SensorModeAll              SensorMode = "all"
)

// ----------------------------------------------------------------------------
// Polarization
// ----------------------------------------------------------------------------

// Polarization represents the radar polarization channels.
type Polarization string

const (
	PolarizationHH   Polarization = "HH"
	PolarizationVV   Polarization = "VV"
	PolarizationHV   Polarization = "HV"
	PolarizationVH   Polarization = "VH"
	PolarizationHHVV Polarization = "HHVV"
	PolarizationHHHV Polarization = "HHHV"
	PolarizationVVVH Polarization = "VVVH"
	PolarizationAll  Polarization = "all"
)

// ----------------------------------------------------------------------------
// Path & Look Direction
// ----------------------------------------------------------------------------

// PathDirection represents the satellite orbit direction.
type PathDirection string

const (
	PathDirectionAscending  PathDirection = "ascending"
	PathDirectionDescending PathDirection = "descending"
	PathDirectionBoth       PathDirection = "both"
)

// LookDirection represents the SAR look direction.
type LookDirection string

const (
	LookDirectionRight LookDirection = "R"
	LookDirectionLeft  LookDirection = "L"
	LookDirectionBoth  LookDirection = "both"
)

// ----------------------------------------------------------------------------
// Product Type & Processing Options
// ----------------------------------------------------------------------------

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
	OrbitTypePremiumNRT OrbitType = "Premium NRT" // Premium Near Real-Time
	OrbitTypeNRT        OrbitType = "NRT"         // Near Real-Time
	OrbitTypeRapid      OrbitType = "rapid"       // Rapid
	OrbitTypeScience    OrbitType = "science"     // Science (highest accuracy)
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

// ----------------------------------------------------------------------------
// Feasibility
// ----------------------------------------------------------------------------

// FeasibilityLevel represents the depth of feasibility analysis.
type FeasibilityLevel string

const (
	FeasibilityLevelSimple   FeasibilityLevel = "simple"
	FeasibilityLevelComplete FeasibilityLevel = "complete"
)

// Priority represents order priority for tasking.
type Priority string

const (
	PriorityStandard  Priority = "standard"
	PriorityPriority  Priority = "priority"
	PriorityExclusive Priority = "exclusive"
)

// Periodicity represents the repeat cycle in days for datastack acquisitions.
type Periodicity int

const (
	Periodicity11 Periodicity = 11
	Periodicity22 Periodicity = 22
	Periodicity33 Periodicity = 33
	Periodicity44 Periodicity = 44
	Periodicity55 Periodicity = 55
	Periodicity66 Periodicity = 66
	Periodicity77 Periodicity = 77
	Periodicity88 Periodicity = 88
	Periodicity99 Periodicity = 99
)

// ----------------------------------------------------------------------------
// Item & Order Status
// ----------------------------------------------------------------------------

// ItemStatus represents the status of an order item.
type ItemStatus string

const (
	ItemStatusPlanned            ItemStatus = "planned"
	ItemStatusAcquired           ItemStatus = "acquired"
	ItemStatusAcquisitionFailed  ItemStatus = "acquisitionFailed"
	ItemStatusProcessing         ItemStatus = "processing"
	ItemStatusProcessed          ItemStatus = "processed"
	ItemStatusProcessingFailed   ItemStatus = "processingFailed"
	ItemStatusDelivering         ItemStatus = "delivering"
	ItemStatusDelivered          ItemStatus = "delivered"
	ItemStatusDeliveryFailed     ItemStatus = "deliveryFailed"
	ItemStatusCancelled          ItemStatus = "cancelled"
	ItemStatusCancellationFailed ItemStatus = "cancellationFailed"
	ItemStatusExpired            ItemStatus = "expired"
)

// OrderType represents the type of order.
type OrderType string

const (
	OrderTypeCatalogue   OrderType = "catalogue"
	OrderTypeFeasibility OrderType = "feasibility"
)

// ----------------------------------------------------------------------------
// Service & Purpose
// ----------------------------------------------------------------------------

// Service represents an SAR-API service.
type Service string

const (
	ServiceRadar    Service = "radar"
	ServiceWorldDEM Service = "worlddem"
	ServiceMgmt     Service = "mgmt"
)

// Purpose represents the order purpose (required for compliance).
type Purpose string

const (
	PurposeAerospaceIndustry         Purpose = "Aerospace Industry Company"
	PurposeAgroCompany               Purpose = "Agro Company"
	PurposeAgroServiceCompany        Purpose = "Agro Service Company"
	PurposeBank                      Purpose = "Bank"
	PurposeConsultingCompany         Purpose = "Consulting Company"
	PurposeConsumer                  Purpose = "Consumer"
	PurposeCooperativeCompany        Purpose = "Cooperative Company"
	PurposeDefenceCompany            Purpose = "Defence Company"
	PurposeDEM                       Purpose = "DEM"
	PurposeEditionCommunication      Purpose = "Edition & Communication Company"
	PurposeEducationResearch         Purpose = "Education / Research"
	PurposeElectronicSystemCompany   Purpose = "Electronic System company"
	PurposeEmergencyResponse         Purpose = "Emergency Response & Crisis Management"
	PurposeEnergyCompany             Purpose = "Energy Company"
	PurposeEnergyServiceCompany      Purpose = "Energy Service Company"
	PurposeEngineeringCompany        Purpose = "Engineering Company"
	PurposeEngineeringServiceCompany Purpose = "Engineering Service Company"
	PurposeEnvironment               Purpose = "Environment"
	PurposeForestCompany             Purpose = "Forest Company"
	PurposeForestServiceCompany      Purpose = "Forest Service Company"
	PurposeGeoservicesCompany        Purpose = "Geoservices Company"
	PurposeGovernment                Purpose = "Government"
	PurposeHumanitarianOrganization  Purpose = "Humanitarian Organization"
	PurposeInfrastructuresMonitoring Purpose = "Infrastructures Monitoring"
	PurposeInsuranceCompany          Purpose = "Insurance Company"
	PurposeLocationBasedServices     Purpose = "Location Based Services"
	PurposeLogisticsCompany          Purpose = "Logistics Company"
	PurposeMaritimeServices          Purpose = "Maritime services"
	PurposeMiningCompany             Purpose = "Mining Company"
	PurposeNetworkOperator           Purpose = "Network Operator"
	PurposeNGO                       Purpose = "NGO"
	PurposeOilGasCompany             Purpose = "Oil and Gas Company"
	PurposeOther                     Purpose = "Other"
	PurposePublicAdministration      Purpose = "Public Administration"
	PurposeRealEstateCompany         Purpose = "Real estate Company"
	PurposeSecurityCompany           Purpose = "Security Company"
	PurposeSpaceAgency               Purpose = "Space Agency"
	PurposeTelecomCompany            Purpose = "Telecom Company"
	PurposeTransportCompany          Purpose = "Transport Company"
	PurposeUrbanPlanningCompany      Purpose = "Urban planning Company"
	PurposeUtilitiesCompany          Purpose = "Utilities Company"
	PurposeWaterCompany              Purpose = "Water Company"
	PurposeWaterServiceCompany       Purpose = "Water Service Company"
	PurposeInternalUse               Purpose = "Internal Use"
	PurposeValueAdding               Purpose = "Value Adding"
)

// ----------------------------------------------------------------------------
// Notification Type
// ----------------------------------------------------------------------------

// NotificationType represents the type of notification.
type NotificationType string

const (
	NotificationTypeOrder       NotificationType = "order"
	NotificationTypeMaintenance NotificationType = "maintenance"
	NotificationTypeSystem      NotificationType = "system"
)

// ----------------------------------------------------------------------------
// Common Types
// ----------------------------------------------------------------------------

// TimeRange represents a time window for searches.
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to,omitempty"`
}

// IncidenceAngleRange represents min/max incidence angles in degrees.
type IncidenceAngleRange struct {
	Minimum float64 `json:"minimum,omitempty"`
	Maximum float64 `json:"maximum,omitempty"`
}

// OrderOptions represents processing options for an order.
type OrderOptions struct {
	ProductType           ProductType       `json:"productType,omitempty"`
	ResolutionVariant     ResolutionVariant `json:"resolutionVariant,omitempty"`
	OrbitType             OrbitType         `json:"orbitType,omitempty"`
	MapProjection         MapProjection     `json:"mapProjection,omitempty"`
	GainAttenuation       GainAttenuation   `json:"gainAttenuation,omitempty"`
	GeocodedIncidenceMask bool              `json:"geocodedIncidenceMask,omitempty"`
}

// Price represents pricing information for an item.
type Price struct {
	Final    bool    `json:"final"`
	Total    float64 `json:"total"`
	Currency string  `json:"currency"`
}

// ----------------------------------------------------------------------------
// User Types
// ----------------------------------------------------------------------------

// UserInfo represents user account information.
type UserInfo struct {
	Username           string    `json:"username"`
	ContractType       string    `json:"contract_type,omitempty"`
	PWChangeNeeded     bool      `json:"pw_change_needed"`
	PWExpirationDate   string    `json:"pw_expiration_date,omitempty"`
	ExpirationDate     string    `json:"expiration_date,omitempty"`
	Services           []Service `json:"services"`
	RegistrationStatus string    `json:"registration_status,omitempty"`
	OneAtlasUsername   *string   `json:"oneatlas_username,omitempty"`
}

// ChangePasswordRequest represents a password change request.
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword,omitempty"` // Base64 encoded
	Password    string `json:"password"`              // Base64 encoded
}

// ResetPasswordRequest represents a password reset request.
type ResetPasswordRequest struct {
	Username string `json:"username"`
}

// Notification represents a user notification.
type Notification struct {
	ID        string           `json:"id"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Read      bool             `json:"read"`
	CreatedAt time.Time        `json:"createdAt"`
	ExpiresAt *time.Time       `json:"expiresAt,omitempty"`
}

// UpdateNotificationRequest represents a notification update.
type UpdateNotificationRequest struct {
	Read bool `json:"read"`
}

// ----------------------------------------------------------------------------
// Health Types
// ----------------------------------------------------------------------------

// HealthStatus represents the API health status.
type HealthStatus struct {
	Status      string                   `json:"status"` // "pass", "fail", "warn"
	Version     string                   `json:"version,omitempty"`
	ServiceID   string                   `json:"serviceId,omitempty"`
	Description string                   `json:"description,omitempty"`
	Checks      map[string][]CheckResult `json:"checks,omitempty"`
	Links       map[string]string        `json:"links,omitempty"`
}

// CheckResult represents a health check result.
type CheckResult struct {
	ComponentID   string     `json:"componentId,omitempty"`
	ComponentType string     `json:"componentType,omitempty"`
	Status        string     `json:"status"`
	Time          *time.Time `json:"time,omitempty"`
	ObservedValue any        `json:"observedValue,omitempty"`
	ObservedUnit  string     `json:"observedUnit,omitempty"`
}

// ----------------------------------------------------------------------------
// Catalogue Types
// ----------------------------------------------------------------------------

// CatalogueRequest represents a catalogue search request.
type CatalogueRequest struct {
	Limit                 int                  `json:"limit,omitempty"`
	AOI                   *geojson.Geometry    `json:"aoi,omitempty"`
	Time                  *TimeRange           `json:"time,omitempty"`
	CatalogueTime         *TimeRange           `json:"catalogueTime,omitempty"`
	Mission               any                  `json:"mission,omitempty"`    // Mission, []Mission, or "all"
	Satellite             any                  `json:"satellite,omitempty"`  // Satellite, []Satellite, or "all"
	SensorMode            any                  `json:"sensorMode,omitempty"` // SensorMode, []SensorMode, or "all"
	PolarizationChannels  any                  `json:"polarizationChannels,omitempty"`
	IncidenceAngle        *IncidenceAngleRange `json:"incidenceAngle,omitempty"`
	PathDirection         any                  `json:"pathDirection,omitempty"` // PathDirection, []PathDirection, or "both"
	LookDirection         any                  `json:"lookDirection,omitempty"` // LookDirection, []LookDirection, or "both"
	BeamID                any                  `json:"beamId,omitempty"`        // string or []string
	RelativeOrbit         any                  `json:"relativeOrbit,omitempty"` // int, []int, or "all"
	OutOfFullPerformance  bool                 `json:"outOfFullPerformance,omitempty"`
	Occurrences           int                  `json:"occurrences,omitempty"` // Minimum datastack size
	Customer              string               `json:"customer,omitempty"`
	OrderTemplate         string               `json:"orderTemplate,omitempty"`
	ProductType           ProductType          `json:"productType,omitempty"`
	ResolutionVariant     ResolutionVariant    `json:"resolutionVariant,omitempty"`
	OrbitType             OrbitType            `json:"orbitType,omitempty"`
	GeocodedIncidenceMask bool                 `json:"geocodedIncidenceMask,omitempty"`
	MapProjection         MapProjection        `json:"mapProjection,omitempty"`
	GainAttenuation       GainAttenuation      `json:"gainAttenuation,omitempty"`
	MinimumCoverage       float64              `json:"minimumCoverage,omitempty"`
}

// ReplicationOptions represents options for catalogue replication.
type ReplicationOptions struct {
	Since time.Time `url:"since,omitempty"`
	Limit int       `url:"limit,omitempty"`
}

// RevocationOptions represents options for getting catalogue revocations.
type RevocationOptions struct {
	Since time.Time `url:"since,omitempty"`
	Limit int       `url:"limit,omitempty"`
}

// RevocationResponse contains revoked acquisition IDs.
type RevocationResponse struct {
	Revocations []string `json:"revocations"`
}

// RetrieveRequest represents a request to retrieve ordered items from catalogue.
type RetrieveRequest struct {
	Items []string `json:"items"` // Order item IDs
}

// ----------------------------------------------------------------------------
// Feasibility Types (Tasking)
// ----------------------------------------------------------------------------

// FeasibilityRequest represents a feasibility/tasking search request.
type FeasibilityRequest struct {
	AOI                   *geojson.Geometry    `json:"aoi"`
	Time                  TimeRange            `json:"time"`
	FeasibilityLevel      FeasibilityLevel     `json:"feasibilityLevel"`
	SensorMode            SensorMode           `json:"sensorMode"`
	Mission               any                  `json:"mission,omitempty"` // Mission, []Mission, or "all"
	Priority              Priority             `json:"priority,omitempty"`
	Periodicity           Periodicity          `json:"periodicity,omitempty"`
	Occurrences           int                  `json:"occurrences,omitempty"` // 2-50 for datastacks
	PolarizationChannels  Polarization         `json:"polarizationChannels,omitempty"`
	IncidenceAngle        *IncidenceAngleRange `json:"incidenceAngle,omitempty"`
	PathDirection         any                  `json:"pathDirection,omitempty"`
	LookDirection         LookDirection        `json:"lookDirection,omitempty"`
	BeamID                any                  `json:"beamId,omitempty"`
	RelativeOrbit         any                  `json:"relativeOrbit,omitempty"`
	OutOfFullPerformance  bool                 `json:"outOfFullPerformance,omitempty"`
	Customer              string               `json:"customer,omitempty"`
	OrderTemplate         string               `json:"orderTemplate,omitempty"`
	ProductType           ProductType          `json:"productType,omitempty"`
	ResolutionVariant     ResolutionVariant    `json:"resolutionVariant,omitempty"`
	AcquisitionOnly       bool                 `json:"acquisitionOnly,omitempty"`
	ReceivingStation      string               `json:"receivingStation,omitempty"`
	OrbitType             OrbitType            `json:"orbitType,omitempty"`
	GeocodedIncidenceMask bool                 `json:"geocodedIncidenceMask,omitempty"`
	MapProjection         MapProjection        `json:"mapProjection,omitempty"`
	GainAttenuation       GainAttenuation      `json:"gainAttenuation,omitempty"`
	MinimumCoverage       float64              `json:"minimumCoverage,omitempty"`
}

// ----------------------------------------------------------------------------
// Prices Types
// ----------------------------------------------------------------------------

// PricesRequest represents a price query request.
type PricesRequest struct {
	Acquisitions  []string `json:"acquisitions,omitempty"` // Acquisition IDs
	Items         []string `json:"items,omitempty"`        // Item UUIDs
	Customer      string   `json:"customer,omitempty"`
	OrderTemplate string   `json:"orderTemplate,omitempty"`
}

// PriceResponse represents pricing information for an item.
type PriceResponse struct {
	ItemID        string `json:"itemId"`
	AcquisitionID string `json:"acquisitionId"`
	Price         Price  `json:"price"`
}

// ----------------------------------------------------------------------------
// Basket Types
// ----------------------------------------------------------------------------

// Basket represents a shopping basket.
type Basket struct {
	BasketID          string     `json:"basketId"`
	CreationTime      time.Time  `json:"creationTime"`
	Customer          string     `json:"customer,omitempty"`
	CustomerReference string     `json:"customerReference,omitempty"`
	NotifyEndpoint    *string    `json:"notifyEndpoint,omitempty"`
	Purpose           Purpose    `json:"purpose,omitempty"`
	OrderTemplate     string     `json:"orderTemplate,omitempty"`
	Items             []Item     `json:"items,omitempty"`
	Price             *Price     `json:"price,omitempty"`
	OrderType         OrderType  `json:"orderType,omitempty"`
	OrderID           string     `json:"orderId,omitempty"`
	SubmissionTime    *time.Time `json:"submissionTime,omitempty"`
}

// Item represents an item in a basket or order.
type Item struct {
	ItemID                string            `json:"itemId"`
	AcquisitionID         string            `json:"acquisitionId"`
	GroupID               int               `json:"groupId,omitempty"`
	Mission               Mission           `json:"mission,omitempty"`
	Satellite             Satellite         `json:"satellite,omitempty"`
	SensorMode            SensorMode        `json:"sensorMode,omitempty"`
	PolarizationChannels  Polarization      `json:"polarizationChannels,omitempty"`
	StartTime             *time.Time        `json:"startTime,omitempty"`
	StopTime              *time.Time        `json:"stopTime,omitempty"`
	BeamID                string            `json:"beamId,omitempty"`
	PathDirection         PathDirection     `json:"pathDirection,omitempty"`
	LookDirection         LookDirection     `json:"lookDirection,omitempty"`
	IncidenceAngle        float64           `json:"incidenceAngle,omitempty"`
	RelativeOrbit         int               `json:"relativeOrbit,omitempty"`
	ProductType           ProductType       `json:"productType,omitempty"`
	ResolutionVariant     ResolutionVariant `json:"resolutionVariant,omitempty"`
	OrbitType             OrbitType         `json:"orbitType,omitempty"`
	MapProjection         MapProjection     `json:"mapProjection,omitempty"`
	GainAttenuation       GainAttenuation   `json:"gainAttenuation,omitempty"`
	GeocodedIncidenceMask bool              `json:"geocodedIncidenceMask,omitempty"`
	Status                ItemStatus        `json:"status,omitempty"`
	Price                 *Price            `json:"price,omitempty"`
	OutOfFullPerformance  bool              `json:"outOfFullPerformance,omitempty"`
}

// CreateBasketRequest represents a basket creation request.
type CreateBasketRequest struct {
	Customer          string  `json:"customer,omitempty"`
	CustomerReference string  `json:"customerReference,omitempty"`
	NotifyEndpoint    *string `json:"notifyEndpoint,omitempty"`
	Purpose           Purpose `json:"purpose,omitempty"`
	OrderTemplate     string  `json:"orderTemplate,omitempty"`
}

// UpdateBasketRequest represents a basket update request (PATCH).
type UpdateBasketRequest struct {
	CustomerReference *string `json:"customerReference,omitempty"`
	NotifyEndpoint    *string `json:"notifyEndpoint,omitempty"`
	Purpose           Purpose `json:"purpose,omitempty"`
	OrderTemplate     string  `json:"orderTemplate,omitempty"`
}

// ReplaceBasketRequest represents a basket replacement request (PUT).
type ReplaceBasketRequest struct {
	CustomerReference string  `json:"customerReference,omitempty"`
	NotifyEndpoint    *string `json:"notifyEndpoint,omitempty"`
	Purpose           Purpose `json:"purpose,omitempty"`
	OrderTemplate     string  `json:"orderTemplate,omitempty"`
}

// AddItemsRequest represents a request to add items to a basket.
type AddItemsRequest struct {
	Acquisitions  []string      `json:"acquisitions,omitempty"` // Acquisition IDs
	Items         []string      `json:"items,omitempty"`        // Existing item UUIDs
	Customer      string        `json:"customer,omitempty"`
	OrderTemplate string        `json:"orderTemplate,omitempty"`
	OrderOptions  *OrderOptions `json:"orderOptions,omitempty"`
}

// RemoveItemsRequest represents a request to remove items from a basket.
type RemoveItemsRequest struct {
	Items []string `json:"items"` // Item UUIDs to remove
}

// RearrangeItemsRequest represents a request to rearrange items in a basket.
type RearrangeItemsRequest struct {
	Items []string `json:"items"` // Item UUIDs in desired order
}

// ----------------------------------------------------------------------------
// Order Types
// ----------------------------------------------------------------------------

// Order represents a submitted order.
type Order struct {
	BasketID          string     `json:"basketId"`
	OrderID           string     `json:"orderId"`
	Owner             string     `json:"owner,omitempty"`
	CreationTime      time.Time  `json:"creationTime"`
	SubmissionTime    time.Time  `json:"submissionTime"`
	Customer          string     `json:"customer,omitempty"`
	CustomerReference string     `json:"customerReference,omitempty"`
	NotifyEndpoint    *string    `json:"notifyEndpoint,omitempty"`
	Purpose           Purpose    `json:"purpose,omitempty"`
	OrderTemplate     string     `json:"orderTemplate,omitempty"`
	OrderType         OrderType  `json:"orderType,omitempty"`
	Items             []Item     `json:"items,omitempty"`
	Price             *Price     `json:"price,omitempty"`
	ItemStatistics    *ItemStats `json:"itemStatistics,omitempty"`
}

// OrderSummary represents an order in a list response.
type OrderSummary struct {
	BasketID          string     `json:"basketId"`
	OrderID           string     `json:"orderId,omitempty"`
	Owner             string     `json:"owner,omitempty"`
	CreationTime      time.Time  `json:"creationTime"`
	SubmissionTime    *time.Time `json:"submissionTime,omitempty"`
	Customer          string     `json:"customer,omitempty"`
	CustomerReference string     `json:"customerReference,omitempty"`
	Purpose           Purpose    `json:"purpose,omitempty"`
	OrderType         OrderType  `json:"orderType,omitempty"`
	ItemCount         int        `json:"itemCount,omitempty"`
	Price             *Price     `json:"price,omitempty"`
	ItemStatistics    *ItemStats `json:"itemStatistics,omitempty"`
}

// ItemStats contains item status statistics.
type ItemStats struct {
	Planned            int `json:"planned,omitempty"`
	Acquired           int `json:"acquired,omitempty"`
	AcquisitionFailed  int `json:"acquisitionFailed,omitempty"`
	Processing         int `json:"processing,omitempty"`
	Processed          int `json:"processed,omitempty"`
	ProcessingFailed   int `json:"processingFailed,omitempty"`
	Delivering         int `json:"delivering,omitempty"`
	Delivered          int `json:"delivered,omitempty"`
	DeliveryFailed     int `json:"deliveryFailed,omitempty"`
	Cancelled          int `json:"cancelled,omitempty"`
	CancellationFailed int `json:"cancellationFailed,omitempty"`
	Expired            int `json:"expired,omitempty"`
}

// UpdateOrderRequest represents an order update request.
type UpdateOrderRequest struct {
	NotifyEndpoint *string `json:"notifyEndpoint"`
}

// GetOrderItemsRequest represents a request to get items by ID.
type GetOrderItemsRequest struct {
	Items []string `json:"items"` // Order item IDs
}

// GetOrderItemsStatusRequest represents a request to get item status.
type GetOrderItemsStatusRequest struct {
	BasketID []string     `json:"basketId,omitempty"`
	Status   []ItemStatus `json:"status,omitempty"`
}

// OrderItemsStatusResponse represents item status response.
type OrderItemsStatusResponse struct {
	Items []ItemStatusEntry `json:"items"`
}

// ItemStatusEntry represents a single item status entry.
type ItemStatusEntry struct {
	ItemID   string     `json:"itemId"`
	BasketID string     `json:"basketId"`
	Status   ItemStatus `json:"status"`
}

// CancelItemsRequest represents a request to cancel items.
type CancelItemsRequest struct {
	Items []string `json:"items"` // Item UUIDs to cancel
}

// CancelItemsResponse represents the cancellation result.
type CancelItemsResponse struct {
	Cancelled []string          `json:"cancelled,omitempty"`
	Failed    []CancellationErr `json:"failed,omitempty"`
}

// CancellationErr represents a failed cancellation.
type CancellationErr struct {
	ItemID string `json:"itemId"`
	Reason string `json:"reason"`
}

// ReorderRequest represents a request to reorder items.
type ReorderRequest struct {
	Items         []string      `json:"items"`
	OrderOptions  *OrderOptions `json:"orderOptions,omitempty"`
	OrderTemplate string        `json:"orderTemplate,omitempty"`
}

// SubmitOrderRequest represents a direct order submission request.
type SubmitOrderRequest struct {
	BasketID string   `json:"basketId,omitempty"`
	Items    []string `json:"items,omitempty"`
}

// ----------------------------------------------------------------------------
// Config Types
// ----------------------------------------------------------------------------

// Config represents the complete user configuration.
type Config struct {
	Permissions       *Permissions       `json:"permissions,omitempty"`
	Settings          *Settings          `json:"settings,omitempty"`
	Customers         []Customer         `json:"customers,omitempty"`
	OrderTemplates    []OrderTemplate    `json:"orderTemplates,omitempty"`
	Associations      []Association      `json:"associations,omitempty"`
	ReceivingStations []ReceivingStation `json:"receivingStations,omitempty"`
}

// Permissions represents user permissions.
type Permissions struct {
	CanOrder                bool `json:"canOrder,omitempty"`
	CanTask                 bool `json:"canTask,omitempty"`
	CanTaskLeftLooking      bool `json:"canTaskLeftLooking,omitempty"`
	CanTaskOutOfFullPerf    bool `json:"canTaskOutOfFullPerformance,omitempty"`
	CanOrderAcquisitionOnly bool `json:"canOrderAcquisitionOnly,omitempty"`
	IsReseller              bool `json:"isReseller,omitempty"`
	IsDirectAccess          bool `json:"isDirectAccess,omitempty"`
}

// Settings represents user settings.
type Settings struct {
	DefaultOrderOptions *OrderOptions `json:"defaultOrderOptions,omitempty"`
	NotificationEmail   string        `json:"notificationEmail,omitempty"`
}

// Customer represents an end customer (for resellers).
type Customer struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// OrderTemplate represents an order template.
type OrderTemplate struct {
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	OrderOptions *OrderOptions `json:"orderOptions,omitempty"`
}

// Association represents an associated user.
type Association struct {
	Username string `json:"username"`
	Role     string `json:"role,omitempty"`
}

// ReceivingStation represents a receiving station (for direct-access customers).
type ReceivingStation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdateOrderOptionsRequest represents a request to update order options.
type UpdateOrderOptionsRequest struct {
	Items        []string      `json:"items"`
	OrderOptions *OrderOptions `json:"orderOptions,omitempty"`
}

// ----------------------------------------------------------------------------
// Feature Collection Types (GeoJSON)
// ----------------------------------------------------------------------------

// AcquisitionProperties represents properties of an acquisition feature.
type AcquisitionProperties struct {
	ItemID               string        `json:"itemId"`
	AcquisitionID        string        `json:"acquisitionId"`
	GroupID              int           `json:"groupId,omitempty"`
	Mission              Mission       `json:"mission"`
	Satellite            Satellite     `json:"satellite,omitempty"`
	SensorMode           SensorMode    `json:"sensorMode"`
	PolarizationChannels Polarization  `json:"polarizationChannels"`
	StartTime            time.Time     `json:"startTime"`
	StopTime             time.Time     `json:"stopTime"`
	BeamID               string        `json:"beamId"`
	PathDirection        PathDirection `json:"pathDirection"`
	LookDirection        LookDirection `json:"lookDirection"`
	IncidenceAngle       float64       `json:"incidenceAngle"`
	RelativeOrbit        int           `json:"relativeOrbit,omitempty"`
	Status               string        `json:"status,omitempty"`
	LastUpdateTime       *time.Time    `json:"lastUpdateTime,omitempty"`
	OutOfFullPerformance bool          `json:"outOfFullPerformance,omitempty"`
	// Feasibility-specific fields
	Coverage float64 `json:"coverage,omitempty"`
}

// Feature represents a GeoJSON Feature with acquisition properties.
type Feature struct {
	Type       string                `json:"type"` // "Feature"
	Geometry   *geojson.Geometry     `json:"geometry"`
	Properties AcquisitionProperties `json:"properties"`
}

// FeatureCollection represents a GeoJSON FeatureCollection.
type FeatureCollection struct {
	Type     string    `json:"type"` // "FeatureCollection"
	Features []Feature `json:"features"`
	Limit    int       `json:"limit,omitempty"`
	Total    int       `json:"total,omitempty"`
}

// ----------------------------------------------------------------------------
// Conflicts & Swath Editing Types
// ----------------------------------------------------------------------------

// ConflictsRequest represents a conflict check request.
type ConflictsRequest struct {
	Items []string `json:"items"` // Item UUIDs
}

// Conflict represents a conflict between items.
type Conflict struct {
	ItemID1 string `json:"itemId1"`
	ItemID2 string `json:"itemId2"`
	Reason  string `json:"reason"`
}

// ConflictsResponse represents the conflicts check result.
type ConflictsResponse struct {
	Conflicts []Conflict `json:"conflicts"`
}

// SwathEditRequest represents a swath editing request.
type SwathEditRequest struct {
	Items []string `json:"items"` // Item UUIDs
}

// SwathEditInfo represents swath editing information for an item.
type SwathEditInfo struct {
	ItemID           string            `json:"itemId"`
	EditableGeometry *geojson.Geometry `json:"editableGeometry,omitempty"`
	MinArea          float64           `json:"minArea,omitempty"`
	MaxArea          float64           `json:"maxArea,omitempty"`
}

// SwathEditResponse represents the swath editing information response.
type SwathEditResponse struct {
	Items []SwathEditInfo `json:"items"`
}

// UpdateSwathRequest represents an update to swath geometry.
type UpdateSwathRequest struct {
	Items []SwathUpdate `json:"items"`
}

// SwathUpdate represents a geometry update for a single item.
type SwathUpdate struct {
	ItemID   string            `json:"itemId"`
	Geometry *geojson.Geometry `json:"geometry"`
}

// ----------------------------------------------------------------------------
// Stacks Types
// ----------------------------------------------------------------------------

// CreateStacksRequest represents a datastack creation request.
type CreateStacksRequest struct {
	Items       []string    `json:"items"` // Template item UUIDs
	Periodicity Periodicity `json:"periodicity"`
	Occurrences int         `json:"occurrences"`
}

// StacksResponse represents the stacks creation response.
type StacksResponse struct {
	Stacks []Stack `json:"stacks"`
}

// Stack represents a datastack.
type Stack struct {
	GroupID int      `json:"groupId"`
	Items   []string `json:"items"` // Item UUIDs in the stack
}
