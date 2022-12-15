package response

import (
	"net/http"
)

type (
	factory struct {
		fw        FormatWriter
		verbosity bool
	}
)

func NewFactory(fw FormatWriter, verbosity bool) Factory {
	return &factory{fw: fw, verbosity: verbosity}
}

func (f *factory) GetFormatWriter() FormatWriter {
	return f.fw
}

func (f *factory) CreateResponse(data any, statusCode int) *DataResponse {
	return NewDataResponse(data, statusCode)
}

func (f *factory) CreateInternalServerErrorResponse(err error) *DataResponse {
	var message string
	if f.verbosity {
		message = err.Error()
	} else {
		message = http.StatusText(http.StatusInternalServerError)
	}

	return f.CreateResponse(message, http.StatusInternalServerError)
}
