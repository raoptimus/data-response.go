/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package dataresponse

import (
	"context"
	"io"
	"net/http"
	"strconv"
)

type contextKey string

const formatterKey contextKey = "formatter"

// Write writes a DataResponse to http.ResponseWriter.
// It handles formatting, headers, and body writing.
func Write(ctx context.Context, w http.ResponseWriter, resp DataResponse, factory *Factory) error {
	formatter, ok := resp.Formatter()
	if !ok {
		formatter = factory.Formatter()
	}

	// If binary and formatter doesn't support it, use binary formatter
	if resp.IsBinary() && !formatter.CanFormatBinary() {
		if binaryFmt := defaultBinaryFormatter(); binaryFmt != nil {
			formatter = binaryFmt
		}
	}

	// Format the response
	formatted, err := formatter.Format(resp)
	if err != nil {
		factory.logger.Error(ctx, "failed to format response",
			"error", err.Error(),
			"status", resp.StatusCode(),
		)

		// Try to return error response
		return writeErrorResponse(ctx, w, factory, formatter, err)
	}

	// Write formatted response
	if writeErr := writeFormattedResponse(w, resp, formatted); writeErr != nil {
		factory.logger.Error(ctx, "failed to write response",
			"error", writeErr.Error(),
			"status", resp.StatusCode(),
		)

		return writeErr
	}

	return nil
}

// writeErrorResponse attempts to write an error response when formatting fails.
func writeErrorResponse(ctx context.Context, w http.ResponseWriter, factory *Factory, formatter Formatter, originalErr error) error {
	// Create internal error response
	errorResp := factory.InternalError(ctx, originalErr)

	// Try to format error response
	formatted, formatErr := formatter.Format(errorResp)
	if formatErr != nil {
		factory.logger.Error(ctx, "failed to format error response",
			"error", formatErr.Error(),
		)

		// Last resort - write minimal error
		if minimalErr := writeMinimalError(w, factory, ctx); minimalErr != nil {
			// Absolutely nothing we can do - connection is probably dead
			factory.logger.Error(ctx, "failed to write minimal error",
				"error", minimalErr.Error(),
			)
		}

		return WrapError(http.StatusInternalServerError, formatErr,
			"failed to format error response")
	}

	// Try to write formatted error response
	if writeErr := writeFormattedResponse(w, errorResp, formatted); writeErr != nil {
		factory.logger.Error(ctx, "failed to write error response",
			"error", writeErr.Error(),
		)

		// Last resort - write minimal error
		if minimalErr := writeMinimalError(w, factory, ctx); minimalErr != nil {
			factory.logger.Error(ctx, "failed to write error",
				"error", minimalErr.Error(),
			)
		}

		return WrapError(http.StatusInternalServerError, writeErr,
			"failed to write formatted error response")
	}

	return originalErr
}

// writeFormattedResponse writes headers and body to http.ResponseWriter.
func writeFormattedResponse(w http.ResponseWriter, resp DataResponse, formatted FormattedResponse) error {
	// Write custom headers from response
	for key, values := range resp.Header() {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Write Content-Type
	if resp.ContentType() != "" {
		w.Header().Set(HeaderContentType, resp.ContentType())
	} else if formatted.ContentType != "" {
		w.Header().Set(HeaderContentType, formatted.ContentType)
	}

	// Binary-specific headers
	if resp.IsBinary() {
		if resp.Filename() != "" {
			w.Header().Set(HeaderContentDisposition, `attachment; filename="`+resp.Filename()+`"`)
		}
		if formatted.StreamSize > 0 {
			w.Header().Set(HeaderContentLength, strconv.FormatInt(formatted.StreamSize, 10))
		}
	} else if len(formatted.Body) > 0 {
		// Set Content-Length for non-binary responses
		w.Header().Set(HeaderContentLength, strconv.Itoa(len(formatted.Body)))
	}

	// Write status code
	w.WriteHeader(resp.StatusCode())

	// Write body
	if formatted.Stream != nil {
		// Stream binary data
		if formatted.StreamSize > 0 {
			_, err := io.CopyN(w, formatted.Stream, formatted.StreamSize)
			return err
		}
		_, err := io.Copy(w, formatted.Stream)
		return err
	}

	// Write buffered body
	if len(formatted.Body) > 0 {
		_, err := w.Write(formatted.Body)

		return err
	}

	return nil
}

// writeMinimalError writes a minimal error response when all else fails.
func writeMinimalError(w http.ResponseWriter, factory *Factory, ctx context.Context) error {
	// Clear any headers that might have been set
	w.Header().Set(HeaderContentType, MimeTypePlainText.String())
	w.WriteHeader(http.StatusInternalServerError)

	// Write minimal error message
	message := []byte("Internal Server Error")
	n, err := w.Write(message)

	if err != nil {
		// Connection is probably dead at this point
		return WrapError(0, err, "failed to write minimal error response")
	}

	if n != len(message) {
		// Partial write
		return NewError(0, "partial write of minimal error: expected %d bytes, wrote %d")
	}

	return nil
}

// FormatterFromContext retrieves formatter from context.
func FormatterFromContext(ctx context.Context) Formatter {
	if f, ok := ctx.Value(formatterKey).(Formatter); ok {
		return f
	}
	return nil
}

// ContextWithFormatter returns a new context with the formatter.
func ContextWithFormatter(ctx context.Context, formatter Formatter) context.Context {
	return context.WithValue(ctx, formatterKey, formatter)
}

var defaultBinaryFormatterFunc func() Formatter

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
