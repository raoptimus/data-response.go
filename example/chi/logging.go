package main

import (
	"net/http"
	"time"

	"github.com/raoptimus/data-response.go/pkg/logger"
	dr "github.com/raoptimus/data-response.go/v2"
)

// Logging middleware logs requests and responses.
type Logging struct {
	logger  logger.Logger
	factory *dr.Factory
}

// NewLogging creates a new Logging middleware.
func NewLogging(logger logger.Logger, factory *dr.Factory) *Logging {
	return &Logging{logger: logger, factory: factory}
}

// ServeHTTP implements Middleware interface.
func (l *Logging) ServeHTTP(r *http.Request, next dr.Handler) dr.DataResponse {
	start := time.Now()
	// Log request
	l.logger.Info(
		r.Context(),
		"incoming request",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
	)

	// Execute handler
	resp := next.Handle(r)

	// Log response
	duration := time.Since(start)
	l.logger.Info(
		r.Context(),
		"request completed",
		"method", r.Method,
		"path", r.URL.Path,
		"status", resp.StatusCode(),
		"duration_ms", duration.Milliseconds(),
	)

	return resp
}
