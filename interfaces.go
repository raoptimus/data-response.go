package response

import (
	"context"
	"net/http"
)

type FormatWriter interface {
	Write(w http.ResponseWriter, statusCode int, data any) error
}

type FactoryWithFormatWriter interface {
	FactoryAPI

	FormatWriter() FormatWriter
}

type Factory interface {
	CreateResponse(ctx context.Context, statusCode int, data any) *DataResponse
	CreateInternalServerErrorResponse(ctx context.Context, err error) *DataResponse
}

type FactoryAPI interface {
	CreateResponse(ctx context.Context, statusCode int, data any) *DataResponse
	CreateSuccessResponse(ctx context.Context, data any) *DataResponse
	CreateInternalServerErrorResponse(ctx context.Context, err error) *DataResponse
	CreateUnprocessableEntityResponse(ctx context.Context, message string, attributesErrors map[string][]string) *DataResponse
	CreateNotFoundEntityResponse(ctx context.Context, message string) *DataResponse
	CreateErrorResponse(ctx context.Context, statusCode int, message string) *DataResponse
}

type Handler interface {
	Handle(f Factory, r *http.Request) *DataResponse
}

type HandlerAPI interface {
	Handle(f FactoryAPI, r *http.Request) *DataResponse
}
