package umbra

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

// FeasibilityStatus represents the status of a feasibility request.
type FeasibilityStatus string

const (
	FeasibilityStatusReceived  FeasibilityStatus = "RECEIVED"
	FeasibilityStatusCompleted FeasibilityStatus = "COMPLETED"
	FeasibilityStatusError     FeasibilityStatus = "ERROR"
)

// Opportunity represents a feasible imaging window.
type Opportunity struct {
	WindowStartAt                  time.Time `json:"windowStartAt"`
	WindowEndAt                    time.Time `json:"windowEndAt"`
	DurationSec                    float64   `json:"durationSec"`
	GrazingAngleStartDegrees       float64   `json:"grazingAngleStartDegrees"`
	GrazingAngleEndDegrees         float64   `json:"grazingAngleEndDegrees"`
	TargetAzimuthAngleStartDegrees float64   `json:"targetAzimuthAngleStartDegrees"`
	TargetAzimuthAngleEndDegrees   float64   `json:"targetAzimuthAngleEndDegrees"`
	SquintAngleStartDegrees        float64   `json:"squintAngleStartDegrees"`
	SquintAngleEndDegrees          float64   `json:"squintAngleEndDegrees"`
	SlantRangeStartKm              float64   `json:"slantRangeStartKm"`
	SlantRangeEndKm                float64   `json:"slantRangeEndKm"`
	GroundRangeStartKm             float64   `json:"groundRangeStartKm"`
	GroundRangeEndKm               float64   `json:"groundRangeEndKm"`
	SatelliteID                    string    `json:"satelliteId,omitempty"`
}

// Feasibility represents a feasibility request and its results.
type Feasibility struct {
	ID                   string                `json:"id"`
	Status               FeasibilityStatus     `json:"status"`
	ImagingMode          ImagingMode           `json:"imagingMode"`
	SpotlightConstraints *SpotlightConstraints `json:"spotlightConstraints,omitempty"`
	ScanConstraints      *ScanConstraints      `json:"scanConstraints,omitempty"`
	WindowStartAt        time.Time             `json:"windowStartAt"`
	WindowEndAt          time.Time             `json:"windowEndAt"`
	Opportunities        []Opportunity         `json:"opportunities,omitempty"`
	CreatedAt            time.Time             `json:"createdAt"`
	UpdatedAt            time.Time             `json:"updatedAt"`
}

// CreateFeasibilityRequest contains parameters for a feasibility check.
type CreateFeasibilityRequest struct {
	ImagingMode          ImagingMode           `json:"imagingMode"`
	SpotlightConstraints *SpotlightConstraints `json:"spotlightConstraints,omitempty"`
	ScanConstraints      *ScanConstraints      `json:"scanConstraints,omitempty"`
	WindowStartAt        time.Time             `json:"windowStartAt"`
	WindowEndAt          time.Time             `json:"windowEndAt"`
}

// FeasibilityListResponse contains a paginated list of feasibility requests.
type FeasibilityListResponse struct {
	Feasibilities []Feasibility `json:"feasibilities"`
	TotalCount    int           `json:"totalCount"`
	Limit         int           `json:"limit"`
	Offset        int           `json:"offset"`
}

// CreateFeasibility submits a new feasibility request.
// POST /tasking/feasibilities
func (c *Client) CreateFeasibility(ctx context.Context, req *CreateFeasibilityRequest) (*Feasibility, error) {
	body, err := common.MarshalBody(req)
	if err != nil {
		return nil, err
	}
	var out Feasibility
	err = c.DoRaw(ctx, http.MethodPost, c.BaseURL().JoinPath("tasking", "feasibilities"), body, http.StatusCreated, &out)
	return &out, err
}

// GetFeasibility retrieves a feasibility request by ID.
// GET /tasking/feasibilities/{id}
func (c *Client) GetFeasibility(ctx context.Context, id string) (*Feasibility, error) {
	var out Feasibility
	err := c.DoRaw(ctx, http.MethodGet, c.BaseURL().JoinPath("tasking", "feasibilities", id), nil, http.StatusOK, &out)
	return &out, err
}

// ListFeasibilities retrieves all feasibility requests.
// GET /tasking/feasibilities
func (c *Client) ListFeasibilities(ctx context.Context, opts *ListOptions) (*FeasibilityListResponse, error) {
	u := c.BaseURL().JoinPath("tasking", "feasibilities")
	if opts != nil {
		q := u.Query()
		if opts.Limit > 0 {
			q.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Offset > 0 {
			q.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
		u.RawQuery = q.Encode()
	}
	var resp FeasibilityListResponse
	err := c.DoRaw(ctx, http.MethodGet, u, nil, http.StatusOK, &resp)
	return &resp, err
}

// WaitForFeasibilityCompletion polls until the feasibility request is complete or times out.
func (c *Client) WaitForFeasibilityCompletion(ctx context.Context, id string, opts *WaitOptions) (*Feasibility, error) {
	if opts == nil {
		opts = &WaitOptions{
			PollInterval: 2 * time.Second,
			Timeout:      5 * time.Minute,
		}
	}

	deadline := time.Now().Add(opts.Timeout)
	ticker := time.NewTicker(opts.PollInterval)
	defer ticker.Stop()

	for {
		f, err := c.GetFeasibility(ctx, id)
		if err != nil {
			return nil, err
		}

		if f.Status == FeasibilityStatusCompleted || f.Status == FeasibilityStatusError {
			return f, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for feasibility %s", id)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
