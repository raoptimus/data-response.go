package middleware

import (
	"net/http"
	"time"
)

type Metrics interface {
	Responded(w http.ResponseWriter, req *http.Request, elapsed time.Duration)
}

func Measurement(next http.Handler, m Metrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		m.Responded(w, req, time.Since(start))
	})
}
