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

func TestOrderService_ReviewOrder(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/orders/review")
		requireAuth(t, r, "test-api-key")
		requireContentType(t, r, "application/json")

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)

		items := req["items"].([]any)
		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}

		jsonResponse(w, http.StatusOK, capella.OrderReviewResponse{
			Items: []capella.OrderItemResponse{
				{CollectionID: "capella-geo", GranuleID: "item-1"},
				{CollectionID: "capella-geo", GranuleID: "item-2"},
			},
			TotalCredits: 100.0,
			TotalCostUSD: 500.0,
		})
	}

	cli, _ := newTestClient(t, handler)
	orders := capella.NewOrderService(cli)

	resp, err := orders.ReviewOrder(context.Background(), capella.OrderReviewRequest{
		Items: []capella.OrderItem{
			{CollectionID: "capella-geo", GranuleID: "item-1"},
			{CollectionID: "capella-geo", GranuleID: "item-2"},
		},
	})
	if err != nil {
		t.Fatalf("ReviewOrder failed: %v", err)
	}

	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}

	if resp.TotalCredits != 100.0 {
		t.Errorf("expected 100 credits, got %f", resp.TotalCredits)
	}
}

func TestOrderService_SubmitOrder(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/orders")
		requireAuth(t, r, "test-api-key")

		jsonResponse(w, http.StatusCreated, capella.Order{
			OrderID:   "order-123",
			Status:    capella.OrderPending,
			CreatedAt: time.Now(),
			Items: []capella.OrderItemResponse{
				{CollectionID: "capella-geo", GranuleID: "item-1"},
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	orders := capella.NewOrderService(cli)

	order, err := orders.SubmitOrder(context.Background(), capella.OrderRequest{
		Items: []capella.OrderItem{
			{CollectionID: "capella-geo", GranuleID: "item-1"},
		},
	})
	if err != nil {
		t.Fatalf("SubmitOrder failed: %v", err)
	}

	if order.OrderID != "order-123" {
		t.Errorf("expected order ID 'order-123', got %q", order.OrderID)
	}

	if order.Status != capella.OrderPending {
		t.Errorf("expected status 'pending', got %q", order.Status)
	}
}

func TestOrderService_GetOrder(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/orders/order-123")
		requireAuth(t, r, "test-api-key")

		jsonResponse(w, http.StatusOK, capella.Order{
			OrderID:     "order-123",
			Status:      capella.OrderCompleted,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			CompletedAt: time.Now(),
			Items: []capella.OrderItemResponse{
				{
					CollectionID: "capella-geo",
					GranuleID:    "item-1",
					Status:       capella.OrderCompleted,
					Assets: map[string]capella.Asset{
						"data": {Href: "https://storage.example.com/item-1.tif"},
					},
				},
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	orders := capella.NewOrderService(cli)

	order, err := orders.GetOrder(context.Background(), "order-123")
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}

	if order.Status != capella.OrderCompleted {
		t.Errorf("expected status 'completed', got %q", order.Status)
	}

	if len(order.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(order.Items))
	}
}

func TestOrderService_GetDownloadURLs(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/orders/order-123/download")
		requireAuth(t, r, "test-api-key")

		jsonResponse(w, http.StatusOK, capella.DownloadURLsResponse{
			OrderID: "order-123",
			Downloads: []capella.DownloadURL{
				{
					GranuleID: "item-1",
					AssetKey:  "data",
					URL:       "https://storage.example.com/item-1.tif?sig=abc",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					Size:      1024 * 1024 * 100, // 100MB
				},
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	orders := capella.NewOrderService(cli)

	resp, err := orders.GetDownloadURLs(context.Background(), "order-123")
	if err != nil {
		t.Fatalf("GetDownloadURLs failed: %v", err)
	}

	if resp.OrderID != "order-123" {
		t.Errorf("expected order ID 'order-123', got %q", resp.OrderID)
	}

	if len(resp.Downloads) != 1 {
		t.Errorf("expected 1 download, got %d", len(resp.Downloads))
	}

	if resp.Downloads[0].URL == "" {
		t.Error("expected non-empty download URL")
	}
}

func TestOrderService_OrderTaskingRequest(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/orders/task/tr-123")
		requireAuth(t, r, "test-api-key")

		jsonResponse(w, http.StatusCreated, capella.Order{
			OrderID: "order-456",
			Status:  capella.OrderPending,
		})
	}

	cli, _ := newTestClient(t, handler)
	orders := capella.NewOrderService(cli)

	order, err := orders.OrderTaskingRequest(context.Background(), "tr-123")
	if err != nil {
		t.Fatalf("OrderTaskingRequest failed: %v", err)
	}

	if order.OrderID != "order-456" {
		t.Errorf("expected order ID 'order-456', got %q", order.OrderID)
	}
}

func TestOrderService_QuickOrder(t *testing.T) {
	requestCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		if r.URL.Path == "/orders/review" {
			jsonResponse(w, http.StatusOK, capella.OrderReviewResponse{
				Items: []capella.OrderItemResponse{
					{CollectionID: "capella-geo", GranuleID: "item-1"},
				},
				TotalCredits: 50.0,
			})
		} else if r.URL.Path == "/orders" {
			jsonResponse(w, http.StatusCreated, capella.Order{
				OrderID: "order-789",
				Status:  capella.OrderPending,
			})
		}
	}

	cli, _ := newTestClient(t, handler)
	orders := capella.NewOrderService(cli)

	order, err := orders.QuickOrder(context.Background(), []capella.OrderItem{
		{CollectionID: "capella-geo", GranuleID: "item-1"},
	})
	if err != nil {
		t.Fatalf("QuickOrder failed: %v", err)
	}

	if order.OrderID != "order-789" {
		t.Errorf("expected order ID 'order-789', got %q", order.OrderID)
	}

	if requestCount != 2 {
		t.Errorf("expected 2 requests (review + submit), got %d", requestCount)
	}
}

func TestOrderService_QuickOrder_ReviewErrors(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusOK, capella.OrderReviewResponse{
			Items: []capella.OrderItemResponse{},
			Errors: []capella.OrderError{
				{GranuleID: "item-1", Message: "Item not available"},
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	orders := capella.NewOrderService(cli)

	_, err := orders.QuickOrder(context.Background(), []capella.OrderItem{
		{CollectionID: "capella-geo", GranuleID: "item-1"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestOrderService_WaitForOrder(t *testing.T) {
	pollCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		pollCount++
		status := capella.OrderProcessing
		if pollCount >= 3 {
			status = capella.OrderCompleted
		}

		jsonResponse(w, http.StatusOK, capella.Order{
			OrderID: "order-123",
			Status:  status,
		})
	}

	cli, _ := newTestClient(t, handler)
	orders := capella.NewOrderService(cli)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	order, err := orders.WaitForOrder(ctx, "order-123", 10*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForOrder failed: %v", err)
	}

	if order.Status != capella.OrderCompleted {
		t.Errorf("expected status 'completed', got %q", order.Status)
	}

	if pollCount < 3 {
		t.Errorf("expected at least 3 polls, got %d", pollCount)
	}
}

func TestOrderService_WaitForOrder_Failed(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusOK, capella.Order{
			OrderID: "order-123",
			Status:  capella.OrderFailed,
		})
	}

	cli, _ := newTestClient(t, handler)
	orders := capella.NewOrderService(cli)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := orders.WaitForOrder(ctx, "order-123", 10*time.Millisecond)
	if err == nil {
		t.Fatal("expected error for failed order, got nil")
	}
}
