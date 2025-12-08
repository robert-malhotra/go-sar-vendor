package capella_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robert-malhotra/go-sar-vendor/pkg/capella"
)

// newTestClient creates a mock server and configured client for testing.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*capella.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cli, err := capella.NewClient(
		capella.WithBaseURL(srv.URL),
		capella.WithAPIKey("test-api-key"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return cli, srv
}

// jsonResponse writes a JSON response to the response writer.
func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// validationError writes a 422 validation error response.
func validationError(w http.ResponseWriter, field, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{
		"detail": []map[string]any{
			{
				"loc":  []string{"body", "properties", field},
				"msg":  message,
				"type": "value_error.missing",
			},
		},
	})
}

// requireMethod checks that the request method matches.
func requireMethod(t *testing.T, r *http.Request, expected string) {
	t.Helper()
	if r.Method != expected {
		t.Fatalf("expected method %s, got %s", expected, r.Method)
	}
}

// requirePath checks that the request path matches.
func requirePath(t *testing.T, r *http.Request, expected string) {
	t.Helper()
	if r.URL.Path != expected {
		t.Fatalf("expected path %s, got %s", expected, r.URL.Path)
	}
}

// requireAuth checks that the request has the expected API key.
func requireAuth(t *testing.T, r *http.Request, expectedKey string) {
	t.Helper()
	expected := "ApiKey " + expectedKey
	got := r.Header.Get("Authorization")
	if got != expected {
		t.Fatalf("expected Authorization header %q, got %q", expected, got)
	}
}

// requireContentType checks that the request has the expected content type.
func requireContentType(t *testing.T, r *http.Request, expected string) {
	t.Helper()
	got := r.Header.Get("Content-Type")
	if got != expected {
		t.Fatalf("expected Content-Type %q, got %q", expected, got)
	}
}
