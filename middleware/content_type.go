/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package middleware

import (
	"net/http"
	"strings"

	dr "github.com/raoptimus/data-response.go/v2"
)

func CheckContentType() dr.Middleware {
	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
			contentType := r.Header.Get("Content-Type")
			if i := strings.Index(contentType, ";"); i > -1 {
				contentType = contentType[0:i]
			}

			if contentType != f.Formatter().ContentType() {
				return f.Error(r.Context(), http.StatusUnsupportedMediaType, "Unsupported Media Type")
			}

			return next.Handle(r, f)
		})
	}
}
