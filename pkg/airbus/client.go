package airbus

import (
	"net/http"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

const (
	// DefaultBaseURL is the production OneAtlas base URL.
	DefaultBaseURL = "https://sar.api.oneatlas.airbus.com/v1"
	// DefaultTokenURL is the production authentication endpoint.
	DefaultTokenURL = "https://authenticate.foundation.api.oneatlas.airbus.com/auth/realms/IDP/protocol/openid-connect/token"

	// DevBaseURL is the development OneAtlas base URL.
	DevBaseURL = "https://dev.sar.api.oneatlas.airbus.com/v1"
	// DevTokenURL is the development authentication endpoint (same as production).
	DevTokenURL = "https://authenticate.foundation.api.oneatlas.airbus.com/auth/realms/IDP/protocol/openid-connect/token"

	// LegacyBaseURL is the legacy production base URL.
	LegacyBaseURL = "https://sar.api.intelligence.airbus.com/v1"
	// LegacyDevBaseURL is the legacy development base URL.
	LegacyDevBaseURL = "https://dev.sar.api.intelligence.airbus.com/v1"

	defaultTimeout = 30 * time.Second
)

// Client is the SAR-API client. It embeds common.Client for HTTP operations.
type Client struct {
	*common.Client
}

// Option configures a Client.
type Option func(*clientConfig)

type clientConfig struct {
	baseURL    string
	tokenURL   string
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
func WithBaseURL(baseURL string) Option {
	return func(c *clientConfig) {
		c.baseURL = baseURL
	}
}

// WithTokenURL overrides the default token URL.
func WithTokenURL(tokenURL string) Option {
	return func(c *clientConfig) {
		c.tokenURL = tokenURL
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

// NewClient creates a new SAR-API client with the given API key.
// By default, it connects to the production OneAtlas environment.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	cfg := &clientConfig{
		baseURL:  DefaultBaseURL,
		tokenURL: DefaultTokenURL,
		timeout:  defaultTimeout,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	httpClient := common.EnsureHTTPClient(cfg.httpClient, cfg.timeout)

	auth := NewAPIKeyAuth(apiKey, cfg.tokenURL, httpClient)

	c, err := common.NewClient(common.ClientConfig{
		BaseURL:    cfg.baseURL,
		HTTPClient: httpClient,
		Auth:       auth,
		UserAgent:  cfg.userAgent,
	})
	if err != nil {
		return nil, err
	}

	return &Client{Client: c}, nil
}

// NewDevClient creates a client configured for the development environment.
func NewDevClient(apiKey string, opts ...Option) (*Client, error) {
	devOpts := []Option{
		WithBaseURL(DevBaseURL),
		WithTokenURL(DevTokenURL),
	}
	return NewClient(apiKey, append(devOpts, opts...)...)
}

// NewLegacyClient creates a client configured for the legacy production environment.
func NewLegacyClient(apiKey string, opts ...Option) (*Client, error) {
	legacyOpts := []Option{
		WithBaseURL(LegacyBaseURL),
	}
	return NewClient(apiKey, append(legacyOpts, opts...)...)
}

// NewLegacyDevClient creates a client configured for the legacy development environment.
func NewLegacyDevClient(apiKey string, opts ...Option) (*Client, error) {
	legacyOpts := []Option{
		WithBaseURL(LegacyDevBaseURL),
	}
	return NewClient(apiKey, append(legacyOpts, opts...)...)
}

