/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package response

import (
	"errors"
	"fmt"
	"io"
)

// StreamFunc writes the response body incrementally.
// Status code and headers are sent to the client with the first non-empty write to w.
// w is not safe for concurrent use and must not be kept after the function returns.
type StreamFunc func(w io.Writer) error

// StreamPhase tells whether the client had already received the status line when a stream failed.
type StreamPhase uint8

const (
	// StreamPhaseBeforeWrite means nothing was sent yet, so an error response can replace the stream.
	StreamPhaseBeforeWrite StreamPhase = iota + 1
	// StreamPhaseAfterWrite means status and a part of the body are sent, the only option is to abort.
	StreamPhaseAfterWrite
)

func (p StreamPhase) String() string {
	switch p {
	case StreamPhaseBeforeWrite:
		return "before_write"
	case StreamPhaseAfterWrite:
		return "after_write"
	default:
		return "unknown"
	}
}

var errStreamNotReadable = errors.New("stream body must be consumed with WriteTo")

// StreamError is returned by Write when a stream body fails.
type StreamError struct {
	Phase StreamPhase
	// Err is the error returned by the stream function, or a *StreamPanicError if it panicked.
	Err error
	// WriteErr is the first error returned by the http.ResponseWriter, nil if writes never failed.
	// A non-nil value usually means the client has gone away.
	WriteErr error
}

func (e *StreamError) Error() string {
	if e.Err == nil {
		return "stream " + e.Phase.String()
	}

	return "stream " + e.Phase.String() + ": " + e.Err.Error()
}

func (e *StreamError) Unwrap() error {
	return e.Err
}

// StreamPanicError is the cause of a StreamError when the stream function panicked after the first write.
type StreamPanicError struct {
	Value any
	Stack []byte
}

func (e *StreamPanicError) Error() string {
	return fmt.Sprintf("panic: %v", e.Value)
}

// Unwrap returns the panic value if it is an error.
func (e *StreamPanicError) Unwrap() error {
	err, _ := e.Value.(error)

	return err
}

// streamBody adapts StreamFunc to io.Reader + io.WriterTo, so io.Copy hands it the writer directly.
type streamBody struct {
	fn StreamFunc
}

func (s streamBody) Read([]byte) (int, error) {
	return 0, errStreamNotReadable
}

func (s streamBody) WriteTo(w io.Writer) (int64, error) {
	cw := countingWriter{w: w}
	err := s.fn(&cw)

	return cw.n, err
}

type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)

	return n, err
}

// WithStream turns the response into a lazily committed stream.
// Body of such a response has unknown size (StreamSize is -1) and is never compressed.
func (r *DataResponse) WithStream(fn StreamFunc) *DataResponse {
	r.stream = fn

	return r.WithFormatted(FormattedResponse{Stream: streamBody{fn: fn}, StreamSize: -1})
}

// IsStream returns true for responses created with WithStream.
func (r *DataResponse) IsStream() bool {
	return r.stream != nil
}

// StreamFunc returns the body writer of a stream response, nil for any other response.
func (r *DataResponse) StreamFunc() StreamFunc {
	return r.stream
}
