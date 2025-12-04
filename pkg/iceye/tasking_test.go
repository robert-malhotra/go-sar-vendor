package iceye_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/robert.malhotra/go-sar-vendor/pkg/iceye"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTasksPagination(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/tasking/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
			cursor := r.URL.Query().Get("cursor")
			switch cursor {
			case "":
				json.NewEncoder(w).Encode(map[string]any{
					"data":   []iceye.Task{{ID: "t1"}},
					"cursor": "next",
				})
			case "next":
				json.NewEncoder(w).Encode(map[string]any{
					"data":   []iceye.Task{{ID: "t2"}},
					"cursor": nil,
				})
			default:
				t.Fatalf("unexpected cursor: %s", cursor)
			}
		})
	})

	var count int
	for tasks, err := range cli.ListTasks(context.Background(), 1, nil) {
		require.NoError(t, err)
		count += len(tasks)
	}

	assert.Equal(t, 2, count)
}

func TestListTasksWithFilters(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/tasking/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "C-123", r.URL.Query().Get("contractID"))
			json.NewEncoder(w).Encode(map[string]any{
				"data":   []iceye.Task{{ID: "t1", Status: iceye.TaskStatusActive}},
				"cursor": nil,
			})
		})
	})

	opts := &iceye.ListTasksOptions{
		ContractID: "C-123",
	}

	for tasks, err := range cli.ListTasks(context.Background(), 10, opts) {
		require.NoError(t, err)
		require.Len(t, tasks, 1)
		assert.Equal(t, iceye.TaskStatusActive, tasks[0].Status)
	}
}

func TestCreateTask(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/tasking/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)

			var req iceye.CreateTaskRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, "C-1", req.ContractID)
			assert.Equal(t, "SPOTLIGHT", req.ImagingMode)

			json.NewEncoder(w).Encode(iceye.Task{
				ID:          "T-123",
				ContractID:  req.ContractID,
				ImagingMode: req.ImagingMode,
				Status:      iceye.TaskStatusReceived,
			})
		})
	})

	now := time.Now()
	req := &iceye.CreateTaskRequest{
		ContractID:        "C-1",
		PointOfInterest:   iceye.Point{Lat: 60.1699, Lon: 24.9384},
		AcquisitionWindow: iceye.TimeWindow{Start: now.Add(48 * time.Hour), End: now.Add(72 * time.Hour)},
		ImagingMode:       "SPOTLIGHT",
		Priority:          iceye.PriorityCommercial,
		EULA:              iceye.EULAStandard,
	}

	task, err := cli.CreateTask(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "T-123", task.ID)
	assert.Equal(t, iceye.TaskStatusReceived, task.Status)
}

func TestGetTask(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/tasking/v1/tasks/T-123", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			json.NewEncoder(w).Encode(iceye.Task{
				ID:     "T-123",
				Status: iceye.TaskStatusActive,
			})
		})
	})

	task, err := cli.GetTask(context.Background(), "T-123")

	require.NoError(t, err)
	assert.Equal(t, "T-123", task.ID)
	assert.Equal(t, iceye.TaskStatusActive, task.Status)
}

func TestCancelTask(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/tasking/v1/tasks/T-1", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPatch, r.Method)

			var body map[string]string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "CANCELED", body["status"])

			json.NewEncoder(w).Encode(iceye.Task{
				ID:     "T-1",
				Status: iceye.TaskStatusCanceled,
			})
		})
	})

	task, err := cli.CancelTask(context.Background(), "T-1")

	require.NoError(t, err)
	assert.Equal(t, "T-1", task.ID)
	assert.Equal(t, iceye.TaskStatusCanceled, task.Status)
}

func TestGetTaskScene(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/tasking/v1/tasks/T-1/scene", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			json.NewEncoder(w).Encode(iceye.TaskScene{
				Duration:      30,
				LookSide:      iceye.LookSideRight,
				PassDirection: iceye.PassDirectionAscending,
			})
		})
	})

	scene, err := cli.GetTaskScene(context.Background(), "T-1")

	require.NoError(t, err)
	assert.Equal(t, iceye.LookSideRight, scene.LookSide)
	assert.Equal(t, iceye.PassDirectionAscending, scene.PassDirection)
	assert.Equal(t, 30, scene.Duration)
}

func TestGetTaskPrice(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/tasking/v1/price", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)

			q := r.URL.Query()
			assert.Equal(t, "C-1", q.Get("contractID"))
			assert.Equal(t, "60.1699", q.Get("pointOfInterest[lat]"))
			assert.Equal(t, "24.9384", q.Get("pointOfInterest[lon]"))
			assert.Equal(t, "SPOTLIGHT", q.Get("imagingMode"))
			assert.Equal(t, "COMMERCIAL", q.Get("priority"))
			assert.Equal(t, "STANDARD", q.Get("eula"))

			json.NewEncoder(w).Encode(iceye.TaskPrice{
				Amount:   500000,
				Currency: "USD",
			})
		})
	})

	price, err := cli.GetTaskPrice(context.Background(), &iceye.TaskPriceRequest{
		ContractID:      "C-1",
		PointOfInterest: iceye.Point{Lat: 60.1699, Lon: 24.9384},
		ImagingMode:     "SPOTLIGHT",
		Priority:        iceye.PriorityCommercial,
		SLA:             "SLA_8H",
		EULA:            iceye.EULAStandard,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(500000), price.Amount)
	assert.Equal(t, "USD", price.Currency)
}

func TestListTaskProducts(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/tasking/v1/tasks/T/products", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			json.NewEncoder(w).Encode([]iceye.TaskProduct{
				{Type: "GRD"},
				{Type: "SLC"},
			})
		})
	})

	products, err := cli.ListTaskProducts(context.Background(), "T")

	require.NoError(t, err)
	assert.Len(t, products, 2)
	assert.Equal(t, "GRD", products[0].Type)
	assert.Equal(t, "SLC", products[1].Type)
}

func TestGetTaskProduct(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/tasking/v1/tasks/T-1/products/GRD", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			json.NewEncoder(w).Encode(iceye.TaskProduct{
				Type: "GRD",
				Assets: map[string]iceye.Asset{
					"data": {Href: "https://example.com/data.tif", Title: "GRD Data"},
				},
			})
		})
	})

	product, err := cli.GetTaskProduct(context.Background(), "T-1", "GRD")

	require.NoError(t, err)
	assert.Equal(t, "GRD", product.Type)
	assert.Contains(t, product.Assets, "data")
}

func TestPaginatorStopsOnError(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/tasking/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"code":   "FATAL",
				"detail": "no",
			})
		})
	})

	var errCount int
	for _, err := range cli.ListTasks(context.Background(), 1, nil) {
		if err != nil {
			errCount++
		}
	}

	assert.Equal(t, 1, errCount, "should receive exactly one error")
}
