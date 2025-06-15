package airbus

// Airbus OneAtlas Radar SAR – Go SDK (typed version)
//
// The client supports **configurable endpoints** at both the package and the
// instance level:
//
//     // global override (affects all new clients)
//     airbus.TokenURL = mock.URL + "/token"
//     airbus.BaseURL  = mock.URL + "/sar"
//
//     // per‑client override
//     cli := airbus.New("APIKEY", nil,
//         airbus.WithTokenURL("https://stage.auth/token"),
//         airbus.WithBaseURL("https://stage.sar/v1/sar"))
//
// The code is intentionally dependency‑free (std‑lib only) and can be tested
// with `httptest.Server`.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"
)

/*──────────────────── default endpoints (overrideable) ───────────────*/

const (
	DefaultTokenURL = "https://authenticate.foundation.api.oneatlas.airbus.com/auth/realms/IDP/protocol/openid-connect/token"
	DefaultBaseURL  = "https://sar.api.oneatlas.airbus.com/v1/sar"
)

/*──────────────────── geometry & helpers ─────────────────────────────*/

type GeoJSONGeometry struct {
	Type        string `json:"type"`
	Coordinates any    `json:"coordinates"`
}

type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

/*──────────────────── enumerations (subset) ──────────────────────────*/

type (
	ProductType       string
	SensorMode        string
	OrbitType         string
	ResolutionVariant string
	MapProjection     string
	Purpose           string
	Polarization      string
	PathDirection     string
)

const (
	PTSSC ProductType = "SSC"
	PTEEC ProductType = "EEC"

	OrbitRapid OrbitType = "rapid"

	ResRadiometric ResolutionVariant = "RE"

	MapAuto MapProjection = "auto"
)

/*──────────────────── request models ─────────────────────────────────*/

type FeasibilityRequest struct {
	AOI              GeoJSONGeometry `json:"aoi"`
	Time             TimeRange       `json:"time"`
	FeasibilityLevel string          `json:"feasibilityLevel"`
	SensorMode       SensorMode      `json:"sensorMode"`
	ProductType      ProductType     `json:"productType"`
	OrbitType        OrbitType       `json:"orbitType"`

	Polarization      *Polarization      `json:"polarizationChannels,omitempty"`
	ResolutionVariant *ResolutionVariant `json:"resolutionVariant,omitempty"`
	PathDirection     *PathDirection     `json:"pathDirection,omitempty"`
}

type CatalogueRequest struct {
	AOI  *GeoJSONGeometry `json:"aoi,omitempty"`
	Time *TimeRange       `json:"time,omitempty"`
}

/*──────────────────── order workflow structs ─────────────────────────*/

type OrderOptions struct {
	ProductType       ProductType       `json:"productType"`
	ResolutionVariant ResolutionVariant `json:"resolutionVariant"`
	OrbitType         OrbitType         `json:"orbitType"`
	MapProjection     MapProjection     `json:"mapProjection"`
	GainAttenuation   int               `json:"gainAttenuation"`
}

type AddItemsRequest struct {
	Items        []string     `json:"items"`
	OrderOptions OrderOptions `json:"orderOptions"`
}

type ShopcartPatch struct {
	Purpose Purpose `json:"purpose"`
}

/*──────────────────── response models ───────────────────────────────*/

type FeatureCollection[T any] struct {
	Type     string `json:"type"`
	Features []T    `json:"features"`
}

type AcquisitionFeature struct {
	Type       string                `json:"type"`
	Geometry   GeoJSONGeometry       `json:"geometry"`
	Properties AcquisitionProperties `json:"properties"`
}

type AcquisitionProperties struct {
	ItemID      string      `json:"itemId"`
	SensorMode  SensorMode  `json:"sensorMode"`
	ProductType ProductType `json:"productType"`
	OrbitType   OrbitType   `json:"orbitType"`
	Status      string      `json:"status"`
}

type Order struct {
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
}

/*──────────────────── client & options ───────────────────────────────*/

type Client struct {
	apiKey     string
	httpClient *http.Client

	tokenURL string
	baseURL  string

	mu    sync.Mutex
	token string
	exp   time.Time
}

type ClientOption func(*Client)

func WithBaseURL(u string) ClientOption  { return func(c *Client) { c.baseURL = u } }
func WithTokenURL(u string) ClientOption { return func(c *Client) { c.tokenURL = u } }

// New returns a configured client; hc==nil uses a 15 s timeout http.Client.
func New(apiKey string, hc *http.Client, opts ...ClientOption) *Client {
	if hc == nil {
		hc = &http.Client{Timeout: 15 * time.Second}
	}
	c := &Client{
		apiKey:     apiKey,
		httpClient: hc,
		tokenURL:   DefaultTokenURL,
		baseURL:    DefaultBaseURL,
	}
	for _, f := range opts {
		f(c)
	}
	return c
}

/*──────────────────── auth & low‑level HTTP ───────────────────────────*/

func (c *Client) ensureToken(ctx context.Context) error {
	c.mu.Lock()
	valid := time.Until(c.exp) > 5*time.Minute
	c.mu.Unlock()
	if valid {
		return nil
	}

	data := url.Values{
		"apikey":     {c.apiKey},
		"grant_type": {"api_key"},
		"client_id":  {"IDP"},
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("airbus auth: %s", resp.Status)
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if json.NewDecoder(resp.Body).Decode(&tok) != nil || tok.AccessToken == "" {
		return errors.New("airbus auth: bad JSON")
	}

	c.mu.Lock()
	c.token = tok.AccessToken
	c.exp = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
	c.mu.Unlock()
	return nil
}

func (c *Client) do(ctx context.Context, method, rel string, in, out any) error {
	if err := c.ensureToken(ctx); err != nil {
		return err
	}
	var body io.Reader
	if in != nil {
		b, _ := json.Marshal(in)
		body = bytes.NewReader(b)
	}
	req, _ := http.NewRequestWithContext(ctx, method, c.baseURL+rel, body)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("airbus %s %s: %s", method, rel, resp.Status)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

/*──────────────────── high‑level API helpers ─────────────────────────*/

func (c *Client) SearchFeasibility(ctx context.Context, r FeasibilityRequest) (FeatureCollection[AcquisitionFeature], error) {
	var fc FeatureCollection[AcquisitionFeature]
	err := c.do(ctx, http.MethodPost, "/feasibility", r, &fc)
	return fc, err
}

func (c *Client) SearchCatalogue(ctx context.Context, r CatalogueRequest) (FeatureCollection[AcquisitionFeature], error) {
	var fc FeatureCollection[AcquisitionFeature]
	err := c.do(ctx, http.MethodPost, "/catalogue", r, &fc)
	return fc, err
}

// AddItems adds archive or feasibility itemIds to the current shopcart.
func (c *Client) AddItems(ctx context.Context, body AddItemsRequest) error {
	return c.do(ctx, http.MethodPost, "/shopcart/addItems", body, nil)
}

// PatchShopcart sets the legally‑required purchase purpose.
func (c *Client) PatchShopcart(ctx context.Context, body ShopcartPatch) error {
	return c.do(ctx, http.MethodPatch, "/shopcart", body, nil)
}

// SubmitShopcart finalises the order and returns the orderId.
func (c *Client) SubmitShopcart(ctx context.Context) (string, error) {
	var resp struct {
		OrderID string `json:"orderId"`
	}
	err := c.do(ctx, http.MethodPost, "/shopcart/submit", nil, &resp)
	return resp.OrderID, err
}

// Order fetches an order by id and returns its status & item list.
func (c *Client) Order(ctx context.Context, id string) (*Order, error) {
	var o Order
	err := c.do(ctx, http.MethodGet, path.Join("/orders", id), nil, &o)
	return &o, err
}
