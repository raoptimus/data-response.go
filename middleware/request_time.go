package middleware

import (
	"context"
	"net/http"
	"time"

	dr "github.com/raoptimus/data-response.go/v2"
)

type contextKey string

const (
	// RequestStartTimeKey is the context key for storing request start time
	RequestStartTimeKey contextKey = "request_start_time"
)

// RequestTimer middleware sets the request start time in context.
// This should be the first middleware in the chain to get accurate timing.
func RequestTimer() dr.Middleware {
	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
			// Store start time in context
			ctx := context.WithValue(r.Context(), RequestStartTimeKey, time.Now())

			return next.Handle(r.WithContext(ctx), f)
		})
	}
}

// RequestStartTime retrieves the request start time from context.
// Returns current time if not found (fallback).
func RequestStartTime(ctx context.Context) time.Time {
	if start, ok := ctx.Value(RequestStartTimeKey).(time.Time); ok {
		return start
	}

	return time.Now() // Fallback to current time
}
