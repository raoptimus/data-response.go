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
	"regexp"
	"time"

	dr "github.com/raoptimus/data-response.go/v2"
)

type MetricsData struct {
	StatusCode int
	Method     string
	Route      string
	Elapsed    time.Duration
}

//go:generate mockery
type MetricsService interface {
	Responded(data MetricsData)
}

type MatchedRoutePatternFunc func(r *http.Request) string

var patternPlaceholdersRegxp = regexp.MustCompile(`{([a-zA-Z0-9]+).*?}`)

func Measurement(serv MetricsService, patternFunc MatchedRoutePatternFunc) dr.Middleware {
	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
			start := RequestStartTime(r.Context())
			resp := next.Handle(r, f)

			var routePattern string
			if patternFunc != nil {
				routePattern = patternFunc(r)
			}

			if len(routePattern) > 0 {
				routePattern = patternPlaceholdersRegxp.ReplaceAllString(routePattern, `$1`)
			}

			serv.Responded(MetricsData{
				StatusCode: resp.StatusCode(),
				Method:     r.Method,
				Route:      routePattern,
				Elapsed:    time.Since(start),
			})

			return resp
		})
	}
}
