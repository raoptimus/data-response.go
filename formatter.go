package dataresponse

import (
	"net/http"
	"strconv"
)

// Formatter defines the interface for response formatting strategies.
type Formatter interface {
	// Format writes the formatted response to http.ResponseWriter.
	Format(w http.ResponseWriter, resp DataResponse) error

	// ContentType returns the default Content-Type for this formatter.
	ContentType() string

	// CanFormatBinary returns true if formatter can handle binary data.
	CanFormatBinary() bool
}

// BaseFormatter provides common functionality for formatters.
type BaseFormatter struct{}

// CanFormatBinary returns false by default.
func (BaseFormatter) CanFormatBinary() bool {
	return false
}

// WriteHeaders writes common headers to the response.
// It handles both single and multiple values per header key using http.Header.
func (BaseFormatter) WriteHeaders(w http.ResponseWriter, resp DataResponse, contentType string) {
	// Write custom headers from response (supporting multiple values per key)
	for key, values := range resp.Header() {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Content-Type
	if resp.ContentType() != "" {
		w.Header().Set("Content-Type", resp.ContentType())
	} else if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	// Binary-specific headers
	if resp.IsBinary() {
		if resp.Filename() != "" {
			w.Header().Set("Content-Disposition", `attachment; filename="`+resp.Filename()+`"`)
		}
		if resp.Size() > 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(resp.Size(), 10))
		}
	}
}
