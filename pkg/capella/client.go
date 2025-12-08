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
	"net/http"
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

// Option is a function that configures a Client.
type Option func(*clientConfig)

// WithBaseURL sets the base URL for API requests.
func WithBaseURL(baseURL string) Option {
	return func(c *clientConfig) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *clientConfig) {
		c.httpClient = httpClient
	}
}

// WithAPIKey sets the API key for authentication.
func WithAPIKey(key string) Option {
	return func(c *clientConfig) {
		c.auth = common.NewAPIKeyAuth(key)
	}
}

// WithAuth sets a custom authenticator.
func WithAuth(auth common.Authenticator) Option {
	return func(c *clientConfig) {
		c.auth = auth
	}
}

// WithUserAgent sets the User-Agent header for API requests.
func WithUserAgent(userAgent string) Option {
	return func(c *clientConfig) {
		c.userAgent = userAgent
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *clientConfig) {
		c.timeout = timeout
	}
}

// NewClient creates a new Capella Space API client.
// It uses sensible defaults which can be overridden with functional options.
func NewClient(opts ...Option) (*Client, error) {
	cfg := &clientConfig{
		baseURL: defaultBaseURL,
		timeout: defaultTimeout,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	httpClient := common.EnsureHTTPClient(cfg.httpClient, cfg.timeout)

	c, err := common.NewClient(common.ClientConfig{
		BaseURL:    cfg.baseURL,
		HTTPClient: httpClient,
		Auth:       cfg.auth,
		UserAgent:  cfg.userAgent,
	})
	if err != nil {
		return nil, err
	}

	return &Client{Client: c}, nil
}

