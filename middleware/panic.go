package middleware

import (
	"net/http"
	"runtime/debug"
)

func PanicRecovery(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				logger.Println(string(debug.Stack()))
			}
		}()
		next.ServeHTTP(w, req)
	})
}
