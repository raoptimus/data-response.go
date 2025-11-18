package middleware

import (
	"net/http"
	"strings"

	dr "github.com/raoptimus/data-response.go/v2"
)

// ContentTypeValidatorOptions configures content type validation.
type ContentTypeValidatorOptions struct {
	// AllowedTypes is a list of allowed Content-Type values.
	// Supports exact match and wildcards (e.g., "application/*").
	AllowedTypes []string

	// Methods is a list of HTTP methods to validate (default: POST, PUT, PATCH).
	Methods []string

	// ErrorMessage is the error message returned when validation fails.
	ErrorMessage string

	// IgnoreEmpty ignores requests without body (Content-Length: 0).
	IgnoreEmpty bool
}

// ContentTypeValidator creates a middleware that validates request Content-Type header.
// It returns 415 Unsupported Media Type if Content-Type doesn't match allowed types.
func ContentTypeValidator(opts ContentTypeValidatorOptions) dr.Middleware {
	// Default methods
	if len(opts.Methods) == 0 {
		opts.Methods = []string{"POST", "PUT", "PATCH"}
	}

	// Default error message
	if opts.ErrorMessage == "" {
		opts.ErrorMessage = "Unsupported Media Type"
	}

	// Build method map for fast lookup
	methodMap := make(map[string]bool)
	for _, method := range opts.Methods {
		methodMap[strings.ToUpper(method)] = true
	}

	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
			// Check if method should be validated
			if !methodMap[r.Method] {
				return next.Handle(r, f)
			}

			// Check if request has body
			if opts.IgnoreEmpty && r.ContentLength == 0 {
				return next.Handle(r, f)
			}

			// Get Content-Type header
			contentType := r.Header.Get(dr.HeaderContentType)

			// Validate Content-Type
			if !isAllowedContentType(contentType, opts.AllowedTypes) {
				f.Logger().Warn(r.Context(), "invalid content type",
					"content_type", contentType,
					"method", r.Method,
					"path", r.URL.Path,
				)

				return f.Error(r.Context(), http.StatusUnsupportedMediaType, opts.ErrorMessage).
					WithHeader("Accept", strings.Join(opts.AllowedTypes, ", "))
			}

			return next.Handle(r, f)
		})
	}
}

// isAllowedContentType checks if content type is in allowed list.
func isAllowedContentType(contentType string, allowedTypes []string) bool {
	if contentType == "" {
		return false
	}

	// Extract media type without parameters (e.g., "application/json; charset=utf-8" -> "application/json")
	if idx := strings.Index(contentType, ";"); idx > -1 {
		contentType = contentType[:idx]
	}
	contentType = strings.TrimSpace(strings.ToLower(contentType))

	// Check against allowed types
	for _, allowed := range allowedTypes {
		allowed = strings.TrimSpace(strings.ToLower(allowed))

		// Exact match
		if contentType == allowed {
			return true
		}

		// Wildcard match (e.g., "application/*")
		if strings.HasSuffix(allowed, "/*") {
			prefix := strings.TrimSuffix(allowed, "/*")
			if strings.HasPrefix(contentType, prefix+"/") {
				return true
			}
		}

		// Match any
		if allowed == "*/*" || allowed == "*" {
			return true
		}
	}

	return false
}

// JSONOnly creates a middleware that only allows application/json Content-Type.
func JSONOnly() dr.Middleware {
	return ContentTypeValidator(ContentTypeValidatorOptions{
		AllowedTypes: []string{dr.ContentTypeJSON},
		IgnoreEmpty:  true,
		ErrorMessage: "Content-Type must be application/json",
	})
}

// XMLOnly creates a middleware that only allows application/xml Content-Type.
func XMLOnly() dr.Middleware {
	return ContentTypeValidator(ContentTypeValidatorOptions{
		AllowedTypes: []string{dr.ContentTypeXML, dr.ContentTypeTextXML},
		IgnoreEmpty:  true,
		ErrorMessage: "Content-Type must be application/xml or text/xml",
	})
}

// JSONOrXML creates a middleware that allows both JSON and XML.
func JSONOrXML() dr.Middleware {
	return ContentTypeValidator(ContentTypeValidatorOptions{
		AllowedTypes: []string{
			dr.ContentTypeJSON,
			dr.ContentTypeXML,
			dr.ContentTypeTextXML,
		},
		IgnoreEmpty:  true,
		ErrorMessage: "Content-Type must be application/json or application/xml",
	})
}

// APIContentTypes creates a middleware for common API content types.
func APIContentTypes() dr.Middleware {
	return ContentTypeValidator(ContentTypeValidatorOptions{
		AllowedTypes: []string{
			dr.ContentTypeJSON,
			dr.ContentTypeXML,
			dr.ContentTypeTextXML,
			dr.ContentTypeForm,
			dr.ContentTypeMultipartForm,
		},
		IgnoreEmpty: true,
	})
}
