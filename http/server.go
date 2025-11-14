package http

import "net/http"

// ResponseRecorder captures response data.
type ResponseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{ResponseWriter: w}
}

// WriteHeader captures the status code.
func (rr *ResponseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.written = true
	rr.ResponseWriter.WriteHeader(statusCode)
}

// Write marks response as written.
func (rr *ResponseRecorder) Write(b []byte) (int, error) {
	rr.written = true
	return rr.ResponseWriter.Write(b)
}

// Written returns true if response was written.
func (rr *ResponseRecorder) Written() bool {
	return rr.written
}

func (rr *ResponseRecorder) StatusCode() int {
	return rr.statusCode
}
