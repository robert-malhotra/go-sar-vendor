package umbra_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/umbra"
)

func TestCreateTask(t *testing.T) {
	expectedTask := umbra.Task{
		ID:          "task-123",
		TaskName:    "Test Task",
		Status:      umbra.TaskStatusReceived,
		ImagingMode: umbra.ImagingModeSpotlight,
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/tasking/tasks")
		requireAuth(t, r, "test-token")
		requireContentType(t, r, "application/json")
		jsonResponse(w, http.StatusCreated, expectedTask)
	})

	req := &umbra.CreateTaskRequest{
		TaskName:    "Test Task",
		ImagingMode: umbra.ImagingModeSpotlight,
		SpotlightConstraints: &umbra.SpotlightConstraints{
			Geometry: umbra.NewPointGeometry(-122.4194, 37.7749),
		},
		WindowStartAt: time.Now().Add(24 * time.Hour),
		WindowEndAt:   time.Now().Add(48 * time.Hour),
	}

	task, err := cli.CreateTask(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != expectedTask.ID {
		t.Errorf("expected task ID %s, got %s", expectedTask.ID, task.ID)
	}
	if task.Status != expectedTask.Status {
		t.Errorf("expected task status %s, got %s", expectedTask.Status, task.Status)
	}
}

func TestGetTask(t *testing.T) {
	expectedTask := umbra.Task{
		ID:          "task-456",
		TaskName:    "Test Task",
		Status:      umbra.TaskStatusActive,
		ImagingMode: umbra.ImagingModeSpotlight,
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/tasking/tasks/task-456")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expectedTask)
	})

	task, err := cli.GetTask(context.Background(), "task-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != expectedTask.ID {
		t.Errorf("expected task ID %s, got %s", expectedTask.ID, task.ID)
	}
	if task.Status != expectedTask.Status {
		t.Errorf("expected task status %s, got %s", expectedTask.Status, task.Status)
	}
}

func TestListTasks(t *testing.T) {
	expectedResp := umbra.TaskListResponse{
		Tasks: []umbra.Task{
			{ID: "task-1", Status: umbra.TaskStatusActive},
			{ID: "task-2", Status: umbra.TaskStatusScheduled},
		},
		TotalCount: 2,
		Limit:      10,
		Offset:     0,
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/tasking/tasks")
		requireAuth(t, r, "test-token")

		// Verify query parameters
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Errorf("expected limit=10, got %s", got)
		}
		if got := r.URL.Query().Get("status"); got != "ACTIVE" {
			t.Errorf("expected status=ACTIVE, got %s", got)
		}

		jsonResponse(w, http.StatusOK, expectedResp)
	})

	opts := &umbra.ListTasksOptions{
		Limit:  10,
		Status: []umbra.TaskStatus{umbra.TaskStatusActive},
	}
	resp, err := cli.ListTasks(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(resp.Tasks))
	}
	if resp.TotalCount != 2 {
		t.Errorf("expected total count 2, got %d", resp.TotalCount)
	}
}

func TestCancelTask(t *testing.T) {
	expectedTask := umbra.Task{
		ID:     "task-789",
		Status: umbra.TaskStatusCancelRequested,
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPatch)
		requirePath(t, r, "/tasking/tasks/task-789/cancel")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expectedTask)
	})

	task, err := cli.CancelTask(context.Background(), "task-789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != umbra.TaskStatusCancelRequested {
		t.Errorf("expected status %s, got %s", umbra.TaskStatusCancelRequested, task.Status)
	}
}

func TestTaskStatusIsTerminal(t *testing.T) {
	tests := []struct {
		status   umbra.TaskStatus
		terminal bool
	}{
		{umbra.TaskStatusReceived, false},
		{umbra.TaskStatusActive, false},
		{umbra.TaskStatusScheduled, false},
		{umbra.TaskStatusDelivered, true},
		{umbra.TaskStatusCanceled, true},
		{umbra.TaskStatusRejected, true},
		{umbra.TaskStatusError, true},
		{umbra.TaskStatusAnomaly, true},
		{umbra.TaskStatusCompleted, true},
		{umbra.TaskStatusExpired, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.terminal {
				t.Errorf("TaskStatus(%s).IsTerminal() = %v, want %v", tt.status, got, tt.terminal)
			}
		})
	}
}

func TestCreateTask_Error(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusBadRequest, "Invalid request")
	})

	req := &umbra.CreateTaskRequest{
		TaskName:    "Test Task",
		ImagingMode: umbra.ImagingModeSpotlight,
	}

	_, err := cli.CreateTask(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !umbra.IsBadRequest(err) {
		t.Errorf("expected bad request error, got %v", err)
	}
}

func TestGetTask_NotFound(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusNotFound, "Task not found")
	})

	_, err := cli.GetTask(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !umbra.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestGetTask_Unauthorized(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusUnauthorized, "Invalid token")
	})

	_, err := cli.GetTask(context.Background(), "task-123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !umbra.IsUnauthorized(err) {
		t.Errorf("expected unauthorized error, got %v", err)
	}
}

func TestSearchTasks(t *testing.T) {
	expected := []umbra.Task{
		{ID: "task-1", Status: umbra.TaskStatusActive},
		{ID: "task-2", Status: umbra.TaskStatusScheduled},
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/tasking/tasks/search")
		requireAuth(t, r, "test-token")
		requireContentType(t, r, "application/json")
		jsonResponse(w, http.StatusOK, expected)
	})

	limit := 10
	req := umbra.TaskSearchRequest{
		Limit: &limit,
		Query: map[string]interface{}{"status": []string{"ACTIVE", "SCHEDULED"}},
	}

	var tasks []umbra.Task
	for task, err := range cli.SearchTasks(context.Background(), req) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tasks = append(tasks, task)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}
