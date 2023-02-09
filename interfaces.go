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
	Response(ctx context.Context, statusCode int, data any) *DataResponse
	InternalServerErrorResponse(ctx context.Context, err error) *DataResponse
}

type FactoryAPI interface {
	Response(ctx context.Context, statusCode int, data any) *DataResponse
	SuccessResponse(ctx context.Context, data any) *DataResponse
	InternalServerErrorResponse(ctx context.Context, err error) *DataResponse
	UnprocessableEntityResponse(ctx context.Context, message string, attributesErrors map[string][]string) *DataResponse
	NotFoundEntityResponse(ctx context.Context, message string) *DataResponse
	ErrorResponse(ctx context.Context, statusCode int, message string) *DataResponse
}

type Handler interface {
	Handle(f Factory, r *http.Request) *DataResponse
}

type HandlerAPI interface {
	Handle(f FactoryAPI, r *http.Request) *DataResponse
}

type HandlerWithFormatWriter interface {
	Handle(f FactoryWithFormatWriter, r *http.Request) *DataResponse
}
