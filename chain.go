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

	"github.com/raoptimus/data-response.go/v2/response"
)

// ChainHandler wraps a middleware and next handler.
type ChainHandler struct {
	middleware Middleware
	next       Handler
}

// Handle executes the middleware.
func (c *ChainHandler) Handle(r *http.Request, f *Factory) *response.DataResponse {
	return c.middleware(c.next).Handle(r, f)
}

// Chain creates a handler chain with multiple middlewares.
//
//nolint:ireturn,nolintlint // its ok
func Chain(h Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = &ChainHandler{
			middleware: middlewares[i],
			next:       h,
		}
	}

	return h
}
