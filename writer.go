package dataresponse

import (
	"context"
	"net/http"
)

type contextKey string

const formatterKey contextKey = "formatter"

// Write writes a DataResponse using the formatter from context or factory default.
// It handles both structured data and binary responses automatically.
func Write(w http.ResponseWriter, r *http.Request, resp DataResponse, factory *Factory) error {
	formatter := FormatterFromContext(r.Context())
	if formatter == nil {
		formatter = factory.Formatter()
	}

	// If binary and formatter doesn't support it, use binary formatter
	if resp.IsBinary() && !formatter.CanFormatBinary() {
		if binaryFmt := defaultBinaryFormatter(); binaryFmt != nil {
			formatter = binaryFmt
		}
	}

	return formatter.Format(w, resp)
}

// FormatterFromContext retrieves formatter from context.
// Returns nil if not found in context.
func FormatterFromContext(ctx context.Context) Formatter {
	if f, ok := ctx.Value(formatterKey).(Formatter); ok {
		return f
	}
	return nil
}

// ContextWithFormatter returns a new context with the formatter.
// Use this in middleware to set formatter based on Accept header.
func ContextWithFormatter(ctx context.Context, formatter Formatter) context.Context {
	return context.WithValue(ctx, formatterKey, formatter)
}

var defaultBinaryFormatterFunc func() Formatter

// defaultBinaryFormatter returns default binary formatter (will be set by formatters package).
func defaultBinaryFormatter() Formatter {
	if defaultBinaryFormatterFunc != nil {
		return defaultBinaryFormatterFunc()
	}
	return nil
}

// SetDefaultBinaryFormatter registers default binary formatter.
func SetDefaultBinaryFormatter(fn func() Formatter) {
	defaultBinaryFormatterFunc = fn
}
