package airbus

import (
	"github.com/robert-malhotra/go-sar-vendor/pkg/common"
)

// APIError is an alias for common.APIError.
type APIError = common.APIError

// Error checking functions from common package.
var (
	// IsNotFound returns true if the error is a 404 Not Found error.
	IsNotFound = common.IsNotFound

	// IsUnauthorized returns true if the error is a 401 Unauthorized error.
	IsUnauthorized = common.IsUnauthorized

	// IsForbidden returns true if the error is a 403 Forbidden error.
	IsForbidden = common.IsForbidden

	// IsBadRequest returns true if the error is a 400 Bad Request error.
	IsBadRequest = common.IsBadRequest

	// IsRateLimited returns true if the error is a 429 Too Many Requests error.
	IsRateLimited = common.IsRateLimited

	// IsServerError returns true if the error is a 5xx server error.
	IsServerError = common.IsServerError

	// IsClientError returns true if the error is a 4xx client error.
	IsClientError = common.IsClientError

	// IsValidationError returns true if the error is a 422 Unprocessable Entity error.
	IsValidationError = common.IsValidationError

	// IsConflict returns true if the error is a 409 Conflict error.
	IsConflict = common.IsConflict
)
