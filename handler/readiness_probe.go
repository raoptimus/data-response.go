package handler

import (
	"context"
	"net/http"

	dr "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/response"
)

type ReadinessService interface {
	Ready(ctx context.Context) error
}

type dummyReadinessService struct{}

func (s *dummyReadinessService) Ready(_ context.Context) error {
	return nil
}

var DummyReadinessService = &dummyReadinessService{}

func ReadinessProbe(serv ReadinessService) dr.HandlerFunc {
	return func(r *http.Request, f *dr.Factory) *response.DataResponse {
		if err := serv.Ready(r.Context()); err != nil {
			return f.ServiceUnavailable(r.Context(), err.Error())
		} else {
			return f.Success(r.Context(), nil)
		}
	}
}
