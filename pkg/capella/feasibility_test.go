package capella_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/robert.malhotra/go-sar-vendor/pkg/capella"
)

func TestFeasibilityService_CreateAccessRequest(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/ma/accessrequests")
		requireAuth(t, r, "test-api-key")
		requireContentType(t, r, "application/json")

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)

		if req["type"] != "Feature" {
			t.Errorf("expected type 'Feature', got %v", req["type"])
		}

		// Return a valid GeoJSON response
		jsonResponse(w, http.StatusCreated, map[string]any{
			"type": "Feature",
			"geometry": map[string]any{
				"type":        "Point",
				"coordinates": []float64{10.0, 20.0},
			},
			"properties": map[string]any{
				"orgId":               "org-1",
				"userId":              "u-1",
				"accessRequestId":     "ar-123",
				"processingStatus":    "queued",
				"accessibilityStatus": "unknown",
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	feasibility := capella.NewFeasibilityService(cli)

	req := capella.AccessRequest{
		Type:     "Feature",
		Geometry: capella.Point(10.0, 20.0),
		Properties: capella.AccessRequestProperties{
			OrgID:       "org-1",
			UserID:      "u-1",
			WindowOpen:  time.Now(),
			WindowClose: time.Now().Add(12 * time.Hour),
		},
	}

	resp, err := feasibility.CreateAccessRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateAccessRequest failed: %v", err)
	}

	if resp.Properties.AccessRequestID != "ar-123" {
		t.Errorf("expected access request ID 'ar-123', got %q", resp.Properties.AccessRequestID)
	}

	if resp.Properties.ProcessingStatus != capella.ProcessingQueued {
		t.Errorf("expected status 'queued', got %q", resp.Properties.ProcessingStatus)
	}
}

func TestFeasibilityService_CreateAccessRequest_ValidationError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		validationError(w, "windowOpen", "field required")
	}

	cli, _ := newTestClient(t, handler)
	feasibility := capella.NewFeasibilityService(cli)

	_, err := feasibility.CreateAccessRequest(context.Background(), capella.AccessRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*capella.APIError)
	if !ok {
		t.Fatalf("expected *capella.APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", apiErr.StatusCode)
	}

	// Check that error message contains validation info
	if apiErr.Message == "" {
		t.Error("expected validation message in error")
	}
}

func TestFeasibilityService_GetAccessRequest(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/ma/accessrequests/ar-123")
		requireAuth(t, r, "test-api-key")

		jsonResponse(w, http.StatusOK, map[string]any{
			"type": "Feature",
			"geometry": map[string]any{
				"type":        "Point",
				"coordinates": []float64{10.0, 20.0},
			},
			"properties": map[string]any{
				"accessRequestId":     "ar-123",
				"processingStatus":    "completed",
				"accessibilityStatus": "accessible",
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	feasibility := capella.NewFeasibilityService(cli)

	resp, err := feasibility.GetAccessRequest(context.Background(), "ar-123")
	if err != nil {
		t.Fatalf("GetAccessRequest failed: %v", err)
	}

	if resp.Properties.AccessRequestID != "ar-123" {
		t.Errorf("expected access request ID 'ar-123', got %q", resp.Properties.AccessRequestID)
	}

	if resp.Properties.ProcessingStatus != capella.ProcessingCompleted {
		t.Errorf("expected status 'completed', got %q", resp.Properties.ProcessingStatus)
	}

	if resp.Properties.AccessibilityStatus != capella.AccessibilityAccessible {
		t.Errorf("expected accessibility 'accessible', got %q", resp.Properties.AccessibilityStatus)
	}
}

func TestFeasibilityService_GetAccessRequestDetail(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/ma/accessrequests/ar-123/detail")

		jsonResponse(w, http.StatusOK, capella.AccessRequestDetailResponse{
			AccessRequestResponse: capella.AccessRequestResponse{
				Properties: capella.AccessRequestPropertiesResponse{
					AccessRequestID:     "ar-123",
					ProcessingStatus:    capella.ProcessingCompleted,
					AccessibilityStatus: capella.AccessibilityAccessible,
				},
			},
			AccessWindows: []capella.AccessWindow{
				{
					WindowOpen:    time.Now(),
					WindowClose:   time.Now().Add(30 * time.Minute),
					OrbitalPlane:  "45",
					LookDirection: "right",
					AscDesc:       "ascending",
					OffNadir:      25.5,
				},
				{
					WindowOpen:    time.Now().Add(2 * time.Hour),
					WindowClose:   time.Now().Add(2*time.Hour + 30*time.Minute),
					OrbitalPlane:  "53",
					LookDirection: "left",
					AscDesc:       "descending",
					OffNadir:      30.0,
				},
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	feasibility := capella.NewFeasibilityService(cli)

	resp, err := feasibility.GetAccessRequestDetail(context.Background(), "ar-123")
	if err != nil {
		t.Fatalf("GetAccessRequestDetail failed: %v", err)
	}

	if len(resp.AccessWindows) != 2 {
		t.Errorf("expected 2 access windows, got %d", len(resp.AccessWindows))
	}

	if resp.AccessWindows[0].OrbitalPlane != "45" {
		t.Errorf("expected orbital plane '45', got %q", resp.AccessWindows[0].OrbitalPlane)
	}
}

func TestFeasibilityService_DeleteAccessRequest(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodDelete)
		requirePath(t, r, "/ma/accessrequests/ar-123")
		requireAuth(t, r, "test-api-key")

		w.WriteHeader(http.StatusNoContent)
	}

	cli, _ := newTestClient(t, handler)
	feasibility := capella.NewFeasibilityService(cli)

	err := feasibility.DeleteAccessRequest(context.Background(), "ar-123")
	if err != nil {
		t.Fatalf("DeleteAccessRequest failed: %v", err)
	}
}

func TestFeasibilityService_WaitForAccessRequest(t *testing.T) {
	pollCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		pollCount++
		status := capella.ProcessingProcessing
		if pollCount >= 3 {
			status = capella.ProcessingCompleted
		}

		jsonResponse(w, http.StatusOK, capella.AccessRequestResponse{
			Properties: capella.AccessRequestPropertiesResponse{
				AccessRequestID:     "ar-123",
				ProcessingStatus:    status,
				AccessibilityStatus: capella.AccessibilityAccessible,
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	feasibility := capella.NewFeasibilityService(cli)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := feasibility.WaitForAccessRequest(ctx, "ar-123", 10*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForAccessRequest failed: %v", err)
	}

	if resp.Properties.ProcessingStatus != capella.ProcessingCompleted {
		t.Errorf("expected status 'completed', got %q", resp.Properties.ProcessingStatus)
	}

	if pollCount < 3 {
		t.Errorf("expected at least 3 polls, got %d", pollCount)
	}
}

func TestFeasibilityService_ListAccessRequests(t *testing.T) {
	page := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		page++

		var resp capella.AccessRequestsPagedResponse
		if page == 1 {
			resp = capella.AccessRequestsPagedResponse{
				Results: []capella.AccessRequestResponse{
					{Properties: capella.AccessRequestPropertiesResponse{AccessRequestID: "ar-1"}},
				},
				CurrentPage: 1,
				TotalPages:  2,
			}
		} else {
			resp = capella.AccessRequestsPagedResponse{
				Results: []capella.AccessRequestResponse{
					{Properties: capella.AccessRequestPropertiesResponse{AccessRequestID: "ar-2"}},
				},
				CurrentPage: 2,
				TotalPages:  2,
			}
		}

		jsonResponse(w, http.StatusOK, resp)
	}

	cli, _ := newTestClient(t, handler)
	feasibility := capella.NewFeasibilityService(cli)

	var requests []capella.AccessRequestResponse
	for req, err := range feasibility.ListAccessRequests(context.Background(), capella.ListAccessRequestsParams{}) {
		if err != nil {
			t.Fatalf("iterator error: %v", err)
		}
		requests = append(requests, req)
	}

	if len(requests) != 2 {
		t.Errorf("expected 2 requests, got %d", len(requests))
	}
}

func TestAccessRequestBuilder(t *testing.T) {
	open := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	close := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	req := capella.NewAccessRequestBuilder().
		Point(-71.097, 42.346).
		Name("Boston Feasibility").
		Description("Test access request").
		Window(open, close).
		LookDirection(capella.LookRight).
		OrbitDirection(capella.OrbitAscending).
		OffNadirRange(20.0, 35.0).
		Build()

	if req.Type != "Feature" {
		t.Errorf("expected type 'Feature', got %q", req.Type)
	}

	if req.Properties.AccessRequestName != "Boston Feasibility" {
		t.Errorf("expected name 'Boston Feasibility', got %q", req.Properties.AccessRequestName)
	}

	if req.Properties.AccessConstraints == nil {
		t.Fatal("expected access constraints, got nil")
	}

	if *req.Properties.AccessConstraints.LookDirection != capella.LookRight {
		t.Errorf("expected look direction 'right', got %q", *req.Properties.AccessConstraints.LookDirection)
	}

	if *req.Properties.AccessConstraints.OffNadirMin != 20.0 {
		t.Errorf("expected off-nadir min 20.0, got %f", *req.Properties.AccessConstraints.OffNadirMin)
	}
}

func TestAccessRequestBuilder_WithPolygon(t *testing.T) {
	coords := [][][]float64{{
		{-71.1, 42.3},
		{-71.0, 42.3},
		{-71.0, 42.4},
		{-71.1, 42.4},
		{-71.1, 42.3},
	}}

	req := capella.NewAccessRequestBuilder().
		Polygon(coords).
		Window(time.Now(), time.Now().Add(12*time.Hour)).
		Build()

	if req.Geometry == nil {
		t.Fatal("expected geometry, got nil")
	}

	if req.Geometry.Type != "Polygon" {
		t.Errorf("expected geometry type 'Polygon', got %q", req.Geometry.Type)
	}
}

func TestAccessRequestBuilder_WithBBox(t *testing.T) {
	bbox := capella.BoundingBox{-71.1, 42.3, -71.0, 42.4}

	req := capella.NewAccessRequestBuilder().
		BBox(bbox).
		Window(time.Now(), time.Now().Add(12*time.Hour)).
		Build()

	if req.Geometry == nil {
		t.Fatal("expected geometry, got nil")
	}

	if req.Geometry.Type != "Polygon" {
		t.Errorf("expected geometry type 'Polygon', got %q", req.Geometry.Type)
	}
}
