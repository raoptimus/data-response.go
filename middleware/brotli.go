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

func BrotliCompressor(quality int) dr.Middleware {
	if quality < 0 {
		quality = 0
	}
	if quality > 11 {
		quality = 11
	}

	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
			if !supportsBrotli(r) {
				return next.Handle(r, f)
			}

			// Note: This is a placeholder. For actual Brotli support, use:
			// github.com/andybalholm/brotli
			return next.Handle(r, f)
		})
	}
}

func supportsBrotli(r *http.Request) bool {
	acceptEncoding := r.Header.Get("Accept-Encoding")

	return strings.Contains(acceptEncoding, "br")
}
