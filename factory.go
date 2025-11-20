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
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/raoptimus/data-response.go/pkg/logger"
	"github.com/raoptimus/data-response.go/v2/response"
)

// Factory creates standardized HTTP responses.
type Factory struct {
	logger            Logger
	verbosity         bool
	formatter         response.Formatter
	debugMode         bool
	errorBuilder      ErrorBuilder
	validationBuilder ValidationErrorBuilder
}

// ErrorBuilder builds error response data structure.
type ErrorBuilder func(ctx context.Context, status int, message string, details any) any

// ValidationErrorBuilder builds validation error response data structure.
type ValidationErrorBuilder func(ctx context.Context, message string, attributeErrors map[string][]string) any

// Option configures a Factory.
type Option func(*Factory)

// WithLogger sets the logger.
func WithLogger(logger Logger) Option {
	return func(f *Factory) {
		f.logger = logger
	}
}

// WithVerbosity controls error detail visibility for security.
// false (production): hides error details from clients.
// true (dev/staging): shows full error details including stack traces.
func WithVerbosity(verbose bool) Option {
	return func(f *Factory) {
		f.verbosity = verbose
	}
}

// WithFormatter sets the default formatter for this factory.
func WithFormatter(formatter response.Formatter) Option {
	return func(f *Factory) {
		f.formatter = formatter
	}
}

// WithDebugMode enables detailed logging (use only in development).
func WithDebugMode(debug bool) Option {
	return func(f *Factory) {
		f.debugMode = debug
	}
}

// WithErrorBuilder sets custom error builder.
func WithErrorBuilder(builder ErrorBuilder) Option {
	return func(f *Factory) {
		f.errorBuilder = builder
	}
}

// WithValidationErrorBuilder sets custom validation error builder.
func WithValidationErrorBuilder(builder ValidationErrorBuilder) Option {
	return func(f *Factory) {
		f.validationBuilder = builder
	}
}

// New creates a new Factory with options.
func New(opts ...Option) *Factory {
	f := &Factory{
		logger:            logger.NoOpLogger{},
		verbosity:         false,
		debugMode:         false,
		errorBuilder:      defaultErrorBuilder,
		validationBuilder: defaultValidationErrorBuilder,
	}

	for _, opt := range opts {
		opt(f)
	}

	if f.formatter == nil {
		f.formatter = defaultFormatter()
	}

	return f
}

//nolint:ireturn,nolintlint // its ok
func (f *Factory) Logger() Logger {
	return f.logger
}

// Success creates a 200 OK response.
func (f *Factory) Success(ctx context.Context, data any) *response.DataResponse {
	if f.debugMode {
		f.logger.Debug(ctx, "success response")
	}

	return f.createDataResponse(http.StatusOK, data)
}

// Created creates a 201 Created response.
func (f *Factory) Created(ctx context.Context, data any, location string) *response.DataResponse {
	resp := f.createDataResponse(http.StatusCreated, data)

	if location != "" {
		resp = resp.WithHeader("Location", location)
	}

	if f.debugMode {
		f.logger.Debug(ctx, "created response", "location", location)
	}

	return resp
}

// Accepted creates a 202 Accepted response.
func (f *Factory) Accepted(ctx context.Context, data any) *response.DataResponse {
	if f.debugMode {
		f.logger.Debug(ctx, "accepted response")
	}

	return f.createDataResponse(http.StatusAccepted, data)
}

// NoContent creates a 204 No Content response.
func (f *Factory) NoContent(ctx context.Context) *response.DataResponse {
	if f.debugMode {
		f.logger.Debug(ctx, "no content response")
	}

	return f.createDataResponse(http.StatusNoContent, nil)
}

// Error creates an error response with custom data builder.
func (f *Factory) Error(ctx context.Context, status int, message string) *response.DataResponse {
	if message == "" {
		message = http.StatusText(status)
	}

	if f.debugMode {
		f.logger.Debug(ctx, "error response", "status", status, "message", message)
	}

	data := f.errorBuilder(ctx, status, message, nil)

	return f.createDataResponse(http.StatusInternalServerError, data)
}

// InternalError creates a 500 Internal Server Error response.
func (f *Factory) InternalError(ctx context.Context, err error) *response.DataResponse {
	f.logger.Error(ctx, "internal server error", "error", err.Error())

	message := "Internal server error"
	var details any

	if f.verbosity {
		errData := map[string]string{
			"error": err.Error(),
		}

		var e *response.Error
		if errors.As(err, &e) {
			if st := e.StackTrace(); st != "" {
				errData["stack_trace"] = st
			}
		}

		details = errData
	}

	data := f.errorBuilder(ctx, http.StatusInternalServerError, message, details)

	return f.createDataResponse(http.StatusInternalServerError, data)
}

// BadRequest creates a 400 Bad Request response.
func (f *Factory) BadRequest(ctx context.Context, message string) *response.DataResponse {
	return f.Error(ctx, http.StatusBadRequest, message)
}

// Unauthorized creates a 401 Unauthorized response.
func (f *Factory) Unauthorized(ctx context.Context, message string) *response.DataResponse {
	return f.Error(ctx, http.StatusUnauthorized, message)
}

// ServiceUnavailable creates a 503 Service Unavailable response
func (f *Factory) ServiceUnavailable(ctx context.Context, message string) *response.DataResponse {
	return f.Error(ctx, http.StatusServiceUnavailable, message)
}

// Forbidden creates a 403 Forbidden response.
func (f *Factory) Forbidden(ctx context.Context, message string) *response.DataResponse {
	return f.Error(ctx, http.StatusForbidden, message)
}

// NotFound creates a 404 Not Found response.
func (f *Factory) NotFound(ctx context.Context, message string) *response.DataResponse {
	return f.Error(ctx, http.StatusNotFound, message)
}

// Conflict creates a 409 Conflict response.
func (f *Factory) Conflict(ctx context.Context, message string) *response.DataResponse {
	return f.Error(ctx, http.StatusConflict, message)
}

// ValidationError creates a 422 Unprocessable Entity response.
func (f *Factory) ValidationError(ctx context.Context, message string, attributeErrors map[string][]string) *response.DataResponse {
	if f.debugMode {
		f.logger.Info(ctx, "validation error", "errors_count", len(attributeErrors))
	}

	if message == "" {
		message = "Validation failed"
	}

	data := f.validationBuilder(ctx, message, attributeErrors)

	return f.createDataResponse(http.StatusUnprocessableEntity, data)
}

// Binary creates a binary file response from io.Reader.
func (f *Factory) Binary(ctx context.Context, reader io.ReadCloser, filename string, size int64) *response.DataResponse {
	if f.debugMode {
		f.logger.Debug(ctx, "binary response", "filename", filename, "size", size)
	}

	// Detect Content-Type from filename
	ext := filepath.Ext(filename)
	contentType := response.MimeTypeFromExtension(ext).String()

	resp := f.createDataResponse(http.StatusOK, nil).
		WithFormatted(response.FormattedResponse{
			Stream:     reader,
			StreamSize: size,
		}).
		WithFile(reader, path.Base(filename)).
		WithContentType(contentType)

	return resp
}

// File creates a response from a file on disk.
func (f *Factory) File(ctx context.Context, filename string) *response.DataResponse {
	file, err := os.Open(filename)
	if err != nil {
		return f.InternalError(ctx, response.WrapError(http.StatusInternalServerError, err, "failed to open file"))
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()

		return f.InternalError(ctx, response.WrapError(http.StatusInternalServerError, err, "failed to stat file"))
	}

	if f.debugMode {
		f.logger.Debug(ctx, "file response", "path", filename, "size", stat.Size())
	}

	resp := f.Binary(ctx, file, stat.Name(), stat.Size()).
		WithFile(file, path.Base(filename))

	return resp
}

// Formatter returns the current default formatter for this factory.
//
//nolint:ireturn,nolintlint // its ok
func (f *Factory) Formatter() response.Formatter {
	return f.formatter
}

// Clone creates a copy of the factory with different options.
func (f *Factory) Clone(opts ...Option) *Factory {
	clone := *f
	for _, opt := range opts {
		opt(&clone)
	}

	return &clone
}

func (f *Factory) createDataResponse(statusCode int, data any) *response.DataResponse {
	return response.NewDataResponse(statusCode, data).WithFormatter(f.formatter)
}

// defaultErrorBuilder creates simple error structure.
func defaultErrorBuilder(_ context.Context, status int, message string, details any) any {
	return Template{
		Code:    response.CodeFromStatus(status),
		Status:  strconv.Itoa(status),
		Title:   message,
		Details: details,
	}
}

// defaultValidationErrorBuilder creates simple validation error structure.
func defaultValidationErrorBuilder(_ context.Context, message string, attributeErrors map[string][]string) any {
	errorsData := make(TemplateErrors, 0, len(attributeErrors))
	for k, v := range attributeErrors {
		for _, m := range v {
			errorsData = append(errorsData, TemplateError{
				Pointer: k,
				Detail:  m,
			})
		}
	}

	return Template{
		Code:   response.CodeFromStatus(http.StatusUnprocessableEntity),
		Status: strconv.Itoa(http.StatusUnprocessableEntity),
		Title:  message,
		Errors: errorsData,
	}
}

// defaultFormatter returns a minimal no-op formatter as fallback.
//
//nolint:ireturn,nolintlint // its ok
func defaultFormatter() response.Formatter {
	return &noopFormatter{}
}

// noopFormatter is a minimal formatter used as fallback when no formatter is configured.
type noopFormatter struct{}

// Format writes only the status code.
func (noopFormatter) Format(_ *response.DataResponse) (response.FormattedResponse, error) {
	return response.FormattedResponse{}, nil
}

// ContentType returns text/plain.
func (noopFormatter) ContentType() string {
	return response.ContentTypePlain
}

// CanFormatBinary returns false.
func (noopFormatter) CanFormatBinary() bool {
	return false
}
