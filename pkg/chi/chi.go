package chi

import (
	"net/http"

	chiorigin "github.com/go-chi/chi/v5"
	dr "github.com/raoptimus/data-response.go/v2"
)

// ChiMiddleware is a chi-compatible middleware type.
type ChiMiddleware func(http.Handler) http.Handler

// ToChiMiddleware converts DataResponse Middleware to chi middleware.
// This allows using DataResponse middleware with chi router.
func ToChiMiddleware(factory *dr.Factory, drMiddleware dr.Middleware) ChiMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			chiorigin.Chain()
			// Wrap http.Handler as DataResponse Handler
			handler := dr.HandlerFunc(func(req *http.Request) DataResponse {
				// Create a response recorder to capture the standard handler's response
				recorder := &dr.responseRecorder{
					ResponseWriter: w,
					statusCode:     http.StatusOK,
				}

				next.ServeHTTP(recorder, req)

				// If handler already wrote response, return empty DataResponse
				if recorder.written {
					return dr.DataResponse{statusCode: recorder.statusCode}
				}

				// Otherwise, return success (shouldn't happen in normal flow)
				return factory.Success(req.Context(), nil)
			})

			// Execute DataResponse middleware
			resp := drMiddleware.ServeHTTP(r, handler)

			// Write the response if not already written
			var written bool
			if ww, ok := w.(interface{ Written() bool }); ok {
				written = ww.Written()
			}

			if !written {
				dr.Write(w, r, resp, factory)
			}
		})
	}
}

// FromChiMiddleware converts chi middleware to DataResponse Middleware.
func FromChiMiddleware(chiMiddleware ChiMiddleware) dr.Middleware {
	return dr.MiddlewareFunc(func(r *http.Request, next dr.Handler) dr.DataResponse {
		// This is complex because we need to bridge between two different paradigms
		// For now, just pass through to next handler
		return next.Handle(r)
	})
}

// ChiAdapter provides helpers for working with chi router.
type ChiAdapter struct {
	factory *dr.Factory
	adapter *dr.Adapter
}

// NewChiAdapter creates a new chi adapter.
func NewChiAdapter(factory *dr.Factory) *ChiAdapter {
	return &ChiAdapter{
		factory: factory,
		adapter: dr.NewAdapter(factory),
	}
}

// Handler converts DataResponse Handler to http.Handler for use with chi.
func (ca *ChiAdapter) Handler(handler dr.Handler) http.Handler {
	return ca.adapter.Handler(handler)
}

// Middleware converts DataResponse Middleware to chi middleware.
func (ca *ChiAdapter) Middleware(drMiddleware dr.Middleware) ChiMiddleware {
	return ToChiMiddleware(ca.factory, drMiddleware)
}

// HandlerFunc is a shortcut for creating chi-compatible handlers.
func (ca *ChiAdapter) HandlerFunc(fn func(*http.Request, Factory) dr.DataResponse) http.HandlerFunc {
	return ca.adapter.Handler(HandlerFunc(fn)).ServeHTTP
}

// ChiChain chains DataResponse middlewares into chi-compatible middleware.
func ChiChain(factory *dr.Factory, middlewares ...dr.Middleware) ChiMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Convert http.Handler to DataResponse Handler
			handler := dr.HandlerFunc(func(req *http.Request) dr.DataResponse {
				recorder := &responseRecorder{ResponseWriter: w}
				next.ServeHTTP(recorder, req)
				return dr.DataResponse{statusCode: recorder.statusCode}
			})

			// Chain DataResponse middlewares
			chainedHandler := dr.Chain(handler, middlewares...)

			// Execute and write response
			resp := chainedHandler.Handle(r)
			dr.Write(w, r, resp, factory)
		})
	}
}
