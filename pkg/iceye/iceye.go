// Package iceye provides a minimal yet extensible Go SDK for the ICEYE Tasking API.
//
// Key features:
//   - OAuth 2 Client‑Credentials token management with auto‑refresh.
//   - Idiomatic Go types mirroring ICEYE v2 company / tasking / price / products paths.
//   - DRY lazy pagination helpers using Go 1.23+ `iter.Seq2`.
//   - Zero non‑std‑lib runtime dependencies.
//   - Thread‑safe; safe for concurrent goroutines.
//   - RFC 7807 errors mapped to *iceye.Error.
//
// Version history
//
//	v0.1.0  – initial skeleton (contracts, basic tasking)
//	v0.2.0  – ListContracts iterator
//	v1.0.0  – full Tasking API coverage (price, scene, products, cancel/patch)
//	v1.0.1  – Align iterator signatures with Go 1.23 `iter` (fixes compiler errors)
//	v1.1.1  – GetTaskProducts converted to iterator (iter.Seq2)
//	v1.2.0  – Introduce shared paginator (compile issue fixed by moving it to a standalone generic func)
//
// Docs: https://docs.iceye.com/api/tasking-v2  (June 2025 release)
// ----------------------------------------------------------------------------
// Quick example (Go 1.23+)
// ----------------------------------------------------------------------------
//
//	for tasks, err := range cli.ListTasks(ctx, 100, nil) {
//	    if err != nil { log.Fatal(err) }
//	    log.Printf("page with %d tasks", len(tasks))
//	}
package iceye

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"iter" // std‑lib as of Go 1.23
)

// ----------------------------------------------------------------------------
// Config & Client
// ----------------------------------------------------------------------------

type Config struct {
	BaseURL      string // ex: https://api.iceye.com/api
	TokenURL     string // ex: https://auth.iceye.com/oauth2/token
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client // defaults to http.DefaultClient
}

type Client struct {
	cfg   Config
	mu    sync.Mutex // guards token+exp
	token string
	exp   time.Time
}

func NewClient(cfg Config) *Client {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	return &Client{cfg: cfg}
}

// ----------------------------------------------------------------------------
// Token management (OAuth2 Client‑Credentials)
// ----------------------------------------------------------------------------

func (c *Client) authenticate(ctx context.Context) error {
	c.mu.Lock()
	if time.Until(c.exp) > 30*time.Second {
		c.mu.Unlock()
		return nil // still valid
	}
	c.mu.Unlock()

	body := []byte("grant_type=client_credentials")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.TokenURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	b64 := base64.StdEncoding.EncodeToString([]byte(c.cfg.ClientID + ":" + c.cfg.ClientSecret))
	req.Header.Set("Authorization", "Basic "+b64)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json, application/problem+json")

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return err
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
		return err
	}
	if tok.AccessToken == "" {
		return errors.New("iceye: empty access_token")
	}

	c.mu.Lock()
	c.token = tok.AccessToken
	c.exp = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
	c.mu.Unlock()
	return nil
}

// do wraps HTTP with auth + JSON encode/decode.
func (c *Client) do(ctx context.Context, method, path string, in any, out any) error {
	if err := c.authenticate(ctx); err != nil {
		return err
	}

	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json, application/problem+json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return parseError(resp)
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Generic pagination helper (stand‑alone generic function)
// ----------------------------------------------------------------------------

// paginate generates an `iter.Seq2` for endpoints following ICEYE's common
// `{ "data": [...], "cursor": "..." }` pattern. Callers supply a `fetch` closure
// receiving the previous cursor (nil on first page) and returning the next slice,
// the next cursor (nil/empty when no more pages), and an error.
func paginate[T any](fetch func(cur *string) ([]T, *string, error)) iter.Seq2[[]T, error] {
	return func(yield func([]T, error) bool) {
		var cur *string
		for {
			data, next, err := fetch(cur)
			if !yield(data, err) {
				return
			}
			if err != nil || next == nil || *next == "" {
				return
			}
			cur = next
		}
	}
}

// ----------------------------------------------------------------------------
// Company API (Contracts)
// ----------------------------------------------------------------------------

type ContractsResponse struct {
	Data   []Contract `json:"data"`
	Cursor string     `json:"cursor,omitempty"`
}

type Contract struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Start        time.Time         `json:"start"`
	End          time.Time         `json:"end"`
	ImagingModes ModePolicy        `json:"imagingModes"`
	Priority     PriorityPolicy    `json:"priority"`
	Exclusivity  ExclusivityPolicy `json:"exclusivity"`
}

type ModePolicy struct {
	Allowed []string `json:"allowed"`
	Default string   `json:"default"`
}

type PriorityPolicy struct {
	Allowed []string `json:"allowed"`
	Default string   `json:"default"`
}

type ExclusivityPolicy struct {
	Allowed []string `json:"allowed"`
	Default string   `json:"default"`
}

func (c *Client) ListContracts(ctx context.Context, pageSize int) iter.Seq2[ContractsResponse, error] {
	return func(yield func(ContractsResponse, error) bool) {
		seq := paginate[Contract](func(cur *string) ([]Contract, *string, error) {
			path := "/company/v1/contracts"
			sep := "?"
			if pageSize > 0 {
				path += fmt.Sprintf("%slimit=%d", sep, pageSize)
				sep = "&"
			}
			if cur != nil {
				path += fmt.Sprintf("%scursor=%s", sep, url.QueryEscape(*cur))
			}
			var resp ContractsResponse
			err := c.do(ctx, http.MethodGet, path, nil, &resp)
			return resp.Data, &resp.Cursor, err
		})
		for data, err := range seq {
			if !yield(ContractsResponse{Data: data}, err) {
				return
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Tasking API – structs
// ----------------------------------------------------------------------------

type Point struct{ Lat, Lon float64 }

type Window struct{ Start, End time.Time }

type ImagingMode string

const (
	IMSpotlight     ImagingMode = "SPOTLIGHT"
	IMSpotlightFine ImagingMode = "SPOTLIGHT_FINE"
	IMStripmap      ImagingMode = "STRIPMAP"
	IMScan          ImagingMode = "SCAN"
)

type CreateTaskRequest struct {
	AcquisitionWindow Window      `json:"acquisitionWindow"`
	ContractID        string      `json:"contractID"`
	ImagingMode       ImagingMode `json:"imagingMode"`
	PointOfInterest   Point       `json:"pointOfInterest"`

	Reference      *string                 `json:"reference,omitempty"`
	Exclusivity    *string                 `json:"exclusivity,omitempty"`
	Priority       *string                 `json:"priority,omitempty"`
	SLA            *string                 `json:"sla,omitempty"`
	IncidenceAngle *struct{ Min, Max int } `json:"incidenceAngle,omitempty"`
	LookSide       *string                 `json:"lookSide,omitempty"`
	PassDirection  *string                 `json:"passDirection,omitempty"`
}

type Task struct {
	ID                string      `json:"id"`
	ContractID        string      `json:"contractID"`
	PointOfInterest   Point       `json:"pointOfInterest"`
	AcquisitionWindow Window      `json:"acquisitionWindow"`
	ImagingMode       ImagingMode `json:"imagingMode"`
	Status            string      `json:"status"`
	CreatedAt         time.Time   `json:"createdAt"`
	UpdatedAt         time.Time   `json:"updatedAt"`
}

type PriceResponse struct {
	Amount   float64            `json:"amount"`
	Currency string             `json:"currency"`
	Items    map[string]float64 `json:"items"`
}

type Scene struct {
	ID           string `json:"id"`
	Start, End   time.Time
	FootprintWKT string `json:"footprintWkt"`
}

type Product struct {
	Type   string  `json:"type"`
	Assets []Asset `json:"assets"`
}

type Asset struct {
	Name, Href string
	Size       int64
}

// ----------------------------------------------------------------------------
// Tasking API methods
// ----------------------------------------------------------------------------

func (c *Client) GetTaskPrice(ctx context.Context, req CreateTaskRequest) (*PriceResponse, error) {
	var resp PriceResponse
	if err := c.do(ctx, http.MethodPost, "/tasking/v2/price", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) CreateTask(ctx context.Context, req CreateTaskRequest) (*Task, error) {
	var resp Task
	if err := c.do(ctx, http.MethodPost, "/tasking/v2/tasks", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetTask(ctx context.Context, taskID string) (*Task, error) {
	var resp Task
	if err := c.do(ctx, http.MethodGet, "/tasking/v2/tasks/"+url.PathEscape(taskID), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) CancelTask(ctx context.Context, taskID string) error {
	return c.do(ctx, http.MethodPatch, "/tasking/v2/tasks/"+url.PathEscape(taskID), map[string]string{"status": "CANCELED"}, nil)
}

func (c *Client) ListTasks(ctx context.Context, pageSize int, contractID *string) iter.Seq2[[]Task, error] {
	return paginate[Task](func(cur *string) ([]Task, *string, error) {
		path := "/tasking/v2/tasks"
		sep := "?"
		if pageSize > 0 {
			path += fmt.Sprintf("%slimit=%d", sep, pageSize)
			sep = "&"
		}
		if contractID != nil {
			path += fmt.Sprintf("%scontractID=%s", sep, url.QueryEscape(*contractID))
			sep = "&"
		}
		if cur != nil {
			path += fmt.Sprintf("%scursor=%s", sep, url.QueryEscape(*cur))
		}
		var resp struct {
			Data   []Task  `json:"data"`
			Cursor *string `json:"cursor"`
		}
		err := c.do(ctx, http.MethodGet, path, nil, &resp)
		return resp.Data, resp.Cursor, err
	})
}

func (c *Client) GetTaskScene(ctx context.Context, taskID string) (*Scene, error) {
	var resp Scene
	if err := c.do(ctx, http.MethodGet, "/tasking/v2/tasks/"+url.PathEscape(taskID)+"/scene", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetTaskProducts(ctx context.Context, taskID string, pageSize int) iter.Seq2[[]Product, error] {
	return paginate[Product](func(cur *string) ([]Product, *string, error) {
		path := fmt.Sprintf("/tasking/v2/tasks/%s/products", url.PathEscape(taskID))
		sep := "?"
		if pageSize > 0 {
			path += fmt.Sprintf("%slimit=%d", sep, pageSize)
			sep = "&"
		}
		if cur != nil {
			path += fmt.Sprintf("%scursor=%s", sep, url.QueryEscape(*cur))
		}
		var resp struct {
			Data   []Product `json:"data"`
			Cursor *string   `json:"cursor"`
		}
		err := c.do(ctx, http.MethodGet, path, nil, &resp)
		return resp.Data, resp.Cursor, err
	})
}

// ----------------------------------------------------------------------------
// Error handling
// ----------------------------------------------------------------------------

type Error struct {
	Status int    `json:"status"`
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("iceye: %s (%d) – %s", e.Code, e.Status, e.Detail)
}

func parseError(resp *http.Response) error {
	var e Error
	e.Status = resp.StatusCode
	if json.NewDecoder(resp.Body).Decode(&e) != nil || e.Code == "" {
		e.Code = http.StatusText(resp.StatusCode)
	}
	return &e
}
