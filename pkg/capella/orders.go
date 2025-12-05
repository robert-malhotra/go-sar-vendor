package capella

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// OrderService provides order management operations.
type OrderService struct {
	client *Client
}

// NewOrderService creates a new order service.
func NewOrderService(client *Client) *OrderService {
	return &OrderService{client: client}
}

// ----------------------------------------------------------------------------
// Order Models
// ----------------------------------------------------------------------------

// OrderItem represents an item in an order.
type OrderItem struct {
	CollectionID string `json:"collectionId"`
	GranuleID    string `json:"granuleId"`
}

// OrderRequest represents an order submission request.
type OrderRequest struct {
	Items []OrderItem `json:"items"`
}

// OrderReviewRequest represents an order review request.
type OrderReviewRequest struct {
	Items []OrderItem `json:"items"`
}

// OrderItemResponse represents an item in an order response.
type OrderItemResponse struct {
	CollectionID string            `json:"collectionId"`
	GranuleID    string            `json:"granuleId"`
	Status       OrderStatus       `json:"status,omitempty"`
	Assets       map[string]Asset  `json:"assets,omitempty"`
}

// OrderReviewResponse represents the cost review response.
type OrderReviewResponse struct {
	Items         []OrderItemResponse `json:"items"`
	TotalCredits  float64             `json:"totalCredits,omitempty"`
	TotalCostUSD  float64             `json:"totalCostUsd,omitempty"`
	Errors        []OrderError        `json:"errors,omitempty"`
}

// OrderError represents an error for a specific order item.
type OrderError struct {
	GranuleID string `json:"granuleId"`
	Message   string `json:"message"`
	Code      string `json:"code,omitempty"`
}

// Order represents an order.
type Order struct {
	OrderID     string              `json:"orderId"`
	Status      OrderStatus         `json:"status"`
	Items       []OrderItemResponse `json:"items"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt,omitempty"`
	CompletedAt time.Time           `json:"completedAt,omitempty"`
}

// DownloadURL represents a presigned download URL for an asset.
type DownloadURL struct {
	GranuleID string    `json:"granuleId"`
	AssetKey  string    `json:"assetKey"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
	Size      int64     `json:"size,omitempty"`
	Checksum  string    `json:"checksum,omitempty"`
}

// DownloadURLsResponse represents the response for download URLs.
type DownloadURLsResponse struct {
	OrderID   string        `json:"orderId"`
	Downloads []DownloadURL `json:"downloads"`
}

// ----------------------------------------------------------------------------
// Order Operations
// ----------------------------------------------------------------------------

// ReviewOrder reviews an order to get cost information before submission.
func (s *OrderService) ReviewOrder(ctx context.Context, req OrderReviewRequest) (*OrderReviewResponse, error) {
	var resp OrderReviewResponse
	if err := s.client.Do(ctx, http.MethodPost, "/orders/review", 0, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SubmitOrder submits an order for processing.
func (s *OrderService) SubmitOrder(ctx context.Context, req OrderRequest) (*Order, error) {
	var resp Order
	if err := s.client.Do(ctx, http.MethodPost, "/orders", 0, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetOrder retrieves an order by ID.
func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	var resp Order
	if err := s.client.Do(ctx, http.MethodGet, "/orders/"+orderID, 0, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDownloadURLs retrieves presigned download URLs for an order.
func (s *OrderService) GetDownloadURLs(ctx context.Context, orderID string) (*DownloadURLsResponse, error) {
	var resp DownloadURLsResponse
	if err := s.client.Do(ctx, http.MethodGet, "/orders/"+orderID+"/download", 0, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// OrderTaskingRequest orders all assets for a tasking request.
func (s *OrderService) OrderTaskingRequest(ctx context.Context, taskingRequestID string) (*Order, error) {
	var resp Order
	if err := s.client.Do(ctx, http.MethodPost, "/orders/task/"+taskingRequestID, 0, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ----------------------------------------------------------------------------
// Download Service
// ----------------------------------------------------------------------------

// DownloadService provides asset download operations.
type DownloadService struct {
	client *Client
}

// NewDownloadService creates a new download service.
func NewDownloadService(client *Client) *DownloadService {
	return &DownloadService{client: client}
}

// DownloadProgress represents download progress information.
type DownloadProgress struct {
	URL           string
	BytesReceived int64
	TotalBytes    int64
	Percent       float64
}

// DownloadOptions configures download behavior.
type DownloadOptions struct {
	// ProgressCallback is called periodically with download progress
	ProgressCallback func(DownloadProgress)

	// Overwrite existing files (default: false)
	Overwrite bool
}

// ToFile downloads an asset from a URL to a local file.
func (s *DownloadService) ToFile(ctx context.Context, downloadURL, destPath string, opts *DownloadOptions) error {
	if opts == nil {
		opts = &DownloadOptions{}
	}

	// Create destination directory if needed
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if file exists
	if !opts.Overwrite {
		if _, err := os.Stat(destPath); err == nil {
			return fmt.Errorf("file already exists: %s", destPath)
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	// Execute request
	resp, err := s.client.HTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create destination file
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Download with progress tracking
	var written int64
	totalBytes := resp.ContentLength

	if opts.ProgressCallback != nil {
		// Use a progress-tracking reader
		buf := make([]byte, 32*1024) // 32KB buffer
		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				nw, writeErr := file.Write(buf[:n])
				if writeErr != nil {
					return fmt.Errorf("failed to write file: %w", writeErr)
				}
				written += int64(nw)

				var percent float64
				if totalBytes > 0 {
					percent = float64(written) / float64(totalBytes) * 100
				}
				opts.ProgressCallback(DownloadProgress{
					URL:           downloadURL,
					BytesReceived: written,
					TotalBytes:    totalBytes,
					Percent:       percent,
				})
			}
			if readErr != nil {
				if readErr == io.EOF {
					break
				}
				return fmt.Errorf("failed to read response: %w", readErr)
			}
		}
	} else {
		// Simple copy without progress tracking
		if _, err = io.Copy(file, resp.Body); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	return nil
}

// ToDirectory downloads an asset to a directory, preserving the filename from the URL.
func (s *DownloadService) ToDirectory(ctx context.Context, downloadURL, destDir string, opts *DownloadOptions) (string, error) {
	// Extract filename from URL
	filename := filepath.Base(downloadURL)
	if filename == "" || filename == "." || filename == "/" {
		filename = "download"
	}

	destPath := filepath.Join(destDir, filename)
	if err := s.ToFile(ctx, downloadURL, destPath, opts); err != nil {
		return "", err
	}

	return destPath, nil
}

// ----------------------------------------------------------------------------
// Convenience Methods
// ----------------------------------------------------------------------------

// QuickOrder is a convenience method that reviews and submits an order in one call.
func (s *OrderService) QuickOrder(ctx context.Context, items []OrderItem) (*Order, error) {
	// First review the order
	review, err := s.ReviewOrder(ctx, OrderReviewRequest{Items: items})
	if err != nil {
		return nil, fmt.Errorf("order review failed: %w", err)
	}

	// Check for errors
	if len(review.Errors) > 0 {
		return nil, fmt.Errorf("order review has errors: %s", review.Errors[0].Message)
	}

	// Submit the order
	return s.SubmitOrder(ctx, OrderRequest{Items: items})
}

// WaitForOrder polls the order status until it completes or times out.
func (s *OrderService) WaitForOrder(ctx context.Context, orderID string, pollInterval time.Duration) (*Order, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			order, err := s.GetOrder(ctx, orderID)
			if err != nil {
				return nil, err
			}

			switch order.Status {
			case OrderCompleted:
				return order, nil
			case OrderFailed, OrderCanceled:
				return order, fmt.Errorf("order %s: %s", order.Status, orderID)
			}
			// Continue polling for pending/processing status
		}
	}
}
