package response

import (
	"context"
	"net/http"
)

type (
	DataResponse struct {
		ctx        context.Context
		data       any
		statusCode int
		header     http.Header
	}
)

func NewDataResponse(data any, code int) *DataResponse {
	return &DataResponse{
		data:       data,
		statusCode: code,
		header:     http.Header{},
	}
}

func (d *DataResponse) Header() http.Header {
	return d.header
}

func (d *DataResponse) GetData() any {
	return d.data
}

func (d *DataResponse) GetStatusCode() int {
	return d.statusCode
}

func (d *DataResponse) GetHeader() http.Header {
	return d.header
}

func (d *DataResponse) GetHeaderValues(key string) []string {
	return d.header.Values(key)
}

func (d *DataResponse) GetHeaderLine(key string) string {
	return d.header.Get(key)
}

func (d *DataResponse) WithHeader(key, value string) *DataResponse {
	d.header.Add(key, value)
	return d
}

func (d *DataResponse) HasHeader(key string) bool {
	return d.header.Get(key) != ""
}
