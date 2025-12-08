package iceye

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

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

// OAuth2Auth implements common.Authenticator for ICEYE OAuth2 Client Credentials flow.
type OAuth2Auth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	httpClient   *http.Client

	mu    sync.Mutex
	token string
	exp   time.Time
}

// NewOAuth2Auth creates an OAuth2 authenticator for ICEYE.
func NewOAuth2Auth(tokenURL, clientID, clientSecret string, httpClient *http.Client) *OAuth2Auth {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &OAuth2Auth{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   httpClient,
	}
}

// Apply implements common.Authenticator.
func (a *OAuth2Auth) Apply(ctx context.Context, req *http.Request) error {
	if err := a.refreshIfNeeded(ctx); err != nil {
		return err
	}
	a.mu.Lock()
	req.Header.Set("Authorization", "Bearer "+a.token)
	a.mu.Unlock()
	return nil
}

func (a *OAuth2Auth) refreshIfNeeded(ctx context.Context) error {
	a.mu.Lock()
	if time.Until(a.exp) > common.TokenExpiryBuffer {
		a.mu.Unlock()
		return nil
	}
	a.mu.Unlock()

	body := []byte("grant_type=client_credentials")
	b64 := base64.StdEncoding.EncodeToString([]byte(a.clientID + ":" + a.clientSecret))

	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create token request: %w", err)
	}
	tokenReq.Header.Set("Authorization", "Basic "+b64)
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("Accept", "application/json, application/problem+json")

	resp, err := a.httpClient.Do(tokenReq)
	if err != nil {
		return fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseError(resp)
	}

	var tok struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return fmt.Errorf("decode token response: %w", err)
	}
	if tok.AccessToken == "" {
		return errors.New("iceye: empty access_token")
	}

	expiresIn := tok.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}

	a.mu.Lock()
	a.token = tok.AccessToken
	a.exp = time.Now().Add(time.Duration(expiresIn) * time.Second)
	a.mu.Unlock()

	return nil
}

// ResourceOwnerAuth implements common.Authenticator for ICEYE Resource Owner Password flow (legacy).
type ResourceOwnerAuth struct {
	tokenURL   string
	apiKey     string // Base64-encoded client credentials
	username   string
	password   string
	httpClient *http.Client

	mu    sync.Mutex
	token string
	exp   time.Time
}

// NewResourceOwnerAuth creates a Resource Owner Password authenticator for ICEYE (legacy).
func NewResourceOwnerAuth(tokenURL, apiKey, username, password string, httpClient *http.Client) *ResourceOwnerAuth {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &ResourceOwnerAuth{
		tokenURL:   tokenURL,
		apiKey:     apiKey,
		username:   username,
		password:   password,
		httpClient: httpClient,
	}
}

// Apply implements common.Authenticator.
func (a *ResourceOwnerAuth) Apply(ctx context.Context, req *http.Request) error {
	if err := a.refreshIfNeeded(ctx); err != nil {
		return err
	}
	a.mu.Lock()
	req.Header.Set("Authorization", "Bearer "+a.token)
	a.mu.Unlock()
	return nil
}

func (a *ResourceOwnerAuth) refreshIfNeeded(ctx context.Context) error {
	a.mu.Lock()
	if time.Until(a.exp) > common.TokenExpiryBuffer {
		a.mu.Unlock()
		return nil
	}
	a.mu.Unlock()

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("username", a.username)
	form.Set("password", a.password)

	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return fmt.Errorf("create token request: %w", err)
	}
	tokenReq.Header.Set("Authorization", "Basic "+a.apiKey)
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("Accept", "application/json, application/problem+json")

	resp, err := a.httpClient.Do(tokenReq)
	if err != nil {
		return fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseError(resp)
	}

	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return fmt.Errorf("decode token response: %w", err)
	}
	if tok.AccessToken == "" {
		return errors.New("iceye: empty access_token")
	}

	expiresIn := tok.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}

	a.mu.Lock()
	a.token = tok.AccessToken
	a.exp = time.Now().Add(time.Duration(expiresIn) * time.Second)
	a.mu.Unlock()

	return nil
}
