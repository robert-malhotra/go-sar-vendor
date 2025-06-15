package umbra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"iter"
)

/*──────────────── ENUMS & BASICS ───────────────*/

type ImagingMode string

const (
	SPOTLIGHT_MODE ImagingMode = "SPOTLIGHT"
)

type Client struct {
	HTTPClient *http.Client
	baseURL    *url.URL
}

func NewClient(rawURL string) (*Client, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    u,
	}, nil
}

/*──────────────── LOW-LEVEL HTTP ──────────────*/

func (c *Client) doRequest(token, method string, u *url.URL, body io.Reader, want int, out any) error {
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if method == http.MethodPost || method == http.MethodPatch {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	if resp.StatusCode != want {
		return fmt.Errorf("unexpected status %s – %s", resp.Status, buf)
	}
	if out != nil {
		if err := json.Unmarshal(buf, out); err != nil {
			return fmt.Errorf("decode: %w", err)
		}
	}
	return nil
}

/*──────────────── GENERIC PAGINATION ──────────
  Seq2[K,V] → yield(K, V) bool , returns bool.
  We yield each value with a nil-error. If the page
  fetch fails, we yield the zero value and the error
  once, then stop.                                       */

func pagedSeq[T any](
	startSkip, limit int,
	fetch func(skip, limit int) ([]T, error),
) iter.Seq2[T, error] {

	if limit <= 0 {
		limit = 100 // Default limit
	}

	return func(yield func(T, error) bool) {
		var zero T
		skip := startSkip

		for {
			page, err := fetch(skip, limit)
			if err != nil {
				yield(zero, err) // Propagate error once and stop.
				return
			}

			for _, v := range page {
				if !yield(v, nil) {
					return // Consumer ended the iteration.
				}
			}

			// If the page is empty or returned fewer items than the limit,
			// it must be the last page.
			if len(page) < limit {
				return // End of data.
			}
			skip += len(page)
		}
	}
}

/*──────────────── DOMAIN TYPES (unchanged) ───*/

type PointGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type SpotlightConstraints struct {
	ImagingMode                    ImagingMode   `json:"imagingMode,omitempty"`
	Geometry                       PointGeometry `json:"geometry"`
	Polarization                   string        `json:"polarization,omitempty"`
	RangeResolutionMinMeters       float64       `json:"rangeResolutionMinMeters,omitempty"`
	MultilookFactor                int           `json:"multilookFactor,omitempty"`
	GrazingAngleMinDegrees         int           `json:"grazingAngleMinDegrees,omitempty"`
	GrazingAngleMaxDegrees         int           `json:"grazingAngleMaxDegrees,omitempty"`
	TargetAzimuthAngleStartDegrees int           `json:"targetAzimuthAngleStartDegrees,omitempty"`
	TargetAzimuthAngleEndDegrees   int           `json:"targetAzimuthAngleEndDegrees,omitempty"`
	SceneSize                      string        `json:"sceneSize,omitempty"`
}

type TaskingRequest struct {
	ImagingMode          ImagingMode          `json:"imagingMode,omitempty"`
	SpotlightConstraints SpotlightConstraints `json:"spotlightConstraints"`
	WindowStartAt        time.Time            `json:"windowStartAt"`
	WindowEndAt          time.Time            `json:"windowEndAt"`
}

type TaskRequest struct {
	TaskingRequest
	ProductTypes []string `json:"productTypes,omitempty"`
	Priority     int      `json:"priority,omitempty"`
}

type TaskStatus string

const (
	TaskActive    TaskStatus = "ACTIVE"
	TaskSubmitted TaskStatus = "SUBMITTED"
	TaskScheduled TaskStatus = "SCHEDULED"
	TaskCanceled  TaskStatus = "CANCELED"
)

type StatusEvent struct {
	Status    TaskStatus `json:"status"`
	Timestamp time.Time  `json:"timestamp"`
}

type Task struct {
	ID           string         `json:"id"`
	Status       TaskStatus     `json:"status"`
	CollectIDs   []string       `json:"collectIds,omitempty"`
	Request      TaskRequest    `json:"taskRequest"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	StatusEvents []StatusEvent  `json:"statusHistory,omitempty"`
	Extra        map[string]any `json:"-"`
}

type Collect struct {
	ID        string         `json:"id"`
	TaskID    string         `json:"taskId"`
	Status    string         `json:"status"`
	Geometry  PointGeometry  `json:"geometry"`
	Satellite string         `json:"satelliteId"`
	CreatedAt time.Time      `json:"createdAt"`
	Extra     map[string]any `json:"-"`
}

type ProductConstraint struct {
	ProductType       string  `json:"productType"`
	SceneSize         string  `json:"sceneSize"`
	MinGrazingDegrees float64 `json:"minGrazingAngle"`
	MaxGrazingDegrees float64 `json:"maxGrazingAngle"`
	RecommendedLooks  int     `json:"recommendedLooks"`
}

/*──────────────── SEARCH REQUEST STRUCTS ─────*/

type SearchRequest[T any] struct {
	Limit  *int      `json:"limit,omitempty"`
	Skip   *int      `json:"skip,omitempty"`
	SortBy *[]string `json:"sortBy,omitempty"`
	Query  T         `json:"query"`
}

type TaskSearchRequest SearchRequest[map[string]any]
type CollectSearchRequest SearchRequest[map[string]any]

/*──────────────── FEASIBILITY (methods unchanged) ─*/

type FeasibilityResponse struct {
	ID      string         `json:"id"`
	Status  string         `json:"status"`
	Request TaskingRequest `json:"feasibilityRequest"`
	Results []Opportunity  `json:"opportunities"`
}

type Opportunity struct {
	WindowStartAt                      time.Time `json:"windowStartAt"`
	WindowEndAt                        time.Time `json:"windowEndAt"`
	DurationSec                        int       `json:"durationSec"`
	GrazingAngleStartDegrees           float64   `json:"grazingAngleStartDegrees"`
	GrazingAngleEndDegrees             float64   `json:"grazingAngleEndDegrees"`
	TargetAzimuthAngleStartDegrees     float64   `json:"targetAzimuthAngleStartDegrees"`
	TargetAzimuthAngleEndDegrees       float64   `json:"targetAzimuthAngleEndDegrees"`
	SquintAngleStartDegrees            float64   `json:"squintAngleStartDegrees"`
	SquintAngleEndDegrees              float64   `json:"squintAngleEndDegrees"`
	SquintAngleEngineeringDegreesStart float64   `json:"squintAngleEngineeringDegreesStart"`
	SquintAngleEngineeringDegreesEnd   float64   `json:"squintAngleEngineeringDegreesEnd"`
	SlantRangeStartKm                  float64   `json:"slantRangeStartKm"`
	SlantRangeEndKm                    float64   `json:"slantRangeEndKm"`
	GroundRangeStartKm                 float64   `json:"groundRangeStartKm"`
	GroundRangeEndKm                   float64   `json:"groundRangeEndKm"`
	SatelliteId                        string    `json:"satelliteId"`
}

func (c *Client) CreateFeasibility(tok string, req *TaskingRequest) (*FeasibilityResponse, error) {
	b, _ := json.Marshal(req)
	var out FeasibilityResponse
	err := c.doRequest(tok, http.MethodPost, c.baseURL.JoinPath("tasking", "feasibilities"), bytes.NewBuffer(b), http.StatusCreated, &out)
	return &out, err
}
func (c *Client) GetFeasibility(tok, id string) (*FeasibilityResponse, error) {
	var out FeasibilityResponse
	err := c.doRequest(tok, http.MethodGet, c.baseURL.JoinPath("tasking", "feasibilities", id), nil, http.StatusOK, &out)
	return &out, err
}

/*──────────────── TASK helpers ───────────────*/

func (c *Client) CreateTask(tok string, req *TaskRequest) (*Task, error) {
	b, _ := json.Marshal(req)
	var t Task
	err := c.doRequest(tok, http.MethodPost, c.baseURL.JoinPath("tasking", "tasks"), bytes.NewBuffer(b), http.StatusCreated, &t)
	return &t, err
}
func (c *Client) GetTask(tok, id string) (*Task, error) {
	var t Task
	err := c.doRequest(tok, http.MethodGet, c.baseURL.JoinPath("tasking", "tasks", id), nil, http.StatusOK, &t)
	return &t, err
}
func (c *Client) CancelTask(tok, id string) (*Task, error) {
	var t Task
	err := c.doRequest(tok, http.MethodPatch, c.baseURL.JoinPath("tasking", "tasks", id, "cancel"), nil, http.StatusOK, &t)
	return &t, err
}

/*──────────────── SEARCH (Seq2) ──────────────*/

func (c *Client) SearchTasks(tok string, req TaskSearchRequest) iter.Seq2[Task, error] {
	fetch := func(skip, limit int) ([]Task, error) {
		b, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		var page []Task
		err = c.doRequest(tok, http.MethodPost, c.baseURL.JoinPath("tasking", "tasks", "search"), bytes.NewBuffer(b), http.StatusOK, &page)
		return page, err
	}
	return pagedSeq(*req.Skip, *req.Limit, fetch)
}

func (c *Client) SearchCollects(tok string, req CollectSearchRequest) iter.Seq2[Collect, error] {
	fetch := func(skip, limit int) ([]Collect, error) {
		b, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		var page []Collect
		err = c.doRequest(tok, http.MethodPost, c.baseURL.JoinPath("tasking", "collects", "search"), bytes.NewBuffer(b), http.StatusOK, &page)
		return page, err
	}
	return pagedSeq(*req.Skip, *req.Limit, fetch)
}

/*──────────────── COLLECT & PRODUCT helpers ──*/

func (c *Client) GetCollect(tok, id string) (*Collect, error) {
	var col Collect
	err := c.doRequest(tok, http.MethodGet, c.baseURL.JoinPath("tasking", "collects", id), nil, http.StatusOK, &col)
	return &col, err
}

func (c *Client) GetProductConstraints(tok string, mode ImagingMode) ([]ProductConstraint, error) {
	var pc []ProductConstraint
	err := c.doRequest(tok, http.MethodGet, c.baseURL.JoinPath("tasking", "products", string(mode), "constraints"), nil, http.StatusOK, &pc)
	return pc, err
}
