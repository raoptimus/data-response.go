package dataresponse

import (
	"io"
	"net/http"
)

// DataResponse represents an HTTP response with data payload.
type DataResponse struct {
	statusCode  int
	data        any
	header      http.Header
	binary      io.Reader
	filename    string
	size        int64
}

// StatusCode returns the HTTP status code.
func (r DataResponse) StatusCode() int {
	return r.statusCode
}

// Data returns the response data payload.
func (r DataResponse) Data() any {
	return r.data
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
	return r.binary
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
	return r.binary != nil
}

// HasData returns true if response has data payload.
func (r DataResponse) HasData() bool {
	return r.data != nil || r.binary != nil
}

// HasHeader returns true if header key exists.
func (r DataResponse) HasHeader(key string) bool {
	return r.header.Get(key) != ""
}

// WithHeader returns a copy of response with an additional header.
// It adds a value to the header list for the given key (supports multiple values).
func (r DataResponse) WithHeader(key, value string) DataResponse {
	if r.header == nil {
		r.header = make(http.Header)
	}
	r.header.Add(key, value)
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

