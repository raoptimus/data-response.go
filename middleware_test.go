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

// TestWrapMiddleware_PassThrough_NoDuplicateContentType covers a regression
// where the captured intermediate DataResponse was pre-populated with a
// formatter (which set its own Content-Type), and the resulting WithHeaders
// merge produced two identical Content-Type values on every response.
func TestWrapMiddleware_PassThrough_NoDuplicateContentType(t *testing.T) {
	factory := New(WithFormatter(formatter.NewJSON()))

	passThroughMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	handler := HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
		return f.Success(r.Context(), map[string]string{"ok": "yes"})
	})

	wrapped := WrapMiddleware(passThroughMW)(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	WrapHandler(wrapped, factory).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, rec.Header().Values(response.HeaderContentType), 1,
		"Content-Type must not be duplicated")
	require.Len(t, rec.Header().Values(response.HeaderContentLength), 1,
		"Content-Length must not be duplicated")
}

// TestWrapMiddleware_HandlerError_NoDuplicateContentType verifies the same
// invariant for an error response returned by the inner handler — a strict
// upstream like Gravitee rejects duplicate body-content headers with 502.
func TestWrapMiddleware_HandlerError_NoDuplicateContentType(t *testing.T) {
	factory := New(WithFormatter(formatter.NewJSON()))

	passThroughMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	handler := HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
		return f.Error(r.Context(), http.StatusBadRequest, "bad")
	})

	wrapped := WrapMiddleware(passThroughMW)(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	WrapHandler(wrapped, factory).ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Len(t, rec.Header().Values(response.HeaderContentType), 1,
		"Content-Type must not be duplicated for error responses")
	require.Len(t, rec.Header().Values(response.HeaderContentLength), 1,
		"Content-Length must not be duplicated for error responses")
}

// TestWrapMiddleware_ShortCircuitViaWrapHandlerFunc_NoDuplicateContentLength
// covers the bug that surfaced with strict upstreams (Gravitee, nginx with
// strict mode) responding 502 to clients. When a chi middleware short-circuits
// using dr.WrapHandlerFunc, the inner dr.Write writes its own Content-Length
// into the captured writer's header. Previously WithHeaders carried that stale
// value into the fallback f.Error response, and the outer dr.Write added its
// own — producing two different Content-Length values in one HTTP response.
func TestWrapMiddleware_ShortCircuitViaWrapHandlerFunc_NoDuplicateContentLength(t *testing.T) {
	factory := New(WithFormatter(formatter.NewJSON()))

	shortCircuitMW := func(_ http.Handler) http.Handler {
		errHandler := WrapHandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
			return f.Error(r.Context(), http.StatusForbidden, "denied")
		}, factory)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			errHandler.ServeHTTP(w, r)
		})
	}

	handler := HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
		return f.Success(r.Context(), map[string]string{"unreachable": "true"})
	})

	wrapped := WrapMiddleware(shortCircuitMW)(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	WrapHandler(wrapped, factory).ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Len(t, rec.Header().Values(response.HeaderContentLength), 1,
		"Content-Length must not be duplicated when chi middleware short-circuits via WrapHandlerFunc")
	require.Len(t, rec.Header().Values(response.HeaderContentType), 1,
		"Content-Type must not be duplicated when chi middleware short-circuits via WrapHandlerFunc")
}

// TestWrapMiddleware_CustomHeadersFromMiddleware_StillPropagate ensures the
// header-filtering fix does not strip legitimate headers the chi middleware
// wants to add (the original purpose of WithHeaders).
func TestWrapMiddleware_CustomHeadersFromMiddleware_StillPropagate(t *testing.T) {
	factory := New(WithFormatter(formatter.NewJSON()))

	customMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Request-ID", "req-123")
			w.Header().Set("X-Trace-Id", "trace-abc")
			next.ServeHTTP(w, r)
		})
	}

	handler := HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
		return f.Success(r.Context(), map[string]string{"ok": "yes"})
	})

	wrapped := WrapMiddleware(customMW)(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	WrapHandler(wrapped, factory).ServeHTTP(rec, req)

	assert.Equal(t, "req-123", rec.Header().Get("X-Request-Id"))
	assert.Equal(t, "trace-abc", rec.Header().Get("X-Trace-Id"))
}
