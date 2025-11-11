package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/raoptimus/data-response.go/pkg/logger"
)

func Logging(next http.Handler, logger logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		logger.Info(req.Context(), fmt.Sprintf("%s %s %s\n", req.Method, req.RequestURI, time.Since(start)))
	})
}

func LoggingN(logger logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return Logging(next, logger)
	}
}
