package handler

import (
	"net/http"

	dr "github.com/raoptimus/data-response.go/v2"
)

func StartupProbe() dr.HandlerFunc {
	return func(r *http.Request, f *dr.Factory) dr.DataResponse {
		return f.Success(r.Context(), nil)
	}
}
