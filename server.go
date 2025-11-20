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
	chained := Chain(handler, s.middlewares...)
	s.ServeMux.Handle(pattern, WrapHandler(chained, s.factory))
}

func (s *ServeMux) HandleFunc(pattern string, handler HandlerFunc) {
	chained := Chain(handler, s.middlewares...)
	s.ServeMux.HandleFunc(pattern, WrapHandlerFunc(chained.Handle, s.factory))
}

func (s *ServeMux) WithMiddleware(m ...Middleware) *ServeMux {
	s.middlewares = append(s.middlewares, m...)

	return s
}
