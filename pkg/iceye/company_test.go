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

func TestListContractsPagination(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/company/v1/contracts", func(w http.ResponseWriter, r *http.Request) {
			cursor := r.URL.Query().Get("cursor")
			switch cursor {
			case "":
				json.NewEncoder(w).Encode(iceye.ContractsResponse{
					Data:   []iceye.Contract{{ID: "c1", Name: "Contract 1"}},
					Cursor: "next",
				})
			case "next":
				json.NewEncoder(w).Encode(iceye.ContractsResponse{
					Data:   []iceye.Contract{{ID: "c2", Name: "Contract 2"}},
					Cursor: "",
				})
			default:
				t.Fatalf("unexpected cursor: %s", cursor)
			}
		})
	})

	var ids []string
	for resp, err := range cli.ListContracts(context.Background(), 1) {
		require.NoError(t, err)
		for _, c := range resp.Data {
			ids = append(ids, c.ID)
		}
	}

	assert.Equal(t, []string{"c1", "c2"}, ids)
}

func TestGetContract(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/company/v1/contracts/C-123", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			json.NewEncoder(w).Encode(iceye.Contract{
				ID:   "C-123",
				Name: "Test Contract",
			})
		})
	})

	contract, err := cli.GetContract(context.Background(), "C-123")

	require.NoError(t, err)
	assert.Equal(t, "C-123", contract.ID)
	assert.Equal(t, "Test Contract", contract.Name)
}

func TestGetSummary(t *testing.T) {
	cli, _, _ := newTestClient(t, func(mux *http.ServeMux, _ *atomic.Int32) {
		mux.HandleFunc("/oauth2/token", mockAuthHandler(nil))
		mux.HandleFunc("/company/v1/contracts/C-123/summary", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			json.NewEncoder(w).Encode(iceye.Summary{
				ContractID:        "C-123",
				ConsolidatedSpent: 50000,
				Currency:          "USD",
				OnHold:            10000,
				SpendLimit:        100000,
			})
		})
	})

	summary, err := cli.GetSummary(context.Background(), "C-123")

	require.NoError(t, err)
	assert.Equal(t, "C-123", summary.ContractID)
	assert.Equal(t, int64(100000), summary.SpendLimit)
	assert.Equal(t, "USD", summary.Currency)
	assert.Equal(t, int64(50000), summary.ConsolidatedSpent)
	assert.Equal(t, int64(10000), summary.OnHold)
}
