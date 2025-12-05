package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second
)

// ClientConfig holds configuration for the HTTP client.
type ClientConfig struct {
	BaseURL    string
	HTTPClient *http.Client
	Auth       Authenticator
	UserAgent  string
	Timeout    time.Duration
}

// Client is a base HTTP client for API requests.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	auth       Authenticator
	userAgent  string
}

// NewClient creates a new HTTP client with the given configuration.
func NewClient(cfg ClientConfig) (*Client, error) {
	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = DefaultTimeout
		}
		httpClient = &http.Client{Timeout: timeout}
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		auth:       cfg.Auth,
		userAgent:  cfg.UserAgent,
	}, nil
}

// BaseURL returns the base URL.
func (c *Client) BaseURL() *url.URL {
	return c.baseURL
}

// HTTPClient returns the underlying HTTP client.
func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

// SetAuth sets the authenticator.
func (c *Client) SetAuth(auth Authenticator) {
	c.auth = auth
}

// BuildURL constructs a URL from the base URL and path.
func (c *Client) BuildURL(path string) *url.URL {
	return c.baseURL.JoinPath(path)
}

// Do performs an HTTP request with JSON encoding/decoding.
// expectedStatus is the expected HTTP status code for success.
// If reqBody is non-nil, it will be JSON-encoded as the request body.
// If respBody is non-nil, the response body will be JSON-decoded into it.
func (c *Client) Do(ctx context.Context, method, path string, expectedStatus int, reqBody, respBody any) error {
	u := c.BuildURL(path)
	return c.DoURL(ctx, method, u, expectedStatus, reqBody, respBody)
}

// DoURL performs an HTTP request with a pre-built URL.
func (c *Client) DoURL(ctx context.Context, method string, u *url.URL, expectedStatus int, reqBody, respBody any) error {
	var body io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		body = bytes.NewReader(b)
	}

	return c.DoRaw(ctx, method, u, body, expectedStatus, respBody)
}

// DoRaw performs an HTTP request with a raw body reader.
func (c *Client) DoRaw(ctx context.Context, method string, u *url.URL, body io.Reader, expectedStatus int, respBody any) error {
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Apply authentication
	if c.auth != nil {
		if err := c.auth.Apply(ctx, req); err != nil {
			return fmt.Errorf("authenticate: %w", err)
		}
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Check for expected status (0 means accept any 2xx)
	if expectedStatus == 0 {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return ParseErrorResponse(resp)
		}
	} else if resp.StatusCode != expectedStatus {
		return ParseErrorResponse(resp)
	}

	// Decode response body
	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// DoRawResponse performs an HTTP request and returns the raw response body.
func (c *Client) DoRawResponse(ctx context.Context, method string, u *url.URL, body io.Reader, expectedStatus int) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Apply authentication
	if c.auth != nil {
		if err := c.auth.Apply(ctx, req); err != nil {
			return nil, fmt.Errorf("authenticate: %w", err)
		}
	}

	// Set headers
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Check for expected status
	if expectedStatus == 0 {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			apiErr := ParseErrorResponse(resp)
			apiErr.RawBody = string(buf)
			return nil, apiErr
		}
	} else if resp.StatusCode != expectedStatus {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(buf),
			RawBody:    string(buf),
		}
		return nil, apiErr
	}

	return buf, nil
}

// NewRequest creates a new HTTP request with authentication headers.
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	u := c.BuildURL(path)

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Apply authentication
	if c.auth != nil {
		if err := c.auth.Apply(ctx, req); err != nil {
			return nil, fmt.Errorf("authenticate: %w", err)
		}
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

// MarshalBody marshals v to JSON and returns a bytes.Buffer.
func MarshalBody(v any) (*bytes.Buffer, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return bytes.NewBuffer(b), nil
}
