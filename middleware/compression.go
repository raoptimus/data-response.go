/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package middleware

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	dr "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/response"
)

var ErrUnknownEncoding = errors.New("unknow encoding")

// CompressionLevel defines compression level.
type CompressionLevel int

const (
	// CompressionLevelDefault uses default compression.
	CompressionLevelDefault CompressionLevel = -1

	// CompressionLevelNone disables compression.
	CompressionLevelNone CompressionLevel = 0

	// CompressionLevelFastest uses fastest compression.
	CompressionLevelFastest CompressionLevel = 1

	// CompressionLevelOptimal uses optimal compression.
	CompressionLevelOptimal CompressionLevel = 6

	// CompressionLevelBest uses best compression.
	CompressionLevelBest CompressionLevel = 9
)

const (
	compressionMethodBR      = "br"
	compressionMethodDeflate = "deflate"
	compressionMethodGzip    = "gzip"
	defaultMinSizeToCompress = 1024 // 1 KB
)

// CompressionOptions configures compression middleware.
type CompressionOptions struct {
	// Level sets compression level (1-9, default is -1 for default compression).
	Level CompressionLevel

	// MinSize sets minimum response size in bytes to compress (default 1024).
	MinSize int64

	// ContentTypes lists MIME types to compress (empty = compress all).
	ContentTypes []string
}

// Compression creates a middleware that compresses response body.
// It supports gzip, deflate, and brotli based on Accept-Encoding header.
func Compression(opts CompressionOptions) dr.Middleware {
	if opts.MinSize == 0 {
		opts.MinSize = defaultMinSizeToCompress // Default 1KB
	}

	if opts.Level == 0 {
		opts.Level = CompressionLevelDefault
	}

	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) *response.DataResponse {
			// Get Accept-Encoding header
			acceptEncoding := r.Header.Get("Accept-Encoding")

			// Select encoding
			encoding := selectEncoding(acceptEncoding)
			if encoding == "" {
				// Client doesn't support compression
				return next.Handle(r, f)
			}

			// Execute handler
			resp := next.Handle(r, f)
			formattedResp, err := resp.Body()
			if err != nil {
				f.Logger().Error(r.Context(), "failed to get formatted response",
					"error", err,
				)

				return resp // Return original on error
			}
			resp = resp.WithFormatted(formattedResp) // Save ready formatted content

			// Check if content type should be compressed
			if !shouldCompress(resp.ContentType(), opts.ContentTypes) {
				return resp
			}

			// Check minimum size
			if formattedResp.StreamSize < opts.MinSize {
				f.Logger().Debug(r.Context(), "body too small to compress",
					"size", formattedResp.StreamSize,
					"min_size", opts.MinSize,
				)

				return resp // Too small to compress
			}

			// Compress the body
			compressedResp, err := compressBody(formattedResp, encoding, int(opts.Level))
			if err != nil {
				if errors.Is(err, ErrUnknownEncoding) {
					f.Logger().Debug(r.Context(), "failed to compress response",
						"error", err.Error(),
						"encoding", encoding,
					)
				} else {
					f.Logger().Warn(r.Context(), "failed to compress response",
						"error", err.Error(),
						"encoding", encoding,
					)
				}

				return resp // Return original on compression error
			}

			// Return compressed response with appropriate headers
			return resp.
				WithFormatted(compressedResp).
				WithHeader("Content-Encoding", encoding).
				WithHeader("Vary", "Accept-Encoding")
		})
	}
}

// selectEncoding selects the best encoding based on Accept-Encoding header.
func selectEncoding(acceptEncoding string) string {
	if acceptEncoding == "" {
		return ""
	}

	// Parse Accept-Encoding and select best supported encoding
	// Priority: brotli > gzip > deflate
	encodings := strings.Split(acceptEncoding, ",")

	supportBrotli := false
	supportGzip := false
	supportDeflate := false

	for _, enc := range encodings {
		enc = strings.TrimSpace(enc)

		// Remove quality factor (e.g., ";q=0.9")
		if idx := strings.Index(enc, ";"); idx > -1 {
			enc = enc[:idx]
		}

		enc = strings.TrimSpace(strings.ToLower(enc))

		switch enc {
		case compressionMethodBR:
			supportBrotli = true
		case compressionMethodGzip:
			supportGzip = true
		case compressionMethodDeflate:
			supportDeflate = true
		}

		if supportBrotli {
			break // because its priority
		}
	}

	// Select best encoding
	if supportBrotli {
		return compressionMethodBR
	}
	if supportGzip {
		return compressionMethodGzip
	}
	if supportDeflate {
		return compressionMethodDeflate
	}

	return ""
}

// shouldCompress checks if content type should be compressed.
func shouldCompress(contentType string, allowedTypes []string) bool {
	if len(allowedTypes) == 0 {
		// Compress all if no restrictions
		return true
	}

	// Extract media type without parameters
	if idx := strings.Index(contentType, ";"); idx > -1 {
		contentType = contentType[:idx]
	}
	contentType = strings.TrimSpace(strings.ToLower(contentType))

	// Check if content type is in allowed list
	for _, allowed := range allowedTypes {
		allowed = strings.TrimSpace(strings.ToLower(allowed))
		if contentType == allowed {
			return true
		}

		// Support wildcards like "text/*"
		if strings.HasSuffix(allowed, "/*") {
			prefix := strings.TrimSuffix(allowed, "/*")
			if strings.HasPrefix(contentType, prefix) {
				return true
			}
		}
	}

	return false
}

// compressBody compresses data using specified encoding.
func compressBody(body response.FormattedResponse, encoding string, level int) (response.FormattedResponse, error) {
	buf := new(bytes.Buffer)
	var writer io.WriteCloser
	var err error

	switch encoding {
	case compressionMethodBR:
		// Brotli compression
		if level < 0 {
			level = brotli.DefaultCompression
		}
		writer = brotli.NewWriterLevel(buf, level)

	case compressionMethodGzip:
		// Gzip compression
		if level < 0 {
			writer, err = gzip.NewWriterLevel(buf, gzip.DefaultCompression)
		} else {
			writer, err = gzip.NewWriterLevel(buf, level)
		}
		if err != nil {
			return response.FormattedResponse{}, err
		}

	case compressionMethodDeflate:
		// Deflate compression
		if level < 0 {
			writer, err = flate.NewWriter(buf, flate.DefaultCompression)
		} else {
			writer, err = flate.NewWriter(buf, level)
		}
		if err != nil {
			return response.FormattedResponse{}, err
		}

	default:
		return response.FormattedResponse{}, errors.WithStack(ErrUnknownEncoding)
	}

	// Write and close
	_, err = io.Copy(writer, body.Stream)
	if err != nil {
		writer.Close()

		return response.FormattedResponse{}, err
	}

	if err := writer.Close(); err != nil {
		return response.FormattedResponse{}, err
	}

	return response.FormattedResponse{
		Stream:     bytes.NewReader(buf.Bytes()),
		StreamSize: int64(buf.Len()),
	}, nil
}

// DefaultCompression creates compression middleware with default settings.
func DefaultCompression() dr.Middleware {
	return Compression(CompressionOptions{
		Level:   CompressionLevelDefault,
		MinSize: defaultMinSizeToCompress,
		ContentTypes: []string{
			"text/*",
			response.ContentTypeJSON,
			response.ContentTypeXML,
			response.ContentTypeJavascript,
		},
	})
}
