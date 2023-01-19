package response

import (
	"context"
	"net/http"
)

func NewDummyFactory(fw FormatWriter, verbosity bool) *DummyFactory {
	return &DummyFactory{fw: fw, verbosity: verbosity}
}

type DummyFactory struct {
	fw        FormatWriter
	verbosity bool
}

func (f *DummyFactory) FormatWriter() FormatWriter {
	return f.fw
}

func (f *DummyFactory) CreateResponse(_ context.Context, statusCode int, data any) *DataResponse {
	return NewDataResponse(statusCode, data)
}

func (f *DummyFactory) CreateSuccessResponse(ctx context.Context, data any) *DataResponse {
	return f.CreateResponse(ctx, http.StatusOK, data)
}

func (f *DummyFactory) CreateInternalServerErrorResponse(ctx context.Context, err error) *DataResponse {
	var message string
	if f.verbosity {
		message = err.Error()
	} else {
		message = http.StatusText(http.StatusInternalServerError)
	}

	return f.CreateResponse(ctx, http.StatusInternalServerError, message)
}

func (f *DummyFactory) CreateUnprocessableEntityResponse(
	ctx context.Context,
	message string,
	attributesErrors map[string][]string,
) *DataResponse {
	// TODO: convert attributes to string
	return f.CreateResponse(ctx, http.StatusUnprocessableEntity, message)
}

func (f *DummyFactory) CreateNotFoundEntityResponse(ctx context.Context, message string) *DataResponse {
	return f.CreateResponse(ctx, http.StatusOK, "NotFound: "+message)
}

func (f *DummyFactory) CreateErrorResponse(ctx context.Context, statusCode int, message string) *DataResponse {
	return f.CreateResponse(ctx, statusCode, message)
}
