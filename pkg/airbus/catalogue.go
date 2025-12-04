package airbus

import (
	"context"
	"fmt"
	"net/http"
)

// SearchCatalogue searches for existing acquisitions in the archive.
// POST /sar/catalogue
func (c *Client) SearchCatalogue(ctx context.Context, req *CatalogueRequest) (*FeatureCollection, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var out FeatureCollection
	err = c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "catalogue"), body, http.StatusOK, &out)
	return &out, err
}

// ReplicateCatalogue retrieves catalogue updates for replication.
// This endpoint is for bulk catalogue synchronization.
// GET /sar/catalogue/replication
func (c *Client) ReplicateCatalogue(ctx context.Context, opts *ReplicationOptions) (*FeatureCollection, error) {
	u := c.BaseURL().JoinPath("sar", "catalogue", "replication")
	if opts != nil {
		q := u.Query()
		if !opts.Since.IsZero() {
			q.Set("since", opts.Since.Format("2006-01-02T15:04:05Z"))
		}
		if opts.Limit > 0 {
			q.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		u.RawQuery = q.Encode()
	}
	var out FeatureCollection
	err := c.doRequest(ctx, http.MethodGet, u, nil, http.StatusOK, &out)
	return &out, err
}

// GetCatalogueRevocations retrieves IDs of revoked catalogue items.
// Items may be revoked due to quality issues or reprocessing.
// GET /sar/catalogue/revocation
func (c *Client) GetCatalogueRevocations(ctx context.Context, opts *RevocationOptions) (*RevocationResponse, error) {
	u := c.BaseURL().JoinPath("sar", "catalogue", "revocation")
	if opts != nil {
		q := u.Query()
		if !opts.Since.IsZero() {
			q.Set("since", opts.Since.Format("2006-01-02T15:04:05Z"))
		}
		if opts.Limit > 0 {
			q.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		u.RawQuery = q.Encode()
	}
	var out RevocationResponse
	err := c.doRequest(ctx, http.MethodGet, u, nil, http.StatusOK, &out)
	return &out, err
}

// RetrieveCatalogueItems retrieves ordered items from the catalogue.
// This returns the updated state of previously ordered items.
// POST /sar/catalogue/retrieve
func (c *Client) RetrieveCatalogueItems(ctx context.Context, req *RetrieveRequest) (*FeatureCollection, error) {
	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}
	var out FeatureCollection
	err = c.doRequest(ctx, http.MethodPost, c.BaseURL().JoinPath("sar", "catalogue", "retrieve"), body, http.StatusOK, &out)
	return &out, err
}
