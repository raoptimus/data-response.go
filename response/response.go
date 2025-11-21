/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package response

import (
	"bytes"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/raoptimus/data-response.go/v2/internal/conv"
)

const defaultHeadersCapacity = 5

// DataResponse represents an HTTP response with data payload.
type DataResponse struct {
	statusCode int
	data       any
	header     http.Header
	formatter  Formatter

	// Pre-formatted response (set by middleware like compression)
	formatted    FormattedResponse
	hasFormatted bool

	// Binary-specific fields
	isBinary bool
	filename string

	closer io.Closer // Close after response is written
}

func NewDataResponse(statusCode int, data any) *DataResponse {
	return &DataResponse{
		statusCode: statusCode,
		data:       data,
		header:     make(http.Header, defaultHeadersCapacity),
	}
}

// StatusCode returns the HTTP status code.
func (r *DataResponse) StatusCode() int {
	return r.statusCode
}

func (r *DataResponse) WithStatusCode(statusCode int) *DataResponse {
	r.statusCode = statusCode

	return r
}

// Data returns the response data payload.
func (r *DataResponse) Data() any {
	return r.data
}

func (r *DataResponse) Body() (FormattedResponse, error) {
	if r.hasFormatted {
		return r.formatted, nil
	}

	if r.formatter != nil {
		formattedResp, err := r.formatter.Format(r)
		if err != nil {
			return FormattedResponse{}, err
		}

		return formattedResp, nil
	}

	data := r.Data()
	dataBytes, err := conv.DataToString(data)
	if err != nil {
		return FormattedResponse{}, err
	}
	var buf bytes.Buffer
	buf.Write(dataBytes)

	return FormattedResponse{
		Stream:     bytes.NewReader(dataBytes),
		StreamSize: int64(len(dataBytes)),
	}, nil
}

// WithFormatted sets pre-formatted response.
// Used by compression middleware after compressing the body.
func (r *DataResponse) WithFormatted(formatted FormattedResponse) *DataResponse {
	r.formatted = formatted
	r.hasFormatted = true

	return r
}

// Header returns the HTTP headers.
func (r *DataResponse) Header() http.Header {
	return r.header
}

// HeaderValues returns all values for the given header key.
func (r *DataResponse) HeaderValues(key string) []string {
	if len(r.header) == 0 {
		var values []string

		return values
	}

	return r.header.Values(key)
}

// HeaderLine returns the first value for the given header key.
func (r *DataResponse) HeaderLine(key string) string {
	if len(r.header) == 0 {
		return ""
	}

	return r.header.Get(key)
}

// ContentType returns the custom content type.
func (r *DataResponse) ContentType() string {
	return r.HeaderLine(HeaderContentType)
}

// Filename returns the filename for binary responses.
func (r *DataResponse) Filename() string {
	return r.filename
}

// IsBinary returns true if this is a binary response.
func (r *DataResponse) IsBinary() bool {
	return r.isBinary
}

// HasHeader returns true if header key exists.
func (r *DataResponse) HasHeader(key string) bool {
	if len(r.header) == 0 {
		return false
	}

	return r.header.Get(key) != ""
}

//nolint:ireturn,nolintlint // its ok
func (r *DataResponse) Formatter() (Formatter, error) {
	if r.formatter != nil {
		return r.formatter, nil
	}

	return nil, errors.WithStack(ErrFormatterMustBeSet)
}

// WithHeader adds a header value (supports multiple values per key).
func (r *DataResponse) WithHeader(key, value string) *DataResponse {
	if r.header == nil {
		r.header = make(http.Header, defaultHeadersCapacity)
	}

	r.header.Add(key, value)

	return r
}

// SetHeader sets a header value (replaces existing values).
func (r *DataResponse) SetHeader(key, value string) *DataResponse {
	if r.header == nil {
		r.header = make(http.Header, defaultHeadersCapacity)
	}

	r.header.Set(key, value)

	return r
}

func (r *DataResponse) WithoutHeader(key string) *DataResponse {
	if len(r.header) == 0 {
		return r
	}

	r.header.Del(key)

	return r
}

// WithHeaders returns a copy of response with additional headers.
// It merges all headers from the provided http.Header.
func (r *DataResponse) WithHeaders(headers http.Header) *DataResponse {
	for key, values := range headers {
		for _, value := range values {
			r.header.Add(key, value)
		}
	}

	return r
}

// WithContentType returns a copy of response with a custom content type.
func (r *DataResponse) WithContentType(contentType string) *DataResponse {
	return r.SetHeader(HeaderContentType, contentType)
}

// WithData returns a copy of response with modified data.
func (r *DataResponse) WithData(data any) *DataResponse {
	r.data = data

	return r
}

// WithCacheControl returns a copy of response with Cache-Control header.
func (r *DataResponse) WithCacheControl(value string) *DataResponse {
	return r.WithHeader(HeaderCacheControl, value)
}

// WithCORS returns a copy of response with CORS headers.
func (r *DataResponse) WithCORS(origin, methods, headers string) *DataResponse {
	r.WithHeader(HeaderAccessControlAllowOrigin, origin)
	if len(methods) > 0 {
		r.WithHeader(HeaderAccessControlAllowMethods, methods)
	}
	if len(headers) > 0 {
		r.WithHeader(HeaderAccessControlAllowHeaders, headers)
	}

	return r
}

// WithSecurityHeaders returns a copy of response with common security headers.
func (r *DataResponse) WithSecurityHeaders() *DataResponse {
	return r.
		WithHeader(HeaderXContentTypeOptions, ContentTypeOptionsNoSniff).
		WithHeader(HeaderXFrameOptions, FrameOptionsDeny).
		WithHeader(HeaderReferrerPolicy, ReferrerPolicyStrictOriginWhenCrossOrigin)
}

func (r *DataResponse) WithContentDisposition(filename string) *DataResponse  {
	return r.WithHeader(HeaderContentDisposition, `attachment; filename="`+filename+`"`)
}

func (r *DataResponse) WithFormatter(formatter Formatter) *DataResponse {
	r.formatter = formatter

	return r.WithContentType(formatter.ContentType())
}

func (r *DataResponse) Close() error {
	if r.closer != nil {
		return r.closer.Close()
	}

	return nil
}

func (r *DataResponse) WithFile(closer io.Closer) *DataResponse {
	r.isBinary = true
	r.closer = closer

	return r
}
