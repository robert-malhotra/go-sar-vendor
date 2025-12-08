package umbra_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/umbra"
)

func TestNewClient(t *testing.T) {
	cli, err := umbra.NewClient("test-token")
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if cli == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewSandboxClient(t *testing.T) {
	cli, err := umbra.NewSandboxClient("test-token")
	if err != nil {
		t.Fatalf("NewSandboxClient returned error: %v", err)
	}
	if cli == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestClientWithOptions(t *testing.T) {
	httpClient := &http.Client{Timeout: 60 * time.Second}

	cli, err := umbra.NewClient("test-token",
		umbra.WithHTTPClient(httpClient),
		umbra.WithBaseURL("https://custom.api.example.com"),
		umbra.WithTimeout(90*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	if cli == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestClientWithBaseURL(t *testing.T) {
	// Create a mock server
	cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Verify the request reaches our mock server
		jsonResponse(w, http.StatusOK, umbra.Task{ID: "test-task"})
	})

	// Make a request to verify the base URL is being used
	task, err := cli.GetTask(context.Background(), "test-task")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != "test-task" {
		t.Errorf("expected task ID test-task, got %s", task.ID)
	}

	// Clean up is handled by t.Cleanup in newTestClient
	_ = srv
}

func TestClientAuthHeader(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", auth)
		}
		jsonResponse(w, http.StatusOK, umbra.Task{ID: "test"})
	})

	_, err := cli.GetTask(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAPIErrorParsing(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code": "VALIDATION_ERROR", "message": "Invalid field"}`))
	})

	_, err := cli.GetTask(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*umbra.APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %s", apiErr.Code)
	}
	if apiErr.Message != "Invalid field" {
		t.Errorf("expected message 'Invalid field', got %s", apiErr.Message)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}
}

func TestAPIErrorString(t *testing.T) {
	tests := []struct {
		name     string
		err      *umbra.APIError
		expected string
	}{
		{
			name: "with code",
			err: &umbra.APIError{
				StatusCode: 400,
				Code:       "VALIDATION_ERROR",
				Message:    "Invalid field",
			},
			expected: "VALIDATION_ERROR (400): Invalid field",
		},
		{
			name: "without code",
			err: &umbra.APIError{
				StatusCode: 500,
				Message:    "Internal error",
			},
			expected: "Internal error (500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		isNotFound bool
		isBadReq   bool
		isUnauth   bool
		isRateLim  bool
		isForbid   bool
		isServer   bool
	}{
		{
			name:       "not found",
			err:        &umbra.APIError{StatusCode: 404},
			isNotFound: true,
		},
		{
			name:     "bad request",
			err:      &umbra.APIError{StatusCode: 400},
			isBadReq: true,
		},
		{
			name:     "unauthorized",
			err:      &umbra.APIError{StatusCode: 401},
			isUnauth: true,
		},
		{
			name:      "rate limited",
			err:       &umbra.APIError{StatusCode: 429},
			isRateLim: true,
		},
		{
			name:     "forbidden",
			err:      &umbra.APIError{StatusCode: 403},
			isForbid: true,
		},
		{
			name:     "server error",
			err:      &umbra.APIError{StatusCode: 500},
			isServer: true,
		},
		{
			name:     "server error 502",
			err:      &umbra.APIError{StatusCode: 502},
			isServer: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := umbra.IsNotFound(tt.err); got != tt.isNotFound {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.isNotFound)
			}
			if got := umbra.IsBadRequest(tt.err); got != tt.isBadReq {
				t.Errorf("IsBadRequest() = %v, want %v", got, tt.isBadReq)
			}
			if got := umbra.IsUnauthorized(tt.err); got != tt.isUnauth {
				t.Errorf("IsUnauthorized() = %v, want %v", got, tt.isUnauth)
			}
			if got := umbra.IsRateLimited(tt.err); got != tt.isRateLim {
				t.Errorf("IsRateLimited() = %v, want %v", got, tt.isRateLim)
			}
			if got := umbra.IsForbidden(tt.err); got != tt.isForbid {
				t.Errorf("IsForbidden() = %v, want %v", got, tt.isForbid)
			}
			if got := umbra.IsServerError(tt.err); got != tt.isServer {
				t.Errorf("IsServerError() = %v, want %v", got, tt.isServer)
			}
		})
	}
}

func TestGeometryHelpers(t *testing.T) {
	t.Run("NewPointGeometry", func(t *testing.T) {
		geom := umbra.NewPointGeometry(-122.4194, 37.7749)
		if geom.Type != "Point" {
			t.Errorf("expected type Point, got %s", geom.Type)
		}
		if geom.Coordinates == nil {
			t.Error("expected coordinates to be set")
		}
	})

	t.Run("NewPolygonGeometry", func(t *testing.T) {
		coords := [][][2]float64{
			{{-122.5, 37.5}, {-122.0, 37.5}, {-122.0, 38.0}, {-122.5, 38.0}, {-122.5, 37.5}},
		}
		geom := umbra.NewPolygonGeometry(coords)
		if geom.Type != "Polygon" {
			t.Errorf("expected type Polygon, got %s", geom.Type)
		}
		if geom.Coordinates == nil {
			t.Error("expected coordinates to be set")
		}
	})
}

func TestCQL2Builder(t *testing.T) {
	cql := umbra.CQL2{}

	t.Run("Equal", func(t *testing.T) {
		filter := cql.Equal("property", "value")
		if filter["op"] != "=" {
			t.Errorf("expected op '=', got %v", filter["op"])
		}
	})

	t.Run("NotEqual", func(t *testing.T) {
		filter := cql.NotEqual("property", "value")
		if filter["op"] != "<>" {
			t.Errorf("expected op '<>', got %v", filter["op"])
		}
	})

	t.Run("GreaterThan", func(t *testing.T) {
		filter := cql.GreaterThan("property", 10)
		if filter["op"] != ">" {
			t.Errorf("expected op '>', got %v", filter["op"])
		}
	})

	t.Run("GreaterThanOrEqual", func(t *testing.T) {
		filter := cql.GreaterThanOrEqual("property", 10)
		if filter["op"] != ">=" {
			t.Errorf("expected op '>=', got %v", filter["op"])
		}
	})

	t.Run("LessThan", func(t *testing.T) {
		filter := cql.LessThan("property", 10)
		if filter["op"] != "<" {
			t.Errorf("expected op '<', got %v", filter["op"])
		}
	})

	t.Run("LessThanOrEqual", func(t *testing.T) {
		filter := cql.LessThanOrEqual("property", 10)
		if filter["op"] != "<=" {
			t.Errorf("expected op '<=', got %v", filter["op"])
		}
	})

	t.Run("And", func(t *testing.T) {
		filter := cql.And(
			cql.Equal("a", 1),
			cql.Equal("b", 2),
		)
		if filter["op"] != "and" {
			t.Errorf("expected op 'and', got %v", filter["op"])
		}
		args, ok := filter["args"].([]interface{})
		if !ok || len(args) != 2 {
			t.Errorf("expected 2 args, got %v", filter["args"])
		}
	})

	t.Run("Or", func(t *testing.T) {
		filter := cql.Or(
			cql.Equal("a", 1),
			cql.Equal("b", 2),
		)
		if filter["op"] != "or" {
			t.Errorf("expected op 'or', got %v", filter["op"])
		}
	})

	t.Run("Not", func(t *testing.T) {
		filter := cql.Not(cql.Equal("a", 1))
		if filter["op"] != "not" {
			t.Errorf("expected op 'not', got %v", filter["op"])
		}
	})

	t.Run("In", func(t *testing.T) {
		filter := cql.In("status", []interface{}{"ACTIVE", "SCHEDULED"})
		if filter["op"] != "in" {
			t.Errorf("expected op 'in', got %v", filter["op"])
		}
	})

	t.Run("Between", func(t *testing.T) {
		filter := cql.Between("value", 10, 20)
		if filter["op"] != "between" {
			t.Errorf("expected op 'between', got %v", filter["op"])
		}
	})

	t.Run("Like", func(t *testing.T) {
		filter := cql.Like("name", "test%")
		if filter["op"] != "like" {
			t.Errorf("expected op 'like', got %v", filter["op"])
		}
	})

	t.Run("IsNull", func(t *testing.T) {
		filter := cql.IsNull("property")
		if filter["op"] != "isNull" {
			t.Errorf("expected op 'isNull', got %v", filter["op"])
		}
	})

	t.Run("Intersects", func(t *testing.T) {
		geom := umbra.NewPointGeometry(-122.4, 37.7)
		filter := cql.Intersects("geometry", geom)
		if filter["op"] != "s_intersects" {
			t.Errorf("expected op 's_intersects', got %v", filter["op"])
		}
	})

	t.Run("Within", func(t *testing.T) {
		geom := umbra.NewPolygonGeometry([][][2]float64{
			{{-122.5, 37.5}, {-122.0, 37.5}, {-122.0, 38.0}, {-122.5, 38.0}, {-122.5, 37.5}},
		})
		filter := cql.Within("geometry", geom)
		if filter["op"] != "s_within" {
			t.Errorf("expected op 's_within', got %v", filter["op"])
		}
	})
}

func TestTaskHelpers(t *testing.T) {
	now := time.Now()
	windowStart := now.Add(24 * time.Hour)
	windowEnd := now.Add(48 * time.Hour)

	t.Run("NewSpotlightTask", func(t *testing.T) {
		req := umbra.NewSpotlightTask(-122.4194, 37.7749, windowStart, windowEnd,
			umbra.WithTaskName("Test Task"),
			umbra.WithProductTypes(umbra.ProductTypeGEC, umbra.ProductTypeSICD),
			umbra.WithResolution(1.0),
			umbra.WithGrazingAngle(30, 70),
			umbra.WithAzimuthAngle(0, 360),
			umbra.WithPolarization(umbra.PolarizationVV),
		)

		if req.ImagingMode != umbra.ImagingModeSpotlight {
			t.Errorf("expected SPOTLIGHT mode, got %s", req.ImagingMode)
		}
		if req.TaskName != "Test Task" {
			t.Errorf("expected task name 'Test Task', got %s", req.TaskName)
		}
		if len(req.ProductTypes) != 2 {
			t.Errorf("expected 2 product types, got %d", len(req.ProductTypes))
		}
		if req.SpotlightConstraints == nil {
			t.Fatal("expected spotlight constraints to be set")
		}
		if req.SpotlightConstraints.RangeResolutionMinMeters != 1.0 {
			t.Errorf("expected resolution 1.0, got %f", req.SpotlightConstraints.RangeResolutionMinMeters)
		}
	})

	t.Run("NewSpotlightTaskWithPolygon", func(t *testing.T) {
		coords := [][][2]float64{
			{{-122.5, 37.5}, {-122.0, 37.5}, {-122.0, 38.0}, {-122.5, 38.0}, {-122.5, 37.5}},
		}
		req := umbra.NewSpotlightTaskWithPolygon(coords, windowStart, windowEnd)

		if req.ImagingMode != umbra.ImagingModeSpotlight {
			t.Errorf("expected SPOTLIGHT mode, got %s", req.ImagingMode)
		}
		if req.SpotlightConstraints.Geometry.Type != "Polygon" {
			t.Errorf("expected Polygon geometry, got %s", req.SpotlightConstraints.Geometry.Type)
		}
	})

	t.Run("NewScanTask", func(t *testing.T) {
		req := umbra.NewScanTask(-122.5, 37.5, -122.0, 38.0, windowStart, windowEnd,
			umbra.WithResolution(2.0),
			umbra.WithPolarization(umbra.PolarizationHH),
		)

		if req.ImagingMode != umbra.ImagingModeScan {
			t.Errorf("expected SCAN mode, got %s", req.ImagingMode)
		}
		if req.ScanConstraints == nil {
			t.Fatal("expected scan constraints to be set")
		}
		if req.ScanConstraints.RangeResolutionMinMeters != 2.0 {
			t.Errorf("expected resolution 2.0, got %f", req.ScanConstraints.RangeResolutionMinMeters)
		}
	})

	t.Run("NewSpotlightTaskFromOpportunity", func(t *testing.T) {
		feas := &umbra.Feasibility{
			ImagingMode: umbra.ImagingModeSpotlight,
			SpotlightConstraints: &umbra.SpotlightConstraints{
				Geometry: umbra.NewPointGeometry(-122.4, 37.7),
			},
			Opportunities: []umbra.Opportunity{
				{
					WindowStartAt: windowStart,
					WindowEndAt:   windowEnd,
					SatelliteID:   "umbra-04",
				},
			},
		}

		req := umbra.NewSpotlightTaskFromOpportunity(feas, 0, "From Opportunity")
		if req == nil {
			t.Fatal("expected non-nil request")
		}
		if req.TaskName != "From Opportunity" {
			t.Errorf("expected task name 'From Opportunity', got %s", req.TaskName)
		}
		if req.WindowStartAt != windowStart {
			t.Error("expected window start from opportunity")
		}
	})

	t.Run("NewSpotlightTaskFromOpportunity_InvalidIndex", func(t *testing.T) {
		feas := &umbra.Feasibility{
			Opportunities: []umbra.Opportunity{},
		}

		req := umbra.NewSpotlightTaskFromOpportunity(feas, 0, "Test")
		if req != nil {
			t.Error("expected nil for invalid index")
		}
	})

	t.Run("WithOptions", func(t *testing.T) {
		req := umbra.NewSpotlightTask(-122.4, 37.7, windowStart, windowEnd,
			umbra.WithUserOrderID("order-123"),
			umbra.WithDeliveryConfig("dc-456"),
			umbra.WithMultilookFactor(2),
			umbra.WithSceneSizeOption("5km"),
			umbra.WithSatelliteIDs("umbra-04", "umbra-05"),
			umbra.WithTags("test", "sample"),
		)

		if req.UserOrderID != "order-123" {
			t.Errorf("expected order ID 'order-123', got %s", req.UserOrderID)
		}
		if req.DeliveryConfigID != "dc-456" {
			t.Errorf("expected delivery config 'dc-456', got %s", req.DeliveryConfigID)
		}
		if req.SpotlightConstraints.MultilookFactor != 2 {
			t.Errorf("expected multilook factor 2, got %d", req.SpotlightConstraints.MultilookFactor)
		}
		if req.SpotlightConstraints.SceneSizeOption != "5km" {
			t.Errorf("expected scene size '5km', got %s", req.SpotlightConstraints.SceneSizeOption)
		}
		if len(req.SatelliteIDs) != 2 {
			t.Errorf("expected 2 satellite IDs, got %d", len(req.SatelliteIDs))
		}
		if len(req.Tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(req.Tags))
		}
	})
}
