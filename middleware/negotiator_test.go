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
	"testing"

	"github.com/stretchr/testify/assert"

	dr "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/formatter"
	"github.com/raoptimus/data-response.go/v2/middleware"
	"github.com/raoptimus/data-response.go/v2/response"
)

func TestContentNegotiator_StreamResponse_KeepsContentType_Successfully(t *testing.T) {
	const (
		streamContentType = "application/x-ndjson"
		streamBody        = "{\"id\":1}\n{\"id\":2}\n"
	)

	tests := []struct {
		name   string
		accept string
	}{
		{name: "accept of another formatter", accept: response.ContentTypeXML},
		{name: "accept of the default formatter", accept: response.ContentTypeJSON},
		{name: "no accept", accept: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := dr.New(dr.WithFormatter(formatter.NewJSON()))
			handler := dr.Chain(
				dr.HandlerFunc(func(r *http.Request, f *dr.Factory) *response.DataResponse {
					return f.Stream(r.Context(), streamContentType, func(_ context.Context, w io.Writer) error {
						_, err := io.WriteString(w, streamBody)

						return err
					})
				}),
				middleware.ContentNegotiator(map[string]response.Formatter{
					response.ContentTypeJSON: formatter.NewJSON(),
					response.ContentTypeXML:  formatter.NewXML(),
				}),
			)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			rec := httptest.NewRecorder()

			dr.WrapHandler(handler, factory).ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, streamContentType, rec.Header().Get(response.HeaderContentType))
			assert.Equal(t, streamBody, rec.Body.String())
		})
	}
}

func TestContentNegotiator_PlainResponse_UsesAcceptedFormatter_Successfully(t *testing.T) {
	factory := dr.New(dr.WithFormatter(formatter.NewJSON()))
	handler := dr.Chain(
		dr.HandlerFunc(func(r *http.Request, f *dr.Factory) *response.DataResponse {
			return f.Success(r.Context(), "ok")
		}),
		middleware.ContentNegotiator(map[string]response.Formatter{
			response.ContentTypeJSON: formatter.NewJSON(),
			response.ContentTypeXML:  formatter.NewXML(),
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", response.ContentTypeXML)
	rec := httptest.NewRecorder()

	dr.WrapHandler(handler, factory).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, response.ContentTypeXML, rec.Header().Get(response.HeaderContentType))
}
