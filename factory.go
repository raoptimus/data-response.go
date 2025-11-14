package dataresponse

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/raoptimus/data-response.go/pkg/logger"
)

type (
	TemplateError struct {
		Pointer string `json:"pointer,omitempty"` // Путь до свойства с проблемой
		NodeID  string `json:"nodeId,omitempty"`  // ID узла(uuid) в котором возникла ошибка
		PortID  string `json:"portId,omitempty"`  // ID порта узла(uuid) в котором возникла ошибка
		Detail  string `json:"detail"`            // Человеко-читаемое описание ошибки
	}
	TemplateErrors []TemplateError
	Template       struct {
		Code    HTTPCode       `json:"code,omitempty"`
		Status  string         `json:"status,omitempty"`
		Title   string         `json:"title,omitempty"`
		Details any            `json:"details,omitempty"`
		Errors  TemplateErrors `json:"errors,omitempty"`
	}
)

// Factory creates standardized HTTP responses.
type Factory struct {
	logger            logger.Logger
	verbosity         bool
	formatter         Formatter
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
func WithLogger(logger logger.Logger) Option {
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
func WithFormatter(formatter Formatter) Option {
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

// Success creates a 200 OK response.
func (f *Factory) Success(ctx context.Context, data any) DataResponse {
	if f.debugMode && f.logger != nil {
		f.logger.Debug(ctx, "success response")
	}

	return DataResponse{
		statusCode: http.StatusOK,
		data:       data,
		header:     make(http.Header),
	}
}

// Created creates a 201 Created response.
func (f *Factory) Created(ctx context.Context, data any, location string) DataResponse {
	resp := DataResponse{
		statusCode: http.StatusCreated,
		data:       data,
		header:     make(http.Header),
	}

	if location != "" {
		resp = resp.WithHeader("Location", location)
	}

	if f.debugMode && f.logger != nil {
		f.logger.Debug(ctx, "created response", "location", location)
	}

	return resp
}

// Accepted creates a 202 Accepted response.
func (f *Factory) Accepted(ctx context.Context, data any) DataResponse {
	if f.debugMode && f.logger != nil {
		f.logger.Debug(ctx, "accepted response")
	}

	return DataResponse{
		statusCode: http.StatusAccepted,
		data:       data,
		header:     make(http.Header),
	}
}

// NoContent creates a 204 No Content response.
func (f *Factory) NoContent(ctx context.Context) DataResponse {
	if f.debugMode && f.logger != nil {
		f.logger.Debug(ctx, "no content response")
	}

	return DataResponse{
		statusCode: http.StatusNoContent,
		header:     make(http.Header),
	}
}

// Error creates an error response with custom data builder.
func (f *Factory) Error(ctx context.Context, status int, message string) DataResponse {
	if f.logger != nil {
		f.logger.Info(ctx, "error response", "status", status, "message", message)
	}

	data := f.errorBuilder(ctx, status, message, nil)

	return DataResponse{
		statusCode: status,
		data:       data,
		header:     make(http.Header),
	}
}

// InternalError creates a 500 Internal Server Error response.
func (f *Factory) InternalError(ctx context.Context, err error) DataResponse {
	if f.logger != nil {
		f.logger.Error(ctx, "internal server error", "error", err.Error())
	}

	message := "Internal server error"
	var details any

	if f.verbosity && err != nil {
		errData := map[string]string{
			"error": err.Error(),
		}

		var e *Error
		if errors.As(err, &e) {
			if st := e.StackTrace(); st != "" {
				errData["stack_trace"] = st
			}
		}

		details = errData
	}

	data := f.errorBuilder(ctx, http.StatusInternalServerError, message, details)

	return DataResponse{
		statusCode: http.StatusInternalServerError,
		data:       data,
		header:     make(http.Header),
	}
}

// BadRequest creates a 400 Bad Request response.
func (f *Factory) BadRequest(ctx context.Context, message string) DataResponse {
	if message == "" {
		message = "Bad request"
	}
	return f.Error(ctx, http.StatusBadRequest, message)
}

// Unauthorized creates a 401 Unauthorized response.
func (f *Factory) Unauthorized(ctx context.Context, message string) DataResponse {
	if message == "" {
		message = "Unauthorized"
	}

	return f.Error(ctx, http.StatusUnauthorized, message)
}

// Forbidden creates a 403 Forbidden response.
func (f *Factory) Forbidden(ctx context.Context, message string) DataResponse {
	if message == "" {
		message = "Forbidden"
	}
	return f.Error(ctx, http.StatusForbidden, message)
}

// NotFound creates a 404 Not Found response.
func (f *Factory) NotFound(ctx context.Context, message string) DataResponse {
	if message == "" {
		message = "Not found"
	}
	return f.Error(ctx, http.StatusNotFound, message)
}

// Conflict creates a 409 Conflict response.
func (f *Factory) Conflict(ctx context.Context, message string) DataResponse {
	if message == "" {
		message = "Conflict"
	}
	return f.Error(ctx, http.StatusConflict, message)
}

// ValidationError creates a 422 Unprocessable Entity response.
func (f *Factory) ValidationError(ctx context.Context, message string, attributeErrors map[string][]string) DataResponse {
	if f.logger != nil {
		f.logger.Info(ctx, "validation error", "errors_count", len(attributeErrors))
	}

	if message == "" {
		message = "Validation failed"
	}

	data := f.validationBuilder(ctx, message, attributeErrors)

	return DataResponse{
		statusCode: http.StatusUnprocessableEntity,
		data:       data,
		header:     make(http.Header),
	}
}

// Binary creates a binary file response from io.Reader.
func (f *Factory) Binary(ctx context.Context, reader io.Reader, filename string, size int64) DataResponse {
	if f.debugMode && f.logger != nil {
		f.logger.Debug(ctx, "binary response", "filename", filename, "size", size)
	}

	return DataResponse{
		statusCode: http.StatusOK,
		binary:     reader,
		filename:   filename,
		size:       size,
		header:     make(http.Header),
	}
}

// File creates a response from a file on disk.
func (f *Factory) File(ctx context.Context, filepath string) DataResponse {
	file, err := os.Open(filepath)
	if err != nil {
		return f.InternalError(ctx, WrapError(http.StatusInternalServerError, err, "failed to open file"))
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return f.InternalError(ctx, WrapError(http.StatusInternalServerError, err, "failed to stat file"))
	}

	if f.debugMode && f.logger != nil {
		f.logger.Debug(ctx, "file response", "path", filepath, "size", stat.Size())
	}

	return DataResponse{
		statusCode: http.StatusOK,
		binary:     file,
		filename:   stat.Name(),
		size:       stat.Size(),
		header:     make(http.Header),
	}
}

// Formatter returns the current default formatter for this factory.
func (f *Factory) Formatter() Formatter {
	return f.formatter
}

// Clone creates a copy of the factory with different options.
func (f *Factory) Clone(opts ...Option) *Factory {
	clone := &Factory{
		logger:            f.logger,
		verbosity:         f.verbosity,
		formatter:         f.formatter,
		debugMode:         f.debugMode,
		errorBuilder:      f.errorBuilder,
		validationBuilder: f.validationBuilder,
	}

	for _, opt := range opts {
		opt(clone)
	}

	return clone
}

// defaultErrorBuilder creates simple error structure.
func defaultErrorBuilder(ctx context.Context, status int, message string, details any) any {
	return Template{
		Code:   CodeFromStatus(status),
		Status: strconv.Itoa(status),
		Title:   message,
		Details: details,
	}
}

// defaultValidationErrorBuilder creates simple validation error structure.
func defaultValidationErrorBuilder(ctx context.Context, message string, attributeErrors map[string][]string) any {
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
		Code:   CodeFromStatus(http.StatusUnprocessableEntity),
		Status: strconv.Itoa(http.StatusUnprocessableEntity),
		Title:  message,
		Errors: errorsData,
	}
}

// defaultFormatter returns a minimal no-op formatter as fallback.
func defaultFormatter() Formatter {
	return &noopFormatter{}
}

// noopFormatter is a minimal formatter used as fallback when no formatter is configured.
type noopFormatter struct{}

// Format writes only the status code.
func (noopFormatter) Format(resp DataResponse) (FormattedResponse, error) {
	return FormattedResponse{}, nil
}

// ContentType returns text/plain.
func (noopFormatter) ContentType() string {
	return "text/plain"
}

// CanFormatBinary returns false.
func (noopFormatter) CanFormatBinary() bool {
	return false
}
