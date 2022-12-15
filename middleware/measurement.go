package middleware

import (
	"net/http"
	"time"
)

type MetricsData struct {
	StatusCode int
	Method     string
	Path       string
	Elapsed    time.Duration
}

//go:generate mockery --name=MetricsService --case underscore --testonly --inpackage

type MetricsService interface {
	Responded(data MetricsData)
}

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

func Measurement(next http.Handler, m MetricsService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		ww := &responseWriter{w: w}
		next.ServeHTTP(ww, req)

		m.Responded(MetricsData{
			StatusCode: ww.statusCode,
			Method:     req.Method,
			Path:       req.URL.Path,
			Elapsed:    time.Since(start),
		})
	})
}
