/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package dataresponse

import (
	"context"
	"errors"
	"fmt"
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
//
// A stream response that fails after its first write ends with panic(http.ErrAbortHandler),
// so recovery middleware placed around the returned handler must let that value through.
func WrapHandler(h Handler, f *Factory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseRecorder{ResponseWriter: w}
		ctx := response.WithRequestStartTime(r.Context())
		resp := h.Handle(r.WithContext(ctx), f)

		err := Write(rw, resp)

		var streamErr *response.StreamError
		if errors.As(err, &streamErr) {
			f.handleStreamError(ctx, w, rw, streamErr)

			return
		}

		if err != nil {
			if rw.Written() { // already written
				f.logger.Error(ctx, "failed to write response", "error", err.Error())
				return
			}

			f.writeErrorResponse(ctx, w, rw, f.InternalError(ctx, err))
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

// writeErrorResponse writes errResp, and a plain 500 if even that fails, so the client never gets an empty 200.
func (f *Factory) writeErrorResponse(
	ctx context.Context,
	w http.ResponseWriter,
	rw *responseRecorder,
	errResp *response.DataResponse,
) {
	err := Write(rw, errResp)
	if err == nil {
		return
	}

	f.logger.Error(ctx, "failed to write error response", "error", err.Error())

	// last chance
	if !rw.Written() {
		w.Header().Set(response.HeaderContentType, response.MimeTypePlainText.String())
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}
}

func (f *Factory) handleStreamError(
	ctx context.Context,
	w http.ResponseWriter,
	rw *responseRecorder,
	streamErr *response.StreamError,
) {
	f.observeStreamError(ctx, streamErr)

	if streamErr.Phase == response.StreamPhaseAfterWrite {
		// The client already has the status line: abort the connection instead of finishing
		// a chunked body that would look complete.
		panic(http.ErrAbortHandler)
	}

	var errResp *response.DataResponse
	if f.streamResponder != nil {
		errResp = f.streamResponder(ctx, streamErr.Err)
	}

	if errResp == nil {
		errResp = f.InternalError(ctx, streamErr.Err)
	}

	f.writeErrorResponse(ctx, w, rw, errResp)
}

// observeStreamError calls the observer; its panic is logged and dropped so it cannot cancel the abort.
func (f *Factory) observeStreamError(ctx context.Context, streamErr *response.StreamError) {
	if f.streamObserver == nil {
		// The default responder logs failures before the first write itself.
		if streamErr.Phase == response.StreamPhaseAfterWrite {
			f.logger.Error(ctx, "stream aborted", "error", streamErr.Error())
		}

		return
	}

	defer func() {
		if rcv := recover(); rcv != nil {
			f.logger.Error(ctx, "stream observer panicked", "panic", fmt.Sprint(rcv))
		}
	}()

	f.streamObserver(ctx, streamErr.Phase, streamErr)
}
