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
	defaultBaseURL = "https://api.capellaspace.com/"
)

//--- API Error ---

type APIError struct {
	StatusCode int
	Body       string
	Validation *HTTPValidationError
}

func (e *APIError) Error() string {
	if e.Validation != nil && len(e.Validation.Detail) > 0 {
		return fmt.Sprintf("API error: status %d - %s", e.StatusCode, e.Validation.Detail[0].Msg)
	}
	return fmt.Sprintf("API error: status %d - %s", e.StatusCode, e.Body)
}

//--- Client ---

// Client is a stateless client for the Capella Tasking Service API.
// It does not hold an API key; the key must be provided with each call.
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string //unused for now
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// WithBaseURL sets the base URL for API requests.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client for the API client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func WithApiKey(key string) ClientOption {
	return func(c *Client) {
		c.apiKey = key
	}
}

// NewClient creates a new Tasking Service client.
// It uses sensible defaults which can be overridden with functional options.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

//--- Helper Methods ---

// newRequest creates an http.Request with the necessary headers.
// Note: It uses context.Background() as context is no longer passed from the caller.
func (c *Client) newRequest(apiKey, method, endpoint string, body []byte) (*http.Request, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}
	u.Path = path.Join(u.Path, endpoint)

	data := bytes.NewReader(body)

	// A non-cancellable context is used as it's not part of the public API.
	req, err := http.NewRequestWithContext(context.Background(), method, u.String(), data)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	req.Header.Set("Authorization", "ApiKey "+apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
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
		return c.handleError(resp)
	}

	if v != nil {
		if err = json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("failed to decode successful response: %w", err)
		}
	}

	return nil
}

// handleError parses an error response and returns a detailed APIError.
func (c *Client) handleError(resp *http.Response) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response body: %w", err)
	}

	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Body:       string(bodyBytes),
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		var validationError HTTPValidationError
		if json.Unmarshal(bodyBytes, &validationError) == nil {
			apiErr.Validation = &validationError
		}
	}

	return apiErr
}

// Feature represents a single GeoJSON feature whose Properties are
// compile-time typed by the generic parameter P.
type Feature[T any] struct {
	Type       string `json:"type"`       // always "Feature"
	Geometry   any    `json:"geometry"`   // any GeoJSON Geometry type
	Properties T      `json:"properties"` // user-supplied metadata
}
