package airbus

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// APIKeyAuth implements common.Authenticator for Airbus API key authentication.
// It exchanges an API key for a Bearer token via the token endpoint using
// the api_key grant type.
type APIKeyAuth struct {
	apiKey     string
	tokenURL   string
	httpClient *http.Client

	mu    sync.Mutex
	token string
	exp   time.Time
}

// NewAPIKeyAuth creates an authenticator that exchanges an API key for a bearer token.
// The httpClient should have appropriate timeouts configured.
func NewAPIKeyAuth(apiKey, tokenURL string, httpClient *http.Client) *APIKeyAuth {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &APIKeyAuth{
		apiKey:     apiKey,
		tokenURL:   tokenURL,
		httpClient: httpClient,
	}
}

// Apply applies authentication to the HTTP request.
// It implements the common.Authenticator interface.
func (a *APIKeyAuth) Apply(ctx context.Context, req *http.Request) error {
	if err := a.refreshIfNeeded(ctx); err != nil {
		return err
	}
	a.mu.Lock()
	req.Header.Set("Authorization", "Bearer "+a.token)
	a.mu.Unlock()
	return nil
}

// refreshIfNeeded obtains or refreshes the bearer token if needed.
func (a *APIKeyAuth) refreshIfNeeded(ctx context.Context) error {
	a.mu.Lock()
	// Check if token is still valid (with 5 minute buffer)
	if time.Until(a.exp) > 5*time.Minute {
		a.mu.Unlock()
		return nil
	}
	a.mu.Unlock()

	// Request new token
	form := url.Values{
		"apikey":     {a.apiKey},
		"grant_type": {"api_key"},
		"client_id":  {"IDP"},
	}

	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return fmt.Errorf("create token request: %w", err)
	}

	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(tokenReq)
	if err != nil {
		return fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("authentication failed: %s - %s", errResp.Error, errResp.ErrorDescription)
		}
		return fmt.Errorf("authentication failed: %s", resp.Status)
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

	a.mu.Lock()
	a.token = tokenResp.AccessToken
	a.exp = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	a.mu.Unlock()

	return nil
}

// Token returns the current access token.
func (a *APIKeyAuth) Token() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.token
}

// Expiry returns the token expiration time.
func (a *APIKeyAuth) Expiry() time.Time {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.exp
}
