package response

import (
	"context"
	"net/http"
)

type FormatWriter interface {
	Write(w http.ResponseWriter, statusCode int, data any) error
}

type Factory interface {
	FactoryAPI

	FormatWriter() FormatWriter
}

type FactoryAPI interface {
	CreateResponse(ctx context.Context, statusCode int, data any) *DataResponse
	CreateInternalServerErrorResponse(ctx context.Context, err error) *DataResponse
	CreateUnprocessableEntityResponse(ctx context.Context, message string, attributesErrors map[string][]string) *DataResponse
	CreateNotFoundEntityResponse(ctx context.Context, message string) *DataResponse
	CreateErrorResponse(ctx context.Context, statusCode int, message string) *DataResponse
}

type Handler interface {
	Handle(f FactoryAPI, r *http.Request) *DataResponse
}
