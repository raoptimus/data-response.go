package response

import (
	"context"
	"time"
)

type contextKey string

const (
	// RequestStartTimeKey is the context key for storing request start time
	RequestStartTimeKey contextKey = "request_start_time"
)

func WithRequestStartTime(ctx context.Context) context.Context {
	return context.WithValue(ctx, RequestStartTimeKey, time.Now())
}

// RequestStartTime retrieves the request start time from context.
// Returns current time if not found (fallback).
func RequestStartTime(ctx context.Context) time.Time {
	if start, ok := ctx.Value(RequestStartTimeKey).(time.Time); ok {
		return start
	}

	return time.Now() // Fallback to current time
}
