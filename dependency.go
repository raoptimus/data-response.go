package response

import (
	"net/http"
)

type (
	Factory interface {
		GetFormatWriter() FormatWriter
		CreateResponse(data any, code int) *DataResponse
		CreateInternalServerErrorResponse(err error) *DataResponse
	}
	Handler interface {
		Handle(f Factory, r *http.Request) *DataResponse
	}
	FormatWriter interface {
		Write(data any, w http.ResponseWriter) error
	}
)
