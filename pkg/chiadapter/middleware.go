/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package chiadapter

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	dr "github.com/raoptimus/data-response.go/v2"
)

// contextKey is the key type for context values.
type contextKey string

const factoryKey contextKey = "dataresponse:factory"

// FactoryMiddleware injects Factory into request context.
// This allows chi middleware to access Factory if needed.
func FactoryMiddleware(factory *dr.Factory) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), factoryKey, factory)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FactoryFromContext retrieves Factory from request context.
func FactoryFromContext(ctx context.Context) (*dr.Factory, bool) {
	factory, ok := ctx.Value(factoryKey).(*dr.Factory)
	return factory, ok
}

// WrapChiMiddleware converts chi middleware to DataResponse middleware.
// The chi middleware will be executed, but DataResponse will be returned from handler.
func WrapChiMiddleware(chiMiddleware func(http.Handler) http.Handler) dr.Middleware {
	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
			// Create a dummy handler that captures the response
			var capturedResp dr.DataResponse

			dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedResp = next.Handle(r, f)
			})

			// Execute chi middleware
			chiMiddleware(dummyHandler).ServeHTTP(nil, r)

			return capturedResp
		})
	}
}

// URLParam returns URL parameter from chi router.
func URLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

// URLParamInt returns URL parameter as int.
func URLParamInt(r *http.Request, key string) (int, error) {
	param := chi.URLParam(r, key)
	var i int
	_, err := fmt.Sscanf(param, "%d", &i)
	return i, err
}

// AllURLParams returns all URL parameters.
func AllURLParams(r *http.Request) map[string]string {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return make(map[string]string)
	}

	params := make(map[string]string)
	for i, key := range rctx.URLParams.Keys {
		params[key] = rctx.URLParams.Values[i]
	}
	return params
}
