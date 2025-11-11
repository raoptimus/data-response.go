package middleware

import (
	"net/http"
	"strings"
)

// AutoCompressor automatically selects the best compression algorithm.
type AutoCompressor struct {
	gzipLevel CompressionLevel
}

// NewAutoCompressor creates a new auto compression middleware.
func NewAutoCompressor(gzipLevel CompressionLevel) *AutoCompressor {
	return &AutoCompressor{gzipLevel: gzipLevel}
}

// Handler returns the middleware handler.
func (ac *AutoCompressor) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptEncoding := r.Header.Get("Accept-Encoding")

		// Prefer Brotli if supported
		if strings.Contains(acceptEncoding, "br") {
			// Use Brotli
			brotli := NewBrotliCompressor(6)
			brotli.Handler(next).ServeHTTP(w, r)
			return
		}

		// Fall back to Gzip
		if strings.Contains(acceptEncoding, "gzip") {
			gzip := NewCompressor(ac.gzipLevel)
			gzip.Handler(next).ServeHTTP(w, r)
			return
		}

		// No compression
		next.ServeHTTP(w, r)
	})
}
