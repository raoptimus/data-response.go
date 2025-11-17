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

	dr "github.com/raoptimus/data-response.go/v2"
)

type capturedResponse struct {
	capturedResp dr.DataResponse
}

func (wr *capturedResponse) Header() http.Header {
	return wr.capturedResp.Header()
}

func (wr *capturedResponse) Write(b []byte) (int, error) {
	// Silently ignore writes - body will be managed by DataResponse
	return len(b), nil
}

func (wr *capturedResponse) WriteHeader(statusCode int) {
	wr.capturedResp = wr.capturedResp.WithStatusCode(statusCode)
}

// WrapChiMiddleware converts chi middleware to DataResponse middleware.
// The chi middleware will be executed, but DataResponse will be returned from handler.
func WrapChiMiddleware(chiMiddleware func(http.Handler) http.Handler) dr.Middleware {
	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
			var captured bool
			capturedWriter := &capturedResponse{
				capturedResp: dr.DataResponse{}, // Empty, status = 0
			}

			// Create a dummy handler that captures the response
			dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedResp := next.Handle(r, f).
					WithStatusCode(capturedWriter.capturedResp.StatusCode()).
					WithHeaders(capturedWriter.capturedResp.Header())
				capturedWriter.capturedResp = capturedResp
				captured = true
			})

			// Execute chi middleware
			chiMiddleware(dummyHandler).ServeHTTP(capturedWriter, r)

			if !captured {
				statusCode := capturedWriter.capturedResp.StatusCode()
				if statusCode == http.StatusOK {
					return f.Success(r.Context(), nil).
						WithHeaders(capturedWriter.capturedResp.Header())
				}
				if statusCode == 0 {
					statusCode = http.StatusInternalServerError
				}

				return f.Error(r.Context(), statusCode, http.StatusText(statusCode)).
					WithHeaders(capturedWriter.capturedResp.Header())
			}

			return capturedWriter.capturedResp
		})
	}
}
