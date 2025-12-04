package capella

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/robert.malhotra/go-sar-vendor/pkg/common"
)

// Common error codes returned by the Capella API.
const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeRateLimit     = "RATE_LIMIT_EXCEEDED"
	ErrCodeInternalError = "INTERNAL_ERROR"
)

// APIError is an alias for common.APIError for backwards compatibility.
type APIError = common.APIError

// HTTPValidationError represents a 422 Unprocessable Entity response.
type HTTPValidationError struct {
	Detail []ValidationError `json:"detail"`
}

// ValidationError provides details about a specific validation failure.
type ValidationError struct {
	Loc  []any  `json:"loc"`
	Msg  string `json:"msg"`
	Type string `json:"type"`
}

// parseError parses an HTTP error response into an APIError.
// This version handles Capella-specific validation errors.
func parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	apiErr := &common.APIError{
		StatusCode: resp.StatusCode,
		Code:       http.StatusText(resp.StatusCode),
		RawBody:    string(body),
	}

	// Try to parse as validation error (422)
	if resp.StatusCode == http.StatusUnprocessableEntity {
		var validationError HTTPValidationError
		if json.Unmarshal(body, &validationError) == nil && len(validationError.Detail) > 0 {
			apiErr.Code = ErrCodeValidation
			apiErr.Message = validationError.Detail[0].Msg
		}
	}

	// Try to parse as generic error response
	var genericError struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Detail  string `json:"detail"`
	}
	if json.Unmarshal(body, &genericError) == nil {
		if genericError.Code != "" {
			apiErr.Code = genericError.Code
		}
		if genericError.Message != "" {
			apiErr.Message = genericError.Message
		} else if genericError.Detail != "" {
			apiErr.Message = genericError.Detail
		}
	}

	return apiErr
}

// Error checking helpers - delegate to common package.
var (
	IsNotFound        = common.IsNotFound
	IsUnauthorized    = common.IsUnauthorized
	IsRateLimited     = common.IsRateLimited
	IsValidationError = common.IsValidationError
)
