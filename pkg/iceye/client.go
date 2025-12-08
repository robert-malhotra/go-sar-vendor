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
//   - Thread-safe; safe for concurrent goroutines.
//   - RFC 7807 errors mapped to *iceye.Error.
//
// Docs: https://docs.iceye.com/api/ (December 2024 release)
package iceye

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

const (
	// DefaultBaseURL is the default ICEYE API base URL.
	DefaultBaseURL = "https://platform.iceye.com/api"

	// DefaultTokenURL is the default ICEYE OAuth2 token endpoint.
	DefaultTokenURL = "https://auth.iceye.com/oauth2/token"

	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "go-sar-vendor/iceye"
)

// Client is the ICEYE API client. It is thread-safe.
type Client struct {
	*common.Client
	userAgent string
}

// Option configures a Client.
type Option func(*clientConfig)

type clientConfig struct {
	baseURL    string
	tokenURL   string
	httpClient *http.Client
	timeout    time.Duration
	userAgent  string
	auth       common.Authenticator
}

// WithBaseURL sets a custom base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *clientConfig) {
		c.baseURL = baseURL
	}
}

// WithTokenURL sets a custom OAuth2 token URL.
func WithTokenURL(tokenURL string) Option {
	return func(c *clientConfig) {
		c.tokenURL = tokenURL
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *clientConfig) {
		c.httpClient = client
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

// WithCredentials sets OAuth2 client credentials for authentication.
func WithCredentials(clientID, clientSecret string) Option {
	return func(c *clientConfig) {
		c.auth = NewOAuth2Auth(c.tokenURL, clientID, clientSecret, c.httpClient)
	}
}

// WithResourceOwner sets legacy Resource Owner Password credentials.
func WithResourceOwner(apiKey, username, password string) Option {
	return func(c *clientConfig) {
		c.auth = NewResourceOwnerAuth(c.tokenURL, apiKey, username, password, c.httpClient)
	}
}

// WithAuth sets a custom authenticator.
func WithAuth(auth common.Authenticator) Option {
	return func(c *clientConfig) {
		c.auth = auth
	}
}

// NewClient creates a new ICEYE API client.
// Credentials must be provided via WithCredentials or WithResourceOwner options.
func NewClient(opts ...Option) (*Client, error) {
	cfg := &clientConfig{
		baseURL:   DefaultBaseURL,
		tokenURL:  DefaultTokenURL,
		timeout:   defaultTimeout,
		userAgent: defaultUserAgent,
	}

	// First pass: set base config values
	for _, opt := range opts {
		opt(cfg)
	}

	httpClient := common.EnsureHTTPClient(cfg.httpClient, cfg.timeout)
	cfg.httpClient = httpClient

	// Second pass: apply auth options that depend on tokenURL and httpClient
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.auth == nil {
		return nil, fmt.Errorf("iceye: credentials required (use WithCredentials or WithResourceOwner)")
	}

	c, err := common.NewClient(common.ClientConfig{
		BaseURL:    cfg.baseURL,
		HTTPClient: httpClient,
		Auth:       cfg.auth,
		UserAgent:  cfg.userAgent,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:    c,
		userAgent: cfg.userAgent,
	}, nil
}

// do performs an HTTP request with JSON encode/decode and ICEYE-specific error handling.
func (c *Client) do(ctx context.Context, method, urlStr string, in any, out any) error {
	// Parse the path as a URL (may contain query string)
	pathURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("parse URL path: %w", err)
	}

	// Resolve against base URL
	fullURL := c.BaseURL().ResolveReference(pathURL)

	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json, application/problem+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	// Apply auth
	if err := c.Client.ApplyAuth(ctx, req); err != nil {
		return err
	}

	resp, err := c.HTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Check for error status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseError(resp)
	}

	// Decode response
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}
