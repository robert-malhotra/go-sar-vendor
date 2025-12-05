package iceye_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/robert-malhotra/go-sar-vendor/pkg/iceye"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCatalogItemsPagination(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/catalog/v1/items", func(w http.ResponseWriter, r *http.Request) {
			cursor := r.URL.Query().Get("cursor")
			switch cursor {
			case "":
				json.NewEncoder(w).Encode(iceye.CatalogResponse{
					Data:   []iceye.STACItem{{ID: "item-1", Type: "Feature"}},
					Cursor: "next",
				})
			case "next":
				json.NewEncoder(w).Encode(iceye.CatalogResponse{
					Data:   []iceye.STACItem{{ID: "item-2", Type: "Feature"}},
					Cursor: "",
				})
			}
		})
	})

	var ids []string
	for resp, err := range cli.ListCatalogItems(context.Background(), 1, nil) {
		require.NoError(t, err)
		for _, item := range resp.Data {
			ids = append(ids, item.ID)
		}
	}

	assert.Equal(t, []string{"item-1", "item-2"}, ids)
}

func TestListCatalogItemsWithFilters(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/catalog/v1/items", func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			assert.NotEmpty(t, q.Get("datetime"))
			assert.NotEmpty(t, q.Get("bbox"))

			json.NewEncoder(w).Encode(iceye.CatalogResponse{
				Data:   []iceye.STACItem{{ID: "filtered-item"}},
				Cursor: "",
			})
		})
	})

	bbox := iceye.BoundingBox{24.0, 59.5, 25.5, 60.5}
	opts := &iceye.ListItemsOptions{
		BBox:     &bbox,
		Datetime: "2024-01-01T00:00:00Z/2024-06-01T00:00:00Z",
	}

	for resp, err := range cli.ListCatalogItems(context.Background(), 10, opts) {
		require.NoError(t, err)
		require.Len(t, resp.Data, 1)
		assert.Equal(t, "filtered-item", resp.Data[0].ID)
	}
}

func TestSearchCatalogItems(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/catalog/v1/search", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)

			var req iceye.SearchRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.NotNil(t, req.BBox)
			assert.Contains(t, req.Query, "product_type")

			json.NewEncoder(w).Encode(map[string]any{
				"data":   []iceye.STACItem{{ID: "search-result-1"}},
				"cursor": "",
			})
		})
	})

	bbox := iceye.BoundingBox{24.0, 59.5, 25.5, 60.5}
	req := &iceye.SearchRequest{
		BBox:     &bbox,
		Datetime: "2024-01-01T00:00:00Z/2024-06-01T00:00:00Z",
		Query: map[string]iceye.QueryFilter{
			"product_type": {In: []any{"GRD", "SLC"}},
		},
		Limit: 20,
	}

	for resp, err := range cli.SearchCatalogItems(context.Background(), req) {
		require.NoError(t, err)
		require.Len(t, resp.Data, 1)
		assert.Equal(t, "search-result-1", resp.Data[0].ID)
	}
}

func TestSearchCatalogItemsPagination(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		callCount := 0
		mux.HandleFunc("/catalog/v1/search", func(w http.ResponseWriter, r *http.Request) {
			callCount++
			json.NewEncoder(w).Encode(map[string]any{
				"data":   []iceye.STACItem{{ID: "page1-item"}},
				"cursor": "cursor2",
			})
		})
		mux.HandleFunc("/catalog/v1/items", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "cursor2", r.URL.Query().Get("cursor"))
			json.NewEncoder(w).Encode(map[string]any{
				"data":   []iceye.STACItem{{ID: "page2-item"}},
				"cursor": "",
			})
		})
	})

	var ids []string
	for resp, err := range cli.SearchCatalogItems(context.Background(), &iceye.SearchRequest{}) {
		require.NoError(t, err)
		for _, item := range resp.Data {
			ids = append(ids, item.ID)
		}
	}

	assert.Equal(t, []string{"page1-item", "page2-item"}, ids)
}

func TestPurchaseCatalogItems(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/catalog/v1/purchases", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)

			var req iceye.PurchaseRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, "C-123", req.ContractID)
			assert.Equal(t, []string{"item-1", "item-2"}, req.ItemIDs)
			assert.Equal(t, "my-ref-123", req.Reference)

			json.NewEncoder(w).Encode(iceye.PurchaseResponse{
				PurchaseID: "P-789",
			})
		})
	})

	resp, err := cli.PurchaseCatalogItems(context.Background(), &iceye.PurchaseRequest{
		ContractID: "C-123",
		ItemIDs:    []string{"item-1", "item-2"},
		Reference:  "my-ref-123",
	})

	require.NoError(t, err)
	assert.Equal(t, "P-789", resp.PurchaseID)
}

func TestGetPurchase(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/catalog/v1/purchases/P-789", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			json.NewEncoder(w).Encode(iceye.Purchase{
				ID:           "P-789",
				CustomerName: "Acme Corp",
				ContractName: "Test Contract",
				Status:       iceye.PurchaseStatusClosed,
				Reference:    "my-ref",
			})
		})
	})

	purchase, err := cli.GetPurchase(context.Background(), "P-789")

	require.NoError(t, err)
	assert.Equal(t, "P-789", purchase.ID)
	assert.Equal(t, "Acme Corp", purchase.CustomerName)
	assert.Equal(t, "Test Contract", purchase.ContractName)
	assert.Equal(t, iceye.PurchaseStatusClosed, purchase.Status)
	assert.Equal(t, "my-ref", purchase.Reference)
}

func TestListPurchases(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/catalog/v1/purchases", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			cursor := r.URL.Query().Get("cursor")
			switch cursor {
			case "":
				json.NewEncoder(w).Encode(iceye.PurchasesResponse{
					Data: []iceye.Purchase{{
						ID:           "P-1",
						CustomerName: "Acme Corp",
						ContractName: "Contract A",
						Status:       iceye.PurchaseStatusActive,
					}},
					Cursor: "next",
				})
			case "next":
				json.NewEncoder(w).Encode(iceye.PurchasesResponse{
					Data: []iceye.Purchase{{
						ID:           "P-2",
						CustomerName: "Acme Corp",
						ContractName: "Contract B",
						Status:       iceye.PurchaseStatusClosed,
					}},
					Cursor: "",
				})
			}
		})
	})

	var ids []string
	for resp, err := range cli.ListPurchases(context.Background(), 1) {
		require.NoError(t, err)
		for _, p := range resp.Data {
			ids = append(ids, p.ID)
		}
	}

	assert.Equal(t, []string{"P-1", "P-2"}, ids)
}

func TestListPurchasedProducts(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/catalog/v1/purchases/P-789/products", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			json.NewEncoder(w).Encode(map[string]any{
				"data": []iceye.STACItem{{ID: "product-item-1"}, {ID: "product-item-2"}},
			})
		})
	})

	items, err := cli.ListPurchasedProducts(context.Background(), "P-789")

	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "product-item-1", items[0].ID)
	assert.Equal(t, "product-item-2", items[1].ID)
}
