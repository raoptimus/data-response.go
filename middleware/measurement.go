/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package middleware

import (
	"net/http"
	"time"

	dr "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/response"
)

type MetricsData struct {
	StatusCode int
	Host       string
	Method     string
	Route      string
	Elapsed    time.Duration
}

//go:generate mockery
type MetricsService interface {
	Responded(data MetricsData)
}

func Measurement(serv MetricsService) dr.Middleware {
	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) *response.DataResponse {
			start := response.RequestStartTime(r.Context())
			resp := next.Handle(r, f)

			serv.Responded(MetricsData{
				StatusCode: resp.StatusCode(),
				Host:       r.Host,
				Method:     r.Method,
				Route:      r.Pattern,
				Elapsed:    time.Since(start),
			})

			return resp
		})
	}
}
