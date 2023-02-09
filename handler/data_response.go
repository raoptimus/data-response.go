package handler

import (
	"net/http"

	response "github.com/raoptimus/data-response.go"
)

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

		if err := writer.Write(w, resp.StatusCode(), resp.Data()); err != nil {
			internalResp := f.InternalServerErrorResponse(req.Context(), err)
			if err := writer.Write(w, resp.StatusCode(), internalResp.Data()); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

// DataResponseWithFormatWriterFunc - оборачивает http вызов в response.Handler и возвращает http.HandlerFunc.
func DataResponseWithFormatWriterFunc(f response.FactoryWithFormatWriter, h response.HandlerWithFormatWriter) http.HandlerFunc {
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
			internalResp := f.InternalServerErrorResponse(req.Context(), err)
			if err := writer.Write(w, resp.StatusCode(), internalResp.Data()); err != nil {
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

// DataResponseWithFormatWriter - оборачивает http вызов в response.HandlerAPI и возвращает http.Handler.
func DataResponseWithFormatWriter(f response.FactoryWithFormatWriter, h response.HandlerWithFormatWriter) http.Handler {
	return DataResponseWithFormatWriterFunc(f, h)
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
