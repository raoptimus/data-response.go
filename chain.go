package dataresponse

import "net/http"

// ChainHandler wraps a middleware and next handler.
type ChainHandler struct {
	middleware func(h Handler) Handler
	next       Handler
}

// Handle executes the middleware.
func (c *ChainHandler) Handle(r *http.Request, f *Factory) DataResponse {
	return c.middleware(c.next).Handle(r, f)
}

// Chain creates a handler chain with multiple middlewares.
func Chain(f *Factory, h Handler, middlewares ...func(h Handler) Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = &ChainHandler{
			middleware: middlewares[i],
			next:       h,
		}
	}

	return WrapHandler(h, f)
}
