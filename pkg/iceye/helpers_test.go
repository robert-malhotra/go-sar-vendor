package iceye_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/robert-malhotra/go-sar-vendor/pkg/iceye"
)

// newTestClient creates a mock auth+API server pair and configured client.
func newTestClient(t *testing.T, handler func(mux *http.ServeMux, authHits *atomic.Int32)) (*iceye.Client, *httptest.Server, *atomic.Int32) {
	t.Helper()
	authHits := &atomic.Int32{}
	mux := http.NewServeMux()
	handler(mux, authHits)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli, err := iceye.NewClient(
		iceye.WithBaseURL(srv.URL),
		iceye.WithTokenURL(srv.URL+"/oauth2/token"),
		iceye.WithHTTPClient(srv.Client()),
		iceye.WithCredentials("test", "secret"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return cli, srv, authHits
}

// mockAuthHandler returns a handler that issues a test token.
func mockAuthHandler(authHits *atomic.Int32) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if authHits != nil {
			authHits.Add(1)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}
}

// mockAuthHandlerShortExpiry returns a handler that issues a token with short expiry.
func mockAuthHandlerShortExpiry(authHits *atomic.Int32) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if authHits != nil {
			authHits.Add(1)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-token",
			"token_type":   "Bearer",
			"expires_in":   1,
		})
	}
}
