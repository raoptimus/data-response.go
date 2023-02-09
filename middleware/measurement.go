package middleware

import (
	"net/http"
	"regexp"
	"time"
)

type MetricsData struct {
	StatusCode int
	Method     string
	Route      string
	Elapsed    time.Duration
}

//go:generate mockery --name=MetricsService --case underscore --testonly --inpackage

type MetricsService interface {
	Responded(data MetricsData)
}

var patternPlaceholdersRegxp = regexp.MustCompile(`{([a-zA-Z0-9]+).*?}`)

type responseWriter struct {
	w http.ResponseWriter

	statusCode int
}

func (w *responseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *responseWriter) Write(data []byte) (int, error) {
	return w.w.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.w.WriteHeader(statusCode)
}

type MatchedRoutePatternFunc func(r *http.Request) string

func Measurement(next http.Handler, serv MetricsService, patternFunc MatchedRoutePatternFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		ww := &responseWriter{w: w}
		next.ServeHTTP(ww, req)

		var routePattern string
		if patternFunc != nil {
			routePattern = patternFunc(req)
		}

		if len(routePattern) > 0 {
			routePattern = patternPlaceholdersRegxp.
				ReplaceAllString(routePattern, `$1`)
		}

		serv.Responded(MetricsData{
			StatusCode: ww.statusCode,
			Method:     req.Method,
			Route:      routePattern,
			Elapsed:    time.Since(start),
		})
	})
}

func MeasurementN(serv MetricsService, patternFunc MatchedRoutePatternFunc) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return Measurement(next, serv, patternFunc)
	}
}
