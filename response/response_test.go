/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package response_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/raoptimus/data-response.go/v2/response"
	"github.com/raoptimus/data-response.go/v2/response/mockresponse"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockCloser is a test double for io.Closer
type mockCloser struct {
	closed     bool
	closeErr   error
	closeCalls int
}

func (m *mockCloser) Close() error {
	m.closed = true
	m.closeCalls++
	return m.closeErr
}

func TestNewDataResponse_CreatesResponseWithStatusCodeAndData_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		data         any
		expectedCode int
		expectedData any
	}{
		{
			name:         "status 200 with string data",
			statusCode:   http.StatusOK,
			data:         "test data",
			expectedCode: http.StatusOK,
			expectedData: "test data",
		},
		{
			name:         "status 201 with struct data",
			statusCode:   http.StatusCreated,
			data:         struct{ ID int }{ID: 42},
			expectedCode: http.StatusCreated,
			expectedData: struct{ ID int }{ID: 42},
		},
		{
			name:         "status 404 with nil data",
			statusCode:   http.StatusNotFound,
			data:         nil,
			expectedCode: http.StatusNotFound,
			expectedData: nil,
		},
		{
			name:         "status 500 with byte slice data",
			statusCode:   http.StatusInternalServerError,
			data:         []byte("error content"),
			expectedCode: http.StatusInternalServerError,
			expectedData: []byte("error content"),
		},
		{
			name:         "status 0 boundary",
			statusCode:   0,
			data:         "zero status",
			expectedCode: 0,
			expectedData: "zero status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(tt.statusCode, tt.data)

			require.NotNil(t, resp)
			require.Equal(t, tt.expectedCode, resp.StatusCode())
			require.Equal(t, tt.expectedData, resp.Data())
			require.NotNil(t, resp.Header())
			require.False(t, resp.IsBinary())
			require.Empty(t, resp.Filename())
		})
	}
}

func TestDataResponse_StatusCode_ReturnsStatusCode_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		expectedCode int
	}{
		{
			name:         "returns 200",
			statusCode:   http.StatusOK,
			expectedCode: http.StatusOK,
		},
		{
			name:         "returns 404",
			statusCode:   http.StatusNotFound,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "returns 500",
			statusCode:   http.StatusInternalServerError,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(tt.statusCode, nil)

			result := resp.StatusCode()

			require.Equal(t, tt.expectedCode, result)
		})
	}
}

func TestDataResponse_WithStatusCode_UpdatesStatusCode_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		initialCode  int
		newCode      int
		expectedCode int
	}{
		{
			name:         "change from 200 to 201",
			initialCode:  http.StatusOK,
			newCode:      http.StatusCreated,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "change from 404 to 500",
			initialCode:  http.StatusNotFound,
			newCode:      http.StatusInternalServerError,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "change to same status code",
			initialCode:  http.StatusOK,
			newCode:      http.StatusOK,
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(tt.initialCode, nil)

			result := resp.WithStatusCode(tt.newCode)

			require.Same(t, resp, result)
			require.Equal(t, tt.expectedCode, result.StatusCode())
		})
	}
}

func TestDataResponse_Data_ReturnsData_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		data         any
		expectedData any
	}{
		{
			name:         "returns string data",
			data:         "test string",
			expectedData: "test string",
		},
		{
			name:         "returns nil data",
			data:         nil,
			expectedData: nil,
		},
		{
			name:         "returns struct data",
			data:         struct{ Name string }{Name: "test"},
			expectedData: struct{ Name string }{Name: "test"},
		},
		{
			name:         "returns map data",
			data:         map[string]int{"key": 123},
			expectedData: map[string]int{"key": 123},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, tt.data)

			result := resp.Data()

			require.Equal(t, tt.expectedData, result)
		})
	}
}

func TestDataResponse_WithData_UpdatesData_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		initialData  any
		newData      any
		expectedData any
	}{
		{
			name:         "change string to string",
			initialData:  "initial",
			newData:      "updated",
			expectedData: "updated",
		},
		{
			name:         "change nil to struct",
			initialData:  nil,
			newData:      struct{ ID int }{ID: 1},
			expectedData: struct{ ID int }{ID: 1},
		},
		{
			name:         "change struct to nil",
			initialData:  struct{ ID int }{ID: 1},
			newData:      nil,
			expectedData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, tt.initialData)

			result := resp.WithData(tt.newData)

			require.Same(t, resp, result)
			require.Equal(t, tt.expectedData, result.Data())
		})
	}
}

func TestDataResponse_Header_ReturnsHeader_Successfully(t *testing.T) {
	t.Run("returns empty header for new response", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)

		result := resp.Header()

		require.NotNil(t, result)
		require.Empty(t, result)
	})

	t.Run("returns header with values after adding", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)
		resp.WithHeader("X-Custom", "value")

		result := resp.Header()

		require.NotNil(t, result)
		require.Equal(t, "value", result.Get("X-Custom"))
	})
}

func TestDataResponse_HeaderValues_ReturnsAllValues_Successfully(t *testing.T) {
	tests := []struct {
		name           string
		setupHeaders   func(r *response.DataResponse)
		key            string
		expectedValues []string
	}{
		{
			name:           "returns empty slice for non-existent key",
			setupHeaders:   func(r *response.DataResponse) {},
			key:            "X-Missing",
			expectedValues: nil,
		},
		{
			name: "returns single value",
			setupHeaders: func(r *response.DataResponse) {
				r.WithHeader("X-Single", "value1")
			},
			key:            "X-Single",
			expectedValues: []string{"value1"},
		},
		{
			name: "returns multiple values",
			setupHeaders: func(r *response.DataResponse) {
				r.WithHeader("X-Multi", "value1")
				r.WithHeader("X-Multi", "value2")
				r.WithHeader("X-Multi", "value3")
			},
			key:            "X-Multi",
			expectedValues: []string{"value1", "value2", "value3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, nil)
			tt.setupHeaders(resp)

			result := resp.HeaderValues(tt.key)

			require.Equal(t, tt.expectedValues, result)
		})
	}
}

func TestDataResponse_HeaderLine_ReturnsFirstValue_Successfully(t *testing.T) {
	tests := []struct {
		name          string
		setupHeaders  func(r *response.DataResponse)
		key           string
		expectedValue string
	}{
		{
			name:          "returns empty string for non-existent key",
			setupHeaders:  func(r *response.DataResponse) {},
			key:           "X-Missing",
			expectedValue: "",
		},
		{
			name: "returns value for existing key",
			setupHeaders: func(r *response.DataResponse) {
				r.WithHeader("X-Existing", "the-value")
			},
			key:           "X-Existing",
			expectedValue: "the-value",
		},
		{
			name: "returns first value when multiple exist",
			setupHeaders: func(r *response.DataResponse) {
				r.WithHeader("X-Multi", "first")
				r.WithHeader("X-Multi", "second")
			},
			key:           "X-Multi",
			expectedValue: "first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, nil)
			tt.setupHeaders(resp)

			result := resp.HeaderLine(tt.key)

			require.Equal(t, tt.expectedValue, result)
		})
	}
}

func TestDataResponse_ContentType_ReturnsContentTypeHeader_Successfully(t *testing.T) {
	tests := []struct {
		name          string
		setupHeaders  func(r *response.DataResponse)
		expectedValue string
	}{
		{
			name:          "returns empty string when not set",
			setupHeaders:  func(r *response.DataResponse) {},
			expectedValue: "",
		},
		{
			name: "returns JSON content type",
			setupHeaders: func(r *response.DataResponse) {
				r.WithContentType(response.ContentTypeJSON)
			},
			expectedValue: response.ContentTypeJSON,
		},
		{
			name: "returns XML content type",
			setupHeaders: func(r *response.DataResponse) {
				r.WithContentType(response.ContentTypeXML)
			},
			expectedValue: response.ContentTypeXML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, nil)
			tt.setupHeaders(resp)

			result := resp.ContentType()

			require.Equal(t, tt.expectedValue, result)
		})
	}
}

func TestDataResponse_Filename_ReturnsFilename_Successfully(t *testing.T) {
	t.Run("returns empty string by default", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)

		result := resp.Filename()

		require.Empty(t, result)
	})
}

func TestDataResponse_IsBinary_ReturnsBinaryFlag_Successfully(t *testing.T) {
	t.Run("returns false by default", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)

		result := resp.IsBinary()

		require.False(t, result)
	})

	t.Run("returns true after WithFile", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)
		closer := &mockCloser{}
		resp.WithFile(closer)

		result := resp.IsBinary()

		require.True(t, result)
	})
}

func TestDataResponse_HasHeader_ChecksHeaderExistence_Successfully(t *testing.T) {
	tests := []struct {
		name           string
		setupHeaders   func(r *response.DataResponse)
		key            string
		expectedResult bool
	}{
		{
			name:           "returns false for non-existent key",
			setupHeaders:   func(r *response.DataResponse) {},
			key:            "X-Missing",
			expectedResult: false,
		},
		{
			name: "returns true for existing key",
			setupHeaders: func(r *response.DataResponse) {
				r.WithHeader("X-Exists", "value")
			},
			key:            "X-Exists",
			expectedResult: true,
		},
		{
			name: "returns false for empty value",
			setupHeaders: func(r *response.DataResponse) {
				r.Header().Set("X-Empty", "")
			},
			key:            "X-Empty",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, nil)
			tt.setupHeaders(resp)

			result := resp.HasHeader(tt.key)

			require.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestDataResponse_WithHeader_AddsHeaderValue_Successfully(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		value          string
		expectedValues []string
	}{
		{
			name:           "adds single header",
			key:            "X-Custom",
			value:          "value1",
			expectedValues: []string{"value1"},
		},
		{
			name:           "adds header with empty value",
			key:            "X-Empty",
			value:          "",
			expectedValues: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, nil)

			result := resp.WithHeader(tt.key, tt.value)

			require.Same(t, resp, result)
			require.Equal(t, tt.expectedValues, result.HeaderValues(tt.key))
		})
	}
}

func TestDataResponse_WithHeader_AddsMultipleValues_Successfully(t *testing.T) {
	t.Run("adds multiple values for same key", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)

		resp.WithHeader("X-Multi", "value1")
		resp.WithHeader("X-Multi", "value2")
		result := resp.WithHeader("X-Multi", "value3")

		require.Same(t, resp, result)
		require.Equal(t, []string{"value1", "value2", "value3"}, result.HeaderValues("X-Multi"))
	})
}

func TestDataResponse_SetHeader_ReplacesHeaderValue_Successfully(t *testing.T) {
	tests := []struct {
		name          string
		setupHeaders  func(r *response.DataResponse)
		key           string
		value         string
		expectedValue string
	}{
		{
			name:          "sets new header",
			setupHeaders:  func(r *response.DataResponse) {},
			key:           "X-New",
			value:         "new-value",
			expectedValue: "new-value",
		},
		{
			name: "replaces existing header",
			setupHeaders: func(r *response.DataResponse) {
				r.WithHeader("X-Existing", "old-value")
			},
			key:           "X-Existing",
			value:         "new-value",
			expectedValue: "new-value",
		},
		{
			name: "replaces multiple values with single",
			setupHeaders: func(r *response.DataResponse) {
				r.WithHeader("X-Multi", "value1")
				r.WithHeader("X-Multi", "value2")
			},
			key:           "X-Multi",
			value:         "single-value",
			expectedValue: "single-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, nil)
			tt.setupHeaders(resp)

			result := resp.SetHeader(tt.key, tt.value)

			require.Same(t, resp, result)
			require.Equal(t, tt.expectedValue, result.HeaderLine(tt.key))
			require.Len(t, result.HeaderValues(tt.key), 1)
		})
	}
}

func TestDataResponse_WithoutHeader_RemovesHeader_Successfully(t *testing.T) {
	tests := []struct {
		name           string
		setupHeaders   func(r *response.DataResponse)
		key            string
		expectedExists bool
	}{
		{
			name: "removes existing header",
			setupHeaders: func(r *response.DataResponse) {
				r.WithHeader("X-Remove", "value")
			},
			key:            "X-Remove",
			expectedExists: false,
		},
		{
			name: "removes all values for key",
			setupHeaders: func(r *response.DataResponse) {
				r.WithHeader("X-Multi", "value1")
				r.WithHeader("X-Multi", "value2")
			},
			key:            "X-Multi",
			expectedExists: false,
		},
		{
			name:           "handles non-existent key gracefully",
			setupHeaders:   func(r *response.DataResponse) {},
			key:            "X-Missing",
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, nil)
			tt.setupHeaders(resp)

			result := resp.WithoutHeader(tt.key)

			require.Same(t, resp, result)
			require.Equal(t, tt.expectedExists, result.HasHeader(tt.key))
		})
	}
}

func TestDataResponse_WithHeaders_MergesHeaders_Successfully(t *testing.T) {
	tests := []struct {
		name            string
		setupHeaders    func(r *response.DataResponse)
		newHeaders      http.Header
		expectedHeaders map[string][]string
	}{
		{
			name:         "merges into empty headers",
			setupHeaders: func(r *response.DataResponse) {},
			newHeaders: http.Header{
				"X-New-1": []string{"value1"},
				"X-New-2": []string{"value2"},
			},
			expectedHeaders: map[string][]string{
				"X-New-1": {"value1"},
				"X-New-2": {"value2"},
			},
		},
		{
			name: "adds to existing headers",
			setupHeaders: func(r *response.DataResponse) {
				r.WithHeader("X-Existing", "existing-value")
			},
			newHeaders: http.Header{
				"X-New": []string{"new-value"},
			},
			expectedHeaders: map[string][]string{
				"X-Existing": {"existing-value"},
				"X-New":      {"new-value"},
			},
		},
		{
			name: "appends to existing key",
			setupHeaders: func(r *response.DataResponse) {
				r.WithHeader("X-Multi", "value1")
			},
			newHeaders: http.Header{
				"X-Multi": []string{"value2", "value3"},
			},
			expectedHeaders: map[string][]string{
				"X-Multi": {"value1", "value2", "value3"},
			},
		},
		{
			name:            "handles nil new headers",
			setupHeaders:    func(r *response.DataResponse) {},
			newHeaders:      nil,
			expectedHeaders: map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, nil)
			tt.setupHeaders(resp)

			result := resp.WithHeaders(tt.newHeaders)

			require.Same(t, resp, result)
			for key, expectedValues := range tt.expectedHeaders {
				require.Equal(t, expectedValues, result.HeaderValues(key))
			}
		})
	}
}

func TestDataResponse_WithContentType_SetsContentTypeHeader_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		expectedType string
	}{
		{
			name:         "sets JSON content type",
			contentType:  response.ContentTypeJSON,
			expectedType: response.ContentTypeJSON,
		},
		{
			name:         "sets XML content type",
			contentType:  response.ContentTypeXML,
			expectedType: response.ContentTypeXML,
		},
		{
			name:         "sets plain text content type",
			contentType:  response.ContentTypePlain,
			expectedType: response.ContentTypePlain,
		},
		{
			name:         "sets custom content type",
			contentType:  "application/custom",
			expectedType: "application/custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, nil)

			result := resp.WithContentType(tt.contentType)

			require.Same(t, resp, result)
			require.Equal(t, tt.expectedType, result.ContentType())
		})
	}
}

func TestDataResponse_WithCacheControl_SetsCacheControlHeader_Successfully(t *testing.T) {
	tests := []struct {
		name          string
		value         string
		expectedValue string
	}{
		{
			name:          "sets no-cache",
			value:         response.CacheControlNoCache,
			expectedValue: response.CacheControlNoCache,
		},
		{
			name:          "sets no-store",
			value:         response.CacheControlNoStore,
			expectedValue: response.CacheControlNoStore,
		},
		{
			name:          "sets max-age",
			value:         "max-age=3600",
			expectedValue: "max-age=3600",
		},
		{
			name:          "sets complex cache control",
			value:         "public, max-age=86400",
			expectedValue: "public, max-age=86400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, nil)

			result := resp.WithCacheControl(tt.value)

			require.Same(t, resp, result)
			require.Equal(t, tt.expectedValue, result.HeaderLine(response.HeaderCacheControl))
		})
	}
}

func TestDataResponse_WithCORS_SetsCORSHeaders_Successfully(t *testing.T) {
	tests := []struct {
		name            string
		origin          string
		methods         string
		headers         string
		expectedOrigin  string
		expectedMethods string
		expectedHeaders string
	}{
		{
			name:            "sets all CORS headers",
			origin:          "https://example.com",
			methods:         "GET, POST, PUT",
			headers:         "Content-Type, Authorization",
			expectedOrigin:  "https://example.com",
			expectedMethods: "GET, POST, PUT",
			expectedHeaders: "Content-Type, Authorization",
		},
		{
			name:            "sets origin only",
			origin:          "*",
			methods:         "",
			headers:         "",
			expectedOrigin:  "*",
			expectedMethods: "",
			expectedHeaders: "",
		},
		{
			name:            "sets origin and methods only",
			origin:          "https://api.example.com",
			methods:         "GET",
			headers:         "",
			expectedOrigin:  "https://api.example.com",
			expectedMethods: "GET",
			expectedHeaders: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, nil)

			result := resp.WithCORS(tt.origin, tt.methods, tt.headers)

			require.Same(t, resp, result)
			require.Equal(t, tt.expectedOrigin, result.HeaderLine(response.HeaderAccessControlAllowOrigin))
			if tt.expectedMethods != "" {
				require.Equal(t, tt.expectedMethods, result.HeaderLine(response.HeaderAccessControlAllowMethods))
			} else {
				require.False(t, result.HasHeader(response.HeaderAccessControlAllowMethods))
			}
			if tt.expectedHeaders != "" {
				require.Equal(t, tt.expectedHeaders, result.HeaderLine(response.HeaderAccessControlAllowHeaders))
			} else {
				require.False(t, result.HasHeader(response.HeaderAccessControlAllowHeaders))
			}
		})
	}
}

func TestDataResponse_WithSecurityHeaders_SetsSecurityHeaders_Successfully(t *testing.T) {
	t.Run("sets all security headers", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)

		result := resp.WithSecurityHeaders()

		require.Same(t, resp, result)
		require.Equal(t, response.ContentTypeOptionsNoSniff, result.HeaderLine(response.HeaderXContentTypeOptions))
		require.Equal(t, response.FrameOptionsDeny, result.HeaderLine(response.HeaderXFrameOptions))
		require.Equal(t, response.ReferrerPolicyStrictOriginWhenCrossOrigin, result.HeaderLine(response.HeaderReferrerPolicy))
	})
}

func TestDataResponse_WithContentDisposition_SetsHeaderAndSanitizesFilename_Successfully(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		expectedHeader string
	}{
		{
			name:           "sets simple filename",
			filename:       "document.pdf",
			expectedHeader: `attachment; filename="document.pdf"`,
		},
		{
			name:           "sets filename with spaces",
			filename:       "my document.pdf",
			expectedHeader: `attachment; filename="my document.pdf"`,
		},
		{
			name:           "sanitizes path with forward slashes",
			filename:       "/path/to/document.pdf",
			expectedHeader: `attachment; filename="document.pdf"`,
		},
		{
			name:           "sanitizes path with backslashes",
			filename:       `C:\Users\test\document.pdf`,
			expectedHeader: `attachment; filename="document.pdf"`,
		},
		{
			name:           "escapes quotes in filename",
			filename:       `file"with"quotes.pdf`,
			expectedHeader: `attachment; filename="file\"with\"quotes.pdf"`,
		},
		{
			name:           "removes carriage return",
			filename:       "file\rwith\rcr.pdf",
			expectedHeader: `attachment; filename="filewithcr.pdf"`,
		},
		{
			name:           "removes newline",
			filename:       "file\nwith\nnl.pdf",
			expectedHeader: `attachment; filename="filewithnl.pdf"`,
		},
		{
			name:           "sanitizes complex malicious filename",
			filename:       "../../../etc/passwd\r\nX-Injected: header",
			expectedHeader: `attachment; filename="passwdX-Injected: header"`,
		},
		{
			name:           "handles empty filename",
			filename:       "",
			expectedHeader: `attachment; filename="."`,
		},
		{
			name:           "handles filename with only dots",
			filename:       "...",
			expectedHeader: `attachment; filename="..."`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, nil)

			result := resp.WithContentDisposition(tt.filename)

			require.Same(t, resp, result)
			require.Equal(t, tt.expectedHeader, result.HeaderLine(response.HeaderContentDisposition))
		})
	}
}

func TestDataResponse_WithFormatter_SetsFormatterAndContentType_Successfully(t *testing.T) {
	t.Run("sets formatter and content type", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)
		formatter := mockresponse.NewFormatter(t)
		formatter.EXPECT().ContentType().Return(response.ContentTypeJSON)

		result := resp.WithFormatter(formatter)

		require.Same(t, resp, result)
		require.Equal(t, response.ContentTypeJSON, result.ContentType())

		returnedFormatter, err := result.Formatter()
		require.NoError(t, err)
		require.Same(t, formatter, returnedFormatter)
	})
}

func TestDataResponse_WithFormatted_SetsPreFormattedContent_Successfully(t *testing.T) {
	t.Run("sets pre-formatted response", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)
		formatted := response.FormattedResponse{
			Stream:     bytes.NewReader([]byte(`{"key":"value"}`)),
			StreamSize: 15,
		}

		result := resp.WithFormatted(formatted)

		require.Same(t, resp, result)

		body, err := result.Body()
		require.NoError(t, err)
		require.Equal(t, int64(15), body.StreamSize)
	})
}

func TestDataResponse_WithFile_SetsBinaryFlagAndCloser_Successfully(t *testing.T) {
	t.Run("sets binary flag and closer", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)
		closer := &mockCloser{}

		result := resp.WithFile(closer)

		require.Same(t, resp, result)
		require.True(t, result.IsBinary())

		err := result.Close()
		require.NoError(t, err)
		require.True(t, closer.closed)
	})
}

func TestDataResponse_Body_ReturnsPreFormattedContent_Successfully(t *testing.T) {
	t.Run("returns pre-formatted content when set", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)
		content := []byte(`{"pre":"formatted"}`)
		formatted := response.FormattedResponse{
			Stream:     bytes.NewReader(content),
			StreamSize: int64(len(content)),
		}
		resp.WithFormatted(formatted)

		body, err := resp.Body()

		require.NoError(t, err)
		require.Equal(t, int64(19), body.StreamSize)
	})
}

func TestDataResponse_Body_UsesFormatter_Successfully(t *testing.T) {
	t.Run("uses formatter when set", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, struct{ Name string }{Name: "test"})
		formatter := mockresponse.NewFormatter(t)
		formattedContent := []byte(`{"name":"test"}`)
		expectedFormatted := response.FormattedResponse{
			Stream:     bytes.NewReader(formattedContent),
			StreamSize: int64(len(formattedContent)),
		}
		formatter.EXPECT().ContentType().Return(response.ContentTypeJSON)
		formatter.EXPECT().Format(mock.Anything).Return(expectedFormatted, nil)
		resp.WithFormatter(formatter)

		body, err := resp.Body()

		require.NoError(t, err)
		require.Equal(t, int64(15), body.StreamSize)
	})
}

func TestDataResponse_Body_FormatterReturnsError_Failure(t *testing.T) {
	t.Run("returns error when formatter fails", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, struct{ Name string }{Name: "test"})
		formatter := mockresponse.NewFormatter(t)
		formatter.EXPECT().ContentType().Return(response.ContentTypeJSON)
		formatter.EXPECT().Format(mock.Anything).Return(response.FormattedResponse{}, errors.New("format error"))
		resp.WithFormatter(formatter)

		body, err := resp.Body()

		require.Error(t, err)
		require.Contains(t, err.Error(), "format error")
		require.Equal(t, response.FormattedResponse{}, body)
	})
}

func TestDataResponse_Body_ConvertsStringData_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		data         any
		expectedSize int64
	}{
		{
			name:         "converts string data",
			data:         "hello world",
			expectedSize: 11,
		},
		{
			name:         "converts byte slice data",
			data:         []byte("byte data"),
			expectedSize: 9,
		},
		{
			name:         "converts nil data",
			data:         nil,
			expectedSize: 4, // "null"
		},
		{
			name:         "converts empty string",
			data:         "",
			expectedSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := response.NewDataResponse(http.StatusOK, tt.data)

			body, err := resp.Body()

			require.NoError(t, err)
			require.Equal(t, tt.expectedSize, body.StreamSize)
		})
	}
}

func TestDataResponse_Body_NonStringableData_Failure(t *testing.T) {
	t.Run("returns error for non-stringable data without formatter", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, struct{ ID int }{ID: 1})

		body, err := resp.Body()

		require.Error(t, err)
		require.Equal(t, response.FormattedResponse{}, body)
	})
}

func TestDataResponse_Formatter_ReturnsFormatter_Successfully(t *testing.T) {
	t.Run("returns formatter when set", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)
		formatter := mockresponse.NewFormatter(t)
		formatter.EXPECT().ContentType().Return(response.ContentTypeJSON)
		resp.WithFormatter(formatter)

		result, err := resp.Formatter()

		require.NoError(t, err)
		require.Same(t, formatter, result)
	})
}

func TestDataResponse_Formatter_FormatterNotSet_Failure(t *testing.T) {
	t.Run("returns error when formatter is not set", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)

		result, err := resp.Formatter()

		require.Error(t, err)
		require.ErrorIs(t, err, response.ErrFormatterMustBeSet)
		require.Nil(t, result)
	})
}

func TestDataResponse_Close_ClosesCloser_Successfully(t *testing.T) {
	t.Run("closes closer when set", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)
		closer := &mockCloser{}
		resp.WithFile(closer)

		err := resp.Close()

		require.NoError(t, err)
		require.True(t, closer.closed)
		require.Equal(t, 1, closer.closeCalls)
	})
}

func TestDataResponse_Close_CloserReturnsError_Failure(t *testing.T) {
	t.Run("returns error from closer", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)
		closer := &mockCloser{closeErr: errors.New("close error")}
		resp.WithFile(closer)

		err := resp.Close()

		require.Error(t, err)
		require.Contains(t, err.Error(), "close error")
		require.True(t, closer.closed)
	})
}

func TestDataResponse_Close_NoCloser_Successfully(t *testing.T) {
	t.Run("returns nil when no closer is set", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, nil)

		err := resp.Close()

		require.NoError(t, err)
	})
}

func TestDataResponse_FluentChaining_Successfully(t *testing.T) {
	t.Run("supports method chaining", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, "initial data")

		result := resp.
			WithStatusCode(http.StatusCreated).
			WithData("updated data").
			WithContentType(response.ContentTypeJSON).
			WithHeader("X-Custom", "value").
			WithCacheControl(response.CacheControlNoCache).
			WithSecurityHeaders()

		require.Same(t, resp, result)
		require.Equal(t, http.StatusCreated, result.StatusCode())
		require.Equal(t, "updated data", result.Data())
		require.Equal(t, response.ContentTypeJSON, result.ContentType())
		require.Equal(t, "value", result.HeaderLine("X-Custom"))
		require.Equal(t, response.CacheControlNoCache, result.HeaderLine(response.HeaderCacheControl))
		require.Equal(t, response.ContentTypeOptionsNoSniff, result.HeaderLine(response.HeaderXContentTypeOptions))
	})
}

func TestDataResponse_BodyStream_ReturnsReadableStream_Successfully(t *testing.T) {
	t.Run("returns readable stream", func(t *testing.T) {
		resp := response.NewDataResponse(http.StatusOK, "test content")

		body, err := resp.Body()

		require.NoError(t, err)

		content, err := io.ReadAll(body.Stream)
		require.NoError(t, err)
		require.Equal(t, "test content", string(content))
	})
}
