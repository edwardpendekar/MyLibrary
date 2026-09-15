// Package apperror defines a transport-agnostic application error type.
// Handlers map Codes to HTTP status codes; services/repositories never import net/http.
package apperror

import (
	"errors"
	"fmt"
)

type Code string

const (
	CodeValidation   Code = "VALIDATION_ERROR"
	CodeNotFound     Code = "NOT_FOUND"
	CodeConflict     Code = "CONFLICT"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeRateLimited  Code = "RATE_LIMITED"
	CodeInternal     Code = "INTERNAL_ERROR"
)

// Error is the canonical application error. It carries a Code the transport layer
// maps to an HTTP status, a human message safe to show to clients, and optional
// field-level validation details.
type Error struct {
	Code    Code
	Message string
	Fields  map[string]string
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

func new(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func Validation(message string, fields map[string]string) *Error {
	return &Error{Code: CodeValidation, Message: message, Fields: fields}
}

func NotFound(message string) *Error {
	return new(CodeNotFound, message)
}

func Conflict(message string) *Error {
	return new(CodeConflict, message)
}

func Unauthorized(message string) *Error {
	return new(CodeUnauthorized, message)
}

func Forbidden(message string) *Error {
	return new(CodeForbidden, message)
}

func RateLimited(message string) *Error {
	return new(CodeRateLimited, message)
}

func Internal(message string, cause error) *Error {
	return &Error{Code: CodeInternal, Message: message, cause: cause}
}

// As extracts an *Error from err, if any.
func As(err error) (*Error, bool) {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
