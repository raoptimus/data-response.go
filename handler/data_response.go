package handler

import (
	"net/http"

	response "github.com/raoptimus/data-response.go"
)

type NoBody struct{}

// DataResponseAPIFunc - оборачивает http вызов в response.Handler и возвращает http.HandlerFunc.
func DataResponseAPIFunc(f response.FactoryWithFormatWriter, h response.HandlerAPI) http.HandlerFunc {
	writer := f.FormatWriter()
	return func(w http.ResponseWriter, req *http.Request) {
		resp := h.Handle(f, req)

		header := w.Header()
		for key, values := range resp.Header() {
			for i := range values {
				header.Add(key, values[i])
			}
		}

		data := resp.Data()
		statusCode := resp.StatusCode()

		if _, ok := data.(NoBody); ok {
			w.WriteHeader(statusCode)

			return
		}

		if err := writer.Write(w, statusCode, data); err != nil {
			internalResp := f.InternalServerErrorResponse(req.Context(), err)
			if err := writer.Write(w, internalResp.StatusCode(), internalResp.Data()); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

// DataResponseAPI - оборачивает http вызов в response.HandlerAPI и возвращает http.Handler.
func DataResponseAPI(f response.FactoryWithFormatWriter, h response.HandlerAPI) http.Handler {
	return DataResponseAPIFunc(f, h)
}

// Func - адаптер, позволяющий использовать обычные функции как обработчики HTTP.
type Func func(f response.Factory, r *http.Request) *response.DataResponse

// Handle - вызывает f(factory, r).
func (f Func) Handle(factory response.Factory, r *http.Request) *response.DataResponse {
	return f(factory, r)
}

// FuncAPI - адаптер, позволяющий использовать обычные функции как обработчики HTTP.
type FuncAPI func(f response.FactoryAPI, r *http.Request) *response.DataResponse

// Handle - вызывает f(factory, r).
func (f FuncAPI) Handle(factory response.FactoryAPI, r *http.Request) *response.DataResponse {
	return f(factory, r)
}
