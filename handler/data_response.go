package handler

import (
	"log"
	"net/http"

	"github.com/pkg/errors"
	response "github.com/raoptimus/data-response.go"
)

// DataResponseAPIFunc - оборачивает http вызов в response.Handler и возвращает http.HandlerFunc.
func DataResponseAPIFunc(f response.FactoryWithFormatWriter, h response.HandlerAPI) http.HandlerFunc {
	encoder := f.FormatWriter()
	wrec, canWrec := f.(response.WriteResponseErrorCallback)

	return func(w http.ResponseWriter, req *http.Request) {
		resp := h.Handle(f, req)

		header := w.Header()
		for key, values := range resp.Header() {
			for i := range values {
				header.Add(key, values[i])
			}
		}

		statusCode := resp.StatusCode()
		data := resp.Data()

		if data == nil {
			w.WriteHeader(statusCode)

			return
		}

		bytes, err := encoder.Marshal(header, data)
		if err != nil {
			internalResp := f.InternalServerErrorResponse(req.Context(), err)
			if bytes, err = encoder.Marshal(header, internalResp.Data()); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}

			statusCode = internalResp.StatusCode()
		}

		w.WriteHeader(statusCode)

		if _, err := w.Write(bytes); err != nil {
			if canWrec {
				wrec.WriteResponseError(req.Context(), errors.Wrap(err, "write response"))
			} else {
				log.Printf("failed to write response: %v\n", err)
			}
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
