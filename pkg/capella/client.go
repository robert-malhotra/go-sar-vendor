// Package capella provides a comprehensive Go SDK for the Capella Space API.
//
// The Capella Space platform provides access to the world's first commercial SAR
// (Synthetic Aperture Radar) satellite constellation, offering high-resolution
// Earth observation imagery through a REST API.
//
// Key features:
//   - API key authentication (recommended) with optional JWT token support.
//   - Full coverage of Catalog, Tasking, Orders, and Access APIs.
//   - STAC (Spatio-Temporal Asset Catalog) compliant catalog search.
//   - Idiomatic Go types with comprehensive type safety.
//   - Lazy pagination helpers using Go 1.23+ `iter.Seq2`.
//   - Thread-safe; safe for concurrent goroutines.
//
// Docs: https://docs.capellaspace.com/
package capella

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

const (
	defaultBaseURL = "https://api.capellaspace.com"
	defaultTimeout = 30 * time.Second
)

// Client is the Capella Space API client. It is thread-safe.
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	userAgent  string
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// WithBaseURL sets the base URL for API requests.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithAPIKey sets the API key for authentication.
func WithAPIKey(key string) ClientOption {
	return func(c *Client) {
		c.apiKey = key
	}
}

// WithUserAgent sets the User-Agent header for API requests.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// NewClient creates a new Capella Space API client.
// It uses sensible defaults which can be overridden with functional options.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// newRequest creates an http.Request with the necessary headers.
func (c *Client) newRequest(ctx context.Context, apiKey, method, endpoint string, body []byte) (*http.Request, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}
	u.Path = path.Join(u.Path, endpoint)

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Use provided apiKey, fallback to client's stored key
	key := apiKey
	if key == "" {
		key = c.apiKey
	}
	if key != "" {
		req.Header.Set("Authorization", "ApiKey "+key)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	return req, nil
}

// do sends an API request and handles the response.
func (c *Client) do(req *http.Request, v any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseError(resp)
	}

	if v != nil {
		if err = json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
