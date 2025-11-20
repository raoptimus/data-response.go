package handler

import (
	"net/http"

	dr "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/response"
)

type ReadinessService interface {
	Ready() error
}

type dummyReadinessService struct{}

func (s *dummyReadinessService) Ready() error {
	return nil
}

var DummyReadinessService = &dummyReadinessService{}

func ReadinessProbe(serv ReadinessService) dr.HandlerFunc {
	return func(r *http.Request, f *dr.Factory) *response.DataResponse {
		if err := serv.Ready(); err != nil {
			return f.ServiceUnavailable(r.Context(), err.Error())
		} else {
			return f.Success(r.Context(), nil)
		}
	}
}
