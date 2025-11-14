package middleware

import (
	"net/http"
	"strings"

	dr "github.com/raoptimus/data-response.go"
)

func ContentNegotiator(formatters map[string]dr.Formatter) func(next dr.Handler) dr.Handler {
	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
			accept := r.Header.Get("Accept")
			formatter := selectFormatter(accept, formatters, f.Formatter())

			if formatter == nil {
				// Return 406 Not Acceptable
				return f.Error(r.Context(), http.StatusNotAcceptable, "Not Acceptable")
			}

			return next.Handle(r, f).WithFormatter(formatter)
		})
	}
}

func selectFormatter(accept string, formatters map[string]dr.Formatter, defaultFormatter dr.Formatter) dr.Formatter {
	if accept == "" {
		return defaultFormatter
	}

	// Split by comma to handle multiple Accept values
	for _, part := range strings.Split(accept, ",") {
		part = strings.TrimSpace(part)
		// Remove quality factor (e.g., ";q=0.9")
		if idx := strings.Index(part, ";"); idx > -1 {
			part = part[:idx]
		}
		part = strings.TrimSpace(part)

		if formatter, ok := formatters[part]; ok {
			return formatter
		}
	}

	return defaultFormatter
}
