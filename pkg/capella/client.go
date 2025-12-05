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
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

const (
	defaultBaseURL = "https://api.capellaspace.com"
	defaultTimeout = 30 * time.Second
)

// Client is the Capella Space API client. It is thread-safe.
type Client struct {
	*common.Client
}

// clientConfig holds configuration for building a Client.
type clientConfig struct {
	baseURL    string
	httpClient *http.Client
	auth       common.Authenticator
	userAgent  string
	timeout    time.Duration
}

// ClientOption is a function that configures a Client.
type ClientOption func(*clientConfig)

// WithBaseURL sets the base URL for API requests.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *clientConfig) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *clientConfig) {
		c.httpClient = httpClient
	}
}

// WithAPIKey sets the API key for authentication.
func WithAPIKey(key string) ClientOption {
	return func(c *clientConfig) {
		c.auth = common.NewAPIKeyAuth(key)
	}
}

// WithAuth sets a custom authenticator.
func WithAuth(auth common.Authenticator) ClientOption {
	return func(c *clientConfig) {
		c.auth = auth
	}
}

// WithUserAgent sets the User-Agent header for API requests.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *clientConfig) {
		c.userAgent = userAgent
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientConfig) {
		c.timeout = timeout
	}
}

// NewClient creates a new Capella Space API client.
// It uses sensible defaults which can be overridden with functional options.
func NewClient(opts ...ClientOption) *Client {
	cfg := &clientConfig{
		baseURL: defaultBaseURL,
		timeout: defaultTimeout,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.timeout}
	}

	c, _ := common.NewClient(common.ClientConfig{
		BaseURL:    cfg.baseURL,
		HTTPClient: httpClient,
		Auth:       cfg.auth,
		UserAgent:  cfg.userAgent,
	})

	return &Client{Client: c}
}

// doRequest performs an HTTP request and decodes the response.
func (c *Client) doRequest(ctx context.Context, method string, u *url.URL, body io.Reader, want int, out any) error {
	return c.Client.DoRaw(ctx, method, u, body, want, out)
}

// doRequestRaw performs an HTTP request and returns the raw response body.
func (c *Client) doRequestRaw(ctx context.Context, method string, u *url.URL, body io.Reader, want int) ([]byte, error) {
	return c.Client.DoRawResponse(ctx, method, u, body, want)
}

// marshalBody marshals v to JSON and returns a bytes.Buffer.
var marshalBody = common.MarshalBody
