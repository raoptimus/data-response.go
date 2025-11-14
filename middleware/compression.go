/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package middleware

//
//// CompressionLevel определяет уровень сжатия.
//type CompressionLevel int
//
//const (
//	// DefaultCompression использует уровень по умолчанию.
//	DefaultCompression CompressionLevel = iota
//	// BestSpeed приоритизирует скорость.
//	BestSpeed
//	// BestCompression приоритизирует размер.
//	BestCompression
//)
//
//// GZIPCompressor middleware for response compression.
//func GZIPCompressor(level CompressionLevel) dr.Middleware {
//	return func(next dr.Handler) dr.Handler {
//		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
//			// Check if client supports gzip
//			if !supportsEncoding(r, "gzip") {
//				return next.Handle(r, f)
//			}
//
//			// Wrap response writer with gzip writer
//			gz := newGzipWriter(w, level)
//			defer gz.Close()
//
//			// Create wrapped response writer
//			wrapped := &gzipResponseWriter{
//				ResponseWriter: w,
//				gzipWriter:     gz,
//			}
//
//			resp := next.Handle(r, f).
//				WithHeader("Content-Encoding", "gzip").
//				WithoutHeader("Content-Length")
//			//next.ServeHTTP(wrapped, r)
//		})
//	}
//}
//
//func supportsEncoding(r *http.Request, encoding string) bool {
//	acceptEncoding := r.Header.Get("Accept-Encoding")
//	return strings.Contains(acceptEncoding, encoding)
//}
//
//func newGzipWriter(w http.ResponseWriter, level CompressionLevel) *gzip.Writer {
//	gzipLevel := gzip.DefaultCompression
//
//	switch level {
//	case BestSpeed:
//		gzipLevel = gzip.BestSpeed
//	case BestCompression:
//		gzipLevel = gzip.BestCompression
//	}
//
//	gz, _ := gzip.NewWriterLevel(w, gzipLevel)
//	return gz
//}
//
//// gzipResponseWriter wraps http.ResponseWriter with gzip compression.
//type gzipResponseWriter struct {
//	http.ResponseWriter
//	gzipWriter *gzip.Writer
//}
//
//// Write writes compressed data.
//func (grw *gzipResponseWriter) Write(b []byte) (int, error) {
//	return grw.gzipWriter.Write(b)
//}
//
//// Close closes the gzip writer.
//func (grw *gzipResponseWriter) Close() error {
//	return grw.gzipWriter.Close()
//}
//
//// Flush flushes the gzip writer.
//func (grw *gzipResponseWriter) Flush() error {
//	return grw.gzipWriter.Flush()
//}
