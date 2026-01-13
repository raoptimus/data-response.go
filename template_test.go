/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package dataresponse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInternalError_String_ErrorWithStackTrace_Successfully(t *testing.T) {
	tests := []struct {
		name       string
		error      string
		stackTrace string
		expected   string
	}{
		{
			name:       "short error with single line stack trace",
			error:      "something went wrong",
			stackTrace: "\n\tat main.go:10",
			expected:   "Error:\nsomething went wrong\nStack Trace:\n\tat main.go:10\n\n",
		},
		{
			name:       "error with multiline stack trace",
			error:      "database connection failed",
			stackTrace: "\n\tat db/connection.go:42\n\tat main.go:15",
			expected:   "Error:\ndatabase connection failed\nStack Trace:\n\tat db/connection.go:42\n\tat main.go:15\n\n",
		},
		{
			name:       "empty error with stack trace",
			error:      "",
			stackTrace: "\n\tat handler.go:99",
			expected:   "Error:\n\nStack Trace:\n\tat handler.go:99\n\n",
		},
		{
			name:       "error with very long stack trace",
			error:      "panic: runtime error",
			stackTrace: "\n\tat level1.go:1\n\tat level2.go:2\n\tat level3.go:3\n\tat level4.go:4\n\tat level5.go:5",
			expected:   "Error:\npanic: runtime error\nStack Trace:\n\tat level1.go:1\n\tat level2.go:2\n\tat level3.go:3\n\tat level4.go:4\n\tat level5.go:5\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ie := &InternalError{
				Error:      tt.error,
				StackTrace: tt.stackTrace,
			}

			result := ie.String()

			require.Equal(t, tt.expected, result)
		})
	}
}

func TestInternalError_String_ErrorWithoutStackTrace_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		error    string
		expected string
	}{
		{
			name:     "simple error message",
			error:    "file not found",
			expected: "Error:\nfile not found",
		},
		{
			name:     "empty error message",
			error:    "",
			expected: "Error:\n",
		},
		{
			name:     "error message with special characters",
			error:    "invalid JSON: unexpected token '<' at position 0",
			expected: "Error:\ninvalid JSON: unexpected token '<' at position 0",
		},
		{
			name:     "multiline error message",
			error:    "validation failed:\n- field1 is required\n- field2 is invalid",
			expected: "Error:\nvalidation failed:\n- field1 is required\n- field2 is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ie := &InternalError{
				Error:      tt.error,
				StackTrace: "",
			}

			result := ie.String()

			require.Equal(t, tt.expected, result)
		})
	}
}

func TestInternalError_String_EmptyStackTrace_DoesNotAppendStackTraceSection(t *testing.T) {
	tests := []struct {
		name       string
		error      string
		stackTrace string
		expected   string
	}{
		{
			name:       "empty stack trace string",
			error:      "test error",
			stackTrace: "",
			expected:   "Error:\ntest error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ie := &InternalError{
				Error:      tt.error,
				StackTrace: tt.stackTrace,
			}

			result := ie.String()

			require.Equal(t, tt.expected, result)
			require.NotContains(t, result, "Stack Trace:")
		})
	}
}

func TestInternalError_String_BothFieldsEmpty_ReturnsMinimalOutput(t *testing.T) {
	ie := &InternalError{
		Error:      "",
		StackTrace: "",
	}

	result := ie.String()

	require.Equal(t, "Error:\n", result)
	require.NotContains(t, result, "Stack Trace:")
}

func TestInternalError_String_StackTraceWithOnlyWhitespace_IncludesStackTraceSection(t *testing.T) {
	tests := []struct {
		name       string
		stackTrace string
		expected   string
	}{
		{
			name:       "stack trace with single space",
			stackTrace: " ",
			expected:   "Error:\ntest error\nStack Trace: \n\n",
		},
		{
			name:       "stack trace with newline only",
			stackTrace: "\n",
			expected:   "Error:\ntest error\nStack Trace:\n\n\n",
		},
		{
			name:       "stack trace with tab only",
			stackTrace: "\t",
			expected:   "Error:\ntest error\nStack Trace:\t\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ie := &InternalError{
				Error:      "test error",
				StackTrace: tt.stackTrace,
			}

			result := ie.String()

			require.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplateError_FieldsAreAccessible(t *testing.T) {
	te := TemplateError{
		Pointer: "/data/attributes/email",
		NodeID:  "550e8400-e29b-41d4-a716-446655440000",
		PortID:  "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Detail:  "Email is required",
	}

	require.Equal(t, "/data/attributes/email", te.Pointer)
	require.Equal(t, "550e8400-e29b-41d4-a716-446655440000", te.NodeID)
	require.Equal(t, "6ba7b810-9dad-11d1-80b4-00c04fd430c8", te.PortID)
	require.Equal(t, "Email is required", te.Detail)
}

func TestTemplateError_EmptyFields(t *testing.T) {
	te := TemplateError{}

	require.Empty(t, te.Pointer)
	require.Empty(t, te.NodeID)
	require.Empty(t, te.PortID)
	require.Empty(t, te.Detail)
}

func TestTemplateErrors_SliceOperations(t *testing.T) {
	tests := []struct {
		name     string
		errors   TemplateErrors
		expected int
	}{
		{
			name:     "empty slice",
			errors:   TemplateErrors{},
			expected: 0,
		},
		{
			name: "single error",
			errors: TemplateErrors{
				{Detail: "error 1"},
			},
			expected: 1,
		},
		{
			name: "multiple errors",
			errors: TemplateErrors{
				{Detail: "error 1"},
				{Detail: "error 2"},
				{Detail: "error 3"},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Len(t, tt.errors, tt.expected)
		})
	}
}

func TestTemplate_FieldsAreAccessible(t *testing.T) {
	tests := []struct {
		name     string
		template Template
	}{
		{
			name: "full template with all fields",
			template: Template{
				Code:    "BAD_REQUEST",
				Status:  "error",
				Title:   "Validation Error",
				Details: map[string]string{"field": "value"},
				Errors: TemplateErrors{
					{Detail: "field is required"},
				},
			},
		},
		{
			name:     "empty template",
			template: Template{},
		},
		{
			name: "template with only code and status",
			template: Template{
				Code:   "OK",
				Status: "success",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := tt.template

			require.Equal(t, tt.template.Code, template.Code)
			require.Equal(t, tt.template.Status, template.Status)
			require.Equal(t, tt.template.Title, template.Title)
			require.Equal(t, tt.template.Details, template.Details)
			require.Equal(t, tt.template.Errors, template.Errors)
		})
	}
}

func TestTemplate_DetailsCanHoldAnyType(t *testing.T) {
	tests := []struct {
		name    string
		details any
	}{
		{
			name:    "string details",
			details: "simple string",
		},
		{
			name:    "map details",
			details: map[string]string{"key": "value"},
		},
		{
			name:    "slice details",
			details: []string{"item1", "item2"},
		},
		{
			name: "struct details",
			details: struct {
				Field1 string
				Field2 int
			}{
				Field1: "test",
				Field2: 42,
			},
		},
		{
			name:    "nil details",
			details: nil,
		},
		{
			name:    "integer details",
			details: 12345,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := Template{
				Details: tt.details,
			}

			require.Equal(t, tt.details, template.Details)
		})
	}
}
