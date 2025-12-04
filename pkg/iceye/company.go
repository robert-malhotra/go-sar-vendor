package iceye

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

// ----------------------------------------------------------------------------
// Company API Types
// ----------------------------------------------------------------------------

// Contract represents an ICEYE contract.
type Contract struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	Start              time.Time          `json:"start"`
	End                time.Time          `json:"end"`
	DeliveryLocations  []DeliveryLocation `json:"deliveryLocations,omitempty"`
	ImagingModes       *OptionConfig      `json:"imagingModes,omitempty"`
	Priority           *OptionConfig      `json:"priority,omitempty"`
	Exclusivity        *OptionConfig      `json:"exclusivity,omitempty"`
	SLA                *OptionConfig      `json:"sla,omitempty"`
	EULA               *OptionConfig      `json:"eula,omitempty"`
	CatalogCollections *OptionConfig      `json:"catalogCollections,omitempty"`
}

// OptionConfig specifies allowed values and default for an option.
type OptionConfig struct {
	Allowed []string `json:"allowed"`
	Default string   `json:"default"`
}

// Summary represents contract budget information.
type Summary struct {
	ContractID        string `json:"contractID"`
	ConsolidatedSpent int64  `json:"consolidatedSpent"` // Minor currency unit
	Currency          string `json:"currency"`
	OnHold            int64  `json:"onHold"`
	SpendLimit        int64  `json:"spendLimit,omitempty"`
}

// ContractsResponse is the paginated response for listing contracts.
type ContractsResponse struct {
	Data   []Contract `json:"data"`
	Cursor string     `json:"cursor,omitempty"`
}

// ----------------------------------------------------------------------------
// Company API Methods
// Endpoints: https://docs.iceye.com/constellation/api/1.0/
// ----------------------------------------------------------------------------

const companyBasePath = "/company/v1"

// ListContracts retrieves all contracts for the authenticated company.
// Returns an iterator that yields pages of contracts.
//
// GET /company/v1/contracts
func (c *Client) ListContracts(ctx context.Context, pageSize int) iter.Seq2[ContractsResponse, error] {
	return func(yield func(ContractsResponse, error) bool) {
		seq := paginate[Contract](func(cur *string) ([]Contract, *string, error) {
			u := &url.URL{Path: path.Join(companyBasePath, "contracts")}
			q := u.Query()
			if pageSize > 0 {
				q.Set("limit", strconv.Itoa(pageSize))
			}
			if cur != nil && *cur != "" {
				q.Set("cursor", *cur)
			}
			u.RawQuery = q.Encode()

			var resp ContractsResponse
			err := c.do(ctx, http.MethodGet, u.String(), nil, &resp)
			return resp.Data, &resp.Cursor, err
		})
		for data, err := range seq {
			if !yield(ContractsResponse{Data: data}, err) {
				return
			}
		}
	}
}

// GetContract retrieves a specific contract by ID.
//
// GET /company/v1/contracts/{contractID}
func (c *Client) GetContract(ctx context.Context, contractID string) (*Contract, error) {
	var resp Contract
	u := &url.URL{Path: path.Join(companyBasePath, "contracts", contractID)}
	if err := c.do(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSummary retrieves budget summary for a contract.
//
// GET /company/v1/contracts/{contractID}/summary
func (c *Client) GetSummary(ctx context.Context, contractID string) (*Summary, error) {
	var resp Summary
	u := &url.URL{Path: path.Join(companyBasePath, "contracts", contractID, "summary")}
	if err := c.do(ctx, http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
