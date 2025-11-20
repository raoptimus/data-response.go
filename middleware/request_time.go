package middleware

import (
	"net/http"

	dr "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/response"
)

// RequestTimer middleware sets the request start time in context.
// This should be the first middleware in the chain to get accurate timing.
func RequestTimer() dr.Middleware {
	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) *response.DataResponse {
			// Store start time in context
			ctx := response.WithRequestStartTime(r.Context())

			return next.Handle(r.WithContext(ctx), f)
		})
	}
}
