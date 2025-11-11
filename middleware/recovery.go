package middleware

import (
	"fmt"
	"net/http"

	dataresponse "github.com/raoptimus/data-response.go"
)

// Recover middleware recovers from panics and returns DataResponse.
type Recover struct {
	factory *dataresponse.Factory
}

// NewRecover creates a new recover middleware.
func NewRecover(factory *dataresponse.Factory) *Recover {
	return &Recover{factory: factory}
}

// Handler returns the middleware handler.
func (rec *Recover) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ctx := r.Context()
				panicErr := dataresponse.NewError(http.StatusInternalServerError, fmt.Sprintf("panic: %v", err))

				resp := rec.factory.InternalError(ctx, panicErr)
				dataresponse.Write(w, r, resp, rec.factory)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
