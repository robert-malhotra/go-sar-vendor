package airbus

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// testServer creates a test server with a token endpoint and API endpoints.
func testServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()

	mux := http.NewServeMux()

	// Token endpoint
	mux.HandleFunc("/auth/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})

	// API endpoints
	mux.HandleFunc("/", handler)

	server := httptest.NewServer(mux)

	client, err := NewClient("test-api-key",
		WithBaseURL(server.URL),
		WithTokenURL(server.URL+"/auth/token"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	return server, client
}

func TestNewClient(t *testing.T) {
	client, err := NewClient("test-api-key")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestNewDevClient(t *testing.T) {
	client, err := NewDevClient("test-api-key")
	if err != nil {
		t.Fatalf("NewDevClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewDevClient() returned nil")
	}
	if client.BaseURL().String() != DevBaseURL {
		t.Errorf("expected base URL %s, got %s", DevBaseURL, client.BaseURL().String())
	}
}

func TestPing(t *testing.T) {
	server, client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sar/ping" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.Ping(context.Background())
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestHealth(t *testing.T) {
	server, client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sar/health" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthStatus{
			Status:  "pass",
			Version: "2.7.0",
		})
	})
	defer server.Close()

	health, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if health.Status != "pass" {
		t.Errorf("expected status 'pass', got %s", health.Status)
	}
}

func TestWhoAmI(t *testing.T) {
	server, client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/whoami" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UserInfo{
			Username:       "test@example.com",
			PWChangeNeeded: false,
			Services:       []Service{ServiceRadar},
		})
	})
	defer server.Close()

	user, err := client.WhoAmI(context.Background())
	if err != nil {
		t.Fatalf("WhoAmI() error = %v", err)
	}
	if user.Username != "test@example.com" {
		t.Errorf("expected username 'test@example.com', got %s", user.Username)
	}
}

func TestSearchCatalogue(t *testing.T) {
	server, client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sar/catalogue" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CatalogueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(FeatureCollection{
			Type:  "FeatureCollection",
			Total: 1,
			Features: []Feature{
				{
					Type: "Feature",
					Properties: AcquisitionProperties{
						ItemID:        "test-item-id",
						AcquisitionID: "TSX-1_ST_S_spot_049R_49677_D31767159_432",
						Mission:       MissionTSX,
						SensorMode:    SensorModeStaringSpotlight,
					},
				},
			},
		})
	})
	defer server.Close()

	result, err := client.SearchCatalogue(context.Background(), &CatalogueRequest{
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("SearchCatalogue() error = %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
	if len(result.Features) != 1 {
		t.Errorf("expected 1 feature, got %d", len(result.Features))
	}
}

func TestSearchFeasibility(t *testing.T) {
	server, client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sar/feasibility" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req FeasibilityRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.FeasibilityLevel == "" {
			http.Error(w, "feasibilityLevel required", http.StatusBadRequest)
			return
		}
		if req.SensorMode == "" {
			http.Error(w, "sensorMode required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(FeatureCollection{
			Type: "FeatureCollection",
			Features: []Feature{
				{
					Type: "Feature",
					Properties: AcquisitionProperties{
						ItemID:     "feasibility-item-id",
						Mission:    MissionTSX,
						SensorMode: SensorModeStaringSpotlight,
						Status:     "planned",
					},
				},
			},
		})
	})
	defer server.Close()

	now := time.Now()
	result, err := client.SearchFeasibility(context.Background(), &FeasibilityRequest{
		AOI: NewPolygonGeometry([][][2]float64{{
			{9.0, 47.0}, {10.0, 47.0}, {10.0, 48.0}, {9.0, 48.0}, {9.0, 47.0},
		}}),
		Time: TimeRange{
			From: now,
			To:   now.AddDate(0, 1, 0),
		},
		FeasibilityLevel: FeasibilityLevelSimple,
		SensorMode:       SensorModeStaringSpotlight,
	})
	if err != nil {
		t.Fatalf("SearchFeasibility() error = %v", err)
	}
	if len(result.Features) != 1 {
		t.Errorf("expected 1 feature, got %d", len(result.Features))
	}
}

func TestListBaskets(t *testing.T) {
	server, client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sar/baskets" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Basket{
			{
				BasketID:          "basket-123",
				CreationTime:      time.Now(),
				CustomerReference: "REF-001",
			},
		})
	})
	defer server.Close()

	baskets, err := client.ListBaskets(context.Background())
	if err != nil {
		t.Fatalf("ListBaskets() error = %v", err)
	}
	if len(baskets) != 1 {
		t.Errorf("expected 1 basket, got %d", len(baskets))
	}
}

func TestCreateBasket(t *testing.T) {
	server, client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sar/baskets" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CreateBasketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Basket{
			BasketID:          "new-basket-id",
			CreationTime:      time.Now(),
			CustomerReference: req.CustomerReference,
			Purpose:           req.Purpose,
		})
	})
	defer server.Close()

	basket, err := client.CreateBasket(context.Background(), &CreateBasketRequest{
		CustomerReference: "PROJECT-001",
		Purpose:           PurposeEducationResearch,
	})
	if err != nil {
		t.Fatalf("CreateBasket() error = %v", err)
	}
	if basket.BasketID != "new-basket-id" {
		t.Errorf("expected basket ID 'new-basket-id', got %s", basket.BasketID)
	}
}

func TestAddItemsToBasket(t *testing.T) {
	server, client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sar/baskets/basket-123/addItems" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req AddItemsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Basket{
			BasketID:     "basket-123",
			CreationTime: time.Now(),
			Items: []Item{
				{
					ItemID:        "item-1",
					AcquisitionID: req.Acquisitions[0],
				},
			},
		})
	})
	defer server.Close()

	basket, err := client.AddItemsToBasket(context.Background(), "basket-123", &AddItemsRequest{
		Acquisitions: []string{"TSX-1_ST_S_spot_049R_49677_D31767159_432"},
		OrderOptions: &OrderOptions{
			ProductType:       ProductTypeEEC,
			ResolutionVariant: ResolutionVariantRE,
			OrbitType:         OrbitTypeScience,
		},
	})
	if err != nil {
		t.Fatalf("AddItemsToBasket() error = %v", err)
	}
	if len(basket.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(basket.Items))
	}
}

func TestSubmitBasket(t *testing.T) {
	server, client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sar/baskets/basket-123/submit" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Order{
			BasketID:       "basket-123",
			OrderID:        "order-456",
			SubmissionTime: time.Now(),
		})
	})
	defer server.Close()

	order, err := client.SubmitBasket(context.Background(), "basket-123")
	if err != nil {
		t.Fatalf("SubmitBasket() error = %v", err)
	}
	if order.OrderID != "order-456" {
		t.Errorf("expected order ID 'order-456', got %s", order.OrderID)
	}
}

func TestListOrders(t *testing.T) {
	server, client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sar/orders" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]OrderSummary{
			{
				BasketID:       "basket-123",
				OrderID:        "order-456",
				SubmissionTime: timePtr(time.Now()),
				OrderType:      OrderTypeCatalogue,
			},
		})
	})
	defer server.Close()

	orders, err := client.ListOrders(context.Background())
	if err != nil {
		t.Fatalf("ListOrders() error = %v", err)
	}
	if len(orders) != 1 {
		t.Errorf("expected 1 order, got %d", len(orders))
	}
}

func TestGetPrices(t *testing.T) {
	server, client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sar/prices" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]PriceResponse{
			{
				ItemID:        "item-1",
				AcquisitionID: "acq-1",
				Price: Price{
					Final:    true,
					Total:    1000.00,
					Currency: "EUR",
				},
			},
		})
	})
	defer server.Close()

	prices, err := client.GetPrices(context.Background(), &PricesRequest{
		Acquisitions: []string{"acq-1"},
	})
	if err != nil {
		t.Fatalf("GetPrices() error = %v", err)
	}
	if len(prices) != 1 {
		t.Errorf("expected 1 price, got %d", len(prices))
	}
	if prices[0].Price.Total != 1000.00 {
		t.Errorf("expected price 1000.00, got %f", prices[0].Price.Total)
	}
}

func TestGetConfig(t *testing.T) {
	server, client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sar/config" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Config{
			Permissions: &Permissions{
				CanOrder: true,
				CanTask:  true,
			},
		})
	})
	defer server.Close()

	config, err := client.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	if config.Permissions == nil {
		t.Fatal("expected permissions, got nil")
	}
	if !config.Permissions.CanOrder {
		t.Error("expected CanOrder to be true")
	}
}

func TestAPIKeyAuth(t *testing.T) {
	tokenCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalled = true
		if r.URL.Path != "/token" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check form values
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if r.FormValue("grant_type") != "api_key" {
			http.Error(w, "invalid grant_type", http.StatusBadRequest)
			return
		}
		if r.FormValue("apikey") != "test-key" {
			http.Error(w, "invalid apikey", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "obtained-token",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	auth := NewAPIKeyAuth("test-key", server.URL+"/token", nil)

	// Create a mock request to test Apply
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/api", nil)
	err := auth.Apply(context.Background(), req)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if !tokenCalled {
		t.Error("token endpoint was not called")
	}
	if auth.Token() != "obtained-token" {
		t.Errorf("expected token 'obtained-token', got %s", auth.Token())
	}
	authHeader := req.Header.Get("Authorization")
	if authHeader != "Bearer obtained-token" {
		t.Errorf("expected header 'Bearer obtained-token', got %s", authHeader)
	}
}

func TestAPIKeyAuth_TokenReuse(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "token",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	auth := NewAPIKeyAuth("test-key", server.URL+"/token", nil)

	// First call should get token
	req1, _ := http.NewRequest(http.MethodGet, "https://example.com/api", nil)
	if err := auth.Apply(context.Background(), req1); err != nil {
		t.Fatalf("first Apply() error = %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	// Second call should reuse token (not expired)
	req2, _ := http.NewRequest(http.MethodGet, "https://example.com/api", nil)
	if err := auth.Apply(context.Background(), req2); err != nil {
		t.Fatalf("second Apply() error = %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call (token reused), got %d", callCount)
	}
}

func TestNewPolygonGeometry(t *testing.T) {
	geom := NewPolygonGeometry([][][2]float64{{
		{9.0, 47.0}, {10.0, 47.0}, {10.0, 48.0}, {9.0, 48.0}, {9.0, 47.0},
	}})
	if geom == nil {
		t.Fatal("NewPolygonGeometry() returned nil")
	}
	if geom.Type != "Polygon" {
		t.Errorf("expected type 'Polygon', got %s", geom.Type)
	}
}

func TestNewPointGeometry(t *testing.T) {
	geom := NewPointGeometry(9.5, 47.5)
	if geom == nil {
		t.Fatal("NewPointGeometry() returned nil")
	}
	if geom.Type != "Point" {
		t.Errorf("expected type 'Point', got %s", geom.Type)
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
