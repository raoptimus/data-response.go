/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package dataresponse

import (
	"bufio"
	"bytes"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/raoptimus/data-response.go/v2/internal/conv"
)

// DataResponse represents an HTTP response with data payload.
type DataResponse struct {
	statusCode int
	data       any
	header     http.Header
	stream   io.Reader
	isBinary bool
	filename   string
	size       int64
	formatter  Formatter
}

// StatusCode returns the HTTP status code.
func (r DataResponse) StatusCode() int {
	return r.statusCode
}

func (r DataResponse) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

// Data returns the response data payload.
func (r DataResponse) Data() any {
	return r.data
}

func (r DataResponse) Body() (int64, io.Reader, error) {
	if r.stream != nil {
		return r.size, r.stream, nil
	}

	if r.formatter != nil {
		formattedResp, err := r.formatter.Format(r)
		if err != nil {
			return 0, nil, err
		}
		r.WithContentType(formattedResp.ContentType)

		return formattedResp.StreamSize, formattedResp.Stream, nil
	}

	data := r.Data()
	dataBytes, err := conv.DataToString(data)
	if err != nil {
		return 0, nil, err
	}
	var buf bytes.Buffer
	buf.Write(dataBytes)

	return int64(buf.Len()), bufio.NewReader(&buf), nil
}

func (r DataResponse) WithStream(stream io.Reader, size int64) DataResponse {
	r.stream = stream
	r.size = size

	return r
}

// Header returns the HTTP headers.
func (r DataResponse) Header() http.Header {
	return r.header
}

// HeaderValues returns all values for the given header key.
func (r DataResponse) HeaderValues(key string) []string {
	return r.header.Values(key)
}

// HeaderLine returns the first value for the given header key.
func (r DataResponse) HeaderLine(key string) string {
	return r.header.Get(key)
}

// ContentType returns the custom content type.
func (r DataResponse) ContentType() string {
	return r.HeaderLine(HeaderContentType)
}

// Binary returns the binary data reader.
func (r DataResponse) Binary() io.Reader {
	return r.stream
}

// Filename returns the filename for binary responses.
func (r DataResponse) Filename() string {
	return r.filename
}

// Size returns the size for binary responses.
func (r DataResponse) Size() int64 {
	return r.size
}

// IsBinary returns true if this is a binary response.
func (r DataResponse) IsBinary() bool {
	return r.isBinary && r.stream != nil
}

// HasData returns true if response has data payload.
func (r DataResponse) HasData() bool {
	return r.data != nil || r.stream != nil
}

// HasHeader returns true if header key exists.
func (r DataResponse) HasHeader(key string) bool {
	return r.header.Get(key) != ""
}

func (r DataResponse) Formatter() (Formatter, error) {
	if r.formatter != nil {
		return r.formatter, nil
	}

	return nil, errors.WithStack(ErrFormatterMustBeSet)
}

// WithHeader returns a copy of response with an additional header.
// It adds a value to the header list for the given key (supports multiple values).
func (r DataResponse) WithHeader(key, value string) DataResponse {
	if r.header == nil {
		r.header = make(http.Header)
	}
	if r.HasHeader(key) {
		r.header.Set(key, value)
	} else {
		r.header.Add(key, value)
	}

	return r
}

func (r DataResponse) WithoutHeader(key string) DataResponse {
	if r.header == nil {
		return r
	}

	r.header.Del(key)

	return r
}

// WithHeaders returns a copy of response with additional headers.
// It merges all headers from the provided http.Header.
func (r DataResponse) WithHeaders(headers http.Header) DataResponse {
	if r.header == nil {
		r.header = make(http.Header)
	}
	for key, values := range headers {
		for _, value := range values {
			r.header.Add(key, value)
		}
	}

	return r
}

// WithContentType returns a copy of response with a custom content type.
func (r DataResponse) WithContentType(contentType string) DataResponse {
	return r.WithHeader(HeaderContentType, contentType)
}

// WithData returns a copy of response with modified data.
func (r DataResponse) WithData(data any) DataResponse {
	r.data = data

	return r
}

// WithCacheControl returns a copy of response with Cache-Control header.
func (r DataResponse) WithCacheControl(value string) DataResponse {
	return r.WithHeader(HeaderCacheControl, value)
}

// WithCORS returns a copy of response with CORS headers.
func (r DataResponse) WithCORS(origin string, methods string, headers string) DataResponse {
	r = r.WithHeader(HeaderAccessControlAllowOrigin, origin)
	if methods != "" {
		r = r.WithHeader(HeaderAccessControlAllowMethods, methods)
	}
	if headers != "" {
		r = r.WithHeader(HeaderAccessControlAllowHeaders, headers)
	}

	return r
}

// WithSecurityHeaders returns a copy of response with common security headers.
func (r DataResponse) WithSecurityHeaders() DataResponse {
	return r.
		WithHeader(HeaderXContentTypeOptions, ContentTypeOptionsNoSniff).
		WithHeader(HeaderXFrameOptions, FrameOptionsDeny).
		WithHeader(HeaderReferrerPolicy, ReferrerPolicyStrictOriginWhenCrossOrigin)
}

func (r DataResponse) WithFormatter(formatter Formatter) DataResponse {
	r.formatter = formatter

	return r
}
