package middleware

import (
	"net/http"
	"time"
)

func Logging(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		logger.Printf("%s %s %s\n", req.Method, req.RequestURI, time.Since(start))
	})
}
