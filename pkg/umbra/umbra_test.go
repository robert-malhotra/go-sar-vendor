package umbra_test

import (
	"encoding/json"
	"iter"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/robert.malhotra/go-sar-vendor/pkg/umbra"
)

// ---------------------- helpers ----------------------
func setupMockServer(t *testing.T, expectedMethod, expectedPath string, expectedStatus int, response any) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectedMethod {
			t.Fatalf("Expected method %s, got %s", expectedMethod, r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, expectedPath) {
			t.Fatalf("Expected path suffix %s, got %s", expectedPath, r.URL.Path)
		}
		w.WriteHeader(expectedStatus)
		_ = json.NewEncoder(w).Encode(response)
	}))
}

func setupErrorServer(t *testing.T, status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

// generic helper to drain iter.Seq2 into slice
func seqToSlice[T any](seq iter.Seq2[T, error]) ([]T, error) {
	var out []T
	for v, err := range seq {
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// ---------------------- task -------------------------
func TestCreateTask(t *testing.T) {
	mock := umbra.Task{ID: "task123", Status: umbra.TaskActive}
	srv := setupMockServer(t, http.MethodPost, "/tasking/tasks", http.StatusCreated, mock)
	defer srv.Close()

	c, _ := umbra.NewClient(srv.URL)
	got, err := c.CreateTask("tok", &umbra.TaskRequest{})
	if err != nil || got.ID != "task123" {
		t.Fatalf("unexpected: %+v %v", got, err)
	}
}

func TestGetTask(t *testing.T) {
	mock := umbra.Task{ID: "t-123", Status: umbra.TaskScheduled}
	srv := setupMockServer(t, http.MethodGet, "/tasking/tasks/t-123", http.StatusOK, mock)
	defer srv.Close()

	c, _ := umbra.NewClient(srv.URL)
	got, err := c.GetTask("tok", "t-123")
	if err != nil || got.Status != umbra.TaskScheduled {
		t.Fatalf("mismatch: %+v %v", got, err)
	}
}

func TestCancelTask(t *testing.T) {
	mock := umbra.Task{ID: "t-123", Status: umbra.TaskCanceled}
	srv := setupMockServer(t, http.MethodPatch, "/tasking/tasks/t-123/cancel", http.StatusOK, mock)
	defer srv.Close()

	c, _ := umbra.NewClient(srv.URL)
	got, err := c.CancelTask("tok", "t-123")
	if err != nil || got.Status != umbra.TaskCanceled {
		t.Fatalf("cancel failed: %+v %v", got, err)
	}
}

// ---------------------- collect ----------------------
func TestGetCollect(t *testing.T) {
	mock := umbra.Collect{ID: "col123", Status: "DELIVERED"}
	srv := setupMockServer(t, http.MethodGet, "/tasking/collects/col123", http.StatusOK, mock)
	defer srv.Close()

	c, _ := umbra.NewClient(srv.URL)
	got, err := c.GetCollect("tok", "col123")
	if err != nil || got.ID != "col123" {
		t.Fatalf("unexpected: %+v %v", got, err)
	}
}

func TestSearchCollects(t *testing.T) {
	mock := []umbra.Collect{{ID: "c1", Status: "DELIVERED"}, {ID: "c2", Status: "PROCESSING"}}
	srv := setupMockServer(t, http.MethodPost, "/tasking/collects/search", http.StatusOK, mock)
	defer srv.Close()

	c, _ := umbra.NewClient(srv.URL)
	limit, skip := 10, 0
	req := umbra.CollectSearchRequest{Limit: &limit, Skip: &skip, Query: map[string]any{"taskIds": []string{"t-123"}}}
	seq := c.SearchCollects("tok", req)

	out, err := seqToSlice(seq)
	if err != nil || len(out) != 2 {
		t.Fatalf("search failed: %+v %v", out, err)
	}
}

// ---------------------- task search ------------------
func TestSearchTasks(t *testing.T) {
	mock := []umbra.Task{{ID: "t1", Status: umbra.TaskActive}, {ID: "t2", Status: umbra.TaskSubmitted}}
	srv := setupMockServer(t, http.MethodPost, "/tasking/tasks/search", http.StatusOK, mock)
	defer srv.Close()

	c, _ := umbra.NewClient(srv.URL)
	limit, skip := 5, 0
	req := umbra.TaskSearchRequest{Limit: &limit, Skip: &skip, Query: map[string]any{"status": []string{"ACTIVE"}}}
	seq := c.SearchTasks("tok", req)

	out, err := seqToSlice(seq)
	if err != nil || len(out) != 2 {
		t.Fatalf("search tasks failed: %+v %v", out, err)
	}
}

// ---------------------- feasibility ------------------
func TestCreateFeasibility(t *testing.T) {
	req := &umbra.TaskingRequest{
		ImagingMode: umbra.SPOTLIGHT_MODE,
		SpotlightConstraints: umbra.SpotlightConstraints{
			Geometry: umbra.PointGeometry{Type: "Point", Coordinates: []float64{12.34, 56.78}},
		},
		WindowStartAt: time.Now(),
		WindowEndAt:   time.Now().Add(30 * time.Minute),
	}

	mock := umbra.FeasibilityResponse{ID: "mock-id", Status: "CREATED", Request: *req}
	srv := setupMockServer(t, http.MethodPost, "/tasking/feasibilities", http.StatusCreated, mock)
	defer srv.Close()

	c, _ := umbra.NewClient(srv.URL)
	got, err := c.CreateFeasibility("tok", req)
	if err != nil || got.ID != "mock-id" {
		t.Fatalf("CreateFeasibility failed: %+v %v", got, err)
	}
}

func TestGetFeasibility(t *testing.T) {
	mock := umbra.FeasibilityResponse{ID: "mock-id", Status: "PROCESSING"}
	srv := setupMockServer(t, http.MethodGet, "/tasking/feasibilities/mock-id", http.StatusOK, mock)
	defer srv.Close()

	c, _ := umbra.NewClient(srv.URL)
	got, err := c.GetFeasibility("tok", "mock-id")
	if err != nil || got.Status != "PROCESSING" {
		t.Fatalf("GetFeasibility failed: %+v %v", got, err)
	}
}

// ---------------------- error path -------------------
func TestCreateTaskError(t *testing.T) {
	srv := setupErrorServer(t, http.StatusInternalServerError, "boom")
	defer srv.Close()

	c, _ := umbra.NewClient(srv.URL)
	_, err := c.CreateTask("tok", &umbra.TaskRequest{})
	if err == nil || !strings.Contains(err.Error(), "unexpected status") {
		t.Fatalf("expected error, got %v", err)
	}
}
