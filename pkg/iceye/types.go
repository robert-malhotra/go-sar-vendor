package iceye

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/robert.malhotra/go-sar-vendor/pkg/common"
)

// ----------------------------------------------------------------------------
// Common Types - using common package where applicable
// ----------------------------------------------------------------------------

// Point represents a WGS84 coordinate point.
type Point struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// TimeWindow is an alias for common.TimeWindow.
type TimeWindow = common.TimeWindow

// IncidenceAngle represents an angle range in degrees.
type IncidenceAngle struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// BoundingBox is an alias for common.BoundingBox.
type BoundingBox = common.BoundingBox

// Price is an alias for common.Price.
type Price = common.Price

// Asset is an alias for common.Asset.
type Asset = common.Asset

// ----------------------------------------------------------------------------
// GeoJSON Types - using orb library
// ----------------------------------------------------------------------------

// Geometry is an alias for geojson.Geometry for backwards compatibility.
type Geometry = geojson.Geometry

// GeoJSONPoint creates a GeoJSON Point geometry.
func GeoJSONPoint(lon, lat float64) *geojson.Geometry {
	return geojson.NewGeometry(orb.Point{lon, lat})
}

// GeoJSONPolygon creates a GeoJSON Polygon from a linear ring.
func GeoJSONPolygon(coordinates [][][]float64) *geojson.Geometry {
	return common.NewGeoJSONPolygon(coordinates)
}

// BBoxToPolygon converts a bounding box to a GeoJSON polygon.
func BBoxToPolygon(bbox BoundingBox) *geojson.Geometry {
	return common.NewGeoJSONFromBBox(bbox)
}

// ----------------------------------------------------------------------------
// Shared Types (used across Company, Tasking, and Delivery APIs)
// ----------------------------------------------------------------------------

// DeliveryLocation specifies where products are delivered.
type DeliveryLocation struct {
	ConfigID string `json:"configID,omitempty"`
	Method   string `json:"method"` // "s3"
	Path     string `json:"path"`
}

// NotificationConfig specifies webhook notifications.
type NotificationConfig struct {
	Webhook *WebhookConfig `json:"webhook,omitempty"`
}

// WebhookConfig specifies a webhook for notifications.
type WebhookConfig struct {
	ID string `json:"id"`
}
