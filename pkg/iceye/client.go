// Package iceye provides a comprehensive Go SDK for the ICEYE API Platform.
//
// The ICEYE platform enables direct access to the world's largest SAR (Synthetic
// Aperture Radar) satellite constellation, providing capabilities for satellite
// tasking, catalog browsing, and purchasing archived imagery.
//
// Key features:
//   - OAuth 2 Client-Credentials token management with auto-refresh.
//   - Full coverage of Company, Tasking, and Catalog APIs.
//   - Idiomatic Go types mirroring ICEYE API domains.
//   - DRY lazy pagination helpers using Go 1.23+ `iter.Seq2`.
//   - Zero non-std-lib runtime dependencies.
//   - Thread-safe; safe for concurrent goroutines.
//   - RFC 7807 errors mapped to *iceye.Error.
//
// Version history
//
//	v0.1.0  – initial skeleton (contracts, basic tasking)
//	v1.0.0  – full Tasking API coverage
//	v2.0.0  – full API coverage (Company, Tasking, Catalog) with unified client
//	v3.0.0  – aligned with ICEYE v1 APIs (Company, Tasking, Catalog, Delivery)
//
// Docs: https://docs.iceye.com/api/ (December 2024 release)
// ----------------------------------------------------------------------------
// Quick example (Go 1.23+)
// ----------------------------------------------------------------------------
//
//	client := iceye.NewClient(iceye.Config{
//	    BaseURL:      "https://api.iceye.com",
//	    TokenURL:     "https://auth.iceye.com/oauth2/token",
//	    ClientID:     "your-client-id",
//	    ClientSecret: "your-client-secret",
//	})
//
//	// List contracts
//	for resp, err := range client.ListContracts(ctx, 100) {
//	    if err != nil { log.Fatal(err) }
//	    for _, c := range resp.Data {
//	        log.Printf("Contract: %s (%s)", c.Name, c.ID)
//	    }
//	}
package iceye

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// ----------------------------------------------------------------------------
// Config & Client
// ----------------------------------------------------------------------------

// Config holds the configuration for the ICEYE API client.
type Config struct {
	// BaseURL is the API base URL (default: https://api.iceye.com)
	BaseURL string

	// TokenURL is the OAuth2 token endpoint URL
	TokenURL string

	// ClientID for OAuth2 Client Credentials flow
	ClientID string

	// ClientSecret for OAuth2 Client Credentials flow
	ClientSecret string

	// ResourceOwner auth (legacy) - if set, uses password grant instead
	ResourceOwner *ResourceOwnerAuth

	// HTTPClient allows custom HTTP client (for testing, retries, etc.)
	HTTPClient *http.Client

	// UserAgent for API requests (optional)
	UserAgent string

	// Timeout for API requests (optional, defaults to 30s)
	Timeout time.Duration
}

// ResourceOwnerAuth contains credentials for the legacy password grant flow.
type ResourceOwnerAuth struct {
	APIKey   string // Base64-encoded client credentials
	Username string
	Password string
}

// Client is the ICEYE API client. It is thread-safe.
type Client struct {
	cfg   Config
	mu    sync.Mutex // guards token+exp
	token string
	exp   time.Time
}

// NewClient creates a new ICEYE API client with the given configuration.
func NewClient(cfg Config) *Client {
	if cfg.HTTPClient == nil {
		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		cfg.HTTPClient = &http.Client{Timeout: timeout}
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://platform.iceye.com/api"
	}
	return &Client{cfg: cfg}
}

// ----------------------------------------------------------------------------
// Token management (OAuth2 Client-Credentials / Resource Owner Password)
// ----------------------------------------------------------------------------

func (c *Client) authenticate(ctx context.Context) error {
	c.mu.Lock()
	if time.Until(c.exp) > 30*time.Second {
		c.mu.Unlock()
		return nil // still valid
	}
	c.mu.Unlock()

	var body []byte
	var authHeader string

	if c.cfg.ResourceOwner != nil {
		// Resource Owner Password flow (legacy)
		body = []byte(fmt.Sprintf("grant_type=password&username=%s&password=%s",
			url.QueryEscape(c.cfg.ResourceOwner.Username),
			url.QueryEscape(c.cfg.ResourceOwner.Password)))
		authHeader = "Basic " + c.cfg.ResourceOwner.APIKey
	} else {
		// Client Credentials flow (recommended)
		body = []byte("grant_type=client_credentials")
		b64 := base64.StdEncoding.EncodeToString([]byte(c.cfg.ClientID + ":" + c.cfg.ClientSecret))
		authHeader = "Basic " + b64
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.TokenURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json, application/problem+json")
	if c.cfg.UserAgent != "" {
		req.Header.Set("User-Agent", c.cfg.UserAgent)
	}

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return parseError(resp)
	}

	var tok struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return err
	}
	if tok.AccessToken == "" {
		return errors.New("iceye: empty access_token")
	}

	c.mu.Lock()
	c.token = tok.AccessToken
	c.exp = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
	c.mu.Unlock()
	return nil
}

// do wraps HTTP with auth + JSON encode/decode.
func (c *Client) do(ctx context.Context, method, path string, in any, out any) error {
	if err := c.authenticate(ctx); err != nil {
		return err
	}

	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json, application/problem+json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.cfg.UserAgent != "" {
		req.Header.Set("User-Agent", c.cfg.UserAgent)
	}

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return parseError(resp)
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Generic pagination helper
// ----------------------------------------------------------------------------

// paginate generates an `iter.Seq2` for endpoints following ICEYE's common
// `{ "data": [...], "cursor": "..." }` pattern.
func paginate[T any](fetch func(cur *string) ([]T, *string, error)) iter.Seq2[[]T, error] {
	return func(yield func([]T, error) bool) {
		var cur *string
		for {
			data, next, err := fetch(cur)
			if !yield(data, err) {
				return
			}
			if err != nil || next == nil || *next == "" {
				return
			}
			cur = next
		}
	}
}
