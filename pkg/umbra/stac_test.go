package umbra_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/robert.malhotra/go-sar-vendor/pkg/umbra"
)

func TestGetSTACItem(t *testing.T) {
	expected := umbra.STACItem{
		ID:         "item-123",
		Type:       "Feature",
		Collection: "umbra-sar",
		Properties: map[string]interface{}{
			"datetime": "2024-01-15T12:00:00Z",
		},
		Assets: map[string]umbra.STACAsset{
			"data": {
				Href: "s3://bucket/data.tif",
				Type: "image/tiff",
			},
		},
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/stac/collections/umbra-sar/items/item-123")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	item, err := cli.GetSTACItem(context.Background(), "umbra-sar", "item-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != expected.ID {
		t.Errorf("expected ID %s, got %s", expected.ID, item.ID)
	}
	if item.Collection != expected.Collection {
		t.Errorf("expected collection %s, got %s", expected.Collection, item.Collection)
	}
}

func TestGetSTACItemV2(t *testing.T) {
	expected := umbra.STACItem{
		ID:         "item-456",
		Type:       "Feature",
		Collection: "umbra-sar",
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/v2/stac/collections/umbra-sar/items/item-456")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	item, err := cli.GetSTACItemV2(context.Background(), "umbra-sar", "item-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != expected.ID {
		t.Errorf("expected ID %s, got %s", expected.ID, item.ID)
	}
}

func TestSearchSTAC(t *testing.T) {
	expected := umbra.STACSearchResponse{
		Type: "FeatureCollection",
		Features: []umbra.STACItem{
			{ID: "item-1", Collection: "umbra-sar"},
			{ID: "item-2", Collection: "umbra-sar"},
		},
		NumberMatched:  2,
		NumberReturned: 2,
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/stac/search")
		requireAuth(t, r, "test-token")
		requireContentType(t, r, "application/json")
		jsonResponse(w, http.StatusOK, expected)
	})

	req := umbra.STACSearchRequest{
		Limit:       10,
		Collections: []string{"umbra-sar"},
		BBox:        []float64{-120, 34, -119, 35},
	}

	var items []umbra.STACItem
	for item, err := range cli.SearchSTAC(context.Background(), req) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		items = append(items, item)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 features, got %d", len(items))
	}
}

func TestSearchSTACV2(t *testing.T) {
	expected := umbra.STACSearchResponse{
		Type: "FeatureCollection",
		Features: []umbra.STACItem{
			{ID: "item-v2", Collection: "umbra-sar"},
		},
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/v2/stac/search")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	req := umbra.STACSearchRequest{
		Limit: 10,
	}

	var items []umbra.STACItem
	for item, err := range cli.SearchSTACV2(context.Background(), req) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		items = append(items, item)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 feature, got %d", len(items))
	}
}

func TestSearchSTAC_WithCQL2Filter(t *testing.T) {
	expected := umbra.STACSearchResponse{
		Type: "FeatureCollection",
		Features: []umbra.STACItem{
			{ID: "filtered-item", Collection: "umbra-sar"},
		},
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/stac/search")
		jsonResponse(w, http.StatusOK, expected)
	})

	cql := umbra.CQL2{}
	req := umbra.STACSearchRequest{
		Limit:      10,
		FilterLang: "cql2-json",
		Filter: cql.And(
			cql.Equal("sar:resolution_range", 1),
			cql.GreaterThan("view:sun_elevation", 50),
		),
	}

	var items []umbra.STACItem
	for item, err := range cli.SearchSTAC(context.Background(), req) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		items = append(items, item)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 feature, got %d", len(items))
	}
}

func TestGetSTACCollection(t *testing.T) {
	expected := umbra.STACCollection{
		ID:          "umbra-sar",
		Type:        "Collection",
		Title:       "Umbra SAR Collection",
		Description: "SAR imagery from Umbra satellites",
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/stac/collections/umbra-sar")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	col, err := cli.GetSTACCollection(context.Background(), "umbra-sar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.ID != expected.ID {
		t.Errorf("expected ID %s, got %s", expected.ID, col.ID)
	}
	if col.Title != expected.Title {
		t.Errorf("expected title %s, got %s", expected.Title, col.Title)
	}
}

func TestListSTACCollections(t *testing.T) {
	expected := struct {
		Collections []umbra.STACCollection `json:"collections"`
	}{
		Collections: []umbra.STACCollection{
			{ID: "umbra-sar", Title: "Umbra SAR"},
			{ID: "umbra-archive", Title: "Umbra Archive"},
		},
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/stac/collections")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	cols, err := cli.ListSTACCollections(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cols) != 2 {
		t.Errorf("expected 2 collections, got %d", len(cols))
	}
}

func TestGetSTACItem_NotFound(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusNotFound, "Item not found")
	})

	_, err := cli.GetSTACItem(context.Background(), "umbra-sar", "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !umbra.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}
