package umbra

import (
	"time"

	"github.com/paulmach/orb/geojson"
)

// TaskOption configures optional task parameters.
type TaskOption func(*CreateTaskRequest)

// WithTaskName sets the task name.
func WithTaskName(name string) TaskOption {
	return func(r *CreateTaskRequest) {
		r.TaskName = name
	}
}

// WithUserOrderID sets the user order ID.
func WithUserOrderID(id string) TaskOption {
	return func(r *CreateTaskRequest) {
		r.UserOrderID = id
	}
}

// WithDeliveryConfig sets the delivery configuration.
func WithDeliveryConfig(configID string) TaskOption {
	return func(r *CreateTaskRequest) {
		r.DeliveryConfigID = configID
	}
}

// WithProductTypes sets the product types to deliver.
func WithProductTypes(types ...ProductType) TaskOption {
	return func(r *CreateTaskRequest) {
		r.ProductTypes = types
	}
}

// WithResolution sets the range resolution.
func WithResolution(meters float64) TaskOption {
	return func(r *CreateTaskRequest) {
		if r.SpotlightConstraints != nil {
			r.SpotlightConstraints.RangeResolutionMinMeters = meters
		}
		if r.ScanConstraints != nil {
			r.ScanConstraints.RangeResolutionMinMeters = meters
		}
	}
}

// WithGrazingAngle sets the grazing angle constraints.
func WithGrazingAngle(min, max float64) TaskOption {
	return func(r *CreateTaskRequest) {
		if r.SpotlightConstraints != nil {
			r.SpotlightConstraints.GrazingAngleMinDegrees = min
			r.SpotlightConstraints.GrazingAngleMaxDegrees = max
		}
		if r.ScanConstraints != nil {
			r.ScanConstraints.GrazingAngleMinDegrees = min
			r.ScanConstraints.GrazingAngleMaxDegrees = max
		}
	}
}

// WithAzimuthAngle sets the target azimuth angle range.
func WithAzimuthAngle(start, end float64) TaskOption {
	return func(r *CreateTaskRequest) {
		if r.SpotlightConstraints != nil {
			r.SpotlightConstraints.TargetAzimuthAngleStartDegrees = start
			r.SpotlightConstraints.TargetAzimuthAngleEndDegrees = end
		}
	}
}

// WithPolarization sets the polarization type.
func WithPolarization(p Polarization) TaskOption {
	return func(r *CreateTaskRequest) {
		if r.SpotlightConstraints != nil {
			r.SpotlightConstraints.Polarization = p
		}
		if r.ScanConstraints != nil {
			r.ScanConstraints.Polarization = p
		}
	}
}

// WithMultilookFactor sets the multilook factor.
func WithMultilookFactor(factor int) TaskOption {
	return func(r *CreateTaskRequest) {
		if r.SpotlightConstraints != nil {
			r.SpotlightConstraints.MultilookFactor = factor
		}
	}
}

// WithSceneSizeOption sets the scene size option.
func WithSceneSizeOption(option string) TaskOption {
	return func(r *CreateTaskRequest) {
		if r.SpotlightConstraints != nil {
			r.SpotlightConstraints.SceneSizeOption = option
		}
	}
}

// WithSatelliteIDs sets the satellite IDs for the task.
func WithSatelliteIDs(ids ...string) TaskOption {
	return func(r *CreateTaskRequest) {
		r.SatelliteIDs = ids
	}
}

// WithTags sets the tags for the task.
func WithTags(tags ...string) TaskOption {
	return func(r *CreateTaskRequest) {
		r.Tags = tags
	}
}

// NewSpotlightTask creates a task request for spotlight imaging.
func NewSpotlightTask(lon, lat float64, windowStart, windowEnd time.Time, opts ...TaskOption) *CreateTaskRequest {
	req := &CreateTaskRequest{
		ImagingMode: ImagingModeSpotlight,
		SpotlightConstraints: &SpotlightConstraints{
			Geometry: NewPointGeometry(lon, lat),
		},
		WindowStartAt: windowStart,
		WindowEndAt:   windowEnd,
	}
	for _, opt := range opts {
		opt(req)
	}
	return req
}

// NewSpotlightTaskWithPolygon creates a task request for spotlight imaging with a polygon AOI.
func NewSpotlightTaskWithPolygon(coords [][][2]float64, windowStart, windowEnd time.Time, opts ...TaskOption) *CreateTaskRequest {
	req := &CreateTaskRequest{
		ImagingMode: ImagingModeSpotlight,
		SpotlightConstraints: &SpotlightConstraints{
			Geometry: NewPolygonGeometry(coords),
		},
		WindowStartAt: windowStart,
		WindowEndAt:   windowEnd,
	}
	for _, opt := range opts {
		opt(req)
	}
	return req
}

// NewScanTask creates a task request for scan mode imaging.
func NewScanTask(startLon, startLat, endLon, endLat float64, windowStart, windowEnd time.Time, opts ...TaskOption) *CreateTaskRequest {
	req := &CreateTaskRequest{
		ImagingMode: ImagingModeScan,
		ScanConstraints: &ScanConstraints{
			StartPoint: NewPointGeometry(startLon, startLat),
			EndPoint:   NewPointGeometry(endLon, endLat),
		},
		WindowStartAt: windowStart,
		WindowEndAt:   windowEnd,
	}
	for _, opt := range opts {
		opt(req)
	}
	return req
}

// NewSpotlightTaskFromOpportunity creates a task request from a feasibility opportunity.
func NewSpotlightTaskFromOpportunity(feasibility *Feasibility, opportunityIndex int, taskName string) *CreateTaskRequest {
	if opportunityIndex >= len(feasibility.Opportunities) {
		return nil
	}
	opp := feasibility.Opportunities[opportunityIndex]

	req := &CreateTaskRequest{
		TaskName:             taskName,
		ImagingMode:          feasibility.ImagingMode,
		SpotlightConstraints: feasibility.SpotlightConstraints,
		ScanConstraints:      feasibility.ScanConstraints,
		WindowStartAt:        opp.WindowStartAt,
		WindowEndAt:          opp.WindowEndAt,
	}
	return req
}

// CQL2 provides a builder for CQL2 filter expressions.
type CQL2 struct{}

// Equal creates an equality filter.
func (CQL2) Equal(property string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"op":   "=",
		"args": []interface{}{map[string]string{"property": property}, value},
	}
}

// NotEqual creates a not-equal filter.
func (CQL2) NotEqual(property string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"op":   "<>",
		"args": []interface{}{map[string]string{"property": property}, value},
	}
}

// GreaterThan creates a greater-than filter.
func (CQL2) GreaterThan(property string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"op":   ">",
		"args": []interface{}{map[string]string{"property": property}, value},
	}
}

// GreaterThanOrEqual creates a greater-than-or-equal filter.
func (CQL2) GreaterThanOrEqual(property string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"op":   ">=",
		"args": []interface{}{map[string]string{"property": property}, value},
	}
}

// LessThan creates a less-than filter.
func (CQL2) LessThan(property string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"op":   "<",
		"args": []interface{}{map[string]string{"property": property}, value},
	}
}

// LessThanOrEqual creates a less-than-or-equal filter.
func (CQL2) LessThanOrEqual(property string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"op":   "<=",
		"args": []interface{}{map[string]string{"property": property}, value},
	}
}

// And combines multiple filters with AND logic.
func (CQL2) And(filters ...map[string]interface{}) map[string]interface{} {
	args := make([]interface{}, len(filters))
	for i, f := range filters {
		args[i] = f
	}
	return map[string]interface{}{
		"op":   "and",
		"args": args,
	}
}

// Or combines multiple filters with OR logic.
func (CQL2) Or(filters ...map[string]interface{}) map[string]interface{} {
	args := make([]interface{}, len(filters))
	for i, f := range filters {
		args[i] = f
	}
	return map[string]interface{}{
		"op":   "or",
		"args": args,
	}
}

// Not negates a filter.
func (CQL2) Not(filter map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"op":   "not",
		"args": []interface{}{filter},
	}
}

// In creates an "in" filter for matching multiple values.
func (CQL2) In(property string, values []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"op":   "in",
		"args": []interface{}{map[string]string{"property": property}, values},
	}
}

// Between creates a "between" filter for range matching.
func (CQL2) Between(property string, lower, upper interface{}) map[string]interface{} {
	return map[string]interface{}{
		"op":   "between",
		"args": []interface{}{map[string]string{"property": property}, lower, upper},
	}
}

// Like creates a "like" filter for pattern matching.
func (CQL2) Like(property string, pattern string) map[string]interface{} {
	return map[string]interface{}{
		"op":   "like",
		"args": []interface{}{map[string]string{"property": property}, pattern},
	}
}

// IsNull creates an "isNull" filter.
func (CQL2) IsNull(property string) map[string]interface{} {
	return map[string]interface{}{
		"op":   "isNull",
		"args": []interface{}{map[string]string{"property": property}},
	}
}

// Intersects creates a spatial intersects filter.
func (CQL2) Intersects(property string, geometry *geojson.Geometry) map[string]interface{} {
	return map[string]interface{}{
		"op": "s_intersects",
		"args": []interface{}{
			map[string]string{"property": property},
			geometry,
		},
	}
}

// Within creates a spatial within filter.
func (CQL2) Within(property string, geometry *geojson.Geometry) map[string]interface{} {
	return map[string]interface{}{
		"op": "s_within",
		"args": []interface{}{
			map[string]string{"property": property},
			geometry,
		},
	}
}
