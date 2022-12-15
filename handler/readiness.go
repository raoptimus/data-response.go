package handler

import (
	"net/http"
)

type ReadinessService interface {
	Ready() error
}

func Readiness(r ReadinessService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := r.Ready(); err != nil {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
	})
}
