package response

import (
	"net/http"
)

type (
	Factory interface {
		GetFormatWriter() FormatWriter
		CreateResponse(data any, statusCode int) *DataResponse
		CreateInternalServerErrorResponse(err error) *DataResponse
	}
	FactoryAPI interface {
		Factory

		CreateUnprocessableEntityResponse(attributesErrors map[string][]string, message error)
		CreateNotFoundEntityResponse(message error) *DataResponse
		CreateErrorResponse(message error, statusCode int) *DataResponse
	}
	Handler interface {
		Handle(f Factory, r *http.Request) *DataResponse
	}
	HandlerAPI interface {
		Handle(f FactoryAPI, r *http.Request) *DataResponse
	}
	FormatWriter interface {
		Write(data any, w http.ResponseWriter) error
	}
)
