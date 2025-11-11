package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// CompressionLevel определяет уровень сжатия.
type CompressionLevel int

const (
	// DefaultCompression использует уровень по умолчанию.
	DefaultCompression CompressionLevel = iota
	// BestSpeed приоритизирует скорость.
	BestSpeed
	// BestCompression приоритизирует размер.
	BestCompression
)

// Compressor middleware for response compression.
type Compressor struct {
	level CompressionLevel
}

// NewCompressor creates a new compression middleware.
func NewCompressor(level CompressionLevel) *Compressor {
	return &Compressor{level: level}
}

// Handler returns the middleware handler.
func (c *Compressor) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client supports gzip
		if !c.supportsEncoding(r, "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Wrap response writer with gzip writer
		gz := c.newGzipWriter(w)
		defer gz.Close()

		// Create wrapped response writer
		wrapped := &gzipResponseWriter{
			ResponseWriter: w,
			gzipWriter:     gz,
		}

		wrapped.Header().Set("Content-Encoding", "gzip")
		wrapped.Header().Del("Content-Length")

		next.ServeHTTP(wrapped, r)
	})
}

func (c *Compressor) supportsEncoding(r *http.Request, encoding string) bool {
	acceptEncoding := r.Header.Get("Accept-Encoding")
	return strings.Contains(acceptEncoding, encoding)
}

func (c *Compressor) newGzipWriter(w http.ResponseWriter) *gzip.Writer {
	gzipLevel := gzip.DefaultCompression

	switch c.level {
	case BestSpeed:
		gzipLevel = gzip.BestSpeed
	case BestCompression:
		gzipLevel = gzip.BestCompression
	}

	gz, _ := gzip.NewWriterLevel(w, gzipLevel)
	return gz
}

// gzipResponseWriter wraps http.ResponseWriter with gzip compression.
type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
}

// Write writes compressed data.
func (grw *gzipResponseWriter) Write(b []byte) (int, error) {
	return grw.gzipWriter.Write(b)
}

// Close closes the gzip writer.
func (grw *gzipResponseWriter) Close() error {
	return grw.gzipWriter.Close()
}

// Flush flushes the gzip writer.
func (grw *gzipResponseWriter) Flush() error {
	return grw.gzipWriter.Flush()
}
