package iceye

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Common error codes returned by the ICEYE API.
const (
	ErrCodeInsufficientFunds = "ERR_INSUFFICIENT_FUNDS"
	ErrCodeInvalidContract   = "ERR_INVALID_CONTRACT"
	ErrCodeTaskNotFound      = "ERR_TASK_NOT_FOUND"
	ErrCodeSceneUnavailable  = "ERR_SCENE_UNAVAILABLE"
)

// Error represents an ICEYE API error (RFC 7807 Problem Details).
type Error struct {
	Status int    `json:"status"`
	Code   string `json:"code"`
	Detail string `json:"detail"`
	Title  string `json:"title,omitempty"`
	Type   string `json:"type,omitempty"`
}

func (e *Error) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("iceye: %s (%d) – %s", e.Code, e.Status, e.Detail)
	}
	return fmt.Sprintf("iceye: %s (%d)", e.Code, e.Status)
}

func parseError(resp *http.Response) error {
	var e Error
	e.Status = resp.StatusCode
	body, _ := io.ReadAll(resp.Body)
	if json.Unmarshal(body, &e) != nil || e.Code == "" {
		e.Code = http.StatusText(resp.StatusCode)
		e.Detail = string(body)
	}
	return &e
}

// Error checking helpers using errors.As for compatibility.

// IsNotFound returns true if the error is a 404 Not Found error.
func IsNotFound(err error) bool {
	var e *Error
	if ok := isError(err, &e); ok {
		return e.Status == http.StatusNotFound
	}
	return false
}

// IsUnauthorized returns true if the error is a 401 Unauthorized error.
func IsUnauthorized(err error) bool {
	var e *Error
	if ok := isError(err, &e); ok {
		return e.Status == http.StatusUnauthorized
	}
	return false
}

// IsRateLimited returns true if the error is a 429 Too Many Requests error.
func IsRateLimited(err error) bool {
	var e *Error
	if ok := isError(err, &e); ok {
		return e.Status == http.StatusTooManyRequests
	}
	return false
}

// IsForbidden returns true if the error is a 403 Forbidden error.
func IsForbidden(err error) bool {
	var e *Error
	if ok := isError(err, &e); ok {
		return e.Status == http.StatusForbidden
	}
	return false
}

// IsBadRequest returns true if the error is a 400 Bad Request error.
func IsBadRequest(err error) bool {
	var e *Error
	if ok := isError(err, &e); ok {
		return e.Status == http.StatusBadRequest
	}
	return false
}

// IsServerError returns true if the error is a 5xx server error.
func IsServerError(err error) bool {
	var e *Error
	if ok := isError(err, &e); ok {
		return e.Status >= 500
	}
	return false
}

func isError(err error, target **Error) bool {
	if e, ok := err.(*Error); ok {
		*target = e
		return true
	}
	return false
}
