/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package chiadapter

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	dr "github.com/raoptimus/data-response.go/v2"
)

// Router wraps chi.Router with DataResponse support.
type Router struct {
	chi.Router
	factory     *dr.Factory
	middlewares []dr.Middleware
}

// NewRouter creates a new chi Router with DataResponse support.
func NewRouter(factory *dr.Factory) *Router {
	return &Router{
		Router:      chi.NewRouter(),
		factory:     factory,
		middlewares: make([]dr.Middleware, 0),
	}
}

// WithMiddleware adds DataResponse middleware.
// These middleware will be applied to all subsequently registered handlers.
func (r *Router) WithMiddleware(middlewares ...dr.Middleware) *Router {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

// Handle registers a DataResponse handler for the given pattern and method.
func (r *Router) Handle(method, pattern string, handler dr.Handler) {
	chained := dr.Chain(handler, r.middlewares...)
	r.Router.Method(method, pattern, dr.WrapHandler(chained, r.factory))
}

// HandleFunc registers a DataResponse handler function.
func (r *Router) HandleFunc(method, pattern string, handlerFunc dr.HandlerFunc) {
	r.Handle(method, pattern, handlerFunc)
}

// Get registers a GET handler.
func (r *Router) Get(pattern string, handlerFunc dr.HandlerFunc) {
	r.HandleFunc("GET", pattern, handlerFunc)
}

// Post registers a POST handler.
func (r *Router) Post(pattern string, handlerFunc dr.HandlerFunc) {
	r.HandleFunc("POST", pattern, handlerFunc)
}

// Put registers a PUT handler.
func (r *Router) Put(pattern string, handlerFunc dr.HandlerFunc) {
	r.HandleFunc("PUT", pattern, handlerFunc)
}

// Patch registers a PATCH handler.
func (r *Router) Patch(pattern string, handlerFunc dr.HandlerFunc) {
	r.HandleFunc("PATCH", pattern, handlerFunc)
}

// Delete registers a DELETE handler.
func (r *Router) Delete(pattern string, handlerFunc dr.HandlerFunc) {
	r.HandleFunc("DELETE", pattern, handlerFunc)
}

// Options registers an OPTIONS handler.
func (r *Router) Options(pattern string, handlerFunc dr.HandlerFunc) {
	r.HandleFunc("OPTIONS", pattern, handlerFunc)
}

// Head registers a HEAD handler.
func (r *Router) Head(pattern string, handlerFunc dr.HandlerFunc) {
	r.HandleFunc("HEAD", pattern, handlerFunc)
}

// Group creates a new router group with optional middleware.
func (r *Router) Group(fn func(r *Router)) *Router {
	// Create new router with same factory and copy of middlewares
	group := &Router{
		Router:      chi.NewRouter(),
		factory:     r.factory,
		middlewares: append([]dr.Middleware{}, r.middlewares...),
	}

	if fn != nil {
		fn(group)
	}

	return group
}

// Route mounts a sub-router along a routing path.
func (r *Router) Route(pattern string, fn func(r *Router)) {
	subRouter := r.Group(fn)
	r.Router.Mount(pattern, subRouter)
}

// Use adds standard chi middleware (not DataResponse middleware).
// For DataResponse middleware, use WithMiddleware.
func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	r.Router.Use(middlewares...)
}

// Mount attaches another Router as a sub-router.
func (r *Router) Mount(pattern string, handler http.Handler) {
	r.Router.Mount(pattern, handler)
}

// Factory returns the Factory associated with this router.
func (r *Router) Factory() *dr.Factory {
	return r.factory
}
