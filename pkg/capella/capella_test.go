package capella

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

// helper to spin up a mock HTTP server and a client wired to it.
func newMockClient(handler http.HandlerFunc) (*Client, func()) {
	srv := httptest.NewServer(handler)
	cli := NewClient(WithBaseURL(srv.URL + "/")) // keep trailing slash so path.Join behaves
	return cli, srv.Close
}

// -----------------------------------------------------------------------------
// Access‑request tests
// -----------------------------------------------------------------------------

func TestCreateAccessRequest_Success(t *testing.T) {
	respBody := `{
        "type": "Feature",
        "geometry": {"type": "Point", "coordinates": [10.0, 20.0]},
        "properties": {
            "accessrequestId": "ar-123",
            "orgId": "org-1",
            "userId": "u-1",
            "windowOpen": "2025-06-15T00:00:00Z",
            "windowClose": "2025-06-15T12:00:00Z",
            "processingStatus": "queued",
            "accessibilityStatus": "unknown"
        }
    }`

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/ma/accessrequests" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "ApiKey test-key" {
			t.Fatalf("missing / wrong auth header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(respBody))
	}

	cli, closeFn := newMockClient(handler)
	defer closeFn()

	req := AccessRequest{
		Type:     "Feature",
		Geometry: GeoJSONGeometry{Type: Point, Coordinates: []float64{10, 20}},
		Properties: AccessRequestProperties{
			OrgID:       "org-1",
			UserID:      "u-1",
			WindowOpen:  time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
			WindowClose: time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC),
		},
	}

	got, err := cli.CreateAccessRequest("test-key", req)
	if err != nil {
		t.Fatalf("CreateAccessRequest returned error: %v", err)
	}
	if got.Properties.AccessRequestID != "ar-123" {
		t.Errorf("unexpected accessrequestId: %s", got.Properties.AccessRequestID)
	}
	if got.Properties.ProcessingStatus != Queued {
		t.Errorf("unexpected processingStatus: %v", got.Properties.ProcessingStatus)
	}
}

func TestCreateAccessRequest_ValidationError(t *testing.T) {
	errBody := `{"detail":[{"loc":["body","properties","windowOpen"],"msg":"field required","type":"value_error.missing"}]}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(errBody))
	}

	cli, closeFn := newMockClient(handler)
	defer closeFn()

	var req AccessRequest
	_, err := cli.CreateAccessRequest("key", req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("unexpected status code: %d", apiErr.StatusCode)
	}
	if apiErr.Validation == nil || len(apiErr.Validation.Detail) == 0 {
		t.Fatalf("expected validation details, got %+v", apiErr.Validation)
	}
	if msg := apiErr.Validation.Detail[0].Msg; msg != "field required" {
		t.Errorf("unexpected validation msg: %q", msg)
	}
}

func TestGetAccessRequest_Success(t *testing.T) {
	respBody := `{
        "type": "Feature",
        "geometry": {"type": "Point", "coordinates": [10.0, 20.0]},
        "properties": {
            "accessrequestId": "ar-123",
            "orgId": "org-1",
            "userId": "u-1",
            "windowOpen": "2025-06-15T00:00:00Z",
            "windowClose": "2025-06-15T12:00:00Z",
            "processingStatus": "completed",
            "accessibilityStatus": "accessible"
        }
    }`

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/ma/accessrequests/ar-123" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "ApiKey test-key" {
			t.Fatalf("missing / wrong auth header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(respBody))
	}

	cli, closeFn := newMockClient(handler)
	defer closeFn()

	got, err := cli.GetAccessRequest("test-key", "ar-123")
	if err != nil {
		t.Fatalf("GetAccessRequest returned error: %v", err)
	}
	if got.Properties.AccessRequestID != "ar-123" {
		t.Errorf("unexpected accessrequestId: %s", got.Properties.AccessRequestID)
	}
	if got.Properties.ProcessingStatus != Completed {
		t.Errorf("unexpected processingStatus: %v", got.Properties.ProcessingStatus)
	}
	if got.Geometry.Type != Point {
		t.Errorf("unexpected geometry type: %s", got.Geometry.Type)
	}
}

// -----------------------------------------------------------------------------
// Tasking‑request tests
// -----------------------------------------------------------------------------

func TestCreateTaskingRequest_Success(t *testing.T) {
	respBody := `{
        "type": "Feature",
        "geometry": {"type": "Point", "coordinates": [0,0]},
        "properties": {
            "taskingrequestId": "tr-1",
            "orgId": "org-1",
            "userId": "u-1",
            "windowOpen": "2025-06-20T00:00:00Z",
            "windowClose": "2025-06-21T00:00:00Z",
            "collectionTier": "urgent",
            "collectionType": "spotlight",
            "processingStatus": "queued"
        }
    }`

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/task" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "\"collectionTier\":\"urgent\"") {
			t.Fatalf("request body missing collectionTier: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(respBody))
	}

	cli, closeFn := newMockClient(handler)
	defer closeFn()

	tr := TaskingRequest{
		Type:     "Feature",
		Geometry: GeoJSONGeometry{Type: Point, Coordinates: []float64{0, 0}},
		Properties: TaskingRequestProperties{
			OrgID:          "org-1",
			UserID:         "u-1",
			WindowOpen:     time.Date(2025, 6, 20, 0, 0, 0, 0, time.UTC),
			WindowClose:    time.Date(2025, 6, 21, 0, 0, 0, 0, time.UTC),
			CollectionTier: TierUrgent,
			CollectionType: "spotlight",
		},
	}

	got, err := cli.CreateTaskingRequest("key", tr)
	if err != nil {
		t.Fatalf("CreateTaskingRequest error: %v", err)
	}
	if got.Properties.TaskingRequestID != "tr-1" {
		t.Errorf("unexpected taskingrequestId: %s", got.Properties.TaskingRequestID)
	}
	if got.Properties.CollectionTier != TierUrgent {
		t.Errorf("unexpected collectionTier: %s", got.Properties.CollectionTier)
	}
}

func TestApproveTaskingRequest_Success(t *testing.T) {
	respBody := `{"properties":{"taskingrequestId":"tr-1","processingStatus":"approved"}}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/task/tr-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		// ensure body contains status approved
		body, _ := io.ReadAll(r.Body)
		if !bytes.Contains(body, []byte("\"approved\"")) {
			t.Fatalf("patch body incorrect: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(respBody))
	}

	cli, closeFn := newMockClient(handler)
	defer closeFn()

	got, err := cli.ApproveTaskingRequest("key", "tr-1")
	if err != nil {
		t.Fatalf("ApproveTaskingRequest error: %v", err)
	}
	if got.Properties.ProcessingStatus != "approved" {
		t.Errorf("unexpected status: %v", got.Properties.ProcessingStatus)
	}
}

func TestSearchTasks_Success(t *testing.T) {
	respBody := `{"results":[],"currentPage":1,"totalPages":1}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/tasks/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "\"query\"") {
			t.Fatalf("missing query in body: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(respBody))
	}

	cli, closeFn := newMockClient(handler)
	defer closeFn()

	resp, err := cli.SearchTasks("key", TaskSearchRequest{Query: map[string]any{"status": "queued"}})
	if err != nil {
		t.Fatalf("SearchTasks error: %v", err)
	}
	if resp.TotalPages != 1 {
		t.Errorf("unexpected totalPages: %d", resp.TotalPages)
	}
}

// -----------------------------------------------------------------------------
// Pagination iterator edge‑cases
// -----------------------------------------------------------------------------

func TestListTasksPaged_ErrorPropagates(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}

	cli, closeFn := newMockClient(handler)
	defer closeFn()

	seq := cli.ListTasksPaged("k", PagedTasksParams{})
	for _, err := range seq {
		// first yield should contain error and break loop implicitly
		if err == nil {
			t.Fatal("expected error from iterator, got nil")
		}
		return // success – we saw an error
	}
	t.Fatal("iterator produced no values")
}

// -----------------------------------------------------------------------------
// Pagination / iterator happy‑path test retained from earlier (now at bottom)
// -----------------------------------------------------------------------------

func TestListTasksPaged_IteratesAcrossPages(t *testing.T) {
	page1 := TaskingRequestsPagedResponse{
		Results: []TaskingRequestResponse{
			{Properties: TaskingRequestPropertiesResponse{TaskingRequestID: "task-1", ProcessingStatus: Queued}},
		},
		CurrentPage: 1,
		TotalPages:  2,
	}
	page2 := TaskingRequestsPagedResponse{
		Results: []TaskingRequestResponse{
			{Properties: TaskingRequestPropertiesResponse{TaskingRequestID: "task-2", ProcessingStatus: Queued}},
		},
		CurrentPage: 2,
		TotalPages:  2,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			json.NewEncoder(w).Encode(page2)
		} else {
			json.NewEncoder(w).Encode(page1)
		}
	}

	cli, closeFn := newMockClient(handler)
	defer closeFn()

	seq := cli.ListTasksPaged("api-key", PagedTasksParams{})
	var gotIDs []string
	for task, err := range seq {
		if err != nil {
			t.Fatalf("iterator yielded error: %v", err)
		}
		gotIDs = append(gotIDs, task.Properties.TaskingRequestID)
	}

	want := []string{"task-1", "task-2"}
	if !reflect.DeepEqual(gotIDs, want) {
		t.Errorf("unexpected task IDs: got %v, want %v", gotIDs, want)
	}
}
