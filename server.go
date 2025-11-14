/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package dataresponse

import (
	"net/http"
)

type Middleware func(next Handler) Handler

// Handler defines a handler that returns DataResponse instead of writing to http.ResponseWriter.
// This allows for more functional and composable middleware.
type Handler interface {
	Handle(r *http.Request, f *Factory) DataResponse
}

// HandlerFunc is a function adapter for Handler interface.
type HandlerFunc func(r *http.Request, f *Factory) DataResponse

// Handle calls f(r).
func (hf HandlerFunc) Handle(r *http.Request, f *Factory) DataResponse {
	return hf(r, f)
}

// WrapHandler converts DataResponse Handler to http.Handler.
func WrapHandler(h Handler, f *Factory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseRecorder{ResponseWriter: w}
		resp := h.Handle(r, f)
		if err := Write(r.Context(), rw, resp, f); err != nil {
			f.logger.Error(r.Context(), "failed to write response", "error", err.Error())
		}
	})
}

// WrapHandlerFunc converts DataResponse HandlerFunc to http.HandlerFunc.
func WrapHandlerFunc(hf HandlerFunc, f *Factory) http.HandlerFunc {
	return WrapHandler(hf, f).ServeHTTP
}

type ServeMux struct {
	*http.ServeMux
	factory     *Factory
	middlewares []Middleware
}

// NewServeMux allocates and returns a new [ServeMux].
func NewServeMux(factory *Factory) *ServeMux {
	return &ServeMux{
		ServeMux:    http.NewServeMux(),
		factory:     factory,
		middlewares: make([]Middleware, 0),
	}
}

func (s *ServeMux) Handle(pattern string, handler Handler) {
	s.ServeMux.Handle(pattern, Chain(s.factory, handler, s.middlewares...))
	//s.ServeMux.Handle(pattern, WrapHandler(handler, s.factory))
}

func (s *ServeMux) HandleFunc(pattern string, handler HandlerFunc) {
	s.ServeMux.HandleFunc(pattern, Chain(s.factory, handler, s.middlewares...).ServeHTTP)
	//s.ServeMux.HandleFunc(pattern, WrapHandlerFunc(handler, s.factory))
}

func (s *ServeMux) WithMiddleware(m ...Middleware) {
	s.middlewares = append(s.middlewares, m...)
}

// responseRecorder captures response data.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader captures the status code.
func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.written = true
	rr.ResponseWriter.WriteHeader(statusCode)
}

// Write marks response as written.
func (rr *responseRecorder) Write(b []byte) (int, error) {
	rr.written = true
	return rr.ResponseWriter.Write(b)
}

// Written returns true if response was written.
func (rr *responseRecorder) Written() bool {
	return rr.written
}

func (rr *responseRecorder) StatusCode() int {
	return rr.statusCode
}
