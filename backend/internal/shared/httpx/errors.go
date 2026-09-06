package httpx

import (
	"errors"
	"fmt"
)

// AppError is a transport-agnostic error carrying an HTTP-ish status code.
// Domain and application layers return these; the HTTP error handler maps them
// onto responses without importing Fiber into business code.
type AppError struct {
	Code    int    // HTTP status
	Message string // safe, user-facing message
	Err     error  // wrapped internal error (never sent to the client)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

// Constructors -------------------------------------------------------------

func NewBadRequest(msg string) *AppError    { return &AppError{Code: 400, Message: msg} }
func NewUnauthorized(msg string) *AppError  { return &AppError{Code: 401, Message: msg} }
func NewForbidden(msg string) *AppError     { return &AppError{Code: 403, Message: msg} }
func NewNotFound(msg string) *AppError      { return &AppError{Code: 404, Message: msg} }
func NewConflict(msg string) *AppError      { return &AppError{Code: 409, Message: msg} }
func NewUnprocessable(msg string) *AppError { return &AppError{Code: 422, Message: msg} }

// NewInternal wraps an unexpected error. The wrapped detail is logged, not sent.
func NewInternal(err error) *AppError {
	return &AppError{Code: 500, Message: "Internal server error", Err: err}
}

// AsAppError extracts an *AppError from err, or synthesises a 500 for anything
// unexpected.
func AsAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return NewInternal(err)
}
