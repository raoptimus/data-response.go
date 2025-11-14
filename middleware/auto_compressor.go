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

// AutoCompressor returns auto compression middleware.
// AutoCompressor automatically selects the best compression algorithm.
func AutoCompressor() dr.Middleware {
	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
			acceptEncoding := r.Header.Get("Accept-Encoding")
			// Prefer Brotli if supported
			if strings.Contains(acceptEncoding, "br") {
				return BrotliCompressor(6)(next).Handle(r, f)
			}

			// Fall back to Gzip
			// if strings.Contains(acceptEncoding, "gzip") {
			//	 return Compressor(BestCompression)(next).Handle(r, f)
			// }

			return next.Handle(r, f)
		})
	}
}

