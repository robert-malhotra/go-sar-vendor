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
	// Apply applies authentication to an HTTP request.
	// It performs any necessary authentication (e.g., token refresh) and sets
	// the appropriate headers on the request.
	Apply(ctx context.Context, req *http.Request) error
}

// BearerAuth implements static bearer token authentication.
type BearerAuth struct {
	token string
}

// NewBearerAuth creates a new bearer token authenticator.
func NewBearerAuth(token string) *BearerAuth {
	return &BearerAuth{token: token}
}

func (b *BearerAuth) Apply(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+b.token)
	return nil
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

func (a *APIKeyAuth) Apply(ctx context.Context, req *http.Request) error {
	if a.prefix != "" {
		req.Header.Set("Authorization", a.prefix+" "+a.key)
	} else {
		req.Header.Set("Authorization", a.key)
	}
	return nil
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

func (o *OAuth2Auth) Apply(ctx context.Context, req *http.Request) error {
	if err := o.refreshIfNeeded(ctx); err != nil {
		return err
	}
	o.mu.Lock()
	req.Header.Set("Authorization", "Bearer "+o.token)
	o.mu.Unlock()
	return nil
}

func (o *OAuth2Auth) refreshIfNeeded(ctx context.Context) error {
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

	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPost, o.cfg.TokenURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return fmt.Errorf("create token request: %w", err)
	}

	// Set Basic auth header
	credentials := o.cfg.ClientID + ":" + o.cfg.ClientSecret
	auth := base64.StdEncoding.EncodeToString([]byte(credentials))
	tokenReq.Header.Set("Authorization", "Basic "+auth)
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("Accept", "application/json")

	resp, err := o.cfg.HTTPClient.Do(tokenReq)
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

func (r *ResourceOwnerAuth) Apply(ctx context.Context, req *http.Request) error {
	if err := r.refreshIfNeeded(ctx); err != nil {
		return err
	}
	r.mu.Lock()
	req.Header.Set("Authorization", "Bearer "+r.token)
	r.mu.Unlock()
	return nil
}

func (r *ResourceOwnerAuth) refreshIfNeeded(ctx context.Context) error {
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

	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPost, r.cfg.TokenURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return fmt.Errorf("create token request: %w", err)
	}

	tokenReq.Header.Set("Authorization", "Basic "+r.cfg.APIKey)
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("Accept", "application/json")

	resp, err := r.cfg.HTTPClient.Do(tokenReq)
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
