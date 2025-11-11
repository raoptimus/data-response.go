package middleware

import (
	"net/http"
	"strings"
)

// BrotliCompressor middleware for Brotli compression.
// Requires: github.com/andybalholm/brotli
type BrotliCompressor struct {
	quality int
}

// NewBrotliCompressor creates a new brotli compression middleware.
func NewBrotliCompressor(quality int) *BrotliCompressor {
	if quality < 0 {
		quality = 0
	}
	if quality > 11 {
		quality = 11
	}
	return &BrotliCompressor{quality: quality}
}

// Handler returns the middleware handler.
func (bc *BrotliCompressor) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !bc.supportsBrotli(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Note: This is a placeholder. For actual Brotli support, use:
		// github.com/andybalholm/brotli
		next.ServeHTTP(w, r)
	})
}

func (bc *BrotliCompressor) supportsBrotli(r *http.Request) bool {
	acceptEncoding := r.Header.Get("Accept-Encoding")
	return strings.Contains(acceptEncoding, "br")
}
