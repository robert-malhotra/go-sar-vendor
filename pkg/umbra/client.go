// Package umbra provides a Go client for the Umbra Canopy API.
//
// The Umbra Canopy platform provides self-service access to Umbra's constellation
// of synthetic aperture radar (SAR) satellites, enabling customers to task satellites
// for new data collection, check feasibility of imaging requests, track task lifecycle,
// manage delivery configurations, and access collected imagery via STAC-compliant endpoints.
//
// Key features:
//   - Bearer token authentication
//   - Full coverage of Tasking, Feasibility, Collects, Delivery, STAC, and Archive APIs
//   - STAC (Spatio-Temporal Asset Catalog) compliant catalog search
//   - CQL2 filter builder for advanced queries
//   - Idiomatic Go types with comprehensive type safety
//   - Thread-safe; safe for concurrent goroutines
//
// Docs: https://docs.canopy.umbra.space/
package umbra

import (
	"net/http"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

const (
	// ProductionBaseURL is the production API endpoint.
	ProductionBaseURL = "https://api.canopy.umbra.space"
	// SandboxBaseURL is the sandbox API endpoint for testing.
	SandboxBaseURL = "https://api.canopy.prod.umbra-sandbox.space"

	defaultTimeout = 30 * time.Second
)

// Client represents a Canopy API client.
type Client struct {
	*common.Client
}

// Option configures a Client.
type Option func(*clientConfig)

type clientConfig struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *clientConfig) {
		c.httpClient = client
	}
}

// WithBaseURL overrides the default base URL.
func WithBaseURL(rawURL string) Option {
	return func(c *clientConfig) {
		c.baseURL = rawURL
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *clientConfig) {
		c.timeout = timeout
	}
}

// NewClient creates a new Canopy API client configured for production.
func NewClient(accessToken string, opts ...Option) (*Client, error) {
	cfg := &clientConfig{
		baseURL: ProductionBaseURL,
		timeout: defaultTimeout,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	httpClient := common.EnsureHTTPClient(cfg.httpClient, cfg.timeout)

	c, err := common.NewClient(common.ClientConfig{
		BaseURL:    cfg.baseURL,
		HTTPClient: httpClient,
		Auth:       common.NewBearerAuth(accessToken),
	})
	if err != nil {
		return nil, err
	}

	return &Client{Client: c}, nil
}

// NewSandboxClient creates a new Canopy API client configured for the sandbox environment.
func NewSandboxClient(accessToken string, opts ...Option) (*Client, error) {
	cfg := &clientConfig{
		baseURL: SandboxBaseURL,
		timeout: defaultTimeout,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	httpClient := common.EnsureHTTPClient(cfg.httpClient, cfg.timeout)

	c, err := common.NewClient(common.ClientConfig{
		BaseURL:    cfg.baseURL,
		HTTPClient: httpClient,
		Auth:       common.NewBearerAuth(accessToken),
	})
	if err != nil {
		return nil, err
	}

	return &Client{Client: c}, nil
}

