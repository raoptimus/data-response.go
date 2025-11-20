/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package middleware

import (
	"fmt"
	"net/http"

	dr "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/response"
)

func Recovery() dr.Middleware {
	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) *response.DataResponse {
			var resp *response.DataResponse
			defer func() {
				if err := recover(); err != nil {
					ctx := r.Context()
					panicErr := response.NewError(http.StatusInternalServerError, fmt.Sprintf("panic: %v", err))
					resp = f.InternalError(ctx, panicErr)
				}
			}()

			resp = next.Handle(r, f)

			return resp
		})
	}
}
