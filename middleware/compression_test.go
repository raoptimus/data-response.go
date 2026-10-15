/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package middleware_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	dr "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/formatter"
	"github.com/raoptimus/data-response.go/v2/middleware"
	"github.com/raoptimus/data-response.go/v2/response"
)

func TestCompression_StreamResponse_IsNotCompressed_Successfully(t *testing.T) {
	body := strings.Repeat("0123456789abcdef", 1024)

	tests := []struct {
		name string
		mw   dr.Middleware
	}{
		{name: "default settings", mw: middleware.DefaultCompression()},
		{
			name: "negative minimum size",
			mw: middleware.Compression(middleware.CompressionOptions{
				MinSize: -1,
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			factory := dr.New(dr.WithFormatter(formatter.NewJSON()))
			handler := dr.Chain(
				dr.HandlerFunc(func(r *http.Request, f *dr.Factory) *response.DataResponse {
					return f.Stream(r.Context(), response.ContentTypeJSON, func(_ context.Context, w io.Writer) error {
						calls++
						_, err := io.WriteString(w, body)

						return err
					})
				}),
				tt.mw,
			)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(response.HeaderAcceptEncoding, "gzip, br")
			rec := httptest.NewRecorder()

			dr.WrapHandler(handler, factory).ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, body, rec.Body.String())
			assert.Empty(t, rec.Header().Get(response.HeaderContentEncoding))
			assert.Equal(t, 1, calls)
		})
	}
}

func TestCompression_PlainResponse_IsCompressed_Successfully(t *testing.T) {
	factory := dr.New(dr.WithFormatter(formatter.NewJSON()))
	handler := dr.Chain(
		dr.HandlerFunc(func(r *http.Request, f *dr.Factory) *response.DataResponse {
			return f.Success(r.Context(), strings.Repeat("0123456789abcdef", 1024))
		}),
		middleware.DefaultCompression(),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(response.HeaderAcceptEncoding, "gzip")
	rec := httptest.NewRecorder()

	dr.WrapHandler(handler, factory).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get(response.HeaderContentEncoding))
}
