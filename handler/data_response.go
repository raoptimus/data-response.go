package handler

import (
	"log"
	"net/http"

	response "github.com/raoptimus/data-response.go"
)

func DataResponse(h response.Handler, f response.Factory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		resp := h.Handle(f, req)

		for key, values := range resp.Header() {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		if err := f.GetFormatWriter().Write(resp.GetData(), w); err != nil {
			log.Println(err)

			if err := f.GetFormatWriter().Write(f.CreateInternalServerErrorResponse(err).GetData(), w); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}
	})
}
