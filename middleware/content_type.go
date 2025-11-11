package middleware

import (
	"net/http"
	"slices"
	"strings"

	dataresponse "github.com/raoptimus/data-response.go"
)

// AllowContentType middleware checks if request Content-Type is allowed.
func AllowContentType(factory *dataresponse.Factory, contentTypes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
			if i := strings.Index(contentType, ";"); i > -1 {
				contentType = contentType[0:i]
			}

			if !slices.Contains(contentTypes, contentType) {
				resp := factory.Error(r.Context(), http.StatusUnsupportedMediaType, "Unsupported Media Type")
				dataresponse.Write(w, r, resp, factory)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
