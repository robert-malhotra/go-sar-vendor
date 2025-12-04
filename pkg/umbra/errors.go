package umbra

import (
	"github.com/robert.malhotra/go-sar-vendor/pkg/common"
)

// APIError is an alias for common.APIError for backwards compatibility.
type APIError = common.APIError

// Error checking helpers - delegate to common package.
var (
	IsNotFound     = common.IsNotFound
	IsRateLimited  = common.IsRateLimited
	IsUnauthorized = common.IsUnauthorized
	IsBadRequest   = common.IsBadRequest
	IsForbidden    = common.IsForbidden
	IsServerError  = common.IsServerError
)
