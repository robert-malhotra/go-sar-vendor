// Comprehensive unit tests for the iceye SDK. They use `httptest.Server` to mock
// the ICEYE OAuth and REST endpoints, so no network calls leave the process.
//
// Focus areas:
//   - Token acquisition & refresh logic.
//   - Error handling (RFC-7807 parsing).
//   - Paginator correctness across Contracts, Tasks, and Products.
//   - JSON request/response correctness for create, get, cancel, price.
//
// The tests avoid 3rd-party deps; only std-lib is used.
//
// Run: `go test -tags=unit ./...` (requires Go 1.23+ for `iter` generics).
package iceye

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// helper to create a mock auth+API server pair and configured client.
func newTestClient(t *testing.T, mux func(mux *http.ServeMux, authHits *atomic.Int32)) (*Client, *httptest.Server, *atomic.Int32) {
	t.Helper()
	authHits := &atomic.Int32{}
	muxrouter := http.NewServeMux()
	mux(muxrouter, authHits)

	srv := httptest.NewServer(muxrouter)
	t.Cleanup(srv.Close)

	cli := NewClient(Config{
		BaseURL:      srv.URL, // same host for API
		TokenURL:     srv.URL + "/oauth2/token",
		ClientID:     "test",
		ClientSecret: "secret",
		HTTPClient:   srv.Client(),
	})
	return cli, srv, authHits
}

// ---------------- Token logic ----------------
func TestAuthenticateSuccess(t *testing.T) {
	cli, _, hits := newTestClient(t, func(mux *http.ServeMux, authHits *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
			authHits.Add(1)
			_ = r.ParseForm()
			if r.FormValue("grant_type") != "client_credentials" {
				http.Error(w, "bad grant", 400)
				return
			}
			json.NewEncoder(w).Encode(map[string]any{
				"access_token": "abc123",
				"expires_in":   3600,
			})
		})
	})

	ctx := context.Background()
	if err := cli.authenticate(ctx); err != nil {
		t.Fatalf("authenticate failed: %v", err)
	}
	if cli.token != "abc123" {
		t.Fatalf("unexpected token: %s", cli.token)
	}
	if hits.Load() != 1 {
		t.Fatalf("auth endpoint not hit exactly once, got %d", hits.Load())
	}

	// second authenticate should be a no-op (<30 s skew)
	if err := cli.authenticate(ctx); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 {
		t.Fatalf("token refreshed unexpectedly: %d calls", hits.Load())
	}
}

// ---------------- Error handling ----------------
func TestDoParsesProblemJSON(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"access_token": "x", "expires_in": 1})
		})
		mux.HandleFunc("/boom", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]any{"code": "ERR_BAD", "detail": "kaput"})
		})
	})
	ctx := context.Background()
	err := cli.do(ctx, http.MethodGet, "/boom", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if e, ok := err.(*Error); !ok || e.Code != "ERR_BAD" {
		t.Fatalf("unexpected error type/value: %#v", err)
	}
}

// ---------------- Pagination helpers ----------------
func TestListContractsPagination(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"access_token": "x", "expires_in": 1})
		})

		pages := []ContractsResponse{
			{Data: []Contract{{ID: "c1"}}, Cursor: "next"},
			{Data: []Contract{{ID: "c2"}}, Cursor: ""},
		}
		mux.HandleFunc("/company/v1/contracts", func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Query().Get("cursor") {
			case "":
				json.NewEncoder(w).Encode(pages[0])
			case "next":
				json.NewEncoder(w).Encode(pages[1])
			default:
				t.Fatalf("unexpected cursor %s", r.URL.Query().Get("cursor"))
			}
		})
	})

	var ids []string
	for resp, err := range cli.ListContracts(context.Background(), 1) {
		if err != nil {
			t.Fatalf("iter error: %v", err)
		}
		for _, c := range resp.Data {
			ids = append(ids, c.ID)
		}
	}
	if len(ids) != 2 || ids[0] != "c1" || ids[1] != "c2" {
		t.Fatalf("unexpected ids: %v", ids)
	}
}

func TestListTasksPagination(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"access_token": "x", "expires_in": 1})
		})
		mux.HandleFunc("/tasking/v2/tasks", func(w http.ResponseWriter, r *http.Request) {
			cur := r.URL.Query().Get("cursor")
			switch cur {
			case "":
				json.NewEncoder(w).Encode(map[string]any{"data": []Task{{ID: "t1"}}, "cursor": "n"})
			case "n":
				json.NewEncoder(w).Encode(map[string]any{"data": []Task{{ID: "t2"}}, "cursor": nil})
			default:
				t.Fatalf("unexpected cursor %s", cur)
			}
		})
	})

	got := 0
	for tasks, err := range cli.ListTasks(context.Background(), 1, nil) {
		if err != nil {
			t.Fatalf("iter err: %v", err)
		}
		got += len(tasks)
	}
	if got != 2 {
		t.Fatalf("expected 2 tasks, got %d", got)
	}
}

// ---------------- CRUD helpers ----------------
func TestCreateAndGetTaskHappyPath(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"access_token": "x", "expires_in": 1})
		})
		mux.HandleFunc("/tasking/v2/tasks", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method %s", r.Method)
			}
			var crt CreateTaskRequest
			json.NewDecoder(r.Body).Decode(&crt)
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(map[string]any{"id": "T-123", "contractID": crt.ContractID})
		})
		mux.HandleFunc("/tasking/v2/tasks/T-123", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"id": "T-123", "status": "SUBMITTED"})
		})
	})

	ctx := context.Background()
	tsk, err := cli.CreateTask(ctx, CreateTaskRequest{ContractID: "C"})
	if err != nil || tsk.ID != "T-123" {
		t.Fatalf("create task failed: %v %+v", err, tsk)
	}
	got, err := cli.GetTask(ctx, "T-123")
	if err != nil || got.ID != "T-123" || got.Status != "SUBMITTED" {
		t.Fatalf("get task failed: %v %+v", err, got)
	}
}

func TestCancelTask(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"access_token": "x", "expires_in": 1})
		})
		mux.HandleFunc("/tasking/v2/tasks/T-1", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				t.Fatalf("expected PATCH, got %s", r.Method)
			}
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if body["status"] != "CANCELED" {
				t.Fatalf("body mismatch: %+v", body)
			}
			w.WriteHeader(204)
		})
	})
	if err := cli.CancelTask(context.Background(), "T-1"); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
}

// ---------------- Products paginator ----------------
func TestGetTaskProductsPagination(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"access_token": "x", "expires_in": 1})
		})
		mux.HandleFunc("/tasking/v2/tasks/T/products", func(w http.ResponseWriter, r *http.Request) {
			cur := r.URL.Query().Get("cursor")
			switch cur {
			case "":
				json.NewEncoder(w).Encode(map[string]any{"data": []Product{{Type: "GRD"}}, "cursor": "next"})
			case "next":
				json.NewEncoder(w).Encode(map[string]any{"data": []Product{{Type: "SLC"}}, "cursor": nil})
			}
		})
	})
	total := 0
	for prods, err := range cli.GetTaskProducts(context.Background(), "T", 1) {
		if err != nil {
			t.Fatal(err)
		}
		total += len(prods)
	}
	if total != 2 {
		t.Fatalf("expected 2 products, got %d", total)
	}
}

// ensure paginator stops on error
func TestPaginatorStopsOnError(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"access_token": "x", "expires_in": 1})
		})
		mux.HandleFunc("/tasking/v2/tasks", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]any{"code": "FATAL", "detail": "no"})
		})
	})
	tasks := cli.ListTasks(context.Background(), 1, nil)
	for _, err := range tasks {
		if err == nil {
			t.Fatal("expected error on first page")
		}
	}
}

// fast-forward time to test auto-refresh logic
func TestTokenRefreshAfterExpiry(t *testing.T) {
	var tokenCalls atomic.Int32
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
			tokenCalls.Add(1)
			json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 1})
		})
		mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	})
	ctx := context.Background()
	_ = cli.do(ctx, http.MethodGet, "/ping", nil, nil)
	time.Sleep(2 * time.Second) // wait for expiry
	_ = cli.do(ctx, http.MethodGet, "/ping", nil, nil)
	if tokenCalls.Load() < 2 {
		t.Fatalf("expected token refresh, got %d calls", tokenCalls.Load())
	}
}
