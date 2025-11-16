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
)

func Recovery() dr.Middleware {
	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
			var resp dr.DataResponse
			defer func() {
				if err := recover(); err != nil {
					ctx := r.Context()
					panicErr := dr.NewError(http.StatusInternalServerError, fmt.Sprintf("panic: %v", err))
					resp = f.InternalError(ctx, panicErr)
				}
			}()

			resp = next.Handle(r, f)

			return resp
		})
	}
}
