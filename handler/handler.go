package handler

import (
	"net/http"

	response "github.com/raoptimus/data-response.go"
)

// NewHTTPHandlerFunc - оборачивает http вызов в response.Handler и возвращает http.HandlerFunc.
func NewHTTPHandlerFunc(f response.Factory, h response.Handler) http.HandlerFunc {
	writer := f.FormatWriter()
	return func(w http.ResponseWriter, req *http.Request) {
		resp := h.Handle(f, req)

		header := w.Header()
		for key, values := range resp.Header() {
			for i := range values {
				header.Add(key, values[i])
			}
		}

		if err := writer.Write(w, resp.StatusCode(), resp.Data()); err != nil {
			internalResp := f.CreateInternalServerErrorResponse(req.Context(), err)
			if err := writer.Write(w, resp.StatusCode(), internalResp.Data()); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

// NewHTTPHandler - оборачивает http вызов в response.Handler и возвращает http.Handler.
func NewHTTPHandler(f response.Factory, h response.Handler) http.Handler {
	return NewHTTPHandlerFunc(f, h)
}

// Func - адаптер, позволяющий использовать обычные функции как обработчики HTTP.
type Func func(f response.FactoryAPI, r *http.Request) *response.DataResponse

// Handle - вызывает f(factory, r).
func (f Func) Handle(factory response.FactoryAPI, r *http.Request) *response.DataResponse {
	return f(factory, r)
}
