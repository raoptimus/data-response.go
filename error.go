package dataresponse

import (
	"fmt"

	"github.com/pkg/errors"
)

// Error wraps an error with stack trace preservation.
// It provides additional context and preserves the call stack for debugging.
type Error struct {
	original error
	message  string
	code     int
}

// NewError creates a new error with stack trace.
func NewError(code int, message string) *Error {
	return &Error{
		original: errors.New(message),
		message:  message,
		code:     code,
	}
}

// WrapError wraps an existing error with additional context.
func WrapError(code int, err error, message string) *Error {
	return &Error{
		original: errors.Wrap(err, message),
		message:  message,
		code:     code,
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.message
}

// Unwrap returns the original error for errors.Is/As.
func (e *Error) Unwrap() error {
	return e.original
}

// Code returns the HTTP status code.
func (e *Error) Code() int {
	return e.code
}

// StackTrace returns a formatted stack trace if available.
func (e *Error) StackTrace() string {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	if err, ok := e.original.(stackTracer); ok {
		return fmt.Sprintf("%+v", err.StackTrace())
	}
	return ""
}
