package dataresponse

import (
	"net/http"
)

// ChiMiddleware is a chi-compatible middleware type.
type ChiMiddleware func(http.Handler) http.Handler

// ToChiMiddleware converts DataResponse Middleware to chi middleware.
// This allows using DataResponse middleware with chi router.
func ToChiMiddleware(factory *Factory, drMiddleware Middleware) ChiMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap http.Handler as DataResponse Handler
			handler := HandlerFunc(func(req *http.Request) DataResponse {
				// Create a response recorder to capture the standard handler's response
				recorder := &responseRecorder{
					ResponseWriter: w,
					statusCode:     http.StatusOK,
				}

				next.ServeHTTP(recorder, req)

				// If handler already wrote response, return empty DataResponse
				if recorder.written {
					return DataResponse{statusCode: recorder.statusCode}
				}

				// Otherwise, return success (shouldn't happen in normal flow)
				return factory.Success(req.Context(), nil)
			})

			// Execute DataResponse middleware
			resp := drMiddleware.ServeHTTP(r, handler)

			// Write the response if not already written
			if !w.(interface{ Written() bool }).Written() {
				Write(w, r, resp, factory)
			}
		})
	}
}

// FromChiMiddleware converts chi middleware to DataResponse Middleware.
func FromChiMiddleware(chiMiddleware ChiMiddleware) Middleware {
	return MiddlewareFunc(func(r *http.Request, next Handler) DataResponse {
		// This is complex because we need to bridge between two different paradigms
		// For now, just pass through to next handler
		return next.Handle(r)
	})
}

// ChiAdapter provides helpers for working with chi router.
type ChiAdapter struct {
	factory *Factory
	adapter *Adapter
}

// NewChiAdapter creates a new chi adapter.
func NewChiAdapter(factory *Factory) *ChiAdapter {
	return &ChiAdapter{
		factory: factory,
		adapter: NewAdapter(factory),
	}
}

// Handler converts DataResponse Handler to http.Handler for use with chi.
func (ca *ChiAdapter) Handler(handler Handler) http.Handler {
	return ca.adapter.Handler(handler)
}

// Middleware converts DataResponse Middleware to chi middleware.
func (ca *ChiAdapter) Middleware(drMiddleware Middleware) ChiMiddleware {
	return ToChiMiddleware(ca.factory, drMiddleware)
}

// HandlerFunc is a shortcut for creating chi-compatible handlers.
func (ca *ChiAdapter) HandlerFunc(fn func(*http.Request) DataResponse) http.HandlerFunc {
	return ca.adapter.Handler(HandlerFunc(fn)).ServeHTTP
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

// ChiChain chains DataResponse middlewares into chi-compatible middleware.
func ChiChain(factory *Factory, middlewares ...Middleware) ChiMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Convert http.Handler to DataResponse Handler
			handler := HandlerFunc(func(req *http.Request) DataResponse {
				recorder := &responseRecorder{ResponseWriter: w}
				next.ServeHTTP(recorder, req)
				return DataResponse{statusCode: recorder.statusCode}
			})

			// Chain DataResponse middlewares
			chainedHandler := Chain(handler, middlewares...)

			// Execute and write response
			resp := chainedHandler.Handle(r)
			Write(w, r, resp, factory)
		})
	}
}
