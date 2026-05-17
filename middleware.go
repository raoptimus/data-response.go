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
	"net/http"

	"github.com/raoptimus/data-response.go/v2/response"
)

// WrapMiddleware converts std middleware to DataResponse middleware.
// The chi middleware will be executed, but DataResponse will be returned from handler.
func WrapMiddleware(stdM func(http.Handler) http.Handler) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(r *http.Request, f *Factory) *response.DataResponse {
			var handlerResp *response.DataResponse
			captured := &capturedResponse{}

			stdM(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				handlerResp = next.Handle(r, f)
			})).ServeHTTP(captured, r)

			resp := finalResponse(r.Context(), f, handlerResp, captured.statusCode)
			mergeNonBodyHeaders(resp.Header(), captured.header)

			return resp
		})
	}
}

// finalResponse picks the DataResponse to return based on whether the chi
// middleware passed through to the inner handler (handlerResp != nil) or
// short-circuited with its own status code.
func finalResponse(
	ctx context.Context,
	f *Factory,
	handlerResp *response.DataResponse,
	capturedStatus int,
) *response.DataResponse {
	if handlerResp != nil {
		// Pass-through middleware that explicitly set a status before calling
		// next — honor it over whatever the inner handler chose.
		if capturedStatus != 0 {
			return handlerResp.WithStatusCode(capturedStatus)
		}

		return handlerResp
	}

	if capturedStatus == http.StatusOK {
		return f.Success(ctx, nil)
	}

	if capturedStatus == 0 {
		capturedStatus = http.StatusInternalServerError
	}

	return f.Error(ctx, capturedStatus, http.StatusText(capturedStatus))
}

// mergeNonBodyHeaders copies headers the chi middleware explicitly set into the
// final response, skipping body-content headers. Those describe the captured
// body that we discard; the outer dr.Write computes its own values from the
// actual body, so propagating the captured ones would duplicate (and conflict
// with) the real headers and strict upstreams reject the response with 502.
func mergeNonBodyHeaders(target, captured http.Header) {
	for key, values := range captured {
		switch http.CanonicalHeaderKey(key) {
		case response.HeaderContentType,
			response.HeaderContentLength,
			response.HeaderContentEncoding,
			response.HeaderTransferEncoding:
			continue
		}
		for _, value := range values {
			target.Add(key, value)
		}
	}
}

// capturedResponse is a minimal http.ResponseWriter that records the status
// code and headers set by the wrapped chi middleware. The body is discarded —
// the final response is always written by the outer dr.Write from a fresh
// DataResponse.
type capturedResponse struct {
	header     http.Header
	statusCode int
}

func (c *capturedResponse) Header() http.Header {
	if c.header == nil {
		c.header = make(http.Header)
	}

	return c.header
}

func (c *capturedResponse) Write(b []byte) (int, error) {
	return len(b), nil
}

func (c *capturedResponse) WriteHeader(statusCode int) {
	c.statusCode = statusCode
}
