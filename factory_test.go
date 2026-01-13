/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package dataresponse_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	dataresponse "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/mockdataresponse"
	"github.com/raoptimus/data-response.go/v2/response"
	"github.com/raoptimus/data-response.go/v2/response/mockresponse"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Factory Creation with Options
// =============================================================================

func TestNew_WithDefaultOptions_Successfully(t *testing.T) {
	t.Run("creates factory with default values", func(t *testing.T) {
		factory := dataresponse.New()

		require.NotNil(t, factory)
		require.NotNil(t, factory.Logger())
		require.NotNil(t, factory.Formatter())
	})
}

func TestNew_WithLogger_Successfully(t *testing.T) {
	t.Run("sets custom logger", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger))

		require.Equal(t, mockLogger, factory.Logger())
	})
}

func TestNew_WithVerbosity_Successfully(t *testing.T) {
	tests := []struct {
		name      string
		verbosity bool
	}{
		{
			name:      "verbosity enabled",
			verbosity: true,
		},
		{
			name:      "verbosity disabled",
			verbosity: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New(dataresponse.WithVerbosity(tt.verbosity))

			// Verify indirectly through InternalError behavior
			require.NotNil(t, factory)
		})
	}
}

func TestNew_WithFormatter_Successfully(t *testing.T) {
	t.Run("sets custom formatter", func(t *testing.T) {
		mockFormatter := mockresponse.NewFormatter(t)
		mockFormatter.EXPECT().ContentType().Return("application/json").Maybe()

		factory := dataresponse.New(dataresponse.WithFormatter(mockFormatter))

		require.Equal(t, mockFormatter, factory.Formatter())
	})
}

func TestNew_WithDebugMode_Successfully(t *testing.T) {
	tests := []struct {
		name      string
		debugMode bool
	}{
		{
			name:      "debug mode enabled",
			debugMode: true,
		},
		{
			name:      "debug mode disabled",
			debugMode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New(dataresponse.WithDebugMode(tt.debugMode))

			require.NotNil(t, factory)
		})
	}
}

func TestNew_WithErrorBuilder_Successfully(t *testing.T) {
	t.Run("sets custom error builder", func(t *testing.T) {
		customBuilder := func(_ context.Context, status int, message string, details any) any {
			return map[string]any{
				"custom_status":  status,
				"custom_message": message,
			}
		}

		factory := dataresponse.New(dataresponse.WithErrorBuilder(customBuilder))
		ctx := context.Background()

		resp := factory.BadRequest(ctx, "test message")

		require.NotNil(t, resp)
		resultMap, ok := resp.Data().(map[string]any)
		require.True(t, ok)
		require.Equal(t, http.StatusBadRequest, resultMap["custom_status"])
		require.Equal(t, "test message", resultMap["custom_message"])
	})
}

func TestNew_WithValidationErrorBuilder_Successfully(t *testing.T) {
	t.Run("sets custom validation error builder", func(t *testing.T) {
		customBuilder := func(_ context.Context, message string, attributeErrors map[string][]string) any {
			return map[string]any{
				"custom_message": message,
				"custom_errors":  attributeErrors,
			}
		}

		factory := dataresponse.New(dataresponse.WithValidationErrorBuilder(customBuilder))
		ctx := context.Background()
		attrErrors := map[string][]string{"field": {"error1"}}

		resp := factory.ValidationError(ctx, "validation failed", attrErrors)

		require.NotNil(t, resp)
		resultMap, ok := resp.Data().(map[string]any)
		require.True(t, ok)
		require.Equal(t, "validation failed", resultMap["custom_message"])
	})
}

func TestNew_WithMultipleOptions_Successfully(t *testing.T) {
	t.Run("applies all options", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		mockFormatter := mockresponse.NewFormatter(t)
		mockFormatter.EXPECT().ContentType().Return("application/json").Maybe()

		factory := dataresponse.New(
			dataresponse.WithLogger(mockLogger),
			dataresponse.WithVerbosity(true),
			dataresponse.WithFormatter(mockFormatter),
			dataresponse.WithDebugMode(true),
		)

		require.Equal(t, mockLogger, factory.Logger())
		require.Equal(t, mockFormatter, factory.Formatter())
	})
}

// =============================================================================
// Test Factory.Logger()
// =============================================================================

func TestFactory_Logger_ReturnsLogger_Successfully(t *testing.T) {
	t.Run("returns the configured logger", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		factory := dataresponse.New(dataresponse.WithLogger(mockLogger))

		result := factory.Logger()

		require.Equal(t, mockLogger, result)
	})

	t.Run("returns NoOpLogger when no logger configured", func(t *testing.T) {
		factory := dataresponse.New()

		result := factory.Logger()

		_, ok := result.(dataresponse.NoOpLogger)
		require.True(t, ok)
	})
}

// =============================================================================
// Test Factory.Formatter()
// =============================================================================

func TestFactory_Formatter_ReturnsFormatter_Successfully(t *testing.T) {
	t.Run("returns custom formatter", func(t *testing.T) {
		mockFormatter := mockresponse.NewFormatter(t)
		mockFormatter.EXPECT().ContentType().Return("application/json").Maybe()
		factory := dataresponse.New(dataresponse.WithFormatter(mockFormatter))

		result := factory.Formatter()

		require.Equal(t, mockFormatter, result)
	})

	t.Run("returns default formatter when none configured", func(t *testing.T) {
		factory := dataresponse.New()

		result := factory.Formatter()

		require.NotNil(t, result)
		require.Equal(t, response.ContentTypePlain, result.ContentType())
	})
}

// =============================================================================
// Test Factory.Clone()
// =============================================================================

func TestFactory_Clone_CreatesIndependentCopy_Successfully(t *testing.T) {
	t.Run("creates copy with same values", func(t *testing.T) {
		original := dataresponse.New(
			dataresponse.WithVerbosity(true),
			dataresponse.WithDebugMode(true),
		)

		clone := original.Clone()

		require.NotNil(t, clone)
		require.NotSame(t, original, clone)
	})

	t.Run("applies new options to clone", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		original := dataresponse.New(dataresponse.WithVerbosity(true))

		clone := original.Clone(dataresponse.WithLogger(mockLogger))

		require.NotEqual(t, original.Logger(), clone.Logger())
		require.Equal(t, mockLogger, clone.Logger())
	})

	t.Run("changes to clone do not affect original", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		original := dataresponse.New()

		clone := original.Clone(dataresponse.WithLogger(mockLogger))

		_, originalIsNoOp := original.Logger().(dataresponse.NoOpLogger)
		require.True(t, originalIsNoOp)
		require.Equal(t, mockLogger, clone.Logger())
	})
}

// =============================================================================
// Test Success Response Methods
// =============================================================================

func TestFactory_Success_ReturnsOKResponse_Successfully(t *testing.T) {
	tests := []struct {
		name string
		data any
	}{
		{
			name: "with map data",
			data: map[string]string{"key": "value"},
		},
		{
			name: "with struct data",
			data: struct{ Name string }{Name: "test"},
		},
		{
			name: "with nil data",
			data: nil,
		},
		{
			name: "with slice data",
			data: []int{1, 2, 3},
		},
		{
			name: "with string data",
			data: "test string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()

			resp := factory.Success(ctx, tt.data)

			require.NotNil(t, resp)
			require.Equal(t, http.StatusOK, resp.StatusCode())
			require.Equal(t, tt.data, resp.Data())
		})
	}
}

func TestFactory_Success_WithDebugMode_LogsMessage_Successfully(t *testing.T) {
	t.Run("logs debug message when debug mode enabled", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		mockLogger.EXPECT().Debug(mock.Anything, "success response").Return()

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger), dataresponse.WithDebugMode(true))
		ctx := context.Background()

		resp := factory.Success(ctx, nil)

		require.NotNil(t, resp)
		require.Equal(t, http.StatusOK, resp.StatusCode())
	})
}

func TestFactory_Created_ReturnsCreatedResponse_Successfully(t *testing.T) {
	tests := []struct {
		name           string
		data           any
		location       string
		expectLocation bool
	}{
		{
			name:           "with location",
			data:           map[string]int{"id": 123},
			location:       "/resources/123",
			expectLocation: true,
		},
		{
			name:           "without location",
			data:           map[string]int{"id": 456},
			location:       "",
			expectLocation: false,
		},
		{
			name:           "with nil data and location",
			data:           nil,
			location:       "/items/789",
			expectLocation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()

			resp := factory.Created(ctx, tt.data, tt.location)

			require.NotNil(t, resp)
			require.Equal(t, http.StatusCreated, resp.StatusCode())
			require.Equal(t, tt.data, resp.Data())

			if tt.expectLocation {
				require.Equal(t, tt.location, resp.HeaderLine("Location"))
			} else {
				require.Empty(t, resp.HeaderLine("Location"))
			}
		})
	}
}

func TestFactory_Created_WithDebugMode_LogsMessage_Successfully(t *testing.T) {
	t.Run("logs debug message with location", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		mockLogger.EXPECT().Debug(mock.Anything, "created response", "location", "/test/123").Return()

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger), dataresponse.WithDebugMode(true))
		ctx := context.Background()

		resp := factory.Created(ctx, nil, "/test/123")

		require.NotNil(t, resp)
	})
}

func TestFactory_Accepted_ReturnsAcceptedResponse_Successfully(t *testing.T) {
	tests := []struct {
		name string
		data any
	}{
		{
			name: "with data",
			data: map[string]string{"status": "processing"},
		},
		{
			name: "with nil data",
			data: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()

			resp := factory.Accepted(ctx, tt.data)

			require.NotNil(t, resp)
			require.Equal(t, http.StatusAccepted, resp.StatusCode())
			require.Equal(t, tt.data, resp.Data())
		})
	}
}

func TestFactory_Accepted_WithDebugMode_LogsMessage_Successfully(t *testing.T) {
	t.Run("logs debug message", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		mockLogger.EXPECT().Debug(mock.Anything, "accepted response").Return()

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger), dataresponse.WithDebugMode(true))
		ctx := context.Background()

		resp := factory.Accepted(ctx, nil)

		require.NotNil(t, resp)
	})
}

func TestFactory_NoContent_ReturnsNoContentResponse_Successfully(t *testing.T) {
	t.Run("returns 204 with nil data", func(t *testing.T) {
		factory := dataresponse.New()
		ctx := context.Background()

		resp := factory.NoContent(ctx)

		require.NotNil(t, resp)
		require.Equal(t, http.StatusNoContent, resp.StatusCode())
		require.Nil(t, resp.Data())
	})
}

func TestFactory_NoContent_WithDebugMode_LogsMessage_Successfully(t *testing.T) {
	t.Run("logs debug message", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		mockLogger.EXPECT().Debug(mock.Anything, "no content response").Return()

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger), dataresponse.WithDebugMode(true))
		ctx := context.Background()

		resp := factory.NoContent(ctx)

		require.NotNil(t, resp)
	})
}

func TestFactory_Found_ReturnsFoundResponse_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		location string
	}{
		{
			name:     "with absolute path",
			location: "/new/location",
		},
		{
			name:     "with full URL",
			location: "https://example.com/redirect",
		},
		{
			name:     "with empty location",
			location: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()

			resp := factory.Found(ctx, tt.location)

			require.NotNil(t, resp)
			require.Equal(t, http.StatusFound, resp.StatusCode())
			require.Equal(t, tt.location, resp.HeaderLine(response.HeaderLocation))
			require.Nil(t, resp.Data())
		})
	}
}

func TestFactory_Found_WithDebugMode_LogsMessage_Successfully(t *testing.T) {
	t.Run("logs debug message", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		mockLogger.EXPECT().Debug(mock.Anything, "found response").Return()

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger), dataresponse.WithDebugMode(true))
		ctx := context.Background()

		resp := factory.Found(ctx, "/redirect")

		require.NotNil(t, resp)
	})
}

func TestFactory_NotModified_ReturnsNotModifiedResponse_Successfully(t *testing.T) {
	tests := []struct {
		name string
		eTag string
	}{
		{
			name: "with standard etag",
			eTag: `"abc123"`,
		},
		{
			name: "with weak etag",
			eTag: `W/"abc123"`,
		},
		{
			name: "with empty etag",
			eTag: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()

			resp := factory.NotModified(ctx, tt.eTag)

			require.NotNil(t, resp)
			require.Equal(t, http.StatusNotModified, resp.StatusCode())
			require.Equal(t, tt.eTag, resp.HeaderLine(response.HeaderETag))
			require.Nil(t, resp.Data())
		})
	}
}

func TestFactory_NotModified_WithDebugMode_LogsMessage_Successfully(t *testing.T) {
	t.Run("logs debug message", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		mockLogger.EXPECT().Debug(mock.Anything, "not modified response").Return()

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger), dataresponse.WithDebugMode(true))
		ctx := context.Background()

		resp := factory.NotModified(ctx, `"etag123"`)

		require.NotNil(t, resp)
	})
}

// =============================================================================
// Test Error Response Methods
// =============================================================================

func TestFactory_Error_ReturnsErrorResponse_Successfully(t *testing.T) {
	tests := []struct {
		name            string
		status          int
		message         string
		expectedMessage string
		expectedCode    response.HTTPCode
	}{
		{
			name:            "bad request with message",
			status:          http.StatusBadRequest,
			message:         "invalid input",
			expectedMessage: "invalid input",
			expectedCode:    response.HTTPCodeBadRequest,
		},
		{
			name:            "not found with empty message uses status text",
			status:          http.StatusNotFound,
			message:         "",
			expectedMessage: "Not Found",
			expectedCode:    response.HTTPCodeNotFound,
		},
		{
			name:            "forbidden with message",
			status:          http.StatusForbidden,
			message:         "access denied",
			expectedMessage: "access denied",
			expectedCode:    response.HTTPCodeForbidden,
		},
		{
			name:            "internal server error with empty message",
			status:          http.StatusInternalServerError,
			message:         "",
			expectedMessage: "Internal Server Error",
			expectedCode:    response.HTTPCodeInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()

			resp := factory.Error(ctx, tt.status, tt.message)

			require.NotNil(t, resp)
			require.Equal(t, tt.status, resp.StatusCode())

			data, ok := resp.Data().(dataresponse.Template)
			require.True(t, ok)
			require.Equal(t, tt.expectedMessage, data.Title)
			require.Equal(t, tt.expectedCode, data.Code)
		})
	}
}

func TestFactory_Error_WithDebugMode_LogsMessage_Successfully(t *testing.T) {
	t.Run("logs debug message with status and message", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		mockLogger.EXPECT().Debug(
			mock.Anything,
			"error response",
			"status", http.StatusBadRequest,
			"message", "test error",
		).Return()

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger), dataresponse.WithDebugMode(true))
		ctx := context.Background()

		resp := factory.Error(ctx, http.StatusBadRequest, "test error")

		require.NotNil(t, resp)
	})
}

func TestFactory_InternalError_ReturnsInternalErrorResponse_Successfully(t *testing.T) {
	t.Run("without verbosity hides error details", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		mockLogger.EXPECT().Error(mock.Anything, "internal server error", "error", "test error").Return()

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger), dataresponse.WithVerbosity(false))
		ctx := context.Background()
		err := errors.New("test error")

		resp := factory.InternalError(ctx, err)

		require.NotNil(t, resp)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode())

		data, ok := resp.Data().(dataresponse.Template)
		require.True(t, ok)
		require.Equal(t, "Internal server error", data.Title)
		require.Nil(t, data.Details)
	})

	t.Run("with verbosity shows error details", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		mockLogger.EXPECT().Error(mock.Anything, "internal server error", "error", "test error").Return()

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger), dataresponse.WithVerbosity(true))
		ctx := context.Background()
		err := errors.New("test error")

		resp := factory.InternalError(ctx, err)

		require.NotNil(t, resp)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode())

		data, ok := resp.Data().(dataresponse.Template)
		require.True(t, ok)
		require.Equal(t, "Internal server error", data.Title)

		details, ok := data.Details.(*dataresponse.InternalError)
		require.True(t, ok)
		require.Equal(t, "test error", details.Error)
	})

	t.Run("with wrapped error shows stack trace when verbose", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		mockLogger.EXPECT().Error(mock.Anything, "internal server error", "error", "wrapped error").Return()

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger), dataresponse.WithVerbosity(true))
		ctx := context.Background()
		err := response.WrapError(http.StatusInternalServerError, errors.New("original"), "wrapped error")

		resp := factory.InternalError(ctx, err)

		require.NotNil(t, resp)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode())

		data, ok := resp.Data().(dataresponse.Template)
		require.True(t, ok)

		details, ok := data.Details.(*dataresponse.InternalError)
		require.True(t, ok)
		require.Equal(t, "wrapped error", details.Error)
		require.NotEmpty(t, details.StackTrace)
	})
}

func TestFactory_BadRequest_ReturnsBadRequestResponse_Successfully(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		expectedMessage string
	}{
		{
			name:            "with message",
			message:         "invalid request body",
			expectedMessage: "invalid request body",
		},
		{
			name:            "with empty message",
			message:         "",
			expectedMessage: "Bad Request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()

			resp := factory.BadRequest(ctx, tt.message)

			require.NotNil(t, resp)
			require.Equal(t, http.StatusBadRequest, resp.StatusCode())

			data, ok := resp.Data().(dataresponse.Template)
			require.True(t, ok)
			require.Equal(t, tt.expectedMessage, data.Title)
		})
	}
}

func TestFactory_Unauthorized_ReturnsUnauthorizedResponse_Successfully(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		expectedMessage string
	}{
		{
			name:            "with message",
			message:         "invalid credentials",
			expectedMessage: "invalid credentials",
		},
		{
			name:            "with empty message",
			message:         "",
			expectedMessage: "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()

			resp := factory.Unauthorized(ctx, tt.message)

			require.NotNil(t, resp)
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode())

			data, ok := resp.Data().(dataresponse.Template)
			require.True(t, ok)
			require.Equal(t, tt.expectedMessage, data.Title)
		})
	}
}

func TestFactory_ServiceUnavailable_ReturnsServiceUnavailableResponse_Successfully(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		expectedMessage string
	}{
		{
			name:            "with message",
			message:         "service temporarily unavailable",
			expectedMessage: "service temporarily unavailable",
		},
		{
			name:            "with empty message",
			message:         "",
			expectedMessage: "Service Unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()

			resp := factory.ServiceUnavailable(ctx, tt.message)

			require.NotNil(t, resp)
			require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode())

			data, ok := resp.Data().(dataresponse.Template)
			require.True(t, ok)
			require.Equal(t, tt.expectedMessage, data.Title)
		})
	}
}

func TestFactory_Forbidden_ReturnsForbiddenResponse_Successfully(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		expectedMessage string
	}{
		{
			name:            "with message",
			message:         "access forbidden",
			expectedMessage: "access forbidden",
		},
		{
			name:            "with empty message",
			message:         "",
			expectedMessage: "Forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()

			resp := factory.Forbidden(ctx, tt.message)

			require.NotNil(t, resp)
			require.Equal(t, http.StatusForbidden, resp.StatusCode())

			data, ok := resp.Data().(dataresponse.Template)
			require.True(t, ok)
			require.Equal(t, tt.expectedMessage, data.Title)
		})
	}
}

func TestFactory_NotFound_ReturnsNotFoundResponse_Successfully(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		expectedMessage string
	}{
		{
			name:            "with message",
			message:         "resource not found",
			expectedMessage: "resource not found",
		},
		{
			name:            "with empty message",
			message:         "",
			expectedMessage: "Not Found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()

			resp := factory.NotFound(ctx, tt.message)

			require.NotNil(t, resp)
			require.Equal(t, http.StatusNotFound, resp.StatusCode())

			data, ok := resp.Data().(dataresponse.Template)
			require.True(t, ok)
			require.Equal(t, tt.expectedMessage, data.Title)
		})
	}
}

func TestFactory_Conflict_ReturnsConflictResponse_Successfully(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		expectedMessage string
	}{
		{
			name:            "with message",
			message:         "resource already exists",
			expectedMessage: "resource already exists",
		},
		{
			name:            "with empty message",
			message:         "",
			expectedMessage: "Conflict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()

			resp := factory.Conflict(ctx, tt.message)

			require.NotNil(t, resp)
			require.Equal(t, http.StatusConflict, resp.StatusCode())

			data, ok := resp.Data().(dataresponse.Template)
			require.True(t, ok)
			require.Equal(t, tt.expectedMessage, data.Title)
		})
	}
}

func TestFactory_ValidationError_ReturnsValidationErrorResponse_Successfully(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		attributeErrors map[string][]string
		expectedMessage string
		expectedErrors  int
	}{
		{
			name:    "with message and single attribute error",
			message: "validation failed",
			attributeErrors: map[string][]string{
				"email": {"email is required"},
			},
			expectedMessage: "validation failed",
			expectedErrors:  1,
		},
		{
			name:    "with empty message uses default",
			message: "",
			attributeErrors: map[string][]string{
				"name": {"name is required"},
			},
			expectedMessage: "Validation failed",
			expectedErrors:  1,
		},
		{
			name:    "with multiple attribute errors",
			message: "input validation error",
			attributeErrors: map[string][]string{
				"email":    {"email is required", "email is invalid"},
				"password": {"password is too short"},
			},
			expectedMessage: "input validation error",
			expectedErrors:  3,
		},
		{
			name:            "with empty attribute errors",
			message:         "validation failed",
			attributeErrors: map[string][]string{},
			expectedMessage: "validation failed",
			expectedErrors:  0,
		},
		{
			name:            "with nil attribute errors",
			message:         "validation failed",
			attributeErrors: nil,
			expectedMessage: "validation failed",
			expectedErrors:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()

			resp := factory.ValidationError(ctx, tt.message, tt.attributeErrors)

			require.NotNil(t, resp)
			require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode())

			data, ok := resp.Data().(dataresponse.Template)
			require.True(t, ok)
			require.Equal(t, tt.expectedMessage, data.Title)
			require.Equal(t, response.HTTPCodeUnprocessableEntity, data.Code)
			require.Len(t, data.Errors, tt.expectedErrors)
		})
	}
}

func TestFactory_ValidationError_WithDebugMode_LogsMessage_Successfully(t *testing.T) {
	t.Run("logs info message with error count", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		mockLogger.EXPECT().Info(mock.Anything, "validation error", "errors_count", 2).Return()

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger), dataresponse.WithDebugMode(true))
		ctx := context.Background()
		attrErrors := map[string][]string{
			"email": {"required"},
			"name":  {"too short"},
		}

		resp := factory.ValidationError(ctx, "validation failed", attrErrors)

		require.NotNil(t, resp)
	})
}

// =============================================================================
// Test Binary Response Methods
// =============================================================================

func TestFactory_Binary_ReturnsBinaryResponse_Successfully(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		size        int64
		content     string
	}{
		{
			name:        "with json content type",
			contentType: "application/json",
			size:        15,
			content:     `{"key":"value"}`,
		},
		{
			name:        "with octet stream",
			contentType: "application/octet-stream",
			size:        10,
			content:     "0123456789",
		},
		{
			name:        "with zero size",
			contentType: "text/plain",
			size:        0,
			content:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()
			reader := io.NopCloser(bytes.NewBufferString(tt.content))

			resp := factory.Binary(ctx, reader, tt.contentType, tt.size)

			require.NotNil(t, resp)
			require.Equal(t, http.StatusOK, resp.StatusCode())
			require.Equal(t, tt.contentType, resp.ContentType())
			require.True(t, resp.IsBinary())
		})
	}
}

func TestFactory_Binary_WithDebugMode_LogsMessage_Successfully(t *testing.T) {
	t.Run("logs debug message with content type and size", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		mockLogger.EXPECT().Debug(
			mock.Anything,
			"binary response",
			"contentType", "application/pdf",
			"size", int64(1024),
		).Return()

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger), dataresponse.WithDebugMode(true))
		ctx := context.Background()
		reader := io.NopCloser(bytes.NewBuffer(nil))

		resp := factory.Binary(ctx, reader, "application/pdf", 1024)

		require.NotNil(t, resp)
	})
}

func TestFactory_BinaryWithFilename_ReturnsBinaryResponseWithContentDisposition_Successfully(t *testing.T) {
	tests := []struct {
		name                       string
		filename                   string
		size                       int64
		expectedContentType        string
		expectedContentDisposition string
	}{
		{
			name:                       "pdf file",
			filename:                   "document.pdf",
			size:                       2048,
			expectedContentType:        "application/pdf",
			expectedContentDisposition: `attachment; filename="document.pdf"`,
		},
		{
			name:                       "json file",
			filename:                   "data.json",
			size:                       512,
			expectedContentType:        "application/json; charset=utf-8",
			expectedContentDisposition: `attachment; filename="data.json"`,
		},
		{
			name:                       "unknown extension uses octet-stream",
			filename:                   "file.xyz",
			size:                       100,
			expectedContentType:        "application/octet-stream",
			expectedContentDisposition: `attachment; filename="file.xyz"`,
		},
		{
			name:                       "no extension uses octet-stream",
			filename:                   "noextension",
			size:                       50,
			expectedContentType:        "application/octet-stream",
			expectedContentDisposition: `attachment; filename="noextension"`,
		},
		{
			name:                       "path with directory extracts basename",
			filename:                   "/path/to/image.png",
			size:                       4096,
			expectedContentType:        "image/png",
			expectedContentDisposition: `attachment; filename="image.png"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dataresponse.New()
			ctx := context.Background()
			reader := io.NopCloser(bytes.NewBuffer(nil))

			resp := factory.BinaryWithFilename(ctx, reader, tt.filename, tt.size)

			require.NotNil(t, resp)
			require.Equal(t, http.StatusOK, resp.StatusCode())
			require.Equal(t, tt.expectedContentType, resp.ContentType())
			require.Equal(t, tt.expectedContentDisposition, resp.HeaderLine(response.HeaderContentDisposition))
			require.True(t, resp.IsBinary())
		})
	}
}

// =============================================================================
// Test File Response Method
// =============================================================================

func TestFactory_File_ReturnsFileResponse_Successfully(t *testing.T) {
	t.Run("returns binary response for existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFilePath := filepath.Join(tempDir, "test.txt")
		testContent := "test file content"
		err := os.WriteFile(testFilePath, []byte(testContent), 0644)
		require.NoError(t, err)

		factory := dataresponse.New()
		ctx := context.Background()

		resp := factory.File(ctx, testFilePath)

		require.NotNil(t, resp)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.True(t, resp.IsBinary())
		require.Contains(t, resp.HeaderLine(response.HeaderContentDisposition), "test.txt")

		// Clean up
		err = resp.Close()
		require.NoError(t, err)
	})

	t.Run("logs debug message when file exists and debug mode enabled", func(t *testing.T) {
		tempDir := t.TempDir()
		testFilePath := filepath.Join(tempDir, "debug_test.txt")
		testContent := "debug test content"
		err := os.WriteFile(testFilePath, []byte(testContent), 0644)
		require.NoError(t, err)

		mockLogger := mockdataresponse.NewLogger(t)
		// File method calls Binary which logs "binary response", then logs "file response"
		mockLogger.EXPECT().Debug(
			mock.Anything,
			"binary response",
			"contentType", "text/plain; charset=utf-8",
			"size", int64(len(testContent)),
		).Return()
		mockLogger.EXPECT().Debug(
			mock.Anything,
			"file response",
			"path", testFilePath,
			"size", int64(len(testContent)),
		).Return()

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger), dataresponse.WithDebugMode(true))
		ctx := context.Background()

		resp := factory.File(ctx, testFilePath)

		require.NotNil(t, resp)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		err = resp.Close()
		require.NoError(t, err)
	})
}

func TestFactory_File_NonExistentFile_ReturnsInternalError_Failure(t *testing.T) {
	t.Run("returns internal error for non-existent file", func(t *testing.T) {
		mockLogger := mockdataresponse.NewLogger(t)
		mockLogger.EXPECT().Error(mock.Anything, "internal server error", "error", "failed to open file").Return()

		factory := dataresponse.New(dataresponse.WithLogger(mockLogger))
		ctx := context.Background()

		resp := factory.File(ctx, "/non/existent/path/file.txt")

		require.NotNil(t, resp)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode())
	})
}

// =============================================================================
// Test CreateDataResponse
// =============================================================================

func TestFactory_CreateDataResponse_ReturnsDataResponseWithFormatter_Successfully(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		data       any
	}{
		{
			name:       "with 200 status and data",
			statusCode: http.StatusOK,
			data:       map[string]string{"key": "value"},
		},
		{
			name:       "with 404 status and nil data",
			statusCode: http.StatusNotFound,
			data:       nil,
		},
		{
			name:       "with 201 status and struct data",
			statusCode: http.StatusCreated,
			data:       struct{ ID int }{ID: 123},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFormatter := mockresponse.NewFormatter(t)
			mockFormatter.EXPECT().ContentType().Return("application/json")

			factory := dataresponse.New(dataresponse.WithFormatter(mockFormatter))

			resp := factory.CreateDataResponse(tt.statusCode, tt.data)

			require.NotNil(t, resp)
			require.Equal(t, tt.statusCode, resp.StatusCode())
			require.Equal(t, tt.data, resp.Data())
			require.Equal(t, "application/json", resp.ContentType())
		})
	}
}

// =============================================================================
// Test Edge Cases and Boundary Conditions
// =============================================================================

func TestFactory_BoundaryConditions_Successfully(t *testing.T) {
	t.Run("handles very long error message", func(t *testing.T) {
		factory := dataresponse.New()
		ctx := context.Background()
		longMessage := string(make([]byte, 10000))
		for i := range longMessage {
			longMessage = longMessage[:i] + "a" + longMessage[i+1:]
		}

		resp := factory.BadRequest(ctx, longMessage)

		require.NotNil(t, resp)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode())
	})

	t.Run("handles unicode in error message", func(t *testing.T) {
		factory := dataresponse.New()
		ctx := context.Background()

		resp := factory.BadRequest(ctx, "Unicode message")

		require.NotNil(t, resp)
		data, ok := resp.Data().(dataresponse.Template)
		require.True(t, ok)
		require.Equal(t, "Unicode message", data.Title)
	})

	t.Run("handles large validation error map", func(t *testing.T) {
		factory := dataresponse.New()
		ctx := context.Background()
		largeErrors := make(map[string][]string)
		for i := 0; i < 100; i++ {
			key := "field" + string(rune('a'+i%26))
			largeErrors[key] = []string{"error1", "error2", "error3"}
		}

		resp := factory.ValidationError(ctx, "many errors", largeErrors)

		require.NotNil(t, resp)
		require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode())
	})
}

func TestFactory_ConcurrentAccess_Successfully(t *testing.T) {
	t.Run("factory is safe for concurrent reads", func(t *testing.T) {
		factory := dataresponse.New()
		ctx := context.Background()

		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				resp := factory.Success(ctx, map[string]int{"value": 1})
				require.NotNil(t, resp)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// =============================================================================
// Test defaultFormatter (noopFormatter)
// =============================================================================

func TestDefaultFormatter_ContentType_ReturnsTextPlain_Successfully(t *testing.T) {
	t.Run("returns text/plain content type", func(t *testing.T) {
		factory := dataresponse.New()

		formatter := factory.Formatter()

		require.Equal(t, response.ContentTypePlain, formatter.ContentType())
	})
}

func TestDefaultFormatter_CanFormatBinary_ReturnsFalse_Successfully(t *testing.T) {
	t.Run("returns false for binary formatting", func(t *testing.T) {
		factory := dataresponse.New()

		formatter := factory.Formatter()

		require.False(t, formatter.CanFormatBinary())
	})
}

func TestDefaultFormatter_Format_ReturnsEmptyResponse_Successfully(t *testing.T) {
	t.Run("returns empty formatted response", func(t *testing.T) {
		factory := dataresponse.New()
		formatter := factory.Formatter()
		ctx := context.Background()
		resp := factory.Success(ctx, nil)

		result, err := formatter.Format(resp)

		require.NoError(t, err)
		require.Nil(t, result.Stream)
		require.Zero(t, result.StreamSize)
	})
}
