package capella_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/robert.malhotra/go-sar-vendor/pkg/capella"
)

func TestNewClient_Defaults(t *testing.T) {
	cli := capella.NewClient()
	if cli == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	customHTTPClient := &http.Client{Timeout: 60 * time.Second}

	cli := capella.NewClient(
		capella.WithBaseURL("https://custom.api.com"),
		capella.WithAPIKey("my-api-key"),
		capella.WithHTTPClient(customHTTPClient),
		capella.WithUserAgent("test-agent/1.0"),
	)

	if cli == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestClient_RequestHeaders(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requireAuth(t, r, "test-api-key")

		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept header 'application/json', got %q", r.Header.Get("Accept"))
		}

		jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
	}

	cli, _ := newTestClient(t, handler)

	// Make a request to verify headers
	catalog := capella.NewCatalogService(cli)
	_, err := catalog.ListCollections(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantStatus int
	}{
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			body:       `{"code": "NOT_FOUND", "message": "Resource not found"}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{"code": "UNAUTHORIZED", "message": "Invalid API key"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "internal error",
			statusCode: http.StatusInternalServerError,
			body:       `{"code": "INTERNAL_ERROR", "message": "Something went wrong"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}

			cli, _ := newTestClient(t, handler)
			catalog := capella.NewCatalogService(cli)

			_, err := catalog.ListCollections(t.Context())
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			apiErr, ok := err.(*capella.APIError)
			if !ok {
				t.Fatalf("expected *capella.APIError, got %T", err)
			}

			if apiErr.StatusCode != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, apiErr.StatusCode)
			}
		})
	}
}

func TestClient_ValidationError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		validationError(w, "windowOpen", "field required")
	}

	cli, _ := newTestClient(t, handler)
	tasking := capella.NewTaskingService(cli)

	_, err := tasking.CreateTask(t.Context(), capella.TaskingRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*capella.APIError)
	if !ok {
		t.Fatalf("expected *capella.APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", apiErr.StatusCode)
	}

	// Check that error message contains validation info
	if apiErr.Message == "" {
		t.Error("expected validation message in error")
	}
}

func TestIsNotFound(t *testing.T) {
	err := &capella.APIError{StatusCode: http.StatusNotFound}
	if !capella.IsNotFound(err) {
		t.Error("expected IsNotFound to return true")
	}

	err = &capella.APIError{StatusCode: http.StatusOK}
	if capella.IsNotFound(err) {
		t.Error("expected IsNotFound to return false")
	}
}

func TestIsValidationError(t *testing.T) {
	err := &capella.APIError{StatusCode: http.StatusUnprocessableEntity}
	if !capella.IsValidationError(err) {
		t.Error("expected IsValidationError to return true")
	}

	err = &capella.APIError{StatusCode: http.StatusOK}
	if capella.IsValidationError(err) {
		t.Error("expected IsValidationError to return false")
	}
}

func TestIsUnauthorized(t *testing.T) {
	err := &capella.APIError{StatusCode: http.StatusUnauthorized}
	if !capella.IsUnauthorized(err) {
		t.Error("expected IsUnauthorized to return true")
	}

	err = &capella.APIError{StatusCode: http.StatusOK}
	if capella.IsUnauthorized(err) {
		t.Error("expected IsUnauthorized to return false")
	}
}

func TestIsRateLimited(t *testing.T) {
	err := &capella.APIError{StatusCode: http.StatusTooManyRequests}
	if !capella.IsRateLimited(err) {
		t.Error("expected IsRateLimited to return true")
	}

	err = &capella.APIError{StatusCode: http.StatusOK}
	if capella.IsRateLimited(err) {
		t.Error("expected IsRateLimited to return false")
	}
}
