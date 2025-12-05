package capella_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/capella"
)

func TestTaskingService_CreateTask(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/task")
		requireAuth(t, r, "test-api-key")
		requireContentType(t, r, "application/json")

		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"collectionTier":"urgent"`) {
			t.Errorf("expected collectionTier 'urgent' in body, got: %s", body)
		}

		jsonResponse(w, http.StatusCreated, map[string]any{
			"type": "Feature",
			"geometry": map[string]any{
				"type":        "Point",
				"coordinates": []float64{-71.097, 42.346},
			},
			"properties": map[string]any{
				"taskingrequestName": "Boston Downtown",
				"collectionTier":     "urgent",
				"collectionType":     "spotlight",
				"taskingrequestId":   "tr-123",
				"status":             "received",
				"processingStatus":   "queued",
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	tasking := capella.NewTaskingService(cli)

	req := capella.TaskingRequest{
		Type:     "Feature",
		Geometry: capella.Point(-71.097, 42.346),
		Properties: capella.TaskingRequestProperties{
			TaskingRequestName: "Boston Downtown",
			WindowOpen:         time.Now(),
			WindowClose:        time.Now().Add(7 * 24 * time.Hour),
			CollectionTier:     capella.TierUrgent,
			CollectionType:     capella.CollectionSpotlight,
		},
	}

	resp, err := tasking.CreateTask(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if resp.Properties.TaskingRequestID != "tr-123" {
		t.Errorf("expected task ID 'tr-123', got %q", resp.Properties.TaskingRequestID)
	}

	if resp.Properties.CollectionTier != capella.TierUrgent {
		t.Errorf("expected tier 'urgent', got %q", resp.Properties.CollectionTier)
	}
}

func TestTaskingService_GetTask(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/task/tr-123")
		requireAuth(t, r, "test-api-key")

		jsonResponse(w, http.StatusOK, capella.TaskingRequestResponse{
			Properties: capella.TaskingRequestPropertiesResponse{
				TaskingRequestID: "tr-123",
				Status:           capella.TaskActive,
				ProcessingStatus: capella.ProcessingProcessing,
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	tasking := capella.NewTaskingService(cli)

	resp, err := tasking.GetTask(context.Background(), "tr-123")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if resp.Properties.TaskingRequestID != "tr-123" {
		t.Errorf("expected task ID 'tr-123', got %q", resp.Properties.TaskingRequestID)
	}
}

func TestTaskingService_ApproveTask(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPatch)
		requirePath(t, r, "/task/tr-123")

		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"status":"approved"`) {
			t.Errorf("expected status 'approved' in body, got: %s", body)
		}

		jsonResponse(w, http.StatusOK, capella.TaskingRequestResponse{
			Properties: capella.TaskingRequestPropertiesResponse{
				TaskingRequestID: "tr-123",
				Status:           capella.TaskApproved,
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	tasking := capella.NewTaskingService(cli)

	resp, err := tasking.ApproveTask(context.Background(), "tr-123")
	if err != nil {
		t.Fatalf("ApproveTask failed: %v", err)
	}

	if resp.Properties.Status != capella.TaskApproved {
		t.Errorf("expected status 'approved', got %q", resp.Properties.Status)
	}
}

func TestTaskingService_CancelTask(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPatch)
		requirePath(t, r, "/task/tr-123")

		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"status":"canceled"`) {
			t.Errorf("expected status 'canceled' in body, got: %s", body)
		}

		jsonResponse(w, http.StatusOK, capella.TaskingRequestResponse{
			Properties: capella.TaskingRequestPropertiesResponse{
				TaskingRequestID: "tr-123",
				Status:           capella.TaskCanceled,
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	tasking := capella.NewTaskingService(cli)

	resp, err := tasking.CancelTask(context.Background(), "tr-123")
	if err != nil {
		t.Fatalf("CancelTask failed: %v", err)
	}

	if resp.Properties.Status != capella.TaskCanceled {
		t.Errorf("expected status 'canceled', got %q", resp.Properties.Status)
	}
}

func TestTaskingService_SearchTasks(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/tasks/search")
		requireContentType(t, r, "application/json")

		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"query"`) {
			t.Errorf("expected query in body, got: %s", body)
		}

		jsonResponse(w, http.StatusOK, capella.TaskingRequestsPagedResponse{
			Results: []capella.TaskingRequestResponse{
				{Properties: capella.TaskingRequestPropertiesResponse{TaskingRequestID: "tr-1"}},
				{Properties: capella.TaskingRequestPropertiesResponse{TaskingRequestID: "tr-2"}},
			},
			CurrentPage: 1,
			TotalPages:  1,
		})
	}

	cli, _ := newTestClient(t, handler)
	tasking := capella.NewTaskingService(cli)

	resp, err := tasking.SearchTasks(context.Background(), capella.TaskSearchRequest{
		Query: map[string]any{"status": "active"},
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("SearchTasks failed: %v", err)
	}

	if len(resp.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(resp.Results))
	}
}

func TestTaskingService_ListTasks_Pagination(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)

		page := r.URL.Query().Get("page")
		var resp capella.TaskingRequestsPagedResponse

		if page == "2" {
			resp = capella.TaskingRequestsPagedResponse{
				Results: []capella.TaskingRequestResponse{
					{Properties: capella.TaskingRequestPropertiesResponse{TaskingRequestID: "tr-2"}},
				},
				CurrentPage: 2,
				TotalPages:  2,
			}
		} else {
			resp = capella.TaskingRequestsPagedResponse{
				Results: []capella.TaskingRequestResponse{
					{Properties: capella.TaskingRequestPropertiesResponse{TaskingRequestID: "tr-1"}},
				},
				CurrentPage: 1,
				TotalPages:  2,
			}
		}

		jsonResponse(w, http.StatusOK, resp)
	}

	cli, _ := newTestClient(t, handler)
	tasking := capella.NewTaskingService(cli)

	var tasks []capella.TaskingRequestResponse
	for task, err := range tasking.ListTasks(context.Background(), capella.ListTasksParams{}) {
		if err != nil {
			t.Fatalf("iterator error: %v", err)
		}
		tasks = append(tasks, task)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}

	if tasks[0].Properties.TaskingRequestID != "tr-1" {
		t.Errorf("expected first task 'tr-1', got %q", tasks[0].Properties.TaskingRequestID)
	}

	if tasks[1].Properties.TaskingRequestID != "tr-2" {
		t.Errorf("expected second task 'tr-2', got %q", tasks[1].Properties.TaskingRequestID)
	}
}

func TestTaskingService_ListTasks_ErrorPropagates(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}

	cli, _ := newTestClient(t, handler)
	tasking := capella.NewTaskingService(cli)

	for _, err := range tasking.ListTasks(context.Background(), capella.ListTasksParams{}) {
		if err == nil {
			t.Fatal("expected error from iterator, got nil")
		}
		return // success - we got an error
	}
	t.Fatal("iterator produced no values")
}

func TestTaskingService_GetCollectionTypes(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/collectiontypes")

		jsonResponse(w, http.StatusOK, []capella.CollectionTypeInfo{
			{ID: "spotlight", Name: "Spotlight", Resolution: 0.5},
			{ID: "stripmap_100", Name: "Stripmap 100", Resolution: 1.0},
		})
	}

	cli, _ := newTestClient(t, handler)
	tasking := capella.NewTaskingService(cli)

	types, err := tasking.GetCollectionTypes(context.Background())
	if err != nil {
		t.Fatalf("GetCollectionTypes failed: %v", err)
	}

	if len(types) != 2 {
		t.Errorf("expected 2 collection types, got %d", len(types))
	}
}

func TestTaskingService_CreateRepeatRequest(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/repeat-requests")

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)

		props := req["properties"].(map[string]any)
		if props["repeatInterval"] != float64(7) {
			t.Errorf("expected repeatInterval 7, got %v", props["repeatInterval"])
		}

		jsonResponse(w, http.StatusCreated, capella.RepeatRequestResponse{
			Properties: capella.RepeatRequestPropertiesResponse{
				RepeatRequestProperties: capella.RepeatRequestProperties{
					RepeatInterval: 7,
				},
				RepeatRequestID: "rr-123",
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	tasking := capella.NewTaskingService(cli)

	req := capella.RepeatRequest{
		Geometry: capella.Point(-71.097, 42.346),
		Properties: capella.RepeatRequestProperties{
			WindowOpen:     time.Now(),
			WindowClose:    time.Now().Add(30 * 24 * time.Hour),
			RepeatInterval: 7,
			CollectionTier: capella.TierStandard,
			CollectionType: capella.CollectionSpotlight,
		},
	}

	resp, err := tasking.CreateRepeatRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateRepeatRequest failed: %v", err)
	}

	if resp.Properties.RepeatRequestID != "rr-123" {
		t.Errorf("expected repeat request ID 'rr-123', got %q", resp.Properties.RepeatRequestID)
	}
}

func TestTaskingRequestBuilder(t *testing.T) {
	open := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	close := time.Date(2024, 6, 8, 0, 0, 0, 0, time.UTC)

	preApproval := true
	req := capella.NewTaskingRequestBuilder().
		Point(-71.097, 42.346).
		Name("Boston Downtown").
		Description("Test tasking request").
		Window(open, close).
		Tier(capella.TierStandard).
		Type(capella.CollectionSpotlight).
		PreApproval(preApproval).
		Products(capella.ProductGEO, capella.ProductSLC).
		Build()

	if req.Type != "Feature" {
		t.Errorf("expected type 'Feature', got %q", req.Type)
	}

	if req.Properties.TaskingRequestName != "Boston Downtown" {
		t.Errorf("expected name 'Boston Downtown', got %q", req.Properties.TaskingRequestName)
	}

	if req.Properties.CollectionTier != capella.TierStandard {
		t.Errorf("expected tier 'standard', got %q", req.Properties.CollectionTier)
	}

	if req.Properties.PreApproval == nil || *req.Properties.PreApproval != true {
		t.Error("expected preApproval true")
	}

	if req.Properties.ProcessingConfig == nil || len(req.Properties.ProcessingConfig.ProductTypes) != 2 {
		t.Error("expected 2 product types")
	}
}
