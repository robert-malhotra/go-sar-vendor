package capella

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// AuthClient provides JWT token-based authentication (deprecated).
// For most use cases, use API key authentication via WithAPIKey() instead.
type AuthClient struct {
	client    *Client
	tokenURL  string
	username  string
	password  string
	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

// AuthConfig holds configuration for JWT authentication.
type AuthConfig struct {
	// TokenURL is the OAuth2 token endpoint (default: /token)
	TokenURL string

	// Username for password grant
	Username string

	// Password for password grant
	Password string
}

// NewAuthClient creates a new authentication client for JWT token management.
// Note: API key authentication is recommended over JWT tokens.
func NewAuthClient(client *Client, cfg AuthConfig) *AuthClient {
	tokenURL := cfg.TokenURL
	if tokenURL == "" {
		tokenURL = client.BaseURL().String() + "/token"
	}
	return &AuthClient{
		client:   client,
		tokenURL: tokenURL,
		username: cfg.Username,
		password: cfg.Password,
	}
}

// TokenResponse represents the OAuth2 token response.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// GetToken returns a valid access token, refreshing if necessary.
func (a *AuthClient) GetToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Return cached token if still valid (with 30s buffer)
	if time.Until(a.expiresAt) > 30*time.Second {
		return a.token, nil
	}

	// Refresh token
	token, err := a.refreshToken(ctx)
	if err != nil {
		return "", err
	}

	return token, nil
}

// refreshToken exchanges credentials for a new JWT token.
func (a *AuthClient) refreshToken(ctx context.Context) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", a.username)
	data.Set("password", a.password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.HTTPClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", parseError(resp)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token in response")
	}

	a.token = tokenResp.AccessToken
	a.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return a.token, nil
}

// APIKeyService provides API key management operations.
type APIKeyService struct {
	client *Client
}

// NewAPIKeyService creates a new API key service.
func NewAPIKeyService(client *Client) *APIKeyService {
	return &APIKeyService{client: client}
}

// APIKey represents an API key.
type APIKey struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt,omitempty"`
	LastUsedAt  time.Time `json:"lastUsedAt,omitempty"`
}

// APIKeyCreateRequest represents the request to create an API key.
type APIKeyCreateRequest struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	ExpiresAt   time.Time `json:"expiresAt,omitempty"`
}

// APIKeyCreateResponse represents the response when creating an API key.
type APIKeyCreateResponse struct {
	APIKey
	Key string `json:"key"` // Only returned on creation
}

// Create creates a new API key.
func (s *APIKeyService) Create(ctx context.Context, req APIKeyCreateRequest) (*APIKeyCreateResponse, error) {
	var resp APIKeyCreateResponse
	if err := s.client.Do(ctx, http.MethodPost, "/keys", 0, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// List lists all API keys.
func (s *APIKeyService) List(ctx context.Context) ([]APIKey, error) {
	var resp []APIKey
	if err := s.client.Do(ctx, http.MethodGet, "/keys", 0, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Delete deletes an API key by ID.
func (s *APIKeyService) Delete(ctx context.Context, keyID string) error {
	return s.client.Do(ctx, http.MethodDelete, "/keys/"+keyID, 0, nil, nil)
}
