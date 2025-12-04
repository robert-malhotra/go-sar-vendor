package common

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Authenticator provides authentication for API requests.
type Authenticator interface {
	// Authenticate performs any necessary authentication (e.g., token refresh).
	Authenticate(ctx context.Context) error

	// AuthHeader returns the Authorization header value.
	AuthHeader() string
}

// BearerAuth implements static bearer token authentication.
type BearerAuth struct {
	token string
}

// NewBearerAuth creates a new bearer token authenticator.
func NewBearerAuth(token string) *BearerAuth {
	return &BearerAuth{token: token}
}

func (b *BearerAuth) Authenticate(ctx context.Context) error {
	return nil // Static token, no refresh needed
}

func (b *BearerAuth) AuthHeader() string {
	return "Bearer " + b.token
}

// APIKeyAuth implements API key authentication.
type APIKeyAuth struct {
	key    string
	prefix string // e.g., "ApiKey", "Bearer", or empty
}

// NewAPIKeyAuth creates a new API key authenticator with "ApiKey" prefix.
func NewAPIKeyAuth(key string) *APIKeyAuth {
	return &APIKeyAuth{key: key, prefix: "ApiKey"}
}

// NewAPIKeyAuthWithPrefix creates a new API key authenticator with custom prefix.
func NewAPIKeyAuthWithPrefix(key, prefix string) *APIKeyAuth {
	return &APIKeyAuth{key: key, prefix: prefix}
}

func (a *APIKeyAuth) Authenticate(ctx context.Context) error {
	return nil // Static key, no refresh needed
}

func (a *APIKeyAuth) AuthHeader() string {
	if a.prefix != "" {
		return a.prefix + " " + a.key
	}
	return a.key
}

// OAuth2Config holds configuration for OAuth2 Client Credentials flow.
type OAuth2Config struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scopes       []string
	HTTPClient   *http.Client
}

// OAuth2Auth implements OAuth2 Client Credentials authentication with auto-refresh.
type OAuth2Auth struct {
	cfg OAuth2Config

	mu    sync.Mutex
	token string
	exp   time.Time
}

// NewOAuth2Auth creates a new OAuth2 authenticator.
func NewOAuth2Auth(cfg OAuth2Config) *OAuth2Auth {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &OAuth2Auth{cfg: cfg}
}

func (o *OAuth2Auth) Authenticate(ctx context.Context) error {
	o.mu.Lock()
	if time.Until(o.exp) > 30*time.Second {
		o.mu.Unlock()
		return nil // Token still valid
	}
	o.mu.Unlock()

	// Build request body
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	if len(o.cfg.Scopes) > 0 {
		scope := ""
		for i, s := range o.cfg.Scopes {
			if i > 0 {
				scope += " "
			}
			scope += s
		}
		form.Set("scope", scope)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.cfg.TokenURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return fmt.Errorf("create token request: %w", err)
	}

	// Set Basic auth header
	credentials := o.cfg.ClientID + ":" + o.cfg.ClientSecret
	auth := base64.StdEncoding.EncodeToString([]byte(credentials))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := o.cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ParseErrorResponse(resp)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("decode token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return errors.New("empty access_token in response")
	}

	o.mu.Lock()
	o.token = tokenResp.AccessToken
	o.exp = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	o.mu.Unlock()

	return nil
}

func (o *OAuth2Auth) AuthHeader() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return "Bearer " + o.token
}

// ResourceOwnerConfig holds configuration for OAuth2 Resource Owner Password flow.
type ResourceOwnerConfig struct {
	TokenURL   string
	APIKey     string // Base64-encoded client credentials
	Username   string
	Password   string
	HTTPClient *http.Client
}

// ResourceOwnerAuth implements OAuth2 Resource Owner Password authentication.
type ResourceOwnerAuth struct {
	cfg ResourceOwnerConfig

	mu    sync.Mutex
	token string
	exp   time.Time
}

// NewResourceOwnerAuth creates a new Resource Owner Password authenticator.
func NewResourceOwnerAuth(cfg ResourceOwnerConfig) *ResourceOwnerAuth {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &ResourceOwnerAuth{cfg: cfg}
}

func (r *ResourceOwnerAuth) Authenticate(ctx context.Context) error {
	r.mu.Lock()
	if time.Until(r.exp) > 30*time.Second {
		r.mu.Unlock()
		return nil
	}
	r.mu.Unlock()

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("username", r.cfg.Username)
	form.Set("password", r.cfg.Password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.cfg.TokenURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return fmt.Errorf("create token request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+r.cfg.APIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := r.cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ParseErrorResponse(resp)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("decode token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return errors.New("empty access_token in response")
	}

	r.mu.Lock()
	r.token = tokenResp.AccessToken
	r.exp = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	r.mu.Unlock()

	return nil
}

func (r *ResourceOwnerAuth) AuthHeader() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return "Bearer " + r.token
}
