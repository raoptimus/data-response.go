package middleware

import (
	"net/http"
	"runtime/debug"
)

func PanicRecovery(next http.Handler, logger LoggerWithContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				logger.WithContext(req.Context()).
					Errorln(string(debug.Stack()))
			}
		}()
		next.ServeHTTP(w, req)
	})
}

func PanicRecoveryN(logger LoggerWithContext) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return PanicRecovery(next, logger)
	}
}
