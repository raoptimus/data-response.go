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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raoptimus/data-response.go/v2/formatter"
	"github.com/raoptimus/data-response.go/v2/response"
)

// TestWrapMiddleware_PassThrough_PreservesHandlerStatusCode covers a regression
// where a pass-through chi middleware (one that calls next.ServeHTTP without
// writing its own status) caused the handler's status code to be overwritten
// with 0, producing "invalid WriteHeader code 0" and an empty response to the
// client.
func TestWrapMiddleware_PassThrough_PreservesHandlerStatusCode(t *testing.T) {
	factory := New(WithFormatter(formatter.NewJSON()))

	passThroughMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "applied")
			next.ServeHTTP(w, r)
		})
	}

	handler := HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
		return f.Success(r.Context(), map[string]string{"result": "ok"})
	})

	wrapped := WrapMiddleware(passThroughMW)(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	httpHandler := WrapHandler(wrapped, factory)
	httpHandler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "handler status 200 must be preserved")
	assert.Equal(t, "applied", rec.Header().Get("X-Test"), "middleware header must propagate")
}

// TestWrapMiddleware_ShortCircuit_UsesMiddlewareStatusCode verifies that when a
// chi middleware writes its own status and does not call next, that status code
// is delivered to the client.
func TestWrapMiddleware_ShortCircuit_UsesMiddlewareStatusCode(t *testing.T) {
	factory := New(WithFormatter(formatter.NewJSON()))

	shortCircuitMW := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		})
	}

	handler := HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
		return f.Success(r.Context(), nil)
	})

	wrapped := WrapMiddleware(shortCircuitMW)(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	httpHandler := WrapHandler(wrapped, factory)
	httpHandler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestWrapMiddleware_MiddlewareSetsStatus_OverridesHandler verifies that when a
// chi middleware explicitly sets a status via WriteHeader, that status overrides
// the one returned by the handler (the library's original behavior).
func TestWrapMiddleware_MiddlewareSetsStatus_OverridesHandler(t *testing.T) {
	factory := New(WithFormatter(formatter.NewJSON()))

	statusOverrideMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			next.ServeHTTP(w, r)
		})
	}

	handler := HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
		return f.Success(r.Context(), nil)
	})

	wrapped := WrapMiddleware(statusOverrideMW)(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	httpHandler := WrapHandler(wrapped, factory)
	httpHandler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusAccepted, rec.Code)
}
