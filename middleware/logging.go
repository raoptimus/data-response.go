/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package middleware

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"text/template"
	"time"

	dr "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/response"
)

const DefaultLogTemplate = `{{.RemoteAddr}} - {{.User}} [{{.Time}}] "{{.Method}} {{.URI}} {{.Proto}}" {{.Status}} {{.Size}} "{{.Referer}}" "{{.UserAgent}}" {{.Duration}}`

// LogData contains all available fields for logging template
type LogData struct {
	// Request fields
	RemoteAddr string
	Method     string
	URI        string
	Proto      string
	Host       string
	Referer    string
	UserAgent  string

	// Response fields
	Status int
	Size   string

	// Timing
	Time     string
	Duration string

	// User identification (from context)
	User string

	// Request ID (from context)
	RequestID string

	// Custom fields from context
	Custom map[string]interface{}
}

// ContextValueFunc extracts a value from request context
type ContextValueFunc func(ctx context.Context) interface{}

// LoggingConfig configures the logging middleware
type LoggingConfig struct {
	// Template string for log format
	// If empty, DefaultLogTemplate is used
	Template string

	// TimeFormat for formatting timestamps
	// Default: time.RFC3339
	TimeFormat string

	// SkipPaths that should not be logged
	SkipPaths []string

	// Custom template functions
	TemplateFuncs template.FuncMap

	// ContextFields maps template field names to context extraction functions
	ContextFields map[string]ContextValueFunc
}

var (
	defaultLoggingMiddleware dr.Middleware
	defaultLoggingOnce       sync.Once
)

// Logging creates a new logging middleware with custom configuration
func Logging(cfg *LoggingConfig) (dr.Middleware, error) {
	if cfg == nil {
		cfg = &LoggingConfig{}
	}

	// Set defaults
	if cfg.Template == "" {
		cfg.Template = DefaultLogTemplate
	}
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}
	if cfg.ContextFields == nil {
		cfg.ContextFields = make(map[string]ContextValueFunc)
	}

	// Compile template
	tmpl := template.New("access_log")
	if cfg.TemplateFuncs != nil {
		tmpl = tmpl.Funcs(cfg.TemplateFuncs)
	}

	tmpl, err := tmpl.Parse(cfg.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log template: %w", err)
	}

	// Create skip paths map for O(1) lookup
	skipPaths := make(map[string]bool, len(cfg.SkipPaths))
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) *response.DataResponse {
			// Skip logging for certain paths
			if skipPaths[r.URL.Path] {
				return next.Handle(r, f)
			}

			start := response.RequestStartTime(r.Context())
			resp := next.Handle(r, f)
			duration := time.Since(start)

			// Build log data
			logData := &LogData{
				RemoteAddr: r.RemoteAddr,
				Method:     r.Method,
				URI:        r.RequestURI,
				Proto:      r.Proto,
				Host:       r.Host,
				Referer:    r.Referer(),
				UserAgent:  r.UserAgent(),
				Status:     resp.StatusCode(),
				Size:       resp.HeaderLine(response.HeaderContentLength),
				Time:       start.Format(cfg.TimeFormat),
				Duration:   duration.String(),
				User:       "-",
				RequestID:  resp.HeaderLine(response.HeaderXRequestID),
				Custom:     make(map[string]any),
			}

			// Extract common values from context
			if requestID := r.Context().Value("requestID"); requestID != nil {
				if id, ok := requestID.(string); ok {
					logData.RequestID = id
				}
			}

			if user := r.Context().Value("user"); user != nil {
				if username, ok := user.(string); ok {
					logData.User = username
				}
			}

			// Extract custom fields from context
			for fieldName, extractFunc := range cfg.ContextFields {
				if value := extractFunc(r.Context()); value != nil {
					logData.Custom[fieldName] = value
				}
			}

			// Render template
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, logData); err != nil {
				f.Logger().Error(r.Context(), "failed to render log template: "+err.Error())
				return resp
			}

			f.Logger().Info(r.Context(), buf.String())

			return resp
		})
	}, nil
}

// LoggingDefault returns a logging middleware with default configuration.
// It is safe for concurrent use and initialized only once.
func LoggingDefault() dr.Middleware {
	defaultLoggingOnce.Do(func() {
		mw, err := Logging(&LoggingConfig{
			Template:   DefaultLogTemplate,
			TimeFormat: time.RFC3339,
		})
		if err != nil {
			// This should never happen with default config, but handle it gracefully
			defaultLoggingMiddleware = func(next dr.Handler) dr.Handler {
				return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) *response.DataResponse {
					f.Logger().Error(r.Context(), "failed to initialize default logging middleware: "+err.Error())

					return next.Handle(r, f)
				})
			}
			return
		}
		defaultLoggingMiddleware = mw
	})

	return defaultLoggingMiddleware
}
