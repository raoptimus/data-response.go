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
	"strconv"
	"strings"
	"time"

	dr "github.com/raoptimus/data-response.go/v2"
)

// todo template
func Logging() dr.Middleware {
	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
			start := time.Now()
			resp := next.Handle(r, f)

			var sb strings.Builder
			sb.WriteString(r.RemoteAddr)
			sb.WriteString(" ")
			sb.WriteString("-") // identity id from context
			sb.WriteString(" ")
			sb.WriteString("[")
			sb.WriteString(start.Format(time.RFC3339))
			sb.WriteString("]")
			sb.WriteString(" ")
			sb.WriteString("\"")
			sb.WriteString(r.Method)
			sb.WriteString(" ")
			sb.WriteString(r.RequestURI)
			sb.WriteString(" ")
			sb.WriteString(r.Proto)
			sb.WriteString("\"")
			sb.WriteString(" ")
			sb.WriteString(strconv.Itoa(resp.StatusCode()))
			sb.WriteString(" ")
			sb.WriteString("-") // resp size
			sb.WriteString(" ")
			sb.WriteString("\"")
			sb.WriteString(r.Referer())
			sb.WriteString("\"")
			sb.WriteString(" ")
			sb.WriteString("\"")
			sb.WriteString(r.UserAgent())
			sb.WriteString("\"")
			sb.WriteString(" ")
			sb.WriteString(time.Since(start).String())

			f.Logger().Info(r.Context(), sb.String())

			return resp
		})
	}
}
