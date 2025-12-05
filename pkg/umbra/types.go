package umbra

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

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

// ProductType represents the type of data product.
type ProductType string

const (
	ProductTypeGEC      ProductType = "GEC"      // Geocoded Ellipsoid Corrected
	ProductTypeSICD     ProductType = "SICD"     // Sensor Independent Complex Data
	ProductTypeSIDD     ProductType = "SIDD"     // Sensor Independent Detected Data
	ProductTypeCPHD     ProductType = "CPHD"     // Compensated Phase History Data
	ProductTypeCRSD     ProductType = "CRSD"     // Compensated Range-Doppler Signal Data
	ProductTypeMetadata ProductType = "METADATA" // JSON metadata
	ProductTypeDIGIF    ProductType = "DI_GIF"   // Display Image GIF
	ProductTypeDITIF    ProductType = "DI_TIF"   // Display Image TIF
)

// GeoJSONGeometry is an alias for geojson.Geometry for backwards compatibility.
type GeoJSONGeometry = geojson.Geometry

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

// SpotlightConstraints defines imaging parameters for spotlight mode.
type SpotlightConstraints struct {
	Geometry                       *geojson.Geometry `json:"geometry"`
	Polarization                   Polarization      `json:"polarization,omitempty"`
	RangeResolutionMinMeters       float64           `json:"rangeResolutionMinMeters,omitempty"`
	MultilookFactor                int               `json:"multilookFactor,omitempty"`
	GrazingAngleMinDegrees         float64           `json:"grazingAngleMinDegrees,omitempty"`
	GrazingAngleMaxDegrees         float64           `json:"grazingAngleMaxDegrees,omitempty"`
	TargetAzimuthAngleStartDegrees float64           `json:"targetAzimuthAngleStartDegrees,omitempty"`
	TargetAzimuthAngleEndDegrees   float64           `json:"targetAzimuthAngleEndDegrees,omitempty"`
	SceneSizeOption                string            `json:"sceneSizeOption,omitempty"`
}

// ScanConstraints defines imaging parameters for scan mode.
type ScanConstraints struct {
	StartPoint               *geojson.Geometry `json:"startPoint"`
	EndPoint                 *geojson.Geometry `json:"endPoint"`
	Polarization             Polarization      `json:"polarization,omitempty"`
	RangeResolutionMinMeters float64           `json:"rangeResolutionMinMeters,omitempty"`
	GrazingAngleMinDegrees   float64           `json:"grazingAngleMinDegrees,omitempty"`
	GrazingAngleMaxDegrees   float64           `json:"grazingAngleMaxDegrees,omitempty"`
}

// ProductConstraint represents constraints for a product type.
type ProductConstraint struct {
	ProductType       string  `json:"productType"`
	SceneSize         string  `json:"sceneSize"`
	MinGrazingDegrees float64 `json:"minGrazingAngle"`
	MaxGrazingDegrees float64 `json:"maxGrazingAngle"`
	RecommendedLooks  int     `json:"recommendedLooks"`
}

// ListOptions contains common pagination options.
type ListOptions = common.ListOptions

// WaitOptions configures polling behavior.
type WaitOptions = common.WaitOptions

// defaultSearchLimit is the default page size for search operations.
const defaultSearchLimit = 100
