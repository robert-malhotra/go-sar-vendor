package planet_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/planet"
)

func TestNewClient(t *testing.T) {
	cli, err := planet.NewClient("test-api-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cli == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClientMissingAPIKey(t *testing.T) {
	// Client should still work with empty API key (auth will fail on requests)
	cli, err := planet.NewClient("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cli == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestClientWithOptions(t *testing.T) {
	httpClient := &http.Client{Timeout: 60 * time.Second}

	cli, err := planet.NewClient("test-api-key",
		planet.WithHTTPClient(httpClient),
		planet.WithBaseURL("https://custom.api.example.com"),
		planet.WithTimeout(90*time.Second),
		planet.WithUserAgent("test-agent/1.0"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cli == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestClientAuthHeader(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "api-key test-api-key" {
			t.Errorf("expected 'api-key test-api-key', got %s", auth)
		}
		jsonResponse(w, http.StatusOK, planet.TaskingOrder{ID: "test"})
	})

	_, err := cli.GetTaskingOrder(context.Background(), "test")
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

	_, err := cli.GetTaskingOrder(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*planet.APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
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
			err:        &planet.APIError{StatusCode: 404},
			isNotFound: true,
		},
		{
			name:     "bad request",
			err:      &planet.APIError{StatusCode: 400},
			isBadReq: true,
		},
		{
			name:     "unauthorized",
			err:      &planet.APIError{StatusCode: 401},
			isUnauth: true,
		},
		{
			name:      "rate limited",
			err:       &planet.APIError{StatusCode: 429},
			isRateLim: true,
		},
		{
			name:     "forbidden",
			err:      &planet.APIError{StatusCode: 403},
			isForbid: true,
		},
		{
			name:     "server error",
			err:      &planet.APIError{StatusCode: 500},
			isServer: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := planet.IsNotFound(tt.err); got != tt.isNotFound {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.isNotFound)
			}
			if got := planet.IsBadRequest(tt.err); got != tt.isBadReq {
				t.Errorf("IsBadRequest() = %v, want %v", got, tt.isBadReq)
			}
			if got := planet.IsUnauthorized(tt.err); got != tt.isUnauth {
				t.Errorf("IsUnauthorized() = %v, want %v", got, tt.isUnauth)
			}
			if got := planet.IsRateLimited(tt.err); got != tt.isRateLim {
				t.Errorf("IsRateLimited() = %v, want %v", got, tt.isRateLim)
			}
			if got := planet.IsForbidden(tt.err); got != tt.isForbid {
				t.Errorf("IsForbidden() = %v, want %v", got, tt.isForbid)
			}
			if got := planet.IsServerError(tt.err); got != tt.isServer {
				t.Errorf("IsServerError() = %v, want %v", got, tt.isServer)
			}
		})
	}
}

func TestGeometryHelpers(t *testing.T) {
	t.Run("NewPointGeometry", func(t *testing.T) {
		geom := planet.NewPointGeometry(-122.4194, 37.7749)
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
		geom := planet.NewPolygonGeometry(coords)
		if geom.Type != "Polygon" {
			t.Errorf("expected type Polygon, got %s", geom.Type)
		}
		if geom.Coordinates == nil {
			t.Error("expected coordinates to be set")
		}
	})

	t.Run("BBoxToPolygon", func(t *testing.T) {
		geom := planet.BBoxToPolygon(-122.5, 37.5, -122.0, 38.0)
		if geom.Type != "Polygon" {
			t.Errorf("expected type Polygon, got %s", geom.Type)
		}
		if geom.Coordinates == nil {
			t.Error("expected coordinates to be set")
		}
	})
}

func TestTaskingOrderStatusIsTerminal(t *testing.T) {
	tests := []struct {
		status   planet.TaskingOrderStatus
		terminal bool
	}{
		{planet.TaskingOrderStatusReceived, false},
		{planet.TaskingOrderStatusPending, false},
		{planet.TaskingOrderStatusInProgress, false},
		{planet.TaskingOrderStatusFulfilled, true},
		{planet.TaskingOrderStatusFailed, true},
		{planet.TaskingOrderStatusCancelled, true},
		{planet.TaskingOrderStatusExpired, true},
		{planet.TaskingOrderStatusRejected, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.terminal {
				t.Errorf("%s.IsTerminal() = %v, want %v", tt.status, got, tt.terminal)
			}
		})
	}
}

func TestOrderStateIsTerminal(t *testing.T) {
	tests := []struct {
		state    planet.OrderState
		terminal bool
	}{
		{planet.OrderStateQueued, false},
		{planet.OrderStateRunning, false},
		{planet.OrderStateSuccess, true},
		{planet.OrderStatePartial, true},
		{planet.OrderStateFailed, true},
		{planet.OrderStateCancelled, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.IsTerminal(); got != tt.terminal {
				t.Errorf("%s.IsTerminal() = %v, want %v", tt.state, got, tt.terminal)
			}
		})
	}
}

func TestImagingWindowSearchStatusIsTerminal(t *testing.T) {
	tests := []struct {
		status   planet.ImagingWindowSearchStatus
		terminal bool
	}{
		{planet.ImagingWindowSearchStatusCreated, false},
		{planet.ImagingWindowSearchStatusInProgress, false},
		{planet.ImagingWindowSearchStatusDone, true},
		{planet.ImagingWindowSearchStatusFailed, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.terminal {
				t.Errorf("%s.IsTerminal() = %v, want %v", tt.status, got, tt.terminal)
			}
		})
	}
}

func TestToolHelpers(t *testing.T) {
	t.Run("ClipTool", func(t *testing.T) {
		aoi := planet.NewPointGeometry(-122.4, 37.7)
		tool := planet.ClipTool(aoi)
		if tool.Type != "clip" {
			t.Errorf("expected type 'clip', got %s", tool.Type)
		}
		if tool.Parameters["aoi"] == nil {
			t.Error("expected aoi parameter to be set")
		}
	})

	t.Run("ReprojectTool", func(t *testing.T) {
		tool := planet.ReprojectTool("EPSG:4326")
		if tool.Type != "reproject" {
			t.Errorf("expected type 'reproject', got %s", tool.Type)
		}
		if tool.Parameters["projection"] != "EPSG:4326" {
			t.Errorf("expected projection 'EPSG:4326', got %v", tool.Parameters["projection"])
		}
	})

	t.Run("CompositeTool", func(t *testing.T) {
		tool := planet.CompositeTool()
		if tool.Type != "composite" {
			t.Errorf("expected type 'composite', got %s", tool.Type)
		}
	})

	t.Run("FileFormatTool", func(t *testing.T) {
		tool := planet.FileFormatTool("COG")
		if tool.Type != "file_format" {
			t.Errorf("expected type 'file_format', got %s", tool.Type)
		}
		if tool.Parameters["format"] != "COG" {
			t.Errorf("expected format 'COG', got %v", tool.Parameters["format"])
		}
	})
}
