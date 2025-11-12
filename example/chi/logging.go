package main

import (
	"net/http"
	"time"

	dataresponse "github.com/raoptimus/data-response.go"
	"github.com/raoptimus/data-response.go/pkg/logger"
)

// Logging middleware logs requests and responses.
type Logging struct {
	logger  logger.Logger
	factory *dataresponse.Factory
}

// NewLogging creates a new Logging middleware.
func NewLogging(logger logger.Logger, factory *dataresponse.Factory) *Logging {
	return &Logging{logger: logger, factory: factory}
}

// ServeHTTP implements Middleware interface.
func (l *Logging) ServeHTTP(r *http.Request, next dataresponse.Handler) dataresponse.DataResponse {
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
