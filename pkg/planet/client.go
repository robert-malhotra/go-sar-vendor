// Package planet provides a Go client for the Planet APIs.
//
// The Planet platform provides access to Planet's constellation of satellites
// for imagery tasking, feasibility assessment via imaging windows, and ordering
// of collected imagery for download or cloud delivery.
//
// Key features:
//   - API key authentication
//   - Full coverage of Tasking API v2, Imaging Windows API, and Orders API v2
//   - GeoJSON geometry support via paulmach/orb
//   - Idiomatic Go types with comprehensive type safety
//   - Thread-safe; safe for concurrent goroutines
//
// Docs: https://developers.planet.com/docs/apis/
package planet

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/robert.malhotra/go-sar-vendor/pkg/common"
)

const (
	// DefaultBaseURL is the default Planet API base URL.
	DefaultBaseURL = "https://api.planet.com"

	// TaskingBasePath is the base path for the Tasking API v2.
	TaskingBasePath = "/tasking/v2"

	// OrdersBasePath is the base path for the Orders API v2.
	OrdersBasePath = "/compute/ops/orders/v2"

	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "go-sar-vendor/planet"
)

// Client represents a Planet API client.
type Client struct {
	*common.Client
	taskingBaseURL *url.URL
	ordersBaseURL  *url.URL
}

// Option configures a Client.
type Option func(*clientConfig)

type clientConfig struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	userAgent  string
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

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(userAgent string) Option {
	return func(c *clientConfig) {
		c.userAgent = userAgent
	}
}

// NewClient creates a new Planet API client.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	cfg := &clientConfig{
		baseURL:   DefaultBaseURL,
		timeout:   defaultTimeout,
		userAgent: defaultUserAgent,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.timeout}
	}

	baseURL, err := url.Parse(cfg.baseURL)
	if err != nil {
		return nil, err
	}

	taskingBaseURL := baseURL.JoinPath(TaskingBasePath)
	ordersBaseURL := baseURL.JoinPath(OrdersBasePath)

	c, err := common.NewClient(common.ClientConfig{
		BaseURL:    cfg.baseURL,
		HTTPClient: httpClient,
		Auth:       newAPIKeyAuth(apiKey),
		UserAgent:  cfg.userAgent,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:         c,
		taskingBaseURL: taskingBaseURL,
		ordersBaseURL:  ordersBaseURL,
	}, nil
}

// TaskingURL returns the full URL for a tasking API path.
func (c *Client) TaskingURL(path ...string) *url.URL {
	return c.taskingBaseURL.JoinPath(path...)
}

// OrdersURL returns the full URL for an orders API path.
func (c *Client) OrdersURL(path ...string) *url.URL {
	return c.ordersBaseURL.JoinPath(path...)
}

// doRequest performs an HTTP request and decodes the response.
func (c *Client) doRequest(ctx context.Context, method string, u *url.URL, body io.Reader, want int, out any) error {
	return c.Client.DoRaw(ctx, method, u, body, want, out)
}

// apiKeyAuth implements the common.Authenticator interface for Planet API key authentication.
type apiKeyAuth struct {
	apiKey string
}

func newAPIKeyAuth(apiKey string) *apiKeyAuth {
	return &apiKeyAuth{apiKey: apiKey}
}

// Apply applies authentication to the HTTP request.
func (a *apiKeyAuth) Apply(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", "api-key "+a.apiKey)
	return nil
}

// marshalBody marshals v to JSON and returns a bytes.Buffer.
var marshalBody = common.MarshalBody
