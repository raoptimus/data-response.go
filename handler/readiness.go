package handler

import (
	"net/http"
)

type ReadinessService interface {
	Ready() error
}

type DummyReadinessService struct{}

func NewDummyReadinessService() *DummyReadinessService {
	return &DummyReadinessService{}
}

func (s *DummyReadinessService) Ready() error {
	return nil
}

func Readiness(serv ReadinessService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		
		if err := serv.Ready(); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}
