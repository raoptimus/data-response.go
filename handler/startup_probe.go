package handler

import (
	"net/http"

	dr "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/response"
)

func StartupProbe() dr.HandlerFunc {
	return func(r *http.Request, f *dr.Factory) *response.DataResponse {
		return f.Success(r.Context(), nil)
	}
}
