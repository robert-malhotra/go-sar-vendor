package umbra_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/robert.malhotra/go-sar-vendor/pkg/umbra"
)

func TestGetCollect(t *testing.T) {
	expected := umbra.Collect{
		ID:          "col-123",
		TaskID:      "task-456",
		Status:      umbra.CollectStatusDelivered,
		SatelliteID: "umbra-04",
		CreatedAt:   time.Now(),
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/tasking/collects/col-123")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	col, err := cli.GetCollect(context.Background(), "col-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.ID != expected.ID {
		t.Errorf("expected ID %s, got %s", expected.ID, col.ID)
	}
	if col.Status != expected.Status {
		t.Errorf("expected status %s, got %s", expected.Status, col.Status)
	}
}

func TestListCollects(t *testing.T) {
	expected := umbra.CollectListResponse{
		Collects: []umbra.Collect{
			{ID: "col-1", TaskID: "task-1", Status: umbra.CollectStatusDelivered},
			{ID: "col-2", TaskID: "task-1", Status: umbra.CollectStatusProcessing},
		},
		TotalCount: 2,
		Limit:      10,
		Offset:     0,
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/tasking/collects")
		requireAuth(t, r, "test-token")

		// Verify query parameters
		if got := r.URL.Query().Get("taskId"); got != "task-1" {
			t.Errorf("expected taskId=task-1, got %s", got)
		}
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Errorf("expected limit=10, got %s", got)
		}

		jsonResponse(w, http.StatusOK, expected)
	})

	opts := &umbra.ListCollectsOptions{
		TaskID: "task-1",
		Limit:  10,
	}
	resp, err := cli.ListCollects(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Collects) != 2 {
		t.Errorf("expected 2 collects, got %d", len(resp.Collects))
	}
}

func TestListCollects_WithStatus(t *testing.T) {
	expected := umbra.CollectListResponse{
		Collects: []umbra.Collect{
			{ID: "col-1", Status: umbra.CollectStatusDelivered},
		},
		TotalCount: 1,
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/tasking/collects")

		// Verify status filter
		if got := r.URL.Query().Get("status"); got != "DELIVERED" {
			t.Errorf("expected status=DELIVERED, got %s", got)
		}

		jsonResponse(w, http.StatusOK, expected)
	})

	opts := &umbra.ListCollectsOptions{
		Status: []umbra.CollectStatus{umbra.CollectStatusDelivered},
	}
	resp, err := cli.ListCollects(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Collects) != 1 {
		t.Errorf("expected 1 collect, got %d", len(resp.Collects))
	}
}

func TestGetProductConstraints(t *testing.T) {
	expected := []umbra.ProductConstraint{
		{
			ProductType:       "GEC",
			SceneSize:         "5km",
			MinGrazingDegrees: 25.0,
			MaxGrazingDegrees: 70.0,
			RecommendedLooks:  1,
		},
		{
			ProductType:       "SICD",
			SceneSize:         "5km",
			MinGrazingDegrees: 25.0,
			MaxGrazingDegrees: 70.0,
			RecommendedLooks:  1,
		},
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/tasking/products/SPOTLIGHT/constraints")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	constraints, err := cli.GetProductConstraints(context.Background(), umbra.ImagingModeSpotlight)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(constraints) != 2 {
		t.Errorf("expected 2 constraints, got %d", len(constraints))
	}
	if constraints[0].ProductType != "GEC" {
		t.Errorf("expected product type GEC, got %s", constraints[0].ProductType)
	}
}

func TestGetCollect_NotFound(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusNotFound, "Collect not found")
	})

	_, err := cli.GetCollect(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !umbra.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestSearchCollects(t *testing.T) {
	expected := []umbra.Collect{
		{ID: "col-1", TaskID: "task-1", Status: umbra.CollectStatusDelivered},
		{ID: "col-2", TaskID: "task-1", Status: umbra.CollectStatusProcessing},
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/tasking/collects/search")
		requireAuth(t, r, "test-token")
		requireContentType(t, r, "application/json")
		jsonResponse(w, http.StatusOK, expected)
	})

	limit := 10
	req := umbra.CollectSearchRequest{
		Limit: &limit,
		Query: map[string]interface{}{"taskIds": []string{"task-1"}},
	}

	var collects []umbra.Collect
	for col, err := range cli.SearchCollects(context.Background(), req) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		collects = append(collects, col)
	}
	if len(collects) != 2 {
		t.Errorf("expected 2 collects, got %d", len(collects))
	}
}

func TestCollectStatusIsTerminal(t *testing.T) {
	tests := []struct {
		status   umbra.CollectStatus
		terminal bool
	}{
		{umbra.CollectStatusScheduled, false},
		{umbra.CollectStatusProcessing, false},
		{umbra.CollectStatusDelivered, false},
		{umbra.CollectStatusCanceled, true},
		{umbra.CollectStatusFailed, true},
		{umbra.CollectStatusCorrupt, true},
		{umbra.CollectStatusSuperseded, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.terminal {
				t.Errorf("CollectStatus(%s).IsTerminal() = %v, want %v", tt.status, got, tt.terminal)
			}
		})
	}
}
