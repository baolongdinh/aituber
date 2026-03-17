package apperror

import "net/http"

// Error codes used across the application
const (
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeBadRequest       = "BAD_REQUEST"
	ErrCodeConflict         = "CONFLICT"
	ErrCodeInternal         = "INTERNAL_ERROR"
	ErrCodeValidationFailed = "VALIDATION_FAILED"
)

// AppError is a structured application error
type AppError struct {
	Code       string
	Message    string
	Details    string
	Err        error
	HTTPStatus int
}

func (e *AppError) Error() string { return e.Message }

// NotFound creates a 404 error
func NotFound(msg string) *AppError {
	return &AppError{Code: ErrCodeNotFound, Message: msg, HTTPStatus: http.StatusNotFound}
}

// Unauthorized creates a 401 error
func Unauthorized(msg string) *AppError {
	return &AppError{Code: ErrCodeUnauthorized, Message: msg, HTTPStatus: http.StatusUnauthorized}
}

// Forbidden creates a 403 error
func Forbidden(msg string) *AppError {
	return &AppError{Code: ErrCodeForbidden, Message: msg, HTTPStatus: http.StatusForbidden}
}

// BadRequest creates a 400 error
func BadRequest(msg string) *AppError {
	return &AppError{Code: ErrCodeBadRequest, Message: msg, HTTPStatus: http.StatusBadRequest}
}

// Conflict creates a 409 error
func Conflict(msg string) *AppError {
	return &AppError{Code: ErrCodeConflict, Message: msg, HTTPStatus: http.StatusConflict}
}

// Internal creates a 500 error
func Internal(err error, msg string) *AppError {
	return &AppError{Code: ErrCodeInternal, Message: msg, Err: err, HTTPStatus: http.StatusInternalServerError}
}

// ValidationFailed creates a 400 validation error with details
func ValidationFailed(msg, details string) *AppError {
	return &AppError{Code: ErrCodeValidationFailed, Message: msg, Details: details, HTTPStatus: http.StatusBadRequest}
}
