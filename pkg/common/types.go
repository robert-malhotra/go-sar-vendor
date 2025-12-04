package common

import (
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

// TimeWindow represents a time range.
type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// BoundingBox represents a geographic bounding box [minLon, minLat, maxLon, maxLat].
type BoundingBox [4]float64

// MinLon returns the minimum longitude.
func (b BoundingBox) MinLon() float64 { return b[0] }

// MinLat returns the minimum latitude.
func (b BoundingBox) MinLat() float64 { return b[1] }

// MaxLon returns the maximum longitude.
func (b BoundingBox) MaxLon() float64 { return b[2] }

// MaxLat returns the maximum latitude.
func (b BoundingBox) MaxLat() float64 { return b[3] }

// ToOrbBound converts to an orb.Bound.
func (b BoundingBox) ToOrbBound() orb.Bound {
	return orb.Bound{
		Min: orb.Point{b[0], b[1]},
		Max: orb.Point{b[2], b[3]},
	}
}

// ToPolygon converts the bounding box to an orb.Polygon.
func (b BoundingBox) ToPolygon() orb.Polygon {
	return orb.Polygon{
		orb.Ring{
			orb.Point{b[0], b[1]}, // SW
			orb.Point{b[2], b[1]}, // SE
			orb.Point{b[2], b[3]}, // NE
			orb.Point{b[0], b[3]}, // NW
			orb.Point{b[0], b[1]}, // Close ring
		},
	}
}

// BoundingBoxFromOrb creates a BoundingBox from an orb.Bound.
func BoundingBoxFromOrb(b orb.Bound) BoundingBox {
	return BoundingBox{b.Min[0], b.Min[1], b.Max[0], b.Max[1]}
}

// Asset represents a downloadable asset (STAC-compatible).
type Asset struct {
	Href        string   `json:"href"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type,omitempty"` // MIME type
	Roles       []string `json:"roles,omitempty"`
}

// Link represents a STAC/web link.
type Link struct {
	Href  string `json:"href"`
	Rel   string `json:"rel"`
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
}

// Price represents a currency amount.
type Price struct {
	Amount   int64  `json:"amount"`   // In minor currency unit (e.g., cents)
	Currency string `json:"currency"` // ISO 4217 code (e.g., "USD")
}

// ListOptions contains common pagination options for list operations.
type ListOptions struct {
	Limit  int `url:"limit,omitempty"`
	Offset int `url:"offset,omitempty"`
}

// WaitOptions configures polling behavior for async operations.
type WaitOptions struct {
	PollInterval time.Duration
	Timeout      time.Duration
}

// Point creates an orb.Point from longitude and latitude.
func Point(lon, lat float64) orb.Point {
	return orb.Point{lon, lat}
}

// Polygon creates an orb.Polygon from coordinate rings.
// Each ring is a slice of [lon, lat] pairs.
func Polygon(rings [][][]float64) orb.Polygon {
	poly := make(orb.Polygon, len(rings))
	for i, ring := range rings {
		orbRing := make(orb.Ring, len(ring))
		for j, coord := range ring {
			orbRing[j] = orb.Point{coord[0], coord[1]}
		}
		poly[i] = orbRing
	}
	return poly
}

// BBoxToPolygon converts a bounding box to an orb.Polygon.
func BBoxToPolygon(bbox BoundingBox) orb.Polygon {
	return bbox.ToPolygon()
}

// NewGeoJSONPoint creates a GeoJSON geometry from a point.
func NewGeoJSONPoint(lon, lat float64) *geojson.Geometry {
	return geojson.NewGeometry(orb.Point{lon, lat})
}

// NewGeoJSONPolygon creates a GeoJSON geometry from coordinate rings.
func NewGeoJSONPolygon(rings [][][]float64) *geojson.Geometry {
	return geojson.NewGeometry(Polygon(rings))
}

// NewGeoJSONFromBBox creates a GeoJSON polygon geometry from a bounding box.
func NewGeoJSONFromBBox(bbox BoundingBox) *geojson.Geometry {
	return geojson.NewGeometry(bbox.ToPolygon())
}

// Feature represents a GeoJSON Feature with typed properties.
type Feature[T any] struct {
	Type       string           `json:"type"` // always "Feature"
	ID         string           `json:"id,omitempty"`
	Geometry   *geojson.Geometry `json:"geometry"`
	Properties T                `json:"properties"`
	BBox       BoundingBox      `json:"bbox,omitempty"`
}

// FeatureCollection represents a GeoJSON FeatureCollection.
type FeatureCollection[T any] struct {
	Type     string       `json:"type"` // always "FeatureCollection"
	Features []Feature[T] `json:"features"`
}
