package capella_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/capella"
)

func TestCatalogService_Search(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/catalog/search")
		requireAuth(t, r, "test-api-key")
		requireContentType(t, r, "application/json")

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var params map[string]any
		if err := json.Unmarshal(body, &params); err != nil {
			t.Fatalf("failed to unmarshal request: %v", err)
		}

		if params["limit"] != float64(10) {
			t.Errorf("expected limit 10, got %v", params["limit"])
		}

		jsonResponse(w, http.StatusOK, capella.SearchResponse{
			Type: "FeatureCollection",
			Features: []capella.STACItem{
				{
					ID:   "item-1",
					Type: "Feature",
					Properties: capella.STACProperties{
						DateTime:       time.Now(),
						InstrumentMode: capella.ModeSpotlight,
						ProductType:    capella.ProductGEO,
					},
				},
			},
			NumberReturned: 1,
			NumberMatched:  1,
		})
	}

	cli, _ := newTestClient(t, handler)
	
	resp, err := cli.CatalogSearch(context.Background(), capella.SearchParams{
		BBox:        []float64{-110, 39.5, -105, 40.5},
		Collections: []string{"capella-geo"},
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(resp.Features) != 1 {
		t.Errorf("expected 1 feature, got %d", len(resp.Features))
	}

	if resp.Features[0].ID != "item-1" {
		t.Errorf("expected item ID 'item-1', got %q", resp.Features[0].ID)
	}
}

func TestCatalogService_SearchItems_Pagination(t *testing.T) {
	page := 0
	var serverURL string

	handler := func(w http.ResponseWriter, r *http.Request) {
		// First request is POST (search), subsequent are GET (pagination links)
		if page == 0 {
			requireMethod(t, r, http.MethodPost)
		} else {
			requireMethod(t, r, http.MethodGet)
		}
		page++

		var nextLink []capella.Link
		if page < 2 {
			// Use full server URL for the next link
			nextLink = []capella.Link{{Rel: "next", Href: serverURL + "/catalog/search?page=2"}}
		}

		jsonResponse(w, http.StatusOK, capella.SearchResponse{
			Type: "FeatureCollection",
			Features: []capella.STACItem{
				{ID: "item-" + string(rune('0'+page))},
			},
			Links: nextLink,
		})
	}

	cli, srv := newTestClient(t, handler)
	serverURL = srv.URL
	
	var items []capella.STACItem
	for item, err := range cli.CatalogSearchItems(context.Background(), capella.SearchParams{Limit: 1}) {
		if err != nil {
			t.Fatalf("iterator error: %v", err)
		}
		items = append(items, item)
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestCatalogService_ListCollections(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/catalog/collections")
		requireAuth(t, r, "test-api-key")

		jsonResponse(w, http.StatusOK, map[string]any{
			"collections": []capella.STACCollection{
				{
					ID:          "capella-geo",
					Type:        "Collection",
					Title:       "Capella GEO",
					Description: "Geocoded imagery",
				},
				{
					ID:          "capella-slc",
					Type:        "Collection",
					Title:       "Capella SLC",
					Description: "Single-look complex imagery",
				},
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	
	collections, err := cli.ListCollections(context.Background())
	if err != nil {
		t.Fatalf("ListCollections failed: %v", err)
	}

	if len(collections) != 2 {
		t.Errorf("expected 2 collections, got %d", len(collections))
	}

	if collections[0].ID != "capella-geo" {
		t.Errorf("expected collection ID 'capella-geo', got %q", collections[0].ID)
	}
}

func TestCatalogService_GetCollection(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/catalog/collections/capella-geo")
		requireAuth(t, r, "test-api-key")

		jsonResponse(w, http.StatusOK, capella.STACCollection{
			ID:          "capella-geo",
			Type:        "Collection",
			Title:       "Capella GEO",
			Description: "Geocoded imagery",
		})
	}

	cli, _ := newTestClient(t, handler)
	
	collection, err := cli.GetCollection(context.Background(), "capella-geo")
	if err != nil {
		t.Fatalf("GetCollection failed: %v", err)
	}

	if collection.ID != "capella-geo" {
		t.Errorf("expected collection ID 'capella-geo', got %q", collection.ID)
	}
}

func TestCatalogService_ListArchiveExports(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/catalog/archive-export/available")

		jsonResponse(w, http.StatusOK, []capella.ArchiveExport{
			{
				ID:        "export-1",
				Name:      "Q4 2024 Export",
				CreatedAt: time.Now(),
			},
		})
	}

	cli, _ := newTestClient(t, handler)
	
	exports, err := cli.ListArchiveExports(context.Background())
	if err != nil {
		t.Fatalf("ListArchiveExports failed: %v", err)
	}

	if len(exports) != 1 {
		t.Errorf("expected 1 export, got %d", len(exports))
	}
}

func TestCatalogService_GetArchiveExportURL(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/catalog/archive-export/presigned")

		jsonResponse(w, http.StatusOK, capella.PresignedURL{
			URL:       "https://storage.example.com/export.zip?sig=abc",
			ExpiresAt: time.Now().Add(1 * time.Hour),
		})
	}

	cli, _ := newTestClient(t, handler)
	
	url, err := cli.GetArchiveExportURL(context.Background(), "export-1")
	if err != nil {
		t.Fatalf("GetArchiveExportURL failed: %v", err)
	}

	if url.URL == "" {
		t.Error("expected non-empty URL")
	}
}

func TestSearchBuilder(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 30, 23, 59, 59, 0, time.UTC)

	params := capella.NewSearchBuilder().
		BBox(-110, 39.5, -105, 40.5).
		Collections("capella-geo", "capella-slc").
		DateTime(start, end).
		InstrumentMode(capella.ModeSpotlight).
		ProductType(capella.ProductGEO).
		OrbitState(capella.OrbitAscending).
		SortBy("properties.datetime", true).
		Limit(50).
		Build()

	if len(params.BBox) != 4 {
		t.Errorf("expected 4 bbox values, got %d", len(params.BBox))
	}

	if len(params.Collections) != 2 {
		t.Errorf("expected 2 collections, got %d", len(params.Collections))
	}

	if params.Limit != 50 {
		t.Errorf("expected limit 50, got %d", params.Limit)
	}

	if params.SortBy != "-properties.datetime" {
		t.Errorf("expected sortBy '-properties.datetime', got %q", params.SortBy)
	}
}
