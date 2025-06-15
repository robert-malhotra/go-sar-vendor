package airbus_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/robert.malhotra/umbra-client/pkg/airbus"
)

/*────────────────── helper to spin up mock server ───────────────────*/

type mockEnv struct {
	srv      *httptest.Server
	authHits atomic.Int32
	apiHits  atomic.Int32
}

func newEnv(t *testing.T, mux func(w http.ResponseWriter, r *http.Request, env *mockEnv)) (*airbus.Client, *mockEnv) {
	env := &mockEnv{}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token": // mock Keycloak
			env.authHits.Add(1)
			json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 2})
		default:
			env.apiHits.Add(1)
			mux(w, r, env)
		}
	})
	env.srv = httptest.NewServer(h)
	t.Cleanup(env.srv.Close)

	// Redirect SDK URLs
	tokenUrl := env.srv.URL + "/token"
	baseUrl := env.srv.URL + "/sar"

	cli := airbus.New("key", env.srv.Client(), airbus.WithBaseURL(baseUrl), airbus.WithTokenURL(tokenUrl))
	return cli, env
}

/*────────────────── tests ───────────────────────────────────────────*/

func TestTokenRefresh(t *testing.T) {
	cli, env := newEnv(t, func(w http.ResponseWriter, r *http.Request, _ *mockEnv) {
		w.Write([]byte(`{"ok":true}`))
	})

	ctx := context.Background()
	req := airbus.FeasibilityRequest{
		AOI:              airbus.GeoJSONGeometry{Type: "Point", Coordinates: []float64{0, 0}},
		Time:             airbus.TimeRange{From: time.Now(), To: time.Now()},
		FeasibilityLevel: "complete",
		SensorMode:       "SAR_SM_S",
		ProductType:      airbus.PTSSC,
		OrbitType:        airbus.OrbitRapid,
	}
	if _, err := cli.SearchFeasibility(ctx, req); err != nil {
		t.Fatal(err)
	}
	if env.authHits.Load() != 1 {
		t.Fatalf("want 1 token call, got %d", env.authHits.Load())
	}
	time.Sleep(3 * time.Second) // token expired → refresh
	if _, err := cli.SearchFeasibility(ctx, req); err != nil {
		t.Fatal(err)
	}
	if env.authHits.Load() < 2 {
		t.Fatalf("token not refreshed, calls=%d", env.authHits.Load())
	}
}

func TestAddItemsSubmitOrder(t *testing.T) {
	cli, _ := newEnv(t, func(w http.ResponseWriter, r *http.Request, _ *mockEnv) {
		switch r.URL.Path {
		case "/sar/shopcart/addItems":
			var body airbus.AddItemsRequest
			json.NewDecoder(r.Body).Decode(&body)
			if len(body.Items) != 1 || body.Items[0] != "itm" {
				t.Fatalf("bad items: %+v", body)
			}
			w.WriteHeader(200)
		case "/sar/shopcart":
			w.WriteHeader(200)
		case "/sar/shopcart/submit":
			json.NewEncoder(w).Encode(map[string]string{"orderId": "ORD"})
		case "/sar/orders/ORD":
			json.NewEncoder(w).Encode(map[string]any{"orderId": "ORD", "status": "completed"})
		default:
			http.NotFound(w, r)
		}
	})

	ctx := context.Background()
	err := cli.AddItems(ctx, airbus.AddItemsRequest{
		Items: []string{"itm"},
		OrderOptions: airbus.OrderOptions{
			ProductType:       airbus.PTSSC,
			ResolutionVariant: airbus.ResRadiometric,
			OrbitType:         airbus.OrbitRapid,
			MapProjection:     airbus.MapAuto,
			GainAttenuation:   0,
		},
	})
	if err != nil {
		t.Fatalf("addItems: %v", err)
	}
	if err := cli.PatchShopcart(ctx, airbus.ShopcartPatch{Purpose: "Research"}); err != nil {
		t.Fatalf("patch: %v", err)
	}
	id, err := cli.SubmitShopcart(ctx)
	if err != nil || id != "ORD" {
		t.Fatalf("submit failed: %v id=%s", err, id)
	}
	ord, err := cli.Order(ctx, id)
	if err != nil || ord.Status != "completed" {
		t.Fatalf("order fetch: %v %+v", err, ord)
	}
}

func TestFeasibilityRequestBody(t *testing.T) {
	var captured []byte
	cli, _ := newEnv(t, func(w http.ResponseWriter, r *http.Request, _ *mockEnv) {
		if r.URL.Path == "/sar/feasibility" {
			captured, _ = io.ReadAll(r.Body)
			w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
		}
	})

	req := airbus.FeasibilityRequest{
		AOI:              airbus.GeoJSONGeometry{Type: "Point", Coordinates: []float64{1, 2}},
		Time:             airbus.TimeRange{From: time.Now(), To: time.Now().Add(1 * time.Hour)},
		FeasibilityLevel: "complete",
		SensorMode:       "SAR_SM_S",
		ProductType:      airbus.PTEEC,
		OrbitType:        airbus.OrbitRapid,
	}
	if _, err := cli.SearchFeasibility(context.Background(), req); err != nil {
		t.Fatalf("call failed: %v", err)
	}
	if !bytes.Contains(captured, []byte("\"feasibilityLevel\":\"complete\"")) {
		t.Error("missing key field in body")
	}
}
