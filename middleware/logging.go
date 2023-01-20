package middleware

import (
	"net/http"
	"time"
)

func Logging(next http.Handler, logger LoggerWithContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		logger.WithContext(req.Context()).
			Printf("%s %s %s\n", req.Method, req.RequestURI, time.Since(start))
	})
}

func LoggingN(logger LoggerWithContext) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return Logging(next, logger)
	}
}
