// Package common provides shared types and utilities for SAR vendor API clients.
package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// APIError represents a standard API error response.
type APIError struct {
	// StatusCode is the HTTP status code.
	StatusCode int `json:"-"`

	// Code is the error code (e.g., "NOT_FOUND", "UNAUTHORIZED").
	Code string `json:"code,omitempty"`

	// Message is the human-readable error message.
	Message string `json:"message,omitempty"`

	// Detail provides additional error details.
	Detail string `json:"detail,omitempty"`

	// RawBody contains the raw response body for debugging.
	RawBody string `json:"-"`
}

func (e *APIError) Error() string {
	if e.Code != "" && e.Message != "" {
		return fmt.Sprintf("%s (%d): %s", e.Code, e.StatusCode, e.Message)
	}
	if e.Message != "" {
		return fmt.Sprintf("%s (%d)", e.Message, e.StatusCode)
	}
	if e.Detail != "" {
		return fmt.Sprintf("%s (%d)", e.Detail, e.StatusCode)
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, http.StatusText(e.StatusCode))
}

// IsNotFound returns true if the error is a 404 Not Found error.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsUnauthorized returns true if the error is a 401 Unauthorized error.
func IsUnauthorized(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsForbidden returns true if the error is a 403 Forbidden error.
func IsForbidden(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusForbidden
	}
	return false
}

// IsBadRequest returns true if the error is a 400 Bad Request error.
func IsBadRequest(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusBadRequest
	}
	return false
}

// IsRateLimited returns true if the error is a 429 Too Many Requests error.
func IsRateLimited(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusTooManyRequests
	}
	return false
}

// IsServerError returns true if the error is a 5xx server error.
func IsServerError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500
	}
	return false
}

// IsClientError returns true if the error is a 4xx client error.
func IsClientError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 400 && apiErr.StatusCode < 500
	}
	return false
}

// IsValidationError returns true if the error is a 422 Unprocessable Entity error.
func IsValidationError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusUnprocessableEntity
	}
	return false
}

// IsConflict returns true if the error is a 409 Conflict error.
func IsConflict(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusConflict
	}
	return false
}

// ParseErrorResponse parses an HTTP error response into an APIError.
func ParseErrorResponse(resp *http.Response) *APIError {
	body, _ := io.ReadAll(resp.Body)

	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Code:       http.StatusText(resp.StatusCode),
		RawBody:    string(body),
	}

	// Try to parse as JSON error response
	var jsonErr struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Detail  string `json:"detail"`
		Error   string `json:"error"`
		Title   string `json:"title"`
	}

	if json.Unmarshal(body, &jsonErr) == nil {
		if jsonErr.Code != "" {
			apiErr.Code = jsonErr.Code
		}
		if jsonErr.Message != "" {
			apiErr.Message = jsonErr.Message
		} else if jsonErr.Error != "" {
			apiErr.Message = jsonErr.Error
		} else if jsonErr.Title != "" {
			apiErr.Message = jsonErr.Title
		}
		if jsonErr.Detail != "" {
			apiErr.Detail = jsonErr.Detail
		}
	} else {
		// If not JSON, use raw body as message
		apiErr.Message = string(body)
	}

	return apiErr
}
