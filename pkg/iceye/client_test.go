package iceye_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/iceye"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticateSuccess(t *testing.T) {
	authHits := &atomic.Int32{}
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		authHits.Add(1)
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "abc123",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"scope":        "catalog.read tasking.write",
		})
	})
	mux.HandleFunc("/company/v1/contracts", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(iceye.ContractsResponse{})
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := iceye.NewClient(iceye.Config{
		BaseURL:      srv.URL,
		TokenURL:     srv.URL + "/oauth2/token",
		ClientID:     "test",
		ClientSecret: "secret",
		HTTPClient:   srv.Client(),
	})

	// First request should trigger auth
	ctx := context.Background()
	for range cli.ListContracts(ctx, 1) {
		break // Just trigger one request
	}
	assert.Equal(t, int32(1), authHits.Load(), "auth endpoint should be hit exactly once")

	// Second request should reuse token (within 30s window)
	for range cli.ListContracts(ctx, 1) {
		break
	}
	assert.Equal(t, int32(1), authHits.Load(), "token should not be refreshed")
}

func TestAuthenticateResourceOwner(t *testing.T) {
	authHits := &atomic.Int32{}
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		authHits.Add(1)
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "password", r.FormValue("grant_type"))
		assert.Equal(t, "testuser", r.FormValue("username"))
		assert.Equal(t, "testpass", r.FormValue("password"))
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "legacy-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})
	mux.HandleFunc("/company/v1/contracts", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(iceye.ContractsResponse{})
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := iceye.NewClient(iceye.Config{
		BaseURL:  srv.URL,
		TokenURL: srv.URL + "/oauth2/token",
		ResourceOwner: &iceye.ResourceOwnerAuth{
			APIKey:   "dGVzdDpzZWNyZXQ=", // base64("test:secret")
			Username: "testuser",
			Password: "testpass",
		},
		HTTPClient: srv.Client(),
	})

	ctx := context.Background()
	for range cli.ListContracts(ctx, 1) {
		break
	}
	assert.Equal(t, int32(1), authHits.Load())
}

func TestDoParsesProblemJSON(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/company/v1/contracts/bad", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"code":   "ERR_BAD",
				"detail": "kaput",
			})
		})
	})

	ctx := context.Background()
	_, err := cli.GetContract(ctx, "bad")

	require.Error(t, err)
	apiErr, ok := err.(*iceye.Error)
	require.True(t, ok, "error should be *iceye.Error")
	assert.Equal(t, "ERR_BAD", apiErr.Code)
	assert.Equal(t, "kaput", apiErr.Detail)
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
}

func TestTokenRefreshAfterExpiry(t *testing.T) {
	tokenCalls := &atomic.Int32{}
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", mockAuthHandlerShortExpiry(tokenCalls))
	mux.HandleFunc("/company/v1/contracts", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(iceye.ContractsResponse{})
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := iceye.NewClient(iceye.Config{
		BaseURL:      srv.URL,
		TokenURL:     srv.URL + "/oauth2/token",
		ClientID:     "test",
		ClientSecret: "secret",
		HTTPClient:   srv.Client(),
	})

	ctx := context.Background()

	// First request
	for range cli.ListContracts(ctx, 1) {
		break
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Second request should trigger re-auth
	for range cli.ListContracts(ctx, 1) {
		break
	}

	assert.GreaterOrEqual(t, tokenCalls.Load(), int32(2), "token should be refreshed after expiry")
}

func TestGeoJSONHelpers(t *testing.T) {
	point := iceye.GeoJSONPoint(24.9384, 60.1699)
	assert.Equal(t, "Point", point.Type)
	// orb geometry stores coordinates differently
	assert.NotNil(t, point.Geometry())

	bbox := iceye.BoundingBox{24.0, 59.5, 25.5, 60.5}
	polygon := iceye.BBoxToPolygon(bbox)
	assert.Equal(t, "Polygon", polygon.Type)
}
