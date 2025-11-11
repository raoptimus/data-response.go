package middleware

import (
	"net/http"
	"strings"

	dataresponse "github.com/raoptimus/data-response.go"
)

// ContentNegotiator is a middleware for content type negotiation.
type ContentNegotiator struct {
	factory          *dataresponse.Factory
	formatters       map[string]dataresponse.Formatter
	defaultFormatter dataresponse.Formatter
}

// NewContentNegotiator creates a new content negotiator middleware.
func NewContentNegotiator(factory *dataresponse.Factory, formatters map[string]dataresponse.Formatter, defaultFormatter dataresponse.Formatter) *ContentNegotiator {
	return &ContentNegotiator{
		factory:          factory,
		formatters:       formatters,
		defaultFormatter: defaultFormatter,
	}
}

// Handler returns the middleware handler.
func (cn *ContentNegotiator) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		formatter := cn.selectFormatter(accept)

		if formatter == nil {
			// Return 406 Not Acceptable
			resp := cn.factory.Error(r.Context(), http.StatusNotAcceptable, "Not Acceptable")
			dataresponse.Write(w, r, resp, cn.factory)
			return
		}

		ctx := dataresponse.ContextWithFormatter(r.Context(), formatter)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (cn *ContentNegotiator) selectFormatter(accept string) dataresponse.Formatter {
	if accept == "" {
		return cn.defaultFormatter
	}

	// Split by comma to handle multiple Accept values
	for _, part := range strings.Split(accept, ",") {
		part = strings.TrimSpace(part)
		// Remove quality factor (e.g., ";q=0.9")
		if idx := strings.Index(part, ";"); idx > -1 {
			part = part[:idx]
		}
		part = strings.TrimSpace(part)

		if formatter, ok := cn.formatters[part]; ok {
			return formatter
		}
	}

	return cn.defaultFormatter
}
