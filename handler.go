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

// Handler defines a handler that returns DataResponse instead of writing to http.ResponseWriter.
// This allows for more functional and composable middleware.
type Handler interface {
	Handle(r *http.Request, f *Factory) *response.DataResponse
}

// HandlerFunc is a function adapter for Handler interface.
type HandlerFunc func(r *http.Request, f *Factory) *response.DataResponse

// Handle calls f(r).
func (hf HandlerFunc) Handle(r *http.Request, f *Factory) *response.DataResponse {
	return hf(r, f)
}

// WrapHandler converts DataResponse Handler to http.Handler.
func WrapHandler(h Handler, f *Factory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseRecorder{ResponseWriter: w}
		ctx := response.WithRequestStartTime(r.Context())
		resp := h.Handle(r.WithContext(ctx), f)

		if err := Write(rw, resp); err != nil {
			if rw.Written() { // already written
				f.logger.Error(ctx, "failed to write response", "error", err.Error())
				return
			}

			errResp := f.InternalError(ctx, err)
			if err := Write(rw, errResp); err != nil {
				f.logger.Error(ctx, "failed to write error response", "error", err.Error())

				// last chance
				if !rw.Written() {
					w.Header().Set(response.HeaderContentType, response.MimeTypePlainText.String())
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte("Internal Server Error"))
				}
			}
		}
	})
}

// WrapHandlerFunc converts DataResponse HandlerFunc to http.HandlerFunc.
func WrapHandlerFunc(hf HandlerFunc, f *Factory) http.HandlerFunc {
	return WrapHandler(hf, f).ServeHTTP
}

// responseRecorder captures response data.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader captures the status code.
func (rr *responseRecorder) WriteHeader(statusCode int) {
	if rr.written {
		return
	}

	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
	rr.written = true
}

// Write marks response as written.
func (rr *responseRecorder) Write(b []byte) (int, error) {
	if !rr.written {
		if rr.statusCode == 0 {
			rr.statusCode = http.StatusOK
		}
		rr.WriteHeader(rr.statusCode)
	}

	return rr.ResponseWriter.Write(b)
}

// Written returns true if response was written.
func (rr *responseRecorder) Written() bool {
	return rr.written
}

func (rr *responseRecorder) StatusCode() int {
	return rr.statusCode
}
