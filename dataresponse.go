package dataresponse

import (
	"net/http"

	drhttp "github.com/raoptimus/data-response.go/http"
)

// Handler defines a handler that returns DataResponse instead of writing to http.ResponseWriter.
// This allows for more functional and composable middleware.
type Handler interface {
	Handle(r *http.Request, f *Factory) DataResponse
}

// HandlerFunc is a function adapter for Handler interface.
type HandlerFunc func(r *http.Request, f *Factory) DataResponse

// Handle calls f(r).
func (hf HandlerFunc) Handle(r *http.Request, f *Factory) DataResponse {
	return hf(r, f)
}

// WrapHandler converts DataResponse Handler to http.Handler.
func WrapHandler(h Handler, f *Factory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := drhttp.NewResponseRecorder(w)
		resp := h.Handle(r, f)
		if err := Write(r.Context(), rw, resp, f); err != nil {
			f.logger.Error(r.Context(), "failed to write response", "error", err.Error())
		}
	})
}
