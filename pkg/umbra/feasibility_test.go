package umbra_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/robert.malhotra/go-sar-vendor/pkg/umbra"
)

func TestCreateFeasibility(t *testing.T) {
	expected := umbra.Feasibility{
		ID:          "feas-123",
		Status:      umbra.FeasibilityStatusReceived,
		ImagingMode: umbra.ImagingModeSpotlight,
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/tasking/feasibilities")
		requireAuth(t, r, "test-token")
		requireContentType(t, r, "application/json")
		jsonResponse(w, http.StatusCreated, expected)
	})

	req := &umbra.CreateFeasibilityRequest{
		ImagingMode: umbra.ImagingModeSpotlight,
		SpotlightConstraints: &umbra.SpotlightConstraints{
			Geometry: umbra.NewPointGeometry(-122.4194, 37.7749),
		},
		WindowStartAt: time.Now().Add(24 * time.Hour),
		WindowEndAt:   time.Now().Add(48 * time.Hour),
	}

	feas, err := cli.CreateFeasibility(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if feas.ID != expected.ID {
		t.Errorf("expected ID %s, got %s", expected.ID, feas.ID)
	}
	if feas.Status != expected.Status {
		t.Errorf("expected status %s, got %s", expected.Status, feas.Status)
	}
}

func TestGetFeasibility(t *testing.T) {
	expected := umbra.Feasibility{
		ID:          "feas-456",
		Status:      umbra.FeasibilityStatusCompleted,
		ImagingMode: umbra.ImagingModeSpotlight,
		Opportunities: []umbra.Opportunity{
			{
				WindowStartAt:            time.Now().Add(24 * time.Hour),
				WindowEndAt:              time.Now().Add(25 * time.Hour),
				GrazingAngleStartDegrees: 30.0,
				GrazingAngleEndDegrees:   45.0,
				SatelliteID:              "umbra-04",
			},
		},
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/tasking/feasibilities/feas-456")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	feas, err := cli.GetFeasibility(context.Background(), "feas-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if feas.ID != expected.ID {
		t.Errorf("expected ID %s, got %s", expected.ID, feas.ID)
	}
	if len(feas.Opportunities) != 1 {
		t.Errorf("expected 1 opportunity, got %d", len(feas.Opportunities))
	}
}

func TestListFeasibilities(t *testing.T) {
	expected := umbra.FeasibilityListResponse{
		Feasibilities: []umbra.Feasibility{
			{ID: "feas-1", Status: umbra.FeasibilityStatusCompleted},
			{ID: "feas-2", Status: umbra.FeasibilityStatusReceived},
		},
		TotalCount: 2,
		Limit:      10,
		Offset:     0,
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/tasking/feasibilities")
		requireAuth(t, r, "test-token")

		// Verify query parameters
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Errorf("expected limit=10, got %s", got)
		}

		jsonResponse(w, http.StatusOK, expected)
	})

	opts := &umbra.ListOptions{Limit: 10}
	resp, err := cli.ListFeasibilities(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Feasibilities) != 2 {
		t.Errorf("expected 2 feasibilities, got %d", len(resp.Feasibilities))
	}
}

func TestCreateFeasibility_Error(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusBadRequest, "Invalid request")
	})

	req := &umbra.CreateFeasibilityRequest{
		ImagingMode: umbra.ImagingModeSpotlight,
	}

	_, err := cli.CreateFeasibility(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !umbra.IsBadRequest(err) {
		t.Errorf("expected bad request error, got %v", err)
	}
}

func TestGetFeasibility_NotFound(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusNotFound, "Feasibility not found")
	})

	_, err := cli.GetFeasibility(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !umbra.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}
