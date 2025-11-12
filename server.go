package dataresponse

import "net/http"

// Handler defines a handler that returns DataResponse instead of writing to http.ResponseWriter.
// This allows for more functional and composable middleware.
type Handler interface {
	Handle(r *http.Request) DataResponse
}

// HandlerFunc is a function adapter for Handler interface.
type HandlerFunc func(r *http.Request) DataResponse

// Handle calls f(r).
func (f HandlerFunc) Handle(r *http.Request) DataResponse {
	return f(r)
}

// Middleware defines a middleware that works with DataResponse.
type Middleware interface {
	ServeHTTP(r *http.Request, next Handler) DataResponse
}

// MiddlewareFunc is a function adapter for Middleware interface.
type MiddlewareFunc func(r *http.Request, next Handler) DataResponse

// ServeHTTP calls f(r, next).
func (f MiddlewareFunc) ServeHTTP(r *http.Request, next Handler) DataResponse {
	return f(r, next)
}

// Adapter adapts DataResponse-based handlers to standard http.Handler.
type Adapter struct {
	factory *Factory
}

// NewAdapter creates a new adapter with the given factory.
func NewAdapter(factory *Factory) *Adapter {
	return &Adapter{factory: factory}
}

// Handler converts DataResponse Handler to http.Handler.
func (a *Adapter) Handler(next Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := next.Handle(r)
		if err := Write(w, r, resp, a.factory); err != nil {
			resp = a.factory.InternalError(r.Context(), err)
			// todo: что если statusCode уже был записан. второй раз нельзя. добавить flag с проверкой
			if err := Write(w, r, resp, a.factory); err != nil {
				a.factory.logger.Error(r.Context(), "failed to write response", "error", err.Error())
			}
		}
	})
}

// Middleware wraps Handler with middleware chain.
func (a *Adapter) Middleware(handler Handler, middlewares ...Middleware) Handler {
	// Build middleware chain from right to left
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = &middlewareHandler{
			middleware: middlewares[i],
			next:       handler,
		}
	}
	return handler
}

// middlewareHandler wraps a middleware and next handler.
type middlewareHandler struct {
	middleware Middleware
	next       Handler
}

// Handle executes the middleware.
func (mh *middlewareHandler) Handle(r *http.Request) DataResponse {
	return mh.middleware.ServeHTTP(r, mh.next)
}

// Chain creates a handler chain with multiple middlewares.
func Chain(handler Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = &middlewareHandler{
			middleware: middlewares[i],
			next:       handler,
		}
	}
	return handler
}
