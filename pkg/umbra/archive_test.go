package umbra_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/robert.malhotra/go-sar-vendor/pkg/umbra"
)

func TestSearchArchive(t *testing.T) {
	expected := umbra.STACSearchResponse{
		Type: "FeatureCollection",
		Features: []umbra.STACItem{
			{ID: "archive-1", Collection: "umbra-archive"},
			{ID: "archive-2", Collection: "umbra-archive"},
		},
		NumberMatched: 2,
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/archive/search")
		requireAuth(t, r, "test-token")
		requireContentType(t, r, "application/json")
		jsonResponse(w, http.StatusOK, expected)
	})

	req := umbra.ArchiveSearchRequest{
		Limit:       10,
		Collections: []string{"umbra-archive"},
		BBox:        []float64{-120, 34, -119, 35},
		Datetime:    "2024-01-01T00:00:00Z/2024-12-31T23:59:59Z",
	}

	var items []umbra.STACItem
	for item, err := range cli.SearchArchive(context.Background(), req) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		items = append(items, item)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 features, got %d", len(items))
	}
}

func TestSearchArchive_WithCQL2(t *testing.T) {
	expected := umbra.STACSearchResponse{
		Type: "FeatureCollection",
		Features: []umbra.STACItem{
			{ID: "filtered-archive", Collection: "umbra-archive"},
		},
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/archive/search")
		jsonResponse(w, http.StatusOK, expected)
	})

	cql := umbra.CQL2{}
	req := umbra.ArchiveSearchRequest{
		Limit:      10,
		FilterLang: "cql2-json",
		Filter: cql.And(
			cql.Equal("sar:polarizations", "VV"),
			cql.LessThan("sar:resolution_range", 2),
		),
	}

	var items []umbra.STACItem
	for item, err := range cli.SearchArchive(context.Background(), req) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		items = append(items, item)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 feature, got %d", len(items))
	}
}

func TestGetArchiveThumbnail(t *testing.T) {
	expectedData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG magic bytes

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/archive/thumbnail/archive-123")
		requireAuth(t, r, "test-token")
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write(expectedData)
	})

	data, err := cli.GetArchiveThumbnail(context.Background(), "archive-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) != len(expectedData) {
		t.Errorf("expected %d bytes, got %d", len(expectedData), len(data))
	}
}

func TestGetArchiveCollectionItems(t *testing.T) {
	expected := umbra.STACSearchResponse{
		Type: "FeatureCollection",
		Features: []umbra.STACItem{
			{ID: "item-1", Collection: "my-collection"},
			{ID: "item-2", Collection: "my-collection"},
		},
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/archive/collections/my-collection/items")
		requireAuth(t, r, "test-token")

		// Verify query parameters
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Errorf("expected limit=10, got %s", got)
		}
		if got := r.URL.Query().Get("offset"); got != "5" {
			t.Errorf("expected offset=5, got %s", got)
		}

		jsonResponse(w, http.StatusOK, expected)
	})

	resp, err := cli.GetArchiveCollectionItems(context.Background(), "my-collection", 10, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Features) != 2 {
		t.Errorf("expected 2 features, got %d", len(resp.Features))
	}
}

func TestGetArchiveItem(t *testing.T) {
	expected := umbra.STACItem{
		ID:         "archive-item-123",
		Type:       "Feature",
		Collection: "umbra-archive",
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/archive/items/archive-item-123")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	item, err := cli.GetArchiveItem(context.Background(), "archive-item-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != expected.ID {
		t.Errorf("expected ID %s, got %s", expected.ID, item.ID)
	}
}

func TestSearchArchive_Error(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusBadRequest, "Invalid search request")
	})

	req := umbra.ArchiveSearchRequest{}

	for _, err := range cli.SearchArchive(context.Background(), req) {
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !umbra.IsBadRequest(err) {
			t.Errorf("expected bad request error, got %v", err)
		}
		break
	}
}

func TestGetArchiveThumbnail_NotFound(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	})

	_, err := cli.GetArchiveThumbnail(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !umbra.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}
